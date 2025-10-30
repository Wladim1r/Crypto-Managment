package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/WWoi/web-parcer/config"
	"github.com/WWoi/web-parcer/internal/aggregator"
	"github.com/WWoi/web-parcer/internal/lib/logger/ownlog"
	"github.com/WWoi/web-parcer/internal/models"
	"github.com/WWoi/web-parcer/internal/processor"
	"github.com/WWoi/web-parcer/internal/websocket"
	"github.com/joho/godotenv"
)

var cfg *config.Config

func init() {
	godotenv.Load()
	cfg = config.MustLoad()
}

func main() {

	slog.Debug("⚙️ yaml config", "params", cfg)
	ownlog.SetupLogger(cfg.Env, cfg.LogLevel)

	slog.Info("🚀 Application starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		slog.Info("⚠️ Received interrupt signal, preparation for graceful sutdown...")
		cancel()
	}()

	// ========== КАНАЛЫ ==========
	rawMessages := make(chan []byte, 100)
	procOut := make(chan models.UniversalTrade, 100)
	dailyStatChan := make(chan *models.DailyStat, 2000) // буфер для ~2000 монет

	// ========== WEBSOCKET ==========
	ws := websocket.New(websocket.MiniTickerAllURL, rawMessages, 5*time.Second)
	go ws.Start(ctx)

	// ========== PROCESSOR ==========
	proc := processor.New(rawMessages, procOut)
	go proc.Start(ctx)

	// ========== AGGREGATOR ==========
	agg := aggregator.NewMetricsProcessor(procOut, dailyStatChan)
	go agg.Start()

	go func() {
		for stat := range dailyStatChan {
			fmt.Printf(
				"📊 24h STATS: %s | Open: %.2f → Close: %.2f | High: %.2f | Low: %.2f | Vol: %.2f | %s\n",
				stat.Symbol,
				stat.OpenPrice,
				stat.ClosePrice,
				stat.HighPrice,
				stat.LowPrice,
				stat.Volume,
				stat.ChangeFormatted(),
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

	slog.Info("⌛ Wait for completion all the processes")
	time.Sleep(1500 * time.Millisecond)

	slog.Info("👋 Sutdown complete. Goodbye!")
}
