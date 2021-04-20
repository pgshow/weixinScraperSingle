// Package fetchWeixin 把要获取的object丢进fetchChan里即了fetch到需要的body
package fetchWeixin

import (
	"github.com/parnurzeal/gorequest"
	"time"
	"weixinScraperSingle/adsl"
	"weixinScraperSingle/util"
)

type WeixinFetcher struct {
	Url  string
	Body string
}

var (
	logger            = util.GetLogger("fetchWeixin")
	WeixinFetchChan   = make(chan WeixinFetcher)
	WeixinResultChan  = make(chan WeixinFetcher)
	fetchErrTimes     = new(int)
	fetchTimeOutTimes = new(int)
	noInternetErrs    = []string{"network is unreachable", "read tcp"}
	timeOutErrs       = []string{"Client.Timeout", "TLS handshake timeout"}
)

func Run() {
	for item := range WeixinFetchChan {
	retry:
		err := fetchBing(&item)

		if err != nil {
			logger.Errorf("访问微信时错误 %s", err)

			if util.ContainAny(err.Error(), noInternetErrs) {
				// 网络无法连接立即重试
				*fetchErrTimes += 1
				if *fetchErrTimes > 5 {
					adsl.ChangeIP()
					*fetchErrTimes = 2
				}
				goto retry
			}

			if util.ContainAny(err.Error(), timeOutErrs) {
				// 访问超时等待几秒在试
				*fetchTimeOutTimes += 1
				if *fetchTimeOutTimes > 3 {
					adsl.ChangeIP()
					*fetchErrTimes = 0
				}
				time.Sleep(6 * time.Second)
				goto retry
			}
		}

		*fetchErrTimes = 0
		*fetchTimeOutTimes = 0

		WeixinResultChan <- item
		time.Sleep(6 * time.Second)
	}
}

func fetchBing(profile *WeixinFetcher) (err error) {
	adsl.WaitChangingIp()

	resp, body, errs := gorequest.New().Get(profile.Url).
		Set("User-Agent", util.GetRandomUA()).
		Set("Content-Type", "text/html; charset=utf-8").
		Set("Pragma", `no-cache`).
		Timeout(30 * time.Second).
		End()

	if err = util.ErrAndStatus(errs, resp); err != nil {
		profile.Body = ""
	} else {
		profile.Body = body
	}

	return
}
