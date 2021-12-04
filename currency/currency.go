package currency

type Pair string

const (
	INVALID  Pair = ""
	BTC_USDT Pair = "BTC_USDT"
	TRX_USDT Pair = "TRX_USDT"
	ETH_USDT Pair = "ETH_USDT"
)

var PairsList = []Pair{BTC_USDT, TRX_USDT, ETH_USDT}

func (p Pair) ToString() string {
	return string(p)
}
