package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/WWoi/web-parcer/internal/models"
	"github.com/WWoi/web-parcer/internal/websocket"
)

type Processor struct {
	inputChan  <-chan []byte
	outputChan chan<- models.UniversalTrade
}

func New(inChan chan []byte, outChan chan models.UniversalTrade) *Processor {
	return &Processor{
		inputChan:  inChan,
		outputChan: outChan,
	}
}

func (p *Processor) Start(ctx context.Context) {
	for range 10 {
		go p.worker(ctx)
	}
}

func (p *Processor) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case rawMsg := <-p.inputChan:
			trades, err := p.parse(rawMsg)
			if err != nil {
				slog.Error("Failed to parse message", "error", err, "raw_message", string(rawMsg))
				continue
			}

			for _, trade := range trades {
				select {
				case <-ctx.Done():
					return
				case p.outputChan <- trade:
				}
			}
		}
	}
}

func (p *Processor) parse(rawMsg []byte) ([]models.UniversalTrade, error) {
	var tickersArray []models.MiniTicker
	if err := json.Unmarshal(rawMsg, &tickersArray); err == nil {
		return p.parseTickerArray(tickersArray, rawMsg)
	}

	// Если не массив, пробуем формат с "stream" и "data"
	var rawMap map[string]interface{}
	if err := json.Unmarshal(rawMsg, &rawMap); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}

	// Проверяем наличие поля data
	dataVal, hasData := rawMap["data"]
	if !hasData {
		return nil, fmt.Errorf("no data field found")
	}

	dataBytes, err := json.Marshal(dataVal)
	if err != nil {
		return nil, fmt.Errorf("could not marshal data: %w", err)
	}

	// Парсим data для определения типа события
	var dataMap map[string]interface{}
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return nil, fmt.Errorf("could not parse data: %w", err)
	}

	eVal, hasE := dataMap["e"]
	if !hasE {
		return nil, fmt.Errorf("no event type field found")
	}

	eventType, ok := eVal.(string)
	if !ok {
		return nil, fmt.Errorf("event type is not a string: %v", eVal)
	}

	var unTrade models.UniversalTrade

	switch eventType {
	case websocket.AggTrade:
		var aggTrade models.AggTrade
		if err := json.Unmarshal(dataBytes, &aggTrade); err != nil {
			return nil, fmt.Errorf("could not parse AggTrade: %w", err)
		}

		unTrade, err = convertAggTradeToUniversalTrade(aggTrade)
		if err != nil {
			return nil, fmt.Errorf("could not convert AggTrade: %w", err)
		}

	case websocket.MiniTicker:
		var miniTicker models.MiniTicker
		if err := json.Unmarshal(dataBytes, &miniTicker); err != nil {
			return nil, fmt.Errorf("could not parse MiniTicker: %w", err)
		}

		unTrade, err = convertMiniTickerToUniversalTrade(miniTicker)
		if err != nil {
			return nil, fmt.Errorf("could not convert MiniTicker: %w", err)
		}

	default:
		slog.Warn("Unknown even type received", "type", eventType)
		return nil, nil
	}

	return []models.UniversalTrade{unTrade}, nil
}

func (p *Processor) parseTickerArray(
	tickers []models.MiniTicker,
	rawMsg []byte,
) ([]models.UniversalTrade, error) {
	trades := make([]models.UniversalTrade, 0, len(tickers))

	for _, ticker := range tickers {
		trade, err := convertMiniTickerToUniversalTrade(ticker)
		if err != nil {
			// Пропускаем невалидные тикеры
			continue
		}
		trades = append(trades, trade)
	}

	if len(trades) == 0 {
		slog.Warn("No valid tickers found in array", "raw_data", string(rawMsg))
		return nil, nil
	}

	return trades, nil
}
