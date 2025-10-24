package aggregator

import (
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
		mp.processMiniTicker(trade)
	}
}

func (mp *MetricsProcessor) processMiniTicker(trade models.UniversalTrade) {
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
}

