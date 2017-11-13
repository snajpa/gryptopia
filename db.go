package main

import (
	"github.com/go-pg/pg"
	"time"
)

func DBCheckSchema(db *pg.DB) error {
	var lasterr error
	var respTmp struct { Exists bool}

	db.QueryOne(&respTmp, "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name = 'cryptopia_markets')")
	if !respTmp.Exists {
		kkt("db.CreateTable(cryptopia_markets)")
		db.CreateTable(&CryptopiaMarket{}, nil)
		db.QueryOne(&respTmp, CryptopiaMarketIdxQuery)
	}

	db.QueryOne(&respTmp, "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name = 'cryptopia_market_logs')")
	if !respTmp.Exists {
		kkt("db.CreateTable(cryptopia_market_logs)")
		db.CreateTable(&CryptopiaMarketLog{}, nil)
		db.QueryOne(&respTmp, "SELECT create_hypertable('cryptopia_market_logs', 'time');")
		db.QueryOne(&respTmp, CryptopiaMarketLogIdxQuery)
	}

	db.QueryOne(&respTmp, "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name = 'cryptopia_market_histories')")
	if !respTmp.Exists {
		kkt("db.CreateTable(cryptopia_market_histories)")
		db.CreateTable(&CryptopiaMarketHistory{}, nil)
		db.QueryOne(&respTmp, "SELECT create_hypertable('cryptopia_market_histories', 'time', chunk_time_interval => interval '1 day');")
		db.QueryOne(&respTmp, CryptopiaMarketHistoryIdxQuery)
	}

	db.QueryOne(&respTmp, "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name = 'cryptopia_market_orders')")
	if !respTmp.Exists {
		kkt("db.CreateTable(cryptopia_market_orders)")
		lasterr = db.CreateTable(&CryptopiaMarketOrder{}, nil)
		db.QueryOne(&respTmp, "SELECT create_hypertable('cryptopia_market_orders', 'time', chunk_time_interval => interval '1 hour');")
		db.QueryOne(&respTmp, CryptopiaMarketOrdersIdxQuery)
	}

	return lasterr
}

func DBUpdateMarkets(db *pg.DB, marketData *[]CryptopiaMarket, tstamp time.Time) error {
	for _, market := range *marketData {

		market.Time = tstamp

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

func DBGetUniqMarkets(db *pg.DB) ([]struct{Label string}, error) {
	var resp []struct {Label string}
	var err error

	err = db.Model(&CryptopiaMarket{}).
		ColumnExpr("DISTINCT label").
		Select(&resp)

	return resp, err
}
