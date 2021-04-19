package similar

import (
	"github.com/antchfx/htmlquery"
	"github.com/antlabs/strsim"
	"regexp"
	"strings"
	"unicode/utf8"
	"weixinScraperSingle/fetchBing"
	"weixinScraperSingle/util"
)

var (
	logger            = util.GetLogger("similar")
	bingFetchErrTimes = new(int)
)

// CheckOrigin 相似度算法来源:https://ginson.wang/138.html
func CheckOrigin(sentences *[]string) float64 {
	var (
		finishWholeSteps = 0 // 完成整个检查的次数
		allCount         = 0
	)

	for _, phase := range *sentences {
		job := fetchBing.BingFetcher{Sentence: phase}
		fetchBing.BingFetchChan <- job

		result := <-fetchBing.BingResultChan
		body := result.Body

		if body == "" {
			continue
		}

		doc, err := htmlquery.Parse(strings.NewReader(body))
		if err != nil {
			continue
		}

		tmpSlice := htmlquery.Find(doc, "//li[@class='b_algo']//strong") // bing rule

		if tmpSlice == nil {
			continue
		}

		// 搜索结果中每个高亮都检查相似度,并计算评分
		hitCount := 0
		for _, item := range tmpSlice {
			var (
				highlightPart = htmlquery.InnerText(item)
				ratio         = strsim.Compare(phase, highlightPart)
			)

			// 相似度0.9的记分
			if ratio > 0.9 {
				//fmt.Printf("%x||%s||%s", ratio, highlightPart, phase)
				hitCount += 1
			}
		}
		if hitCount >= 1 {
			allCount += 1
		}

		finishWholeSteps += 1
	}

	// 至少需要完成一次检查, 否则可能是整个文章检查失败
	if finishWholeSteps > 0 {
		//print("原创度 %s%%" % round(100 - (a_count / len(samples) * 100), 2))
		originRate := 100 - (float64(allCount)/float64(len(*sentences)))*100
		return originRate
	}

	return 0
}

// ChooseSentences 从文章总随机选择几个长句子
func ChooseSentences(content *string) []string {
	var (
		//sep           = `。|.|,|，|？|！|“|”|：|；|…| |\n`
		sampleNum     = 10     // 随机取长句数量
		longSentences []string // 长句子
		randomChoose  []string // 随机选择部分句子用百度检查
	)

	sentences := regexp.MustCompile(`[。.,，？！“”：；…\r\n]`).Split(*content, -1)
	//spaceRe, _ := regexp.Compile(sep)
	//sentences := spaceRe.Split(*content, -1)

	for _, x := range sentences {
		if utf8.RuneCountInString(x) >= 10 && utf8.RuneCountInString(x) < 20 {
			longSentences = append(longSentences, x)
		}
	}

	if len(longSentences) < sampleNum {
		// 句子比较少直接返回
		return longSentences
	}

	// 句子比较多, 随机选择几个
	randomNums := util.UniqueRandomNum(0, len(longSentences)-1, sampleNum)

	for _, j := range randomNums {
		randomChoose = append(randomChoose, longSentences[j])
	}

	return randomChoose
}
