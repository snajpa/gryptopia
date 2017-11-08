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

func Scanner(res *ScannerItem, d time.Duration) {


	for {
		var tstamp = time.Now()
		var tmpLogData, _= CryptopiaGetMarketLogData(res.Label)
		var tmpHistoryData, _, _ = CryptopiaGetMarketHistoryData(res.Label)

		tmpLogData.Time = tstamp
		for _, i := range tmpHistoryData {
			i.Time = time.Unix(i.Timestamp, 0)
		}

		//kkt("res.Mutex.Lock() "+res.Label)
		res.Mutex.Lock()
		res.LogData = tmpLogData
		res.HistoryData = tmpHistoryData
		res.LastRun = tstamp
		res.Mutex.Unlock()

		//kkt("res.Mutex.Unlock() "+res.Label)

		time.Sleep(d)
	}


}

func ScannersWait(atleast time.Time, scanners map[string]*ScannerItem) {
	var done = false

	for _, v := range scanners {
		//kkt("v.Mutex.RLock() "+v.Label)
		for !done {
			time.Sleep(50 * time.Millisecond)
			v.Mutex.RLock()
			done = v.LastRun.After(atleast)
			v.Mutex.RUnlock()
		}
		//kkt("v.Mutex.RUnlock() "+v.Label)
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
		scanners[ticker.Label] = &ScannerItem{lastRun, sync.RWMutex{}, ticker.Label, CryptopiaMarketLog{}, []CryptopiaMarketHistory{}}

		go Scanner(scanners[ticker.Label], 30*time.Second)
	}


	for {
		var insertLogs []CryptopiaMarketLog
		var insertHistories []CryptopiaMarketHistory

		kkt("ScannersWait()")
		ScannersWait(lastRun, scanners)
		lastRun = time.Now()

		for _, ticker := range uniqMarkets {
			var tkr ScannerItem

			tkr = *scanners[ticker.Label]

			tkr.Mutex.RLock()
			insertLogs = append(insertLogs, tkr.LogData)
			insertHistories = append(insertHistories, tkr.HistoryData...)
			tkr.Mutex.RUnlock()
		}

		kkt("db.Insert()")
		fmt.Printf("Timecheck: %s\n", time.Since(tbegin))
		db.Insert(&insertLogs)
		db.Model(&insertHistories).
			OnConflict("DO NOTHING").
			Insert()

		time.Sleep(20 * time.Second)
	}

}

