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

// PrettyHandler — наш кастомный обработчик логов.
type prettyHandler struct {
	out io.Writer
	l   *log.Logger
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

// WithAttrs и WithGroup нужны для совместимости с slog.Handler интерфейсом
func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *prettyHandler) WithGroup(name string) slog.Handler       { return h }
func (h *prettyHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func newPrettyHandler(out io.Writer) slog.Handler {
	return &prettyHandler{
		out: out,
		l:   log.New(out, "", 0),
	}
}

func Init() {
	handler := newPrettyHandler(os.Stdout)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
