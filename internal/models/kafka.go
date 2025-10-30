package models

import "time"

type KafkaMiniTicker struct {
    MessageID string `json:"message_id"`

    // :coin: coin data
    Symbol             string    `json:"symbol"`
    OpenPrice          float64   `json:"open_price"`
    HighPrice          float64   `json:"high_price"`
    LowPrice           float64   `json:"low_price"`
    ClosePrice         float64   `json:"close_price"`
    Volume             float64   `json:"volume"`
    QuoteVolume        float64   `json:"quote_volume"`
    ChangePriceMoney   float64   `json:"change_price_money"`
    ChangePricePercent float64   `json:"change_price_percent"`
    Timestamp          time.Time `json:"timestamp"`
}

// type KafkaBatchModel struct {
//     BatchID     string    `json:"batch_id"`
//     BatchSize   int       `json:"batch_size"`
//     CreatedAt   time.Time `json:"created_at"`
//     FirstSymbol string    `json:"first_symbol"`
//     LastSymbol  string    `json:"last_symbol"`
// }

func FromDailyStatIntoKafkaMiniTicker(stat *DailyStat, messageID string) *KafkaMiniTicker {
    return &KafkaMiniTicker{
        MessageID:          messageID,
        Symbol:             stat.Symbol,
        OpenPrice:          stat.OpenPrice,
        HighPrice:          stat.HighPrice,
        LowPrice:           stat.LowPrice,
        ClosePrice:         stat.ClosePrice,
        Volume:             stat.Volume,
        QuoteVolume:        stat.QuoteVolume,
        ChangePriceMoney:   stat.ChangePrice(),
        ChangePricePercent: stat.ChangePercent(),
        Timestamp:          stat.Timestamp,
    }
}