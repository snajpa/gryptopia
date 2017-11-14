package main

import (
	"fmt"
	"time"
	"sync"
	"net/http"
)

const SyncerSleep =			5 * time.Second
const ScannerSleep =		15 * time.Second
const HTTPDialTimeout = 	30 * time.Second
const HTTPTLSTimeout = 		60 * time.Second
const HTTPClientTimeout = 	90 * time.Second

/*type Trade struct {
	Time	time.Time
	Ticker	*Ticker
	IsBuy	bool
	Price	float64
	Amount	float64
	Total	float64
}

func (t Trade) String() string {
	return fmt.Sprintf("Trade<%s %s>", t.Time.String(), t.Ticker.Name)
}*/

type CryptopiaMarketsResponse struct {
	Success string
	Message string
	Data    []CryptopiaMarket
}

type CryptopiaMarketResponse struct {
	Success string
	Message string
	Data    CryptopiaMarketLog
}

type CryptopiaUniqMarket struct {Label string}

type CryptopiaMarket struct {
	Id             int			`json:"tradePairId"`
	Label          string		`sql:"type:varchar(32)"`
	AskPrice       float64
	BidPrice       float64
	Low            float64
	High           float64
	Volume         float64
	LastPrice      float64
	BuyVolume      float64
	SellVolume     float64
	Change         float64
	Open           float64
	Close          float64
	BaseVolume     float64
	BaseBuyVolume  float64
	BaseSellVolume float64
	Time           time.Time
}

const CryptopiaMarketIdxQuery string =
	`CREATE INDEX cryptopia_markets_label_idx ON cryptopia_markes(label);`

func (h CryptopiaMarket) String() string {
	return fmt.Sprintf("CryptopiaMarket<%s open %f change %f last %f @ %s>", h.Label, h.Open, h.Change, h.LastPrice, h.Time)
}

type CryptopiaMarketLog struct {
	TradePairId    int
	Label          string		`sql:"type:varchar(32)"`
	AskPrice       float64
	BidPrice       float64
	Low            float64
	High           float64
	Volume         float64
	LastPrice      float64
	BuyVolume      float64
	SellVolume     float64
	Change         float64
	Open           float64
	Close          float64
	BaseVolume     float64
	BaseBuyVolume  float64
	BaseSellVolume float64
	Time           time.Time
}

const CryptopiaMarketLogIdxQuery string =
	`CREATE INDEX cryptopia_market_logs_label_idx ON cryptopia_market_logs(label);`

func (h CryptopiaMarketLog) String() string {
	return fmt.Sprintf("CryptopiaMarketLog<%s close %f @ %s>", h.Label, h.Close, h.Time)
}

type CryptopiaMarketHistoryResponse struct {
	Success string
	Message string
	Data    []CryptopiaMarketHistory
}

type CryptopiaMarketHistory struct {
	TradePairId		int
	Label			string		`sql:"type:varchar(32)"`
	Type			string		`sql:"type:varchar(8)"`
	Price			float64
	Amount			float64
	Total			float64
	Timestamp 		int64
	Time			time.Time
}

const CryptopiaMarketHistoryIdxQuery string =
	`CREATE INDEX cryptopia_market_histories_label_idx ON cryptopia_market_histories(label);
	CREATE INDEX cryptopia_market_histories_type_idx ON cryptopia_market_histories(type);`

func (h CryptopiaMarketHistory) String() string {
	return fmt.Sprintf("CryptopiaMarketHistory<%s last %f @ %s>", h.Label, h.Price, h.Time)
}

type CryptopiaMarketOrdersResponse struct {
	Success string
	Message string
	Data    struct {
		Buy		[]CryptopiaMarketOrder
		Sell	[]CryptopiaMarketOrder
	}
}

type CryptopiaMarketOrder struct {
	TradePairId		int
	Label			string		`sql:"type:varchar(32)"`
	Type			string		`sql:"type:varchar(8)"`
	Price			float64
	Total			float64
	Time			time.Time
}
func (h CryptopiaMarketOrder) String() string {
	return fmt.Sprintf("CryptopiaMarketOrder<%s price %f total %f @ %s>", h.Label, h.Price, h.Total, h.Time)
}
const CryptopiaMarketOrdersIdxQuery string =
	`CREATE INDEX cryptopia_market_orders_label_idx ON cryptopia_market_orders(label);
	CREATE INDEX cryptopia_market_orders_type_idx ON cryptopia_market_orders(type);
	CREATE UNIQUE INDEX cryptopia_market_orders_unique_idx ON
		cryptopia_market_orders(trade_pair_id,price,total,type,time);`


type ScannerItem struct {
	LastScan    	time.Time
	LastSync		time.Time
	LastFailed		bool
	Mutex			sync.RWMutex
	Label			string
	netTransport	http.Transport
	LogData			[]CryptopiaMarketLog
	HistoryData		[]CryptopiaMarketHistory
	OrderData		[]CryptopiaMarketOrder
}
