package postArticle

import (
	"github.com/parnurzeal/gorequest"
	"time"
	"weixinScraperSingle/model"
	"weixinScraperSingle/util"
)

var (
	logger = util.GetLogger("postArticle")
)

func Run(ch chan *model.Article) {
	for {
		for item := range ch {
			postJson(item)
		}
	}

}

func postJson(article *model.Article) {
	var retryTimes = 1
	defer func() {
		// defer无论所在的函数是正常结束,还是异常结束都会被执行
		// recover可以捕获panic异常
		if err := recover(); err != nil {
			logger.Debugf("提交文章时错误 ", err)
		}
	}()

retry:
	request := gorequest.New()
	_, body, _ := request.Post(util.RemoteServerUrl).
		Send(article).
		Timeout(10 * time.Second).
		End()

	if body != "roger" {
		logger.Error("提交文章但没有返回 roger")
		if retryTimes > 0 {
			time.Sleep(3 * time.Second)
			retryTimes -= 1
			goto retry
		}
	} else {
		logger.Debug("文章提交成功")
	}
}
