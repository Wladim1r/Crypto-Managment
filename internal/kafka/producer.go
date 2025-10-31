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

type Producer struct {
	writer      *kafka.Writer
	config      KafkaConfig
	inputChan   <-chan *models.DailyStat
	batchBuffer []*models.KafkaMiniTicker
	batchTimer  *time.Timer

	// metrics

	messagesSent   int64
	messagesFailed int64
	batchesSent    int64
}

func NewProducer(cfg KafkaConfig, inputChan <-chan *models.DailyStat) *Producer {
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

func (p *Producer) Start(ctx context.Context) {
	defer p.close()

	for {
		select {
		case <-ctx.Done():
			if len(p.batchBuffer) > 0 {
				p.flushBatch(ctx)
				return
			}
		case stat := <-p.inputChan:

			msgKafka := models.FromDailyStatIntoKafkaMiniTicker(stat, uuid.New().String())
			p.batchBuffer = append(p.batchBuffer, msgKafka)

			if len(p.batchBuffer) >= p.config.BatchSize {
				p.flushBatch(ctx)
				p.batchTimer.Reset(p.config.BatchTimeOut)
			}

		case <-p.batchTimer.C:
			if len(p.batchBuffer) > 0 {
				p.flushBatch(ctx)
			}
			p.batchTimer.Reset(p.config.BatchTimeOut)
		}
	}
}

func (p *Producer) close() {
	if p.batchTimer != nil {
		p.batchTimer.Stop()
	}

	if p.writer != nil {
		p.writer.Close()
	}
}

func (p *Producer) flushBatch(ctx context.Context) {
	kafkaMsgs := make([]kafka.Message, 0, p.config.BatchSize)

	for _, msg := range p.batchBuffer {
		jsonData, err := json.Marshal(msg)
		if err != nil {
			slog.Error("Could not parse msgKafka to JSON", "error", err)
			continue
		}

		msgKafka := kafka.Message{
			Key:   []byte(msg.Symbol),
			Value: jsonData,
			Time:  msg.Timestamp,
			Headers: []kafka.Header{
				{Key: "message_id", Value: []byte(msg.MessageID)},
			},
		}
		kafkaMsgs = append(kafkaMsgs, msgKafka)
	}

	start := time.Now()
	err := p.writer.WriteMessages(ctx, kafkaMsgs...)
	end := time.Since(start)

	if err != nil {
		p.messagesFailed += int64(len(kafkaMsgs))
		slog.Error("Failed to send batch to Kafka",
			"error", err,
			"time sending", end,
			"batch size", len(kafkaMsgs))
	} else {
		p.messagesSent += int64(len(kafkaMsgs))
		p.batchesSent++
		slog.Debug("Sent batch to Kafka",
			"time sending", end,
			"batch size", len(kafkaMsgs),
			"messages sent", p.messagesSent,
			"batches sent", p.batchesSent)
	}

	p.batchBuffer = p.batchBuffer[:0]
}
