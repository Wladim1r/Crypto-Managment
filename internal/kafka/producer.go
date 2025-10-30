package kafka

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/WWoi/web-parcer/internal/models"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type ProducerConfig struct {
	Brockers         []string // ["localhost:9092"]
	Topic            string   // binance.miniticker
	BatchSize        int
	BatchTimeOut     time.Duration
	CompressionCodec int // Сжатие
	RequiredAcks     int
	MaxAttempts      int
	WriteTimeOut     time.Duration
}

type Producer struct {
	writer      *kafka.Writer
	config      ProducerConfig
	inputChan   <-chan *models.DailyStat
	batchBuffer []*models.KafkaMiniTicker
	batchTimer  *time.Timer

	// metrics

	messagesSent   int64
	messagesFailed int64
	batchesSent    int64
}

func NewProducer(cfg ProducerConfig, inputChan <-chan *models.DailyStat) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brockers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Hash{}, // partition
		Compression:  compress.Snappy,
		RequiredAcks: kafka.RequireOne,
		MaxAttempts:  cfg.MaxAttempts,
		WriteTimeout: cfg.WriteTimeOut,
		ReadTimeout:  10 * time.Second,

		// Batch
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeOut,

		//Asynchronous sending for performance
		Async: false,

		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			slog.Debug("Kafka writer", "message", fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			slog.Error("Kafka writer error", "message", fmt.Sprintf(msg, args...))
		}),
	}

	return &Producer{
		writer:      writer,
		config:      cfg,
		inputChan:   inputChan,
		batchBuffer: make([]*models.KafkaMiniTicker, 0, cfg.BatchSize),
		batchTimer:  time.NewTimer(cfg.BatchTimeOut),
	}
}

func (p *Producer) Start() {

}
