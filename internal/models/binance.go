package models

type AggTrade struct {
	EventType        string `json:"e"`
	EventTime        int64  `json:"E"`
	Symbol           string `json:"s"`
	AggregateTradeID int64  `json:"a"`
	Price            string `json:"p"`
	Qantity          string `json:"q"`
	FirstTradeID     int64  `json:"f"`
	LastTradeID      int64  `json:"l"`
	TradeTime        int64  `json:"T"`
	IsBuyer          bool   `json:"m"`
	Ignore           bool   `json:"M"`
}

type MiniTicker struct {
	EventType     string `json:"e"`
	EventTime     int64  `json:"E"`
	Symbol        string `json:"s"`
	ClosePrice    string `json:"c"`
	OpenPrice     string `json:"o"`
	HighPrice     string `json:"h"`
	LowPrice      string `json:"l"`
	TotalBaseVol  string `json:"v"`
	TotalQuoteVol string `json:"q"`
}

type BookTicker struct {
	OrderBookUpdateID int64  `json:"u"`
	Symbol            string `json:"s"`
	BestBidPrice      string `json:"b"`
	BestBidQty        string `json:"B"`
	BestAskPrice      string `json:"a"`
	BestAskQty        string `json:"A"`
}
