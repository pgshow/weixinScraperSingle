// Package fetchBing 把要获取的object丢进fetchChan里即了fetch到需要的body
package fetchBing

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"time"
	"weixinScraperSingle/adsl"
	"weixinScraperSingle/util"
)

type BingFetcher struct {
	Sentence string
	Body     string
	Err      error
}

var (
	logger            = util.GetLogger("bingSoGou")
	BingFetchChan     = make(chan BingFetcher)
	BingResultChan    = make(chan BingFetcher)
	fetchErrTimes     = new(int)
	fetchTimeOutTimes = new(int)
	noInternetErrs    = []string{"network is unreachable", "read tcp"}
	timeOutErrs       = []string{"Client.Timeout", "TLS handshake timeout"}
)

func Run() {
	for item := range BingFetchChan {
	retry:
		fetchBing(&item)

		if item.Err != nil {
			logger.Errorf("访问必应时错误 %s", item.Err)

			if util.ContainAny(item.Err.Error(), noInternetErrs) {
				// 网络无法连接立即重试
				*fetchErrTimes += 1
				if *fetchErrTimes > 5 {
					adsl.ChangeIP()
					*fetchErrTimes = 2
				}
				goto retry
			}

			if util.ContainAny(item.Err.Error(), timeOutErrs) {
				// 访问超时等待几秒在试
				*fetchTimeOutTimes += 1
				if *fetchTimeOutTimes > 3 {
					adsl.ChangeIP()
					*fetchErrTimes = 1
				}
				time.Sleep(4 * time.Second)
				goto retry
			}
		}

		*fetchErrTimes = 0
		*fetchTimeOutTimes = 0

		BingResultChan <- item
		time.Sleep(2 * time.Second)
	}
}

func fetchBing(profile *BingFetcher) (status string) {
	adsl.WaitChangingIp()

	url := fmt.Sprintf("https://cn.bing.com/search?q=%s&qs=HS&sc=2-0&FORM=QBLH&sp=1", profile.Sentence)
	resp, body, errs := gorequest.New().Get(url).
		Set("User-Agent", util.GetRandomUA()).
		Set("Content-Type", "text/html; charset=utf-8").
		Set("Pragma", `no-cache`).
		Set("Referer", "https://cn.bing.com").
		Timeout(30 * time.Second).
		End()

	if err := util.ErrAndStatus(errs, resp); err != nil {
		profile.Body = ""
		profile.Err = err
	} else {
		profile.Body = body
	}

	return
}
