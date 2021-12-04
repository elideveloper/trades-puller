package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elideveloper/trades-puller/currency"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type TradeType string

const (
	InvalidTradeType TradeType = "INVALID"
	Buy              TradeType = "BUY"
	Sell             TradeType = "SELL"
)

type RecentTrade struct {
	ID        string    `json:"id"`        // ID транзакции
	Pair      string    `json:"pair"`      // Торговая пара (из списка выше)
	Price     float64   `json:"price"`     // Цена транзакции
	Amount    float64   `json:"amount"`    // Объем транзакции
	Side      string    `json:"side"`      // Как биржа засчитала эту сделку (как buy или как sell)
	Timestamp time.Time `json:"timestamp"` // Время транзакции
}

func (t *RecentTrade) PrettyPrint(ctx context.Context) {
	pretty, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		ctxzap.Extract(ctx).Error(fmt.Sprintf("failed marshaling RecentTrade into json: %s", err.Error()))
		return
	}

	fmt.Println(string(pretty))
}

type Fields interface {
	GetID() string
	GetPair() currency.Pair
	GetPrice() float64
	GetAmount() float64
	GetTradeType() TradeType
	GetTimestamp() time.Time
}

func NewRecentTrade(fields Fields) RecentTrade {
	return RecentTrade{
		ID:        fields.GetID(),
		Pair:      fields.GetPair().ToString(),
		Price:     fields.GetPrice(),
		Amount:    fields.GetAmount(),
		Side:      string(fields.GetTradeType()),
		Timestamp: fields.GetTimestamp(),
	}
}
