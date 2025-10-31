package kafka

import (
	"time"

	"github.com/WWoi/web-parcer/internal/lib/getenv"
)

type KafkaConfig struct {
	Brockers         []string // ["localhost:9092"]
	Topic            string   // binance.miniticker
	BatchSize        int
	BatchTimeOut     time.Duration
	CompressionCodec string // Сжатие
	RequiredAcks     int
	MaxAttempts      int
	WriteTimeOut     time.Duration
}

func LoadKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brockers:         getenv.GetSlice("BROKERS", []string{"localhost:9091"}),
		Topic:            getenv.GetString("TOPIC", "binance.miniticker"),
		BatchSize:        getenv.GetInt("BATCH_SIZE", 120),
		BatchTimeOut:     getenv.GetTime("BATCH_TIMEOUT", 2*time.Second),
		CompressionCodec: getenv.GetString("COMPRESSION", "snappy"),
		RequiredAcks:     getenv.GetInt("ACK", 1),
		MaxAttempts:      getenv.GetInt("MAX_ATTEMPTS", 3),
		WriteTimeOut:     getenv.GetTime("WRITE_TIMEOUT", 10*time.Second),
	}
}
