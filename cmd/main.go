package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/WWoi/web-parcer/internal/aggregator"
	"github.com/WWoi/web-parcer/internal/models"
	"github.com/WWoi/web-parcer/internal/ownlog"
	"github.com/WWoi/web-parcer/internal/processor"
	"github.com/WWoi/web-parcer/internal/websocket"
)

func init() {
	ownlog.Init()
}

func main() {

	slog.Info("Starting application")

	// Channels
	rawMessagesChan := make(chan []byte, 1000)
	universalTradeChan := make(chan models.UniversalTrade, 1000)
	windowChan := make(chan *models.Window, 100)
	dailyStatChan := make(chan *models.DailyStat, 100)

	// Websocket client
	slog.Info("Initializing websocket client")
	wsClient := websocket.New(
		"wss://stream.binance.com:9443/stream?streams=btcusdt@aggTrade/ethusdt@aggTrade/bnbusdt@aggTrade",
		rawMessagesChan,
		5*time.Second,
	)
	go wsClient.Start()

	// Processor
	slog.Info("Initializing processor")
	proc := processor.New(rawMessagesChan, universalTradeChan)
	proc.Start()

	// Aggregators
	slog.Info("Initializing window aggregator")
	windowAggregator := aggregator.NewWindowAggregator(universalTradeChan, windowChan)
	windowAggregator.Start()

	slog.Info("Initializing metrics processor")
	metricsProcessor := aggregator.NewMetricsProcessor(universalTradeChan, dailyStatChan)
	metricsProcessor.Start()

	// Printer
	go func() {
		for {
			select {
			case window := <-windowChan:
				fmt.Printf("Window: %+v\n", window)
			case dailyStat := <-dailyStatChan:
				fmt.Printf("Daily Stat: %+v\n", dailyStat)
			}
		}
	}()

	// Keep the main goroutine alive
	select {}
}
