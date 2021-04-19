package model

type WeixinMp struct {
	Mp        string
	MpName    string
	RenewTime int
}

type SearchInfo struct {
	MpName string
	MpPage string
	Brief  string
}

type Article struct {
	Title   string
	Content string
}
