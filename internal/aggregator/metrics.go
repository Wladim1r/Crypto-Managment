package aggregator

import (
	"log/slog"

	"github.com/WWoi/web-parcer/internal/models"
)

// MetricsProcessor handles the processing of daily statistics.
type MetricsProcessor struct {
	inputChan           <-chan models.UniversalTrade
	outputChanDailyStat chan<- *models.DailyStat
}

// NewMetricsProcessor creates a new MetricsProcessor.
func NewMetricsProcessor(
	inChan <-chan models.UniversalTrade,
	outDayilyStat chan<- *models.DailyStat,
) *MetricsProcessor {
	return &MetricsProcessor{
		inputChan:           inChan,
		outputChanDailyStat: outDayilyStat,
	}
}

func (mp *MetricsProcessor) Start() {
	go mp.processIncoming()
}

func (mp *MetricsProcessor) processIncoming() {
	for trade := range mp.inputChan {
		// ВАЖНО: Обрабатываем только miniTicker события
		if trade.EventType == "24hrMiniTicker" {
			mp.processMiniTicker(trade)
		}
	}
}

func (mp *MetricsProcessor) processMiniTicker(trade models.UniversalTrade) {
	// Проверяем, что у нас есть данные miniTicker
	if trade.OpenPrice == 0 && trade.HighPrice == 0 && trade.LowPrice == 0 {
		slog.Warn("Received miniTicker with empty OHLC data", "symbol", trade.Symbol)
		return
	}

	stat := &models.DailyStat{
		Symbol:      trade.Symbol,
		OpenPrice:   trade.OpenPrice,
		HighPrice:   trade.HighPrice,
		LowPrice:    trade.LowPrice,
		ClosePrice:  trade.Price,
		Volume:      trade.Volume,
		QuoteVolume: trade.QuoteVolume,
		Timestamp:   trade.Timestamp,
	}

	mp.outputChanDailyStat <- stat
	slog.Info("📊 Daily stat processed",
		"symbol", stat.Symbol,
		"close", stat.ClosePrice,
		"24h_change", stat.ChangeFormatted())
}
