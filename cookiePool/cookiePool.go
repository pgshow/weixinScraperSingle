package cookiePool

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
	"weixinScraperSingle/adsl"
	"weixinScraperSingle/util"
)

type queue struct {
	cookie    []*http.Cookie
	restTimes int
}

var (
	logger              = util.GetLogger("cookiePool")
	cookieFetchErrTimes = new(int)
	Pool                = new([]queue)
	mutex               sync.Mutex
)

func Init() {
	for {
		if len(*Pool) < 3 {
			cookie := fetchCookie()
			if cookie == nil {
				time.Sleep(5 * time.Second)
				continue
			}

			*Pool = append(*Pool, queue{cookie, 50})
			logger.Infof("池中有 cookie %d 个", len(*Pool))
		} else {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func Run() {
	for {
		time.Sleep(10 * time.Second)
		if len(*Pool) < 20 {
			cookie := fetchCookie()
			if cookie == nil {
				logger.Error("fetch 新 cookie 为 nil")
				continue
			}

			logger.Debugf("fetch 新 cookie %s", cookie[0].Value)

			//makeCookie(cookieTmp)

			item := queue{cookie, 50}

			mutex.Lock()
			*Pool = append(*Pool, item)
			mutex.Unlock()
		}
	}
}

func GetCookie() []*http.Cookie {
	for {
		if len(*Pool) <= 0 {
			logger.Debugf("等待 cookie 池补充")
			time.Sleep(5 * time.Second)
			continue
		}

		rand.Seed(time.Now().Unix())
		id := rand.Intn(len(*Pool))
		item := (*Pool)[id]
		if item.restTimes <= 0 {
			remove(id)
			continue
		}

		mutex.Lock()
		(*Pool)[id].restTimes -= 1
		mutex.Unlock()
		return item.cookie
	}
}

func remove(id int) {
	mutex.Lock()
	*Pool = append((*Pool)[:id], (*Pool)[id+1:]...)
	mutex.Unlock()
}

func fetchCookie() (cookie []*http.Cookie) {
	req, err := http.NewRequest("GET", "https://v.sogou.com/", nil)
	if err != nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			req.Body.Close()
			logger.Debugf("fetchCookie 错误 ", err)
		}
	}()

	req.Header.Set("User-Agent", util.GetRandomUA())

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	adsl.WaitChangingIp()

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("获取Snuid时错误 %s", err)

		if strings.Contains(err.Error(), "network is unreachable") || strings.Contains(err.Error(), "Client.Timeout") || strings.Contains(err.Error(), "TLS handshake timeout") {
			*cookieFetchErrTimes += 1
			if *cookieFetchErrTimes > 5 {
				adsl.ChangeIP()
				*cookieFetchErrTimes = 0
			}
		}

		return nil
	}

	*cookieFetchErrTimes = 0

	if resp.StatusCode != 200 {
		logger.Errorf("获取cookie时Status错误 %s", resp.StatusCode)
		// return if any error from goRequest or body is empty
		return nil
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(body), "验证码") {
		println("coookie要验证码")
	}

	return resp.Cookies()
}
