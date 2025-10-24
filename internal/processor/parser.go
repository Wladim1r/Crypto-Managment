package processor

import (
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

func (p *Processor) Start() {
	for range 10 {
		go p.worker()
	}
}

func (p *Processor) worker() {
	for rawMsg := range p.inputChan {
		trade, err := p.parse(rawMsg)
		if err != nil {
			slog.Error("Failed to parse message", slog.String("error", err.Error()))
			continue
		}
		slog.Info("Parsed trade",
			slog.String("symbol", trade.Symbol),
			slog.Float64("price", trade.Price),
			slog.String("type", trade.EventType))
		p.outputChan <- trade
	}
}

func (p *Processor) parse(rawMsg []byte) (models.UniversalTrade, error) {
	// Парсим в map для гибкости
	var rawMap map[string]interface{}
	if err := json.Unmarshal(rawMsg, &rawMap); err != nil {
		return models.UniversalTrade{}, fmt.Errorf("could not parse JSON: %w", err)
	}

	// Извлекаем data
	dataVal, hasData := rawMap["data"]
	if !hasData {
		return models.UniversalTrade{}, fmt.Errorf("no data field found")
	}

	// Преобразуем data обратно в JSON bytes
	dataBytes, err := json.Marshal(dataVal)
	if err != nil {
		return models.UniversalTrade{}, fmt.Errorf("could not marshal data: %w", err)
	}

	// Парсим data как map для определения типа события
	var dataMap map[string]interface{}
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return models.UniversalTrade{}, fmt.Errorf("could not parse data: %w", err)
	}

	// Извлекаем тип события
	eVal, hasE := dataMap["e"]
	if !hasE {
		return models.UniversalTrade{}, fmt.Errorf("no event type field found")
	}

	eventType, ok := eVal.(string)
	if !ok {
		return models.UniversalTrade{}, fmt.Errorf("event type is not a string: %v", eVal)
	}

	// Парсим в нужную структуру в зависимости от типа
	var unTrade models.UniversalTrade

	switch eventType {
	case websocket.AggTrade:
		var aggTrade models.AggTrade
		if err := json.Unmarshal(dataBytes, &aggTrade); err != nil {
			return models.UniversalTrade{}, fmt.Errorf("could not parse AggTrade: %w", err)
		}

		unTrade, err = convertAggTradeToUniversalTrade(aggTrade)
		if err != nil {
			return models.UniversalTrade{}, fmt.Errorf("could not convert AggTrade: %w", err)
		}

	case websocket.MiniTicker:
		var miniTicker models.MiniTicker
		if err := json.Unmarshal(dataBytes, &miniTicker); err != nil {
			return models.UniversalTrade{}, fmt.Errorf("could not parse MiniTicker: %w", err)
		}

		unTrade, err = convertMiniTickerToUniversalTrade(miniTicker)
		if err != nil {
			return models.UniversalTrade{}, fmt.Errorf("could not convert MiniTicker: %w", err)
		}

	default:
		return models.UniversalTrade{}, fmt.Errorf("unknown event type: %s", eventType)
	}

	return unTrade, nil
}
