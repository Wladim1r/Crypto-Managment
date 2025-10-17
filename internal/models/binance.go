package models

type AggTrade struct {
	EventType        string `json:"e"`
	EventTime        string `json:"E"`
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
	EventTime     string `json:"E"`
	Symbol        string `json:"s"`
	ClosePrice    string `json:"c"`
	OpenPrice     string `json:"o"`
	HighPrice     string `json:"h"`
	LowPrice      string `json:"l"`
	TotalBaseVol  string `json:"v"`
	TotalQuoteVol string `json:"q"`
}

type Ticker24hr struct {
	EventType              string `json:"e"`
	EventTime              int64  `json:"E"`
	Symbol                 string `json:"s"`
	PriceChange            string `json:"p"`
	PriceChangePercent     string `json:"P"`
	WeightedAveragePrice   string `json:"w"`
	FirstTradePrice        string `json:"x"`
	LastPrice              string `json:"c"`
	LastQuantity           string `json:"Q"`
	BestBidPrice           string `json:"b"`
	BestBidQuantity        string `json:"B"`
	BestAskPrice           string `json:"a"`
	BestAskQuantity        string `json:"A"`
	OpenPrice              string `json:"o"`
	HighPrice              string `json:"h"`
	LowPrice               string `json:"l"`
	TotalTradedBaseVolume  string `json:"v"`
	TotalTradedQuoteVolume string `json:"q"`
	StatisticsOpenTime     int64  `json:"O"`
	StatisticsCloseTime    int64  `json:"C"`
	FirstTradeID           int64  `json:"F"`
	LastTradeID            int64  `json:"L"`
	TotalNumberOfTrades    int64  `json:"n"`
}

type TickerWindow struct {
	EventType              string `json:"e"`
	EventTime              int64  `json:"E"`
	Symbol                 string `json:"s"`
	PriceChange            string `json:"p"`
	PriceChangePercent     string `json:"P"`
	OpenPrice              string `json:"o"`
	HighPrice              string `json:"h"`
	LowPrice               string `json:"l"`
	LastPrice              string `json:"c"`
	WeightedAveragePrice   string `json:"w"`
	TotalTradedBaseVolume  string `json:"v"`
	TotalTradedQuoteVolume string `json:"q"`
	StatisticsOpenTime     int64  `json:"O"`
	StatisticsCloseTime    int64  `json:"C"`
	FirstTradeID           int64  `json:"F"`
	LastTradeID            int64  `json:"L"`
	TotalNumberOfTrades    int64  `json:"n"`
}

type BookTicker struct {
	OrderBookUpdateID int64  `json:"u"`
	Symbol            string `json:"s"`
	BestBidPrice      string `json:"b"`
	BestBidQty        string `json:"B"`
	BestAskPrice      string `json:"a"`
	BestAskQty        string `json:"A"`
}

type AvgPrice struct {
	EventType            string `json:"e"`
	EventTime            int64  `json:"E"`
	Symbol               string `json:"s"`
	AveragePriceInterval string `json:"i"`
	AveragePrice         string `json:"w"`
	LastTradeTime        int64  `json:"T"`
}
