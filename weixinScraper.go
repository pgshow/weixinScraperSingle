package main

import (
	"time"
	"weixinScraperSingle/adsl"
	"weixinScraperSingle/cookiePool"
	"weixinScraperSingle/fetchBing"
	"weixinScraperSingle/fetchSoGou"
	"weixinScraperSingle/sqlite"
)

func main() {
	sqlite.DbInit()

	adsl.ChangeIP()

	cookiePool.Init()
	go cookiePool.Run()

	go fetchSoGou.Run()
	go fetchBing.Run()
	//
	//go collect.Collect()
	//
	//var NewsChan = make(chan *model.Article, 100) // 保存采集到的新闻信息
	//go postArticle.Run(NewsChan)

	for {
		//sogou.Run(NewsChan)
		time.Sleep(time.Minute)
	}
}
