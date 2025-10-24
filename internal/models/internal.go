package models

import (
	"sync"
	"time"
)

type UniversalTrade struct {
	// ОБЩИЕ ПОЛЯ (есть у всех типов)
	Symbol    string    `json:"symbol"`     // "BTCUSDT"
	Timestamp time.Time `json:"timestamp"`  // Когда произошло
	EventType string    `json:"event_type"` // "aggTrade", "miniTicker"

	// ОСНОВНАЯ ЦЕНА (всегда заполнено)
	Price float64 `json:"price"` // Текущая/последняя цена

	// ДЛЯ aggTrade
	Quantity     float64 `json:"quantity,omitempty"`       // Объем сделки
	IsBuyerMaker bool    `json:"is_buyer_maker,omitempty"` // Направление

	// ДЛЯ miniTicker (24ч статистика)
	OpenPrice   float64 `json:"open_price,omitempty"`   // Цена открытия 24ч
	HighPrice   float64 `json:"high_price,omitempty"`   // Максимум 24ч
	LowPrice    float64 `json:"low_price,omitempty"`    // Минимум 24ч
	Volume      float64 `json:"volume,omitempty"`       // Объем 24ч
	QuoteVolume float64 `json:"quote_volume,omitempty"` // Объем в quote asset
}

// Window for aggregator @aggTrade
type Window struct {
	Symbol    string
	Interval  string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Quantity  float64
	Trades    int
	StartTime time.Time
	EndTime   time.Time
	TimeStamp time.Time
	Mu        sync.Mutex
}

// DailyStat for aggregator @miniTicker
type DailyStat struct {
	Symbol      string
	OpenPrice   float64
	HighPrice   float64
	LowPrice    float64
	ClosePrice  float64
	Volume      float64
	QuoteVolume float64
	Timestamp   time.Time
}
