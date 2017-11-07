package main

import (
	"github.com/go-pg/pg"
	"fmt"
	"time"
)


func kkt (msg string) {
	fmt.Printf("log: %s\n", msg)
}

func printErr(err error) {
	if err != nil {
		panic(err)
		fmt.Printf("err pyco <%v>\n", err)
	}
}

func main() {
	var err error
	var tstamp, tbegin time.Time
	var respExist struct { Exists bool}
	var newHistory []CryptopiaMarketLog


	tbegin = time.Now()

	db := pg.Connect(&pg.Options{
		Addr:                  "172.16.0.9:5432",
		User:                  "postgres",
		Password:              "",
		Database:              "test",
	})

	db.QueryOne(&respExist, "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name = 'cryptopia_markets')")

	if !respExist.Exists {
		kkt("DBcreateSchema()")
		err = DBcreateSchema(db)
		printErr(err)
	}

	kkt("CryptopiaGetMarketsData()")
	tstamp = time.Now()
	marketData, err := CryptopiaGetMarketsData()
	printErr(err)

	kkt("DBUpdateMarkets()")
	err = DBUpdateMarkets(db, &marketData, &newHistory, tstamp)
	printErr(err)


	kkt("DBInsertMarketLogs()")
	err = DBInsertMarketLogs(db, &newHistory)
	printErr(err)

	kkt("Traverse Markets and Update Histories")
	uniqMarkets, err := DBGetUniqMarkets(db)
	fmt.Printf("uniqMarkets pyco <%v>\n", uniqMarkets)


	for _, ticker := range uniqMarkets {
		kkt("CryptopiaGetMarketHistoryData("+ticker.Label+")")
		historyData, _, _ := CryptopiaGetMarketHistoryData(ticker.Label)
		fmt.Printf("CryptopiaGetMarketHistoryData pyco <%v>\n", historyData)

	}

	fmt.Printf("Time to completion: %s\n", time.Since(tbegin))
}

