package main

import (
	"github.com/go-pg/pg"
	"time"
	"fmt"
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

	db.QueryOne(&respHT, "SELECT create_hypertable('cryptopia_market_logs', 'time');")
	db.QueryOne(&respHT, "SELECT create_hypertable('cryptopia_market_histories', 'timex');")

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
		if tmpitem.Label == "ZEC/BTC" {
			fmt.Printf("DBUpdateMarkets pyco updating <%v>\n", tmpitem)
		}

		*newHistory = append(*newHistory, tmpitem)

		_, err := db.Model(&market).
			OnConflict("(id) DO UPDATE").
			Set("ask_price = EXCLUDED.ask_price").
			Set("bid_price = EXCLUDED.bid_price").
			Set("low = EXCLUDED.low").
			Set("high = EXCLUDED.high").
			Set("volume = EXCLUDED.volume").
			Set("buy_volume = EXCLUDED.buy_volume").
			Set("sell_volume = EXCLUDED.sell_volume").
			Set("change = EXCLUDED.change").
			Set("open = EXCLUDED.open").
			Set("base_volume = EXCLUDED.base_volume").
			Set("base_buy_volume = EXCLUDED.base_buy_volume").
			Set("base_sell_volume = EXCLUDED.base_sell_volume").
			Insert()

		if err != nil {
			panic(err)
		}


	}
	return nil
}

func DBInsertMarketLogs(db *pg.DB, log *[]CryptopiaMarketLog) error {
	for _, i := range *log {
			if i.Label == "ZEC/BTC" {
				fmt.Printf("DBInsertMarketLogs pyco inserting <%v>\n", i)
			}
	}
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
