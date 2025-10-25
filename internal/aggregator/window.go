// Package aggregator
package aggregator

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/WWoi/web-parcer/internal/models"
)

const (
	Interval10s = "10s"
	Interval1h  = "1h"
	Interval1d  = "1d"
)

type WindowAggregator struct {
	inputChan <-chan models.UniversalTrade

	windowsMap sync.Map // key: <coin_name>:<interval>:<start_time_unix> value: *models.Window

	lastPrices       sync.Map // key: <coin_name> value: <cost>
	outputChanWindow chan<- *models.Window
}

func NewWindowAggregator(
	inChan <-chan models.UniversalTrade,
	outWindown chan<- *models.Window,
) *WindowAggregator {
	return &WindowAggregator{
		inputChan:        inChan,
		outputChanWindow: outWindown,
	}
}

func (wa *WindowAggregator) Start(ctx context.Context) {
	// Запускаем обработчик входящих трейдов
	go wa.processIncoming()

	// Запускаем периодическую проверку завершенных свечей
	go wa.periodicWindowCloser(ctx)

	<-ctx.Done()
}

// periodicWindowCloser периодически проверяет и закрывает завершенные свечи
func (wa *WindowAggregator) periodicWindowCloser(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			wa.closeExpiredWindows(now)
		}
	}
}

// closeExpiredWindows проверяет все активные окна и закрывает те, что уже завершились
func (wa *WindowAggregator) closeExpiredWindows(now time.Time) {
	wa.windowsMap.Range(func(key, value any) bool {
		keyStr := key.(string)
		parts := strings.Split(keyStr, ":")

		if len(parts) != 3 {
			return true
		}

		window := value.(*models.Window)
		interval := parts[1]
		intervalDuration := getIntervalDuration(interval)

		// Время окончания свечи = время начала + длительность интервала
		windowEndTime := window.StartTime.Add(intervalDuration)

		// Если текущее время больше времени окончания, закрываем свечу
		if now.After(windowEndTime) || now.Equal(windowEndTime) {
			window.Mu.Lock()
			if window.Trades > 0 {
				window.EndTime = windowEndTime
				wa.outputChanWindow <- window
			}
			window.Mu.Unlock()
			wa.windowsMap.Delete(keyStr)
		}

		return true
	})
}

func (wa *WindowAggregator) processIncoming() {
	for trade := range wa.inputChan {
		wa.processAggTrade(trade)
	}
}

func (wa *WindowAggregator) processAggTrade(trade models.UniversalTrade) {
	// Всегда обновляем окна для всех трейдов
	wa.updateWindow(trade, Interval10s)
	wa.updateWindow(trade, Interval1h)
	wa.updateWindow(trade, Interval1d)

	// Проверяем, нужно ли обновить lastPrice для уведомлений
	if wa.shouldUpdateLastPrice(&trade) {
		wa.lastPrices.Store(trade.Symbol, trade.Price)
	}
}

// shouldUpdateLastPrice проверяет, достаточно ли изменилась цена для обновления
func (wa *WindowAggregator) shouldUpdateLastPrice(trade *models.UniversalTrade) bool {
	lastPrice, exist := wa.lastPrices.Load(trade.Symbol)
	if !exist {
		return true
	}

	percentForCoin := getPercent(trade.Price)
	lastPrice64 := lastPrice.(float64)

	change := math.Abs((trade.Price - lastPrice64) / lastPrice64)

	return change >= percentForCoin
}

func getIntervalDuration(interval string) time.Duration {
	switch interval {
	case Interval10s:
		return 10 * time.Second
	case Interval1h:
		return time.Hour
	case Interval1d:
		return 24 * time.Hour
	default:
		return 0
	}
}

func (wa *WindowAggregator) updateWindow(trade models.UniversalTrade, interval string) {
	intervalDuration := getIntervalDuration(interval)
	if intervalDuration == 0 {
		return
	}

	// ВАЖНО: Используем текущее время, а не время из трейда
	now := time.Now()
	windowStartTime := now.Truncate(intervalDuration)
	key := fmt.Sprintf("%s:%s:%d", trade.Symbol, interval, windowStartTime.Unix())

	windowInterface, isNew := wa.windowsMap.LoadOrStore(key, &models.Window{
		Symbol:    trade.Symbol,
		Interval:  interval,
		StartTime: windowStartTime,
	})

	window := windowInterface.(*models.Window)

	window.Mu.Lock()
	defer window.Mu.Unlock()

	if window.Open == 0 {
		window.Open = trade.Price
		if isNew {
			slog.Info("📊 Window started",
				"symbol", trade.Symbol,
				"interval", interval,
				"start", windowStartTime.Format("15:04:05"))
		}
	}

	window.Close = trade.Price

	if trade.Price > window.High {
		window.High = trade.Price
	}

	if trade.Price < window.Low || window.Low == 0 {
		window.Low = trade.Price
	}

	window.Quantity += trade.Quantity
	window.Trades++
}
