package util

import (
	"database/sql"
)

const (
	GoLimit         = 1 // 并发数
	TitleFileDir    = "./txt/cnNews/title/"
	ContentFileDir  = "./txt/cnNews/content/"
	MinOriginRate   = 60 // 文章原创度采用限制
	RemoteServerUrl = "https://www.sojson.com/httpRequest/"
)

var (
	DB = sql.DB{}

	// 抛弃新浪采集中含有以下关键词的标题
	SinaFilter = []string{
		"新浪",
		"习近平",
		"共产党",
		"政治",
		"美国",
		"中国",
		"学习",
		"强奸",
		"公安",
		"警察",
		"民警",
	}

	// 抛弃含有以下关键词的段落
	ArticleFilter = []string{
		"长按",
		"二维码",
		"点击下图",
		"扫码",
		"转发",
		"本文来源",
		"http://",
		"https://",
		"阅读全文",
		"投稿",
		"版权",
		"军",
	}

	MpFilter = []string{
		"警",
		"公安",
		"党",
		"政治",
		"法",
	}

	// 含以下关键词段落, 其后的内容将全部放弃
	EndHere = []string{
		"分享",
	}
)
