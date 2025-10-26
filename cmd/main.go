package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/WWoi/web-parcer/internal/aggregator"
	"github.com/WWoi/web-parcer/internal/lib/logger/ownlog"
	"github.com/WWoi/web-parcer/internal/models"
	"github.com/WWoi/web-parcer/internal/processor"
	"github.com/WWoi/web-parcer/internal/websocket"
)

const (
	aggTrade = "wss://stream.binance.com:9443/stream?streams=btcusdt@aggTrade/ethusdt@aggTrade/bnbusdt@aggTrade"

	// All - все монеты, Several - определенные
	// @3000 -> присылает окно каждые 3 секунды (хотя по факту куда реже)
	miniTickerAll     = "wss://stream.binance.com:9443/ws/!miniTicker@arr@3000ms"
	miniTickerSeveral = "wss://stream.binance.com:9443/stream?streams=btcusdt@miniTicker/ethusdt@miniTicker/bnbusdt@miniTicker"
)

func init() {
	ownlog.Init()
}

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

	ws := websocket.New(miniTickerSeveral, rawMessages, 5*time.Second)
	go ws.Start(ctx)

	proc := processor.New(rawMessages, procOut)
	go proc.Start(ctx)

	dailyStatChan := make(chan *models.DailyStat)
	agg := aggregator.NewMetricsProcessor(procOut, dailyStatChan)
	go agg.Start()

	go func() {
		for stat := range dailyStatChan {
			// Вычисляем изменение за 24ч
			change := 0.0
			if stat.OpenPrice > 0 {
				change = ((stat.ClosePrice - stat.OpenPrice) / stat.OpenPrice) * 100
			}

			changeStr := ""
			if change >= 0 {
				changeStr = fmt.Sprintf("📈 +%.2f%%", change)
			} else {
				changeStr = fmt.Sprintf("📉 %.2f%%", change)
			}

			fmt.Printf(
				"📊 24h STATS: %s | Open: %.2f → Close: %.2f | High: %.2f | Low: %.2f | Vol: %.2f | %s\n",
				stat.Symbol,
				stat.OpenPrice,
				stat.ClosePrice,
				stat.HighPrice,
				stat.LowPrice,
				stat.Volume,
				changeStr,
			)
		}
	}()

	// windowsChan := make(chan *models.Window)
	// agg := aggregator.NewWindowAggregator(procOut, windowsChan)
	// go agg.Start(ctx)
	//
	// // Убираем дублирование вывода
	// go func() {
	// 	for window := range windowsChan {
	// 		fmt.Printf(
	// 			"🕯️ CANDLE: %s [%s] | Open: %.2f → Close: %.2f | High: %.2f | Low: %.2f | Vol: %.4f | Trades: %d\n",
	// 			window.Symbol,
	// 			window.Interval,
	// 			window.Open,
	// 			window.Close,
	// 			window.High,
	// 			window.Low,
	// 			window.Quantity,
	// 			window.Trades,
	// 		)
	// 	}
	// }()

	<-ctx.Done()
	fmt.Println("\nshutting down")
	time.Sleep(100 * time.Millisecond)
}
