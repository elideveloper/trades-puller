package poloniex

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/elideveloper/trades-puller/currency"
	"github.com/elideveloper/trades-puller/trade"
	"github.com/elideveloper/trades-puller/wsclient"
	"golang.org/x/net/websocket"
)

const (
	host         = "api2.poloniex.com"
	subscribeCmd = "subscribe"

	defaultBuffSize = 512

	buy  = "1"
	sell = "0"

	tradeChar    = 't'
	blockEndChar = ']'
	delimChar    = ','
)

var (
	ErrInvalidTradeType = errors.New("invalid trade type")
	ErrUndefinedPairID  = errors.New("undefined pair id")

	pairToPoloniexVal = map[currency.Pair]string{
		currency.BTC_USDT: "USDT_BTC",
		currency.TRX_USDT: "USDT_TRX",
		currency.ETH_USDT: "USDT_ETH",
	}

	idToPair = map[string]currency.Pair{
		"121": currency.BTC_USDT,
		"265": currency.TRX_USDT,
		"149": currency.ETH_USDT,
	}
)

type Subscribe struct {
	Command string `json:"command"`
	Channel string `json:"channel"`
}

// tradeFields implements trade.Fields interface
type tradeFields struct {
	id        string
	pair      currency.Pair
	tradeType trade.TradeType
	price     float64
	amount    float64
	ts        time.Time
}

// Client implements wsclient.WsClient interface
type Client struct {
	*websocket.Conn
}

func NewClient() (wsclient.WsClient, error) {
	conn, err := wsclient.GetConnection(host)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn,
	}, nil
}

func (c *Client) Send(msg []byte) error {
	_, err := c.Write(msg)
	return err
}

func (c *Client) Receive() (string, error) {
	msg := make([]byte, defaultBuffSize)
	n, err := c.Read(msg)
	if err != nil {
		return "", err
	}

	return string(msg[:n]), nil
}

func (c *Client) Shutdown() error {
	return c.Close()
}

func NewSubscribeCmd(chanID string) ([]byte, error) {
	subCmd := Subscribe{
		Command: subscribeCmd,
		Channel: chanID,
	}
	cmd, err := json.Marshal(subCmd)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// FetchTradeFields parses poloniex message format https://docs.poloniex.com/#price-aggregated-book
func FetchTradeFields(msg string) ([]trade.Fields, error) {
	fields := make([]trade.Fields, 0)
	startIndex := 0
	pairID := ""
	for i := range msg {
		if i > 1 && pairID == "" && msg[i] == delimChar {
			// first integer value is currency pair id
			pairID = strings.TrimSpace(msg[1:i])
		}

		if msg[i] == tradeChar {
			// found start of trade block
			startIndex = i + 3
		}

		if msg[i] == blockEndChar {
			// if it is end of trade block, then parse it
			if startIndex != 0 && i > startIndex {
				rawFields := strings.Split(msg[startIndex:i], ",")
				f, err := newTradeFields(rawFields, pairID)
				if err != nil {
					return nil, err
				}

				fields = append(fields, f)
				startIndex = 0
			}
		}
	}

	return fields, nil
}

func FormatPair(p currency.Pair) string {
	return pairToPoloniexVal[p]
}

func PairFromID(id string) (currency.Pair, error) {
	if p, found := idToPair[id]; found {
		return p, nil
	}

	return currency.INVALID, ErrUndefinedPairID
}

func newTradeFields(fields []string, pairID string) (*tradeFields, error) {
	rt := &tradeFields{
		id: trimDoubleQuotes(strings.TrimSpace(fields[0])),
	}

	unixTime, err := strconv.ParseInt(strings.TrimSpace(fields[4]), 10, 64)
	if err != nil {
		return nil, err
	}
	rt.ts = time.Unix(unixTime, 0)

	rt.pair, err = PairFromID(pairID)
	if err != nil {
		return nil, err
	}

	rt.tradeType, err = newTradeType(strings.TrimSpace(fields[1]))
	if err != nil {
		return nil, err
	}

	rt.amount, err = strconv.ParseFloat(trimDoubleQuotes(strings.TrimSpace(fields[3])), 64)
	if err != nil {
		return nil, err
	}

	rt.price, err = strconv.ParseFloat(trimDoubleQuotes(strings.TrimSpace(fields[2])), 64)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

func (f *tradeFields) GetID() string {
	return f.id
}
func (f *tradeFields) GetPair() currency.Pair {
	return f.pair
}
func (f *tradeFields) GetPrice() float64 {
	return f.price
}
func (f *tradeFields) GetAmount() float64 {
	return f.amount
}
func (f *tradeFields) GetTradeType() trade.TradeType {
	return f.tradeType
}
func (f *tradeFields) GetTimestamp() time.Time {
	return f.ts
}

func newTradeType(tradeType string) (trade.TradeType, error) {
	switch tradeType {
	case buy:
		return trade.Buy, nil
	case sell:
		return trade.Sell, nil
	}

	return trade.InvalidTradeType, ErrInvalidTradeType
}

func trimDoubleQuotes(s string) string {
	return strings.Trim(s, "\"")
}
