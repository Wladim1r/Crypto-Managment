// Package ownlog
package ownlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/fatih/color"
)

const (
	envLocal      = "local"
	envProduction = "prod"

	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

type prettyHandler struct {
	out      io.Writer
	l        *log.Logger
	minLevel slog.Level
}

func (h *prettyHandler) Handle(ctx context.Context, r slog.Record) error {
	var levelColor func(format string, a ...interface{}) string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = color.MagentaString
	case slog.LevelInfo:
		levelColor = color.BlueString
	case slog.LevelWarn:
		levelColor = color.YellowString
	case slog.LevelError:
		levelColor = color.RedString
	default:
		levelColor = color.WhiteString
	}

	timestamp := r.Time.Format("[ 15:04:05 ]")
	level := fmt.Sprintf("%-5s", r.Level.String())
	msg := color.CyanString(r.Message)

	attrs := ""
	r.Attrs(func(a slog.Attr) bool {
		attrs += fmt.Sprintf("%s=%v ", color.HiBlackString(a.Key), a.Value)
		return true
	})

	h.l.Printf("%s %s %s %s", color.HiBlackString(timestamp), levelColor(level), msg, attrs)
	return nil
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &prettyHandler{
		out:      h.out,
		l:        h.l,
		minLevel: h.minLevel,
	}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func newPrettyHandler(out io.Writer, minLevel slog.Level) slog.Handler {
	return &prettyHandler{
		out:      out,
		l:        log.New(out, "", 0),
		minLevel: minLevel,
	}
}

func initColorLog(logLevel string) *slog.Logger {
	level := parseLogLevel(logLevel)
	handler := newPrettyHandler(os.Stdout, level)
	return slog.New(handler)
}

// parseLogLevel преобразует строку в slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case levelDebug:
		return slog.LevelDebug
	case levelInfo:
		return slog.LevelInfo
	case levelWarn:
		return slog.LevelWarn
	case levelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func SetupLogger(env, logLevel string) {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = initColorLog(logLevel)

	case envProduction:
		level := parseLogLevel(logLevel)
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
		logger = slog.New(handler)
	}

	slog.SetDefault(logger)
}
