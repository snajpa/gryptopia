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
		db.QueryOne(&respTmp, "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;")

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
		db.QueryOne(&respTmp, "SELECT create_hypertable('cryptopia_market_orders', 'time', 3);")
		db.QueryOne(&respTmp, "SELECT add_dimension('cryptopia_market_orders', 'cryptopia_market_id', number_partitions => 2048);")
		db.QueryOne(&respTmp, "SELECT add_dimension('cryptopia_market_orders', 'buy', interval_length => 86400000000);")
		db.QueryOne(&respTmp, "SELECT add_dimension('cryptopia_market_orders', 'sell', interval_length => 86400000000);")
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

func DBGetUniqMarkets(db *pg.DB) ([]CryptopiaUniqMarket, error) {
	var resp []CryptopiaUniqMarket
	var err error

	/*resp = append(resp, CryptopiaUniqMarket{"HUSH/BTC"})
	return resp, err*/

	err = db.Model(&CryptopiaMarket{}).
		ColumnExpr("DISTINCT label").
		Select(&resp)

	return resp, err
}
