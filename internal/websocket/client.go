// Package websocket
package websocket

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	AggTrade   = "aggTrade"
	MiniTicker = "miniTicker"
)

type WSclient struct {
	url            string
	conn           *websocket.Conn
	outputChan     chan<- []byte
	stopChan       chan struct{}
	reconnectDelay time.Duration
}

func New(url string, output chan<- []byte, time time.Duration) *WSclient {
	return &WSclient{
		url:            url,
		outputChan:     output,
		reconnectDelay: time,
	}
}

func (c *WSclient) Start() {
	slog.Info("Starting websocket client")
	for {
		err := c.connect()
		if err != nil {
			slog.Error("Websocket connection error", slog.String("error", err.Error()))
			time.Sleep(c.reconnectDelay)
			continue
		}

		go c.setupPingPong()

		c.readMessage()
	}
}

func (c *WSclient) connect() error {
	slog.Info("Connecting to websocket", "url", c.url)
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	slog.Info("Websocket connected")
	c.conn = conn

	return nil
}

func (c *WSclient) readMessage() {
	slog.Info("Waiting for messages")
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			slog.Error("Read message error", slog.String("error", err.Error()))
		}
		slog.Info("Received message from websocket")
		c.outputChan <- msg
	}
}
