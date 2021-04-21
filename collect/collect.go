package collect

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/go-ego/gse"
	"github.com/parnurzeal/gorequest"
	"regexp"
	"strconv"
	"strings"
	"time"
	"weixinScraperSingle/fetchSoGou"
	"weixinScraperSingle/golimit"
	"weixinScraperSingle/model"
	"weixinScraperSingle/similar"
	"weixinScraperSingle/sqlite"
	"weixinScraperSingle/util"
)

var (
	logger  = util.GetLogger("collect")
	GoLimit = 2 // 并发数
)

func Collect() {
	g := golimit.NewGoLimit(GoLimit)

	for {
		body := fetchSina()

		if body == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		titles := getTitles(body)
		words := getWords(titles)

		// 通过关键词抓取公众号
		for _, word := range words {
			g.Add()
			go func(g *golimit.GoLimit, keyword string) {
				defer func() {
					if err := recover(); err != nil {
						logger.Infof("%s 任务失败: %s", keyword, err)
					}
					g.Done()
				}()

				getMps(keyword)

			}(g, word)

		}

		time.Sleep(time.Minute)
	}
}

func fetchSina() (body string) {
	var request = gorequest.New()
	defer func() {
		if err := recover(); err != nil {
			logger.Debugf("fetchSina 错误 ", err)
		}
	}()

	resp, body, errs := request.Get("https://news.sina.com.cn/").
		Set("User-Agent", util.GetRandomUA()).
		Set("Content-Type", "text/html; charset=utf-8").
		Set("Referer", "https://news.sina.com.cn/").
		Set("Pragma", `no-cache`).
		Timeout(30 * time.Second).
		End()

	if err := util.ErrAndStatus(errs, resp); err != nil {
		logger.Errorf("访问新浪时错误 %s", err)
		// return if any error from goRequest or body is empty
		return
	}

	return body
}

func getTitles(body string) (title []string) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	tmpSlice := htmlquery.Find(doc, "//ul/li/a")
	if tmpSlice == nil {
		return
	}

	for _, item := range tmpSlice {
		t := htmlquery.InnerText(item)

		// 丢弃禁词和太短的
		if util.ContainAny(t, util.SinaFilter) || util.TooShort(t, 6) {
			continue
		}
		title = append(title, t)
	}
	return
}

func getWords(titles []string) (words []string) {
	var (
		seg gse.Segmenter
	)

	// 加载词典
	err := seg.LoadDict("dictionary.txt")
	if err != nil {
		logger.Errorf("分词时发生错误 %s", err)
		return
	}

	for _, item := range titles {
		words = append(words, seg.CutAll(item)...)
	}

	words = util.RemoveRepByLoop(words) // 去重

	var longerWords []string
	for _, word := range words {
		if util.TooShort(word, 2) {
			continue
		}

		longerWords = append(longerWords, word)
	}

	return longerWords
}

// 通过搜索关键词得到今日发布原创信息的公众号
func getMps(word string) {
	for i := 1; i <= 3; i++ {
		logger.Infof("抓取关键词: %s 第 %d 页", word, i)

		// 公众号关键词搜索结果
		job := fetchSoGou.SoGouFetcher{Url: fmt.Sprintf("https://weixin.sogou.com/weixin?type=2&s_from=input&query=%s&ie=utf8&_sug_=n&_sug_type_=&page=%d", word, i)}
		fetchSoGou.SoGouFetchChan <- job

		result := <-fetchSoGou.SoGouResultChan

		// 提取每页的公众号并检查是否原创
		mps := extractAndCheck(result.Body)

		for _, item := range mps {
			sqlite.AddMp(item)
		}
	}
	return
}

// 提炼搜狗搜索信息(公众号主页, 最近发布时间, 最近发布文章简述)
func extractAndCheck(body string) (mps []model.WeixinMp) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	tmpSlice := htmlquery.Find(doc, "//div[@class='txt-box']")
	if tmpSlice == nil {
		return
	}

	for _, item := range tmpSlice {
		html := htmlquery.OutputHTML(item, true)
		html = strings.Replace(html, "\n", "", -1)

		re := regexp.MustCompile(`<p class="txt-info.+?>(.+)</p>.+<a class="account" .+?href="(.+?)".+?uigs="article_account_1">(.+?)</a>.+?timeConvert\('(\d+)'`)
		match := re.FindStringSubmatch(html)

		if match == nil {
			continue
		}

		var (
			briefTmp      = match[1]
			mpName        = match[3]
			createTime, _ = strconv.ParseInt(match[4], 10, 64)
		)

		// 文字简介为空的不要
		if briefTmp == "" {
			continue
		}

		// 超过24小时的不要
		if (time.Now().Unix() - createTime) > 86400 {
			continue
		}

		// 已经存在的不检查
		if sqlite.SelectMpByName(mpName) {
			continue
		}

		//去除所有尖括号内的HTML代码
		re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
		brief := re.ReplaceAllString(briefTmp, "")

		// 原创度检查
		sentences := similar.ChooseSentences(&brief) // 随机提取文章里的一些长句

		if sentences == nil {
			logger.Infof("%s 没有提取到句子", mpName)
			continue
		}

		originRate := similar.CheckOrigin(&sentences) // 逐个搜索长句,检查原创度

		logger.Infof("%s 原创度为 %.0f", mpName, originRate)

		if originRate < 50 {
			continue
		}

		mpId := getMpsInfo(mpName)

		if mpId == "" {
			continue
		}

		mp := model.WeixinMp{Mp: mpId, MpName: mpName}
		mps = append(mps, mp)

		time.Sleep(5 * time.Second)
	}

	return mps
}

func getMpsInfo(mpName string) (mp string) {
	url := fmt.Sprintf("https://weixin.sogou.com/weixin?type=1&s_from=input&query=%s&ie=utf8&_sug_=n&_sug_type_=", mpName)
	job := fetchSoGou.SoGouFetcher{Url: url}
	fetchSoGou.SoGouFetchChan <- job

	result := <-fetchSoGou.SoGouResultChan

	body := result.Body

	if body == "" {
		return
	}

	mp = extractInfo(body, mpName)
	return mp
}

// 提取使用公众号名称搜索到的信息,并提取符合条件的公众号ID
func extractInfo(body string, mpName string) (mp string) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	tmpSlice := htmlquery.Find(doc, "//li[starts-with(@id,'sogou_vr_')]")
	if tmpSlice == nil {
		return
	}

	// 从多个条目里面找到对应的公众号
	for _, item := range tmpSlice {
		html := htmlquery.OutputHTML(item, true)
		str := fmt.Sprintf("<em><!--red_beg-->%s<!--red_end--></em>", mpName)
		if strings.Contains(html, str) {
			if !strings.Contains(html, "最近文章") {
				// 没有主体的公众号是无法搜索到最新文章的,抛弃
				return
			}

			mpTmp := htmlquery.FindOne(item, "//label[@name='em_weixinhao']")
			if mpTmp == nil {
				return
			}

			return htmlquery.InnerText(mpTmp)
		}
	}

	return
}
