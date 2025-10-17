package ws

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
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
	for {
		err := c.connect()
		if err != nil {
			time.Sleep(c.reconnectDelay)
			continue
		}

		// TODO: witek's job
		go c.handlePinPong()

		c.readMessage()
	}
}

func (c *WSclient) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *WSclient) handlePinPong() {
	// ...
}

func (c *WSclient) readMessage() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			slog.Error("Read message", slog.String("error", err.Error()))
		}

		c.outputChan <- msg
	}
}
