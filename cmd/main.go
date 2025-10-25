package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/WWoi/web-parcer/internal/aggregator"
	"github.com/WWoi/web-parcer/internal/models"
	"github.com/WWoi/web-parcer/internal/processor"
	"github.com/WWoi/web-parcer/internal/websocket"
)

func main() {
	fmt.Println("crypto-asset-tracker starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		cancel()
	}()

	rawMessages := make(chan []byte, 100)
	procOut := make(chan models.UniversalTrade, 100)

	ws := websocket.New("wss://stream.binance.com:9443/stream?streams=btcusdt@aggTrade/ethusdt@aggTrade/bnbusdt@aggTrade", rawMessages, 5*time.Second)
	go ws.Start(ctx)

	proc := processor.New(rawMessages, procOut)
	go proc.Start(ctx)

	windowsChan := make(chan *models.Window)
	agg := aggregator.NewWindowAggregator(procOut, windowsChan)
	go agg.Start(ctx)

	// Убираем дублирование вывода
	go func() {
		for window := range windowsChan {
			fmt.Printf(
				"🕯️ CANDLE: %s [%s] | Open: %.2f → Close: %.2f | High: %.2f | Low: %.2f | Vol: %.4f | Trades: %d\n",
				window.Symbol,
				window.Interval,
				window.Open,
				window.Close,
				window.High,
				window.Low,
				window.Quantity,
				window.Trades,
			)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nshutting down")
	time.Sleep(100 * time.Millisecond)
}
