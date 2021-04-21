package sogou

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/parnurzeal/gorequest"
	"regexp"
	"strconv"
	"strings"
	"time"
	"weixinScraperSingle/fetchSoGou"
	"weixinScraperSingle/fetchWeixin"
	"weixinScraperSingle/golimit"
	"weixinScraperSingle/model"
	"weixinScraperSingle/similar"
	"weixinScraperSingle/sqlite"
	"weixinScraperSingle/util"
	"weixinScraperSingle/wash"
)

var (
	logger  = util.GetLogger("soGou")
	GoLimit = 2 // 并发数
)

//func Run(newsChan chan *model.Article) {
//	items := sqlite.SelectMps()
//
//	for _, item := range items {
//		getContent(&item, newsChan)
//	}
//}

func Run(newsChan chan *model.Article) {
	g := golimit.NewGoLimit(GoLimit)

	items := sqlite.SelectMps()

	for _, item := range items {
		g.Add()

		go func(g *golimit.GoLimit, mp model.WeixinMp, ch chan *model.Article) {
			defer func() {
				if err := recover(); err != nil {
					logger.Infof("%s 任务失败: %s", mp.MpName, err)
				}
				g.Done()
			}()

			getContent(&mp, ch)

		}(g, item, newsChan)
	}
}

// 抓取公众号最新文章的内容
func getContent(mp *model.WeixinMp, inputChan chan *model.Article) {
	logger.Infof("抓取公众号: %s - %s", mp.Mp, mp.MpName)

	// 公众号搜索结果
	url := fmt.Sprintf("https://weixin.sogou.com/weixin?type=1&s_from=input&query=%s&ie=utf8&_sug_=n&_sug_type_=", mp.Mp)
	job := fetchSoGou.SoGouFetcher{Url: url}
	fetchSoGou.SoGouFetchChan <- job

	result := <-fetchSoGou.SoGouResultChan

	mpSearchBody := result.Body

	if mpSearchBody == "" {
		return
	}

	//获取搜狗跳转页网址和发布时间
	jumpUrl, title, postTime := extractJumpUrl(mpSearchBody)

	// 公众号没有搜到结果
	if jumpUrl == "" {
		return
	}

	sqlite.UpdateCheckTime(mp.Mp, time.Now().Unix()) // 更新检查时间

	// 与上次发布时间相同,没有新文章
	if strconv.Itoa(mp.RenewTime) == postTime {
		logger.Debugf("%s - %s 没有发布新文章", mp.Mp, mp.MpName)
		return
	}

	sqlite.UpdatePostTime(mp.Mp, postTime) // 存储最新的post时间

	request := gorequest.New()

	// 从跳转页里组合微信文章页网址, 由于解析另一页的参数, 需要带有 request
	url = "https://weixin.sogou.com/" + jumpUrl
	job = fetchSoGou.SoGouFetcher{Url: url, Request: request}
	fetchSoGou.SoGouFetchChan <- job

	result = <-fetchSoGou.SoGouResultChan
	jumpBody := result.Body

	if jumpBody == "" {
		return
	}

	articleUrl := extractArticleUrl(jumpBody)

	// 获取微信文章内容
	jobMp := fetchWeixin.WeixinFetcher{Url: articleUrl}
	fetchWeixin.WeixinFetchChan <- jobMp

	resultMp := <-fetchWeixin.WeixinResultChan
	articleBody := resultMp.Body

	if articleBody == "" {
		return
	}

	// 获取文章的内容
	content := extractArticle(articleBody)

	logger.Infof("抓到文章: %s", title)

	if content == "" {
		return
	}

	// 原创度检查
	sentences := similar.ChooseSentences(&content) // 随机提取文章里的一些长句

	if sentences == nil {
		logger.Infof("%s 没有提取到句子", title)
		return
	}

	originRate := similar.CheckOrigin(&sentences) // 在百度逐个搜索长句,检查原创度

	logger.Infof("%s - 原创度为 %.0f", title, originRate)

	// 放弃原创度低的
	if originRate < util.MinOriginRate {
		return
	}

	article := model.Article{
		Title:   title,
		Content: strings.TrimSpace(content),
	}

	//println(article.Content)

	inputChan <- &article

	return
}

// 提炼搜狗跳转页地址
func extractJumpUrl(body string) (url string, title string, time string) {
	re := regexp.MustCompile(`account_article_0" href="(.+?)">(.+?)</a><span><script>document.write\(timeConvert\('(\d+)'`)
	match := re.FindStringSubmatch(body)

	if match == nil {
		return
	}

	return match[1], match[2], match[3]
}

// 提炼微信文章页地址
func extractArticleUrl(body string) (url string) {
	re := regexp.MustCompile(`url \+= '(.+?)';`)
	match := re.FindAllStringSubmatch(body, -1)

	if match == nil {
		return
	}

	for _, v := range match {
		url += v[1]
	}

	return url
}

// 提取文章内容
func extractArticle(body string) (content string) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return
	}

	tmp := htmlquery.FindOne(doc, "//div[@id='js_content']")

	if tmp == nil {
		return
	}

	content = htmlquery.OutputHTML(tmp, true)

	// 清洗文章
	content = wash.Wash(content)

	return content
}
