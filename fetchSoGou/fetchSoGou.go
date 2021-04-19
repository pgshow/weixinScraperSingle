// Package fetchSoGou 把要获取的object丢进fetchChan里即了fetch到需要的body
package fetchSoGou

import (
	"github.com/parnurzeal/gorequest"
	"strings"
	"time"
	"weixinScraperSingle/adsl"
	"weixinScraperSingle/cookiePool"
	"weixinScraperSingle/util"
)

type SoGouFetcher struct {
	FetchType string
	Url       string
	Request   *gorequest.SuperAgent
	Body      string
	Err       error
}

var (
	logger            = util.GetLogger("fetchSoGou")
	SoGouFetchChan    = make(chan SoGouFetcher)
	SoGouResultChan   = make(chan SoGouFetcher)
	fetchErrTimes     = new(int)
	fetchTimeOutTimes = new(int)
	noInternetErrs    = []string{"network is unreachable", "read tcp"}
	timeOutErrs       = []string{"Client.Timeout", "TLS handshake timeout"}
)

func Run() {
	for item := range SoGouFetchChan {
	retry:
		fetchSoGou(&item)

		if item.Err != nil {
			logger.Errorf("访问搜狗时错误 %s", item.Err)

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
				}
				time.Sleep(6 * time.Second)
				goto retry
			}
		}

		*fetchErrTimes = 0
		*fetchTimeOutTimes = 0

		if strings.Contains(item.Body, "请输入验证码") {
			logger.Errorf("搜狗要求输入验证码")
			adsl.ChangeIP()
			time.Sleep(2 * time.Second)
			goto retry
		}

		SoGouResultChan <- item
		time.Sleep(6 * time.Second)
	}
}

func fetchSoGou(profile *SoGouFetcher) (status string) {
	var (
		request *gorequest.SuperAgent
	)

	adsl.WaitChangingIp()

	if profile.Request != nil {
		// 采用上一次访问的 cookie，用于搜狗解析微信文章页地址
		request = profile.Request
	} else {
		request = gorequest.New()
	}

	resp, body, errs := request.Get(profile.Url).
		Set("User-Agent", util.GetRandomUA()).
		Set("Content-Type", "text/html; charset=utf-8").
		Set("Referer", "https://weixin.sogou.com").
		Set("Pragma", `no-cache`).
		AddCookies(cookiePool.GetCookie()).
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
