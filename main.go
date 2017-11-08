package main

import (
	"github.com/go-pg/pg"
	"fmt"
	"time"
	"sync"
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

func FetchTicker(wg *sync.WaitGroup, res *ResultMapItem, label string) {
	var tstamp = time.Now()

	defer wg.Done()

	res.Mutex.Lock()

	res.LogData, _ = CryptopiaGetMarketLogData(label)
	res.HistoryData, _, _ = CryptopiaGetMarketHistoryData(label)

	res.LogData.Time = tstamp
	for _, i := range res.HistoryData {
		i.Time = time.Unix(i.Timestamp, 0)
	}

	res.Mutex.Unlock()
}

func main() {
	var err error

	var tstamp, tbegin time.Time

	var respExist struct { Exists bool}

	var resultMap = make(map[string]*ResultMapItem)
	var wg sync.WaitGroup

	var insertLogs []CryptopiaMarketLog
	var insertHistories []CryptopiaMarketHistory

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
	err = DBUpdateMarkets(db, &marketData, tstamp)
	printErr(err)

	kkt("Traverse Markets and Update Histories")
	uniqMarkets, err := DBGetUniqMarkets(db)

	for _, ticker := range uniqMarkets {

		resultMap[ticker.Label] = &ResultMapItem{ sync.RWMutex{}, CryptopiaMarketLog{}, []CryptopiaMarketHistory{}}

		kkt("go FetchTicker("+ticker.Label+")")
		wg.Add(1)
		go FetchTicker(&wg, resultMap[ticker.Label], ticker.Label)
	}

	kkt("wg.Wait()")
	wg.Wait()

	for _, ticker := range uniqMarkets {
		var tkr ResultMapItem

		tkr = *resultMap[ticker.Label]

		insertLogs = append(insertLogs, tkr.LogData)
		insertHistories = append(insertHistories, tkr.HistoryData...)
	}

	kkt("db.Insert()")
	db.Insert(&insertLogs)
	db.Insert(&insertHistories)

	fmt.Printf("Time to completion: %s\n", time.Since(tbegin))
}

