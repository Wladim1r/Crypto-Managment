package aggregator

import (
	"math"
	"sync"
	"time"

	"github.com/WWoi/web-parcer/internal/models"
)

type Aggregator struct {
	inputChan <-chan models.UniversalTrade

	// Хранилище текущих окон
	windows sync.Map // key: "BTCUSDT:1s", value: *Window

	// Для фильтрации
	lastPrices           sync.Map // key: "BTCUSDT", value: float64
	priceChangeThreshold float64  // например 0.0001 (0.01%)
}

type Window struct {
	Symbol    string
	Interval  string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Trades    int
	StartTime time.Time
	EndTime   time.Time
	mu        sync.Mutex
}

func (a *Aggregator) Start() {

}

func (a *Aggregator) processIncoming() {

	for trade := range a.inputChan {

	}

}

// filtering
func (a *Aggregator) ShouldProcess(trade *models.UniversalTrade) bool {
	lastPrice, exists := a.lastPrices.Load(trade.Symbol)
	if !exists {
		return true
	}
	priceInFloat := lastPrice.(float64)
	change := math.Abs((trade.Price - priceInFloat) / priceInFloat)
	return change >= a.priceChangeThreshold
}
