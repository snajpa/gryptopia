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

func Scanner(res *ScannerItem, label string, d time.Duration) {

	var tstamp time.Time

	for {
		res.Mutex.Lock()
		res.LogData, _ = CryptopiaGetMarketLogData(label)
		res.HistoryData, _, _ = CryptopiaGetMarketHistoryData(label)

		res.LogData.Time = tstamp
		for _, i := range res.HistoryData {
			i.Time = time.Unix(i.Timestamp, 0)
		}
		res.LastRun = time.Now()
		res.Mutex.Unlock()

		time.Sleep(d)
	}


}

func ScannersWait(atleast time.Time, scanners map[string]*ScannerItem) {
	for _, v := range scanners {
		v.Mutex.RLock()
		for !v.LastRun.After(atleast) {
			time.Sleep(250 * time.Millisecond)
		}
		v.Mutex.RUnlock()
	}
}

func main() {
	var err error

	var lastRun time.Time
	var tbegin time.Time

	var respExist struct { Exists bool}

	var scanners = make(map[string]*ScannerItem)

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
	lastRun = time.Now()
	marketData, err := CryptopiaGetMarketsData()
	printErr(err)

	kkt("DBUpdateMarkets()")
	err = DBUpdateMarkets(db, &marketData, lastRun)
	printErr(err)

	kkt("Init scanners")
	uniqMarkets, err := DBGetUniqMarkets(db)

	lastRun = time.Now()

	for _, ticker := range uniqMarkets {
		scanners[ticker.Label] = &ScannerItem{lastRun, sync.RWMutex{}, CryptopiaMarketLog{}, []CryptopiaMarketHistory{}}

		go Scanner(scanners[ticker.Label], ticker.Label, 30*time.Second)
	}


	for {
		var insertLogs []CryptopiaMarketLog
		var insertHistories []CryptopiaMarketHistory

		kkt("ScannersWait()")
		ScannersWait(lastRun, scanners)

		for _, ticker := range uniqMarkets {

			if ticker.Label != "HUSH/BTC" {
				continue
			}

			var tkr ScannerItem

			tkr = *scanners[ticker.Label]

			insertLogs = append(insertLogs, tkr.LogData)
			insertHistories = append(insertHistories, tkr.HistoryData...)
		}

		kkt("db.Insert()")
		fmt.Printf("Timecheck: %s\n", time.Since(tbegin))
		db.Insert(&insertLogs)
		db.Model(&insertHistories).
			OnConflict("DO NOTHING").
			Insert()

		time.Sleep(20 * time.Second)
		lastRun = time.Now()
	}

}

