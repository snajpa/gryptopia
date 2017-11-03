package main

import (
	"fmt"
	"time"
)


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

type CryptopiaMarket struct {
	Id				int			`json:"tradePairId"`
	Label 			string
	AskPrice 		float64
	BidPrice		float64
	Low				float64
	High			float64
	Volume			float64
	LastPrice		float64
	BuyVolume		float64
	SellVolume		float64
	Change			float64
	Open			float64
	Close			float64
	BaseVolume		float64
	BaseBuyVolume	float64
	BaseSellVolume	float64
	Time	 		time.Time
}

type CryptopiaMarketHistory struct {
	TradePairId		int
	Label 			string
	AskPrice 		float64
	BidPrice		float64
	Low				float64
	High			float64
	Volume			float64
	LastPrice		float64
	BuyVolume		float64
	SellVolume		float64
	Change			float64
	Open			float64
	Close			float64
	BaseVolume		float64
	BaseBuyVolume	float64
	BaseSellVolume	float64
	Time	 		time.Time
}

func (h CryptopiaMarket) String() string {
	return fmt.Sprintf("CryptopiaMarket<%s open %f change %f last %f @ %s>", h.Label, h.Open, h.Change, h.LastPrice, h.Time)
}

func (h CryptopiaMarketHistory) String() string {
	return fmt.Sprintf("CryptopiaMarketHistory<%s last %f @ %s>", h.Label, h.LastPrice, h.Time)
}

