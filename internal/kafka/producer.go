package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/WWoi/web-parcer/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type ProducerConfig struct {
	BrokersURL   []string // url'ы брокеров
	Topic        string
	BatchSize    int           // количество сообщений в батче
	BatchTimeout time.Duration // таймаут батча (1-3 сeк)
	Compression  int           // тип сжатия: Snappy, Gzip, LZ4
	RequiredAcks int           // -1, 0, 1
	MaxAttemps   int           // количество попыток отправки
	WriteTimeout time.Duration // таймаут записи (10s)
}

type Producer struct {
	writer      *kafka.Writer
	config      ProducerConfig
	inputChan   <-chan *models.DailyStat
	batchBuffer []*models.KafkaMiniTicker
	batchTimer  *time.Timer

	// метрики
	messegesSent   int64
	messagesFailed int64
	batchesSent    int64
}

func NewProducer(cfg ProducerConfig, inChan <-chan *models.DailyStat) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.BrokersURL...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Hash{},
		Compression:  compress.Snappy,
		RequiredAcks: kafka.RequireOne, // at least once
		MaxAttempts:  cfg.MaxAttemps,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  10 * time.Second,

		// батчинг
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,

		// асинхронная отправка
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
		inputChan:   inChan,
		batchBuffer: make([]*models.KafkaMiniTicker, 0, cfg.BatchSize),
		batchTimer:  time.NewTimer(cfg.BatchTimeout),
	}
}

func (p *Producer) Start(ctx context.Context) {
	slog.Info("✴️ Kafka producer starting",
		"topic", p.config.Topic,
		"brokers", p.config.BrokersURL,
		"batch_size", p.config.BatchSize,
		"batch_timeout", p.config.BatchTimeout)

	defer p.close()

	for {
		select {
		case <-ctx.Done():
			if len(p.batchBuffer) > 0 {
				p.flushBatch(ctx)
			}
			slog.Info("Kafka producer stopped")

		case stat := <-p.inputChan:
			if stat == nil {
				continue
			}

			msg := models.FromDailyStatIntoKafkaMiniTicker(stat, uuid.New().String())
			p.batchBuffer = append(p.batchBuffer, msg)

			if len(p.batchBuffer) >= p.config.BatchSize {
				p.flushBatch(ctx)
				p.batchTimer.Reset(p.config.BatchTimeout)
			}

		case <-p.batchTimer.C:
			if len(p.batchBuffer) > 0 {
				p.flushBatch(ctx)
			}
			p.batchTimer.Reset(p.config.BatchTimeout)
		}
	}
}

func (p *Producer) flushBatch(ctx context.Context) {
	if len(p.batchBuffer) == 0 {
		return
	}

	batchSize := len(p.batchBuffer)
	messages := make([]kafka.Message, 0, batchSize)

	// prepare batch
	for _, msg := range p.batchBuffer {
		jsonData, err := json.Marshal(msg)
		if err != nil {
			slog.Error("Could not convert into JSON", "error", err, "symbol", msg.Symbol)
			continue
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(msg.Symbol),
			Value: jsonData,
			Time:  msg.Timestamp,
			Headers: []kafka.Header{
				{Key: "message_id", Value: []byte(msg.MessageID)},
			},
		})
	}

	// send batch
	start := time.Now()
	err := p.writer.WriteMessages(ctx, messages...)
	duration := time.Since(start)

	if err != nil {
		p.messagesFailed += int64(len(messages))
		slog.Error("❌ Failed to sent batch to Kafka",
			"error", err,
			"batch_size", len(messages),
			"duration", duration)
	} else {
		p.messegesSent += int64(len(messages))
		p.batchesSent++
		slog.Info("✅ Batch sent to Kafka",
			"batch_size", len(messages),
			"duration", duration,
			"messages_sent", p.messegesSent,
			"batches_sent", p.batchesSent,
		)
	}

	p.batchBuffer = p.batchBuffer[:0]
}

func (p *Producer) close() {
	slog.Info("🚪 Closing Kafka producer",
		"total_messages_sent", p.messegesSent,
		"total_batches_sent", p.batchesSent,
		"messages_failed", p.messagesFailed)

	if p.batchTimer != nil {
		p.batchTimer.Stop()
	}

	if p.writer != nil {
		if err := p.writer.Close(); err != nil {
			slog.Error("Could not close writer",
				"error", err)
		}
	}
}
