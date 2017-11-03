package main

import (
	"github.com/go-pg/pg"
	"time"
)

func DBcreateSchema(db *pg.DB) error {
	var lasterr error
	var respHT struct { Exists bool}

	for _, model := range []interface{}{&CryptopiaMarket{}, &CryptopiaMarketLog{}, &CryptopiaMarketHistory{}} {

		err := db.CreateTable(model, nil)
		if err != nil {
			lasterr = err
		}
	}

	db.QueryOne(&respHT, "SELECT create_hypertable('cryptopia_market_logs', 'timestamp');")
	db.QueryOne(&respHT, "SELECT create_hypertable('cryptopia_market_histories', 'timestamp');")

	db.QueryOne(&respHT, CryptopiaMarketIdxQuery)
	db.QueryOne(&respHT, CryptopiaMarketLogIdxQuery)
	db.QueryOne(&respHT, CryptopiaMarketHistoryIdxQuery)

	return lasterr
}

func DBUpdateMarkets(db *pg.DB, marketData *[]CryptopiaMarket, newHistory *[]CryptopiaMarketLog, tstamp time.Time) error {
	for _, market := range *marketData {
		var tmpitem CryptopiaMarketLog

		market.Time = tstamp

		tmpitem.TradePairId		= market.Id
		tmpitem.Label 			= market.Label
		tmpitem.AskPrice 		= market.AskPrice
		tmpitem.BidPrice		= market.BidPrice
		tmpitem.Low				= market.Low
		tmpitem.High			= market.High
		tmpitem.Volume			= market.Volume
		tmpitem.LastPrice		= market.LastPrice
		tmpitem.BuyVolume		= market.BuyVolume
		tmpitem.SellVolume		= market.SellVolume
		tmpitem.Change			= market.Change
		tmpitem.Open			= market.Open
		tmpitem.Close			= market.Close
		tmpitem.BaseVolume		= market.BaseVolume
		tmpitem.BaseBuyVolume	= market.BaseBuyVolume
		tmpitem.BaseSellVolume	= market.BaseSellVolume
		tmpitem.Time 			= market.Time

		*newHistory = append(*newHistory, tmpitem)

		_, err := db.Model(&market).
			OnConflict("(id) DO UPDATE").
			Set("ask_price = EXCLUDED.ask_price").
			Insert()

		if err != nil {
			panic(err)
		}


	}
	return nil
}

func DBInsertMarketLogs(db *pg.DB, log *[]CryptopiaMarketLog) error {
	return db.Insert(log)
}

func DBGetUniqMarkets(db *pg.DB) ([]struct{Label string}, error) {
	var resp []struct{Label string}
	var err error

	err = db.Model(&CryptopiaMarket{}).
		ColumnExpr("DISTINCT label").
		Select(&resp)

	return resp, err
}

func DBInsertMarketHistory(db *pg.DB, hist *[]CryptopiaMarketHistory) error {
	var err error

	for _, i := range *hist {
		err = db.Insert(&i)
	}

	return err
}