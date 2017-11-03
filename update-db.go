package main

import (
	"github.com/go-pg/pg"
	"fmt"
	"time"
)


func createSchema(db *pg.DB) error {
	var lasterr error
	var respExist struct { Exists bool}
	var respHT struct { Exists bool}

	respExist.Exists = true

	db.QueryOne(&respExist, "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name = 'cryptopia_markets')")

	if !respExist.Exists {
		for _, model := range []interface{}{&CryptopiaMarket{}, &CryptopiaMarketHistory{}} {

			err := db.CreateTable(model, nil)
			if err != nil {
				lasterr = err
			}
		}

		db.QueryOne(&respHT, "SELECT create_hypertable('cryptopia_market_histories', 'time');")
	}

	return lasterr
}
func kkt (msg string) {
	fmt.Printf("log: %s\n", msg)
}

func printErr(err error) {
	if err != nil {
		fmt.Printf("err pyco <%v>\n", err)
	}
}

func main() {
	var err error
	var tstamp, tbegin time.Time
	var newhistory []CryptopiaMarketHistory

	
	tbegin = time.Now()

	db := pg.Connect(&pg.Options{
		Addr:                  "172.16.0.9:5432",
		User:                  "postgres",
		Password:              "",
		Database:              "test",
	})

	err = createSchema(db)
	printErr(err)

	tstamp = time.Now()
	marketData, err := CryptopiaGetMarketsData()
	printErr(err)

	fmt.Printf("<%v>\n", marketData)

	for _, market := range marketData {
		var tmpitem CryptopiaMarketHistory

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
		tmpitem.Time	 		= market.Time

		newhistory = append(newhistory, tmpitem)
		kkt(fmt.Sprintf("update <%v>\n", tmpitem))
		err = db.Update(&market)
		printErr(err)

	}

	kkt("newhistory")
	err = db.Insert(&newhistory)
	printErr(err)


	fmt.Printf("Time to completion: %s\n", time.Since(tbegin))
}

