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

func sleepAtLeast(torig time.Time, d time.Duration) {
	elapsed := time.Since(torig)

	if (elapsed > d) {
		return
	}

	time.Sleep(d - elapsed)
}

func Scanner(res *ScannerItem) {


	for {
		var tbegin = time.Now()

		res.Mutex.RLock()
		var label = res.Label
		res.Mutex.RUnlock()

		var tstamp = time.Now()
		var tmpLogData, err1 = CryptopiaGetMarketLogData(label)
		var tmpHistoryData, err2 = CryptopiaGetMarketHistoryData(label)

		tmpLogData.Time = tstamp
		for _, i := range tmpHistoryData {
			i.Time = time.Unix(i.Timestamp, 0)
		}

		//kkt("res.Mutex.Lock() "+res.Label)
		res.Mutex.Lock()
		res.LogData = tmpLogData
		res.HistoryData = tmpHistoryData
		res.LastRun = tstamp

		if (err1 != nil) || (err2 != nil) {
			res.LastFailed = true
		} else {
			res.LastFailed = false
		}

		res.Mutex.Unlock()

		//kkt("res.Mutex.Unlock() "+res.Label)

		sleepAtLeast(tbegin, ScannerSleep)

	}


}

func ScannersWait(atleast time.Time, scanners map[string]*ScannerItem) {
	var done = false

	for _, v := range scanners {
		//kkt("v.Mutex.RLock() "+v.Label)
		for !done {
			time.Sleep(10 * time.Millisecond)
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

	var scanners = make(map[string]*ScannerItem)

	tbegin = time.Now()

	db := pg.Connect(&pg.Options{
		Addr:                  "172.16.0.9:5432",
		User:                  "postgres",
		Password:              "",
		Database:              "test",
	})

	kkt("DBCheckSchema()")
	err = DBCheckSchema(db)
	printErr(err)

	kkt("CryptopiaGetMarketsData()")
	lastRun = time.Now()
	marketData, err := CryptopiaGetMarketsData()
	printErr(err)

	kkt("DBUpdateMarkets()")
	err = DBUpdateMarkets(db, &marketData, lastRun)
	printErr(err)

	kkt("Init scanners")
	uniqMarkets, err := DBGetUniqMarkets(db)
	fmt.Printf("log: Creating %d scanners...\n", len(uniqMarkets))

	for _, ticker := range uniqMarkets {
		/*if ticker.Label != "HUSH/BTC" {
			continue
		}*/
		scanners[ticker.Label] = &ScannerItem{lastRun, false,sync.RWMutex{}, ticker.Label, CryptopiaMarketLog{}, []CryptopiaMarketHistory{}, []CryptopiaMarketOrder{}}

		time.Sleep(10 * time.Millisecond)
		go Scanner(scanners[ticker.Label])
	}

	kkt("ScannersWait()")
	lastRun = time.Now()
	ScannersWait(lastRun, scanners)

	for {
		var insertLogs []CryptopiaMarketLog
		var insertHistories []CryptopiaMarketHistory
		var insertOrders []CryptopiaMarketOrder
		var failedTickers []string
		var failedNum = 0
		var upToDateCtr = 0

		lastEnd := lastRun
		lastRun = time.Now()

		kkt("===================== mainFor {")
		for _, ticker := range uniqMarkets {
			var tkr ScannerItem

			/*if ticker.Label != "HUSH/BTC" {
				continue
			}*/

			tkr = *scanners[ticker.Label]

			tkr.Mutex.RLock()
			if (tkr.LastRun.After(lastEnd)) {
				upToDateCtr++
				insertLogs = append(insertLogs, tkr.LogData)
				insertHistories = append(insertHistories, tkr.HistoryData...)
				insertOrders = append(insertOrders, tkr.OrderData...)
			}

			if (tkr.LastFailed) {
				failedNum++
				failedTickers = append(failedTickers, ticker.Label)
			}

			tkr.Mutex.RUnlock()
		}

		kkt("db.Insert()")

		db.Insert(&insertLogs)
		db.Model(&insertHistories).
			OnConflict("DO NOTHING").
			Insert()

		db.Model(&insertOrders).
			OnConflict("DO NOTHING").
			Insert()

		fmt.Printf("log: Fials: %v\n", failedTickers)
		fmt.Printf("log: FialCount: %v\n", failedNum)
		fmt.Printf("log: upToDateCtr: %d\n", upToDateCtr)
		fmt.Printf("log: Timecheck: %s, update took %s\n", time.Since(tbegin), lastRun.Sub(lastEnd))
		sleepAtLeast(lastRun, MainSleep)
		kkt("}")
	}

}

