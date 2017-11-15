package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

const StatusSleep = 5 * time.Second
const ScannerSleep = 15 * time.Second
const HTTPDialTimeout = 30 * time.Second
const HTTPTLSTimeout = 60 * time.Second
const HTTPClientTimeout = 90 * time.Second

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

type CryptopiaUniqMarket struct{ Label string }

type CryptopiaMarket struct {
	Id             int    `json:"tradePairId"`
	Label          string `sql:"type:varchar(32)"`
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

const CryptopiaMarketIdxQuery string = `CREATE INDEX cryptopia_markets_label_idx ON cryptopia_markes(label);`

func (h CryptopiaMarket) String() string {
	return fmt.Sprintf("CryptopiaMarket<%s open %f change %f last %f @ %s>", h.Label, h.Open, h.Change, h.LastPrice, h.Time)
}

type CryptopiaMarketLog struct {
	CryptopiaMarketId int `json:"tradePairId"`
	CryptopiaMarket   *CryptopiaMarket
	AskPrice          float64
	BidPrice          float64
	Low               float64
	High              float64
	Volume            float64
	LastPrice         float64
	BuyVolume         float64
	SellVolume        float64
	Change            float64
	Open              float64
	Close             float64
	BaseVolume        float64
	BaseBuyVolume     float64
	BaseSellVolume    float64
	Time              time.Time
}

const CryptopiaMarketLogIdxQuery string = ``

func (h CryptopiaMarketLog) String() string {
	return fmt.Sprintf("CryptopiaMarketLog<%s close %f @ %s>", h.CryptopiaMarket.Label, h.Close, h.Time)
}

type CryptopiaMarketHistoryResponse struct {
	Success string
	Message string
	Data    []CryptopiaMarketHistory
}

type CryptopiaMarketHistory struct {
	CryptopiaMarketId int `json:"tradePairId"`
	CryptopiaMarket   *CryptopiaMarket
	Type              string `sql:"-"`
	Sell            bool
	Buy            bool
	Price             float64
	Amount            float64
	Total             float64
	Timestamp         int64
	Time              time.Time
}

const CryptopiaMarketHistoryIdxQuery string = ``

func (h CryptopiaMarketHistory) String() string {
	return fmt.Sprintf("CryptopiaMarketHistory<%s last %f @ %s>", h.CryptopiaMarket.Label, h.Price, h.Time)
}

type CryptopiaMarketOrdersResponse struct {
	Success string
	Message string
	Data    struct {
		Buy  []CryptopiaMarketOrder
		Sell []CryptopiaMarketOrder
	}
}

type CryptopiaMarketOrder struct {
	CryptopiaMarketId int `json:"tradePairId"`
	Type              string `sql:"-"`
	Sell            bool
	Buy            bool
	Price             float64
	Total             float64
	Time              time.Time
}

/*func (h CryptopiaMarketOrder) String() string {
	return fmt.Sprintf("CryptopiaMarketOrder<%d price %f total %f @ %s>", h.CryptopiaMarketId, h.Price, h.Total, h.Time)
}*/

const CryptopiaMarketOrdersIdxQuery string = ``

type ScannerItem struct {
	LastScan       time.Time
	LastSync       time.Time
	LastFinish     time.Time
	LastOK         bool
	LastScanTook   time.Duration
	LastSyncTook   time.Duration
	Mutex          sync.RWMutex
	Label          string
	NetTransport   http.Transport
	LogDataLen     CryptopiaMarketLog
	HistoryDataLen int
	OrderDataLen   int
}
