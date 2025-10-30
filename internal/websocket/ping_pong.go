package websocket

import (
	"context"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

func (c *WSclient) setupPingPong(ctx context.Context) {

	c.conn.SetPingHandler(func(appData string) error {

		slog.Info("Ping from Binance, answer pong")

		// Отправляем pong обратно

		err := c.conn.WriteControl(

			websocket.PongMessage, // Тип: PONG

			[]byte(appData), // Тот же payload

			time.Now().Add(1*time.Second), // Deadline

		)

		if err != nil {

			slog.Error("Failed to send pong", "error", err)

			return err

		}

		return nil

	})

	// health check

	ticker := time.NewTicker(3 * time.Second)

	defer ticker.Stop()

	for {

		select {

		case <-ticker.C:

			slog.Debug("Sending ping to Binance")

			if err := c.conn.WriteControl(

				websocket.PingMessage,

				[]byte{},

				time.Now().Add(5*time.Second),
			); err != nil {

				slog.Warn("Failed to send ping", "error", err)

				return // выход из горутины

			}

		case <-ctx.Done():

			slog.Info("Ping-pong handler stopped")

			return

		}

	}

}
