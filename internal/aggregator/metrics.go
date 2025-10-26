package aggregator

import (
	"fmt"
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
		} else {
			slog.Debug("Skipping non-miniTicker event", "type", trade.EventType, "symbol", trade.Symbol)
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
		"24h_change", calculateChange(stat.OpenPrice, stat.ClosePrice))
}

func calculateChange(open, close float64) string {
	if open == 0 {
		return "N/A"
	}
	change := ((close - open) / open) * 100
	if change >= 0 {
		return fmt.Sprintf("+%.2f%%", change)
	}
	return fmt.Sprintf("%.2f%%", change)
}
