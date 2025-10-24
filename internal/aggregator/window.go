// Package aggregator
package aggregator

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/WWoi/web-parcer/internal/models"
)

const (
	Interval1h = "1h"
	Interval1d = "1d"
)

type WindowAggregator struct {
	inputChan <-chan models.UniversalTrade

	// windowsMap хранит активные временные окна (свечи).
	// Ключ имеет формат "символ:интервал:время_начала_unix", например, "BTCUSDT:1h:1672531200".
	// Это позволяет иметь уникальное окно для каждого интервала времени.
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

func (wa *WindowAggregator) Start() {
	go wa.processIncoming()
}

func (wa *WindowAggregator) processIncoming() {
	for trade := range wa.inputChan {
		slog.Info("WindowAggregator received trade", "symbol", trade.Symbol, "price", trade.Price)
		// Обрабатываем каждую сделку для построения свечей.
		wa.processAggTrade(trade)
	}
}

func (wa *WindowAggregator) processAggTrade(trade models.UniversalTrade) {
	slog.Info("WindowAggregator processing trade", "symbol", trade.Symbol, "price", trade.Price)
	if !wa.shouldProcess(&trade) {
		return
	}

	wa.updateWindow(trade, Interval1h)
	wa.updateWindow(trade, Interval1d)

	wa.lastPrices.Store(trade.Symbol, trade.Price)
}

func (wa *WindowAggregator) shouldProcess(trade *models.UniversalTrade) bool {
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

	// Округляем время сделки до начала текущего интервала (часа, дня).
	// Например, 10:47:15 станет 10:00:00 для часового интервала.
	// Это гарантирует, что все сделки в одном интервале попадут в одну свечу.
	windowStartTime := trade.Timestamp.Truncate(intervalDuration)

	// --- Логика закрытия старых окон ---
	// Перед обновлением текущего окна, проверяем, не завершились ли другие окна для этого же символа и интервала.
	wa.closeCompletedWindows(trade.Symbol, interval, windowStartTime)

	// --- Обновление текущего окна ---
	// Создаем уникальный ключ для окна, включающий время его начала.
	key := fmt.Sprintf("%s:%s:%d", trade.Symbol, interval, windowStartTime.Unix())

	// Атомарно загружаем или создаем новое окно.
	windowInterface, _ := wa.windowsMap.LoadOrStore(key, &models.Window{
		Symbol:    trade.Symbol,
		Interval:  interval,
		StartTime: windowStartTime,
	})

	window := windowInterface.(*models.Window)

	window.Mu.Lock()
	defer window.Mu.Unlock()

	if window.Open == 0 {
		window.Open = trade.Price
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
	slog.Debug("Window updated", "symbol", trade.Symbol, "interval", interval, "start_time", window.StartTime, "trades", window.Trades)
}

// closeCompletedWindows ищет и закрывает завершенные окна.
// Окно считается завершенным, если его время начала меньше, чем время начала текущего обрабатываемого окна.
func (wa *WindowAggregator) closeCompletedWindows(
	symbol, interval string,
	currentWindowStartTime time.Time,
) {
	wa.windowsMap.Range(func(key, value any) bool {
		keyStr := key.(string)
		parts := strings.Split(keyStr, ":")

		// Проверяем, что ключ соответствует формату и относится к нужному символу и интервалу.
		if len(parts) != 3 || parts[0] != symbol || parts[1] != interval {
			return true
		}

		window := value.(*models.Window)

		// Если время начала сохраненного окна меньше, чем у текущего, значит, оно завершено.
		if window.StartTime.Before(currentWindowStartTime) {
			if window.Trades > 0 {
				wa.outputChanWindow <- window
				slog.Info("Window closed and sent to output", "symbol", window.Symbol, "interval", window.Interval, "start_time", window.StartTime, "trades", window.Trades)
			}
			wa.windowsMap.Delete(keyStr)
		}

		return true
	})
}
