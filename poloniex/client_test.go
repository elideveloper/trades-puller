package poloniex

import (
	"testing"
	"time"

	"github.com/elideveloper/trades-puller/currency"
	"github.com/elideveloper/trades-puller/trade"
	"github.com/stretchr/testify/assert"
)

func TestFetchTradeFields(t *testing.T) {
	testCases := []struct {
		name       string
		msg        string
		needFields []trade.Fields
		needErr    error
	}{
		{
			name:       "empty message, no trade fields",
			msg:        "",
			needFields: []trade.Fields{},
			needErr:    nil,
		},
		{
			name: "one trade is discovered",
			msg:  `[ 121, 8768, [ ["o", 1, "0.00001823", "5534.6474", "1552877119341"], ["o", 0, "0.00001824", "6575.464","1552877119341"], ["t", "42706057", 1, "0.05567134", "0.00181421", 1552877119, "1552877119341"] ] ]`,
			needFields: []trade.Fields{
				&tradeFields{
					id:        "42706057",
					pair:      currency.BTC_USDT,
					tradeType: trade.Buy,
					price:     0.05567134,
					amount:    0.00181421,
					ts: func() time.Time {
						return time.Unix(1552877119, 0)
					}(),
				},
			},
			needErr: nil,
		},
		{
			name: "a few trades are discovered",
			msg:  `[149, 8768, [["o", 1, "0.00001823", "5534.6474", "1562877119341"], ["o", 0, "0.00001824", "6575.464","1562877119341"], ["t", "7854698457", 1, "112.00087", "0.00000081", 1562877119, "1562877119341"],["t", "38748937", 0, "111.09044", "0.00000066", 1562877119, "1562877119341"]]]`,
			needFields: []trade.Fields{
				&tradeFields{
					id:        "7854698457",
					pair:      currency.ETH_USDT,
					tradeType: trade.Buy,
					price:     112.00087,
					amount:    0.00000081,
					ts: func() time.Time {
						return time.Unix(1562877119, 0)
					}(),
				},
				&tradeFields{
					id:        "38748937",
					pair:      currency.ETH_USDT,
					tradeType: trade.Sell,
					price:     111.09044,
					amount:    0.00000066,
					ts: func() time.Time {
						return time.Unix(1562877119, 0)
					}(),
				},
			},
			needErr: nil,
		},
		{
			name:       "no trades in message",
			msg:        `[ 121, 8768, [ ["o", 1, "0.00001823", "5534.6474", "1552877119341"], ["o", 0, "0.00001824", "6575.464","1552877119341"] ] ]`,
			needFields: []trade.Fields{},
			needErr:    nil,
		},
		{
			name:       "invalid trade type",
			msg:        `[ 121, 8768, [ ["o", 1, "0.00001823", "5534.6474", "1552877119341"], ["o", 0, "0.00001824", "6575.464","1552877119341"], ["t", "42706057", 9, "0.05567134", "0.00181421", 1552877119, "1552877119341"] ] ]`,
			needFields: nil,
			needErr:    ErrInvalidTradeType,
		},
		{
			name:       "invalid trade type",
			msg:        `[ 144, 8768, [ ["o", 1, "0.00001823", "5534.6474", "1552877119341"], ["o", 0, "0.00001824", "6575.464","1552877119341"], ["t", "42706057", 0, "0.05567134", "0.00181421", 1552877119, "1552877119341"] ] ]`,
			needFields: nil,
			needErr:    ErrUndefinedPairID,
		},
		{
			name:       "corrupted message with trade block starting",
			msg:        `[ 144, 8768, [ ["o", 1, "0.00001823", "5534.6474", "1552877119341"], ["t", "42706057", 0, "0.0556`,
			needFields: []trade.Fields{},
			needErr:    nil,
		},
		{
			name:       "corrupted message with trade block ending",
			msg:        `6057", 0, "0.05567134", "0.00181421", 1552877119, "1552877119341"] ] ]`,
			needFields: []trade.Fields{},
			needErr:    nil,
		},
		{
			name:       "corrupted message including trade",
			msg:        ` ["t", "42706057", 0, "0.05567134", "0.00181421", 1552877119, "1552877119341"] ]`,
			needFields: nil,
			needErr:    ErrUndefinedPairID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields, err := FetchTradeFields(tc.msg)
			assert.Equal(t, tc.needErr, err)
			assert.Equal(t, tc.needFields, fields)
		})
	}
}
