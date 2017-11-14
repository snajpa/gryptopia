package main

import (
	"github.com/go-pg/pg"
	"fmt"
	"time"
	"sync"
	"net/http"
	"net"
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

func SleepAtLeast(torig time.Time, d time.Duration) {
	elapsed := time.Since(torig)

	if (elapsed > d) {
		return
	}

	time.Sleep(d - elapsed)
}

func main() {
	var err error

	var thisRun time.Time
	var tbegin time.Time

	var netTransport = &http.Transport{
		MaxIdleConns: 16000,
		MaxIdleConnsPerHost: 8000,
		Dial: (&net.Dialer{
			Timeout: HTTPDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: HTTPTLSTimeout,
	}
	var mainHttpClient = &http.Client{Timeout: HTTPClientTimeout, Transport: netTransport}
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
	thisRun = time.Now()
	marketData, err := CryptopiaGetMarketsData(mainHttpClient)
	printErr(err)

	kkt("DBUpdateMarkets()")
	err = DBUpdateMarkets(db, &marketData, thisRun)
	printErr(err)

	kkt("Init scanners")
	uniqMarkets, err := DBGetUniqMarkets(db)
	fmt.Printf("log: Creating %d scanners...\n", len(uniqMarkets))

	for _, ticker := range uniqMarkets {

		scanners[ticker.Label] = &ScannerItem{
			LastScan: thisRun,
			LastSync: thisRun,
			Mutex:    sync.RWMutex{},
			Label:    ticker.Label,
			netTransport: http.Transport{
				MaxIdleConns: 16000,
				MaxIdleConnsPerHost: 8000,
				Dial: (&net.Dialer{
					Timeout: HTTPDialTimeout,
				}).Dial,
				TLSHandshakeTimeout: HTTPTLSTimeout,
			},
			LogData:     nil,
			HistoryData: nil,
			OrderData:   nil,
		}

		time.Sleep(10 * time.Millisecond)
		go Scanner(scanners[ticker.Label])
	}

	kkt("ScannersWait()")
	thisRun = time.Now()
	ScannersWait(thisRun, scanners)

	for {
		var insertLogs []CryptopiaMarketLog
		var insertHistories []CryptopiaMarketHistory
		var insertOrders []CryptopiaMarketOrder
		var failedTickers []string
		var failedNum = 0
		var upToDateCtr = 0

		lastRun := thisRun
		thisRun = time.Now()

		kkt("===================== mainFor {")
		for _, ticker := range uniqMarkets {

			scanners[ticker.Label].Mutex.Lock()

			if (thisRun.After(scanners[ticker.Label].LastScan)) {
				upToDateCtr += len(scanners[ticker.Label].LogData)
				insertLogs = append(insertLogs, scanners[ticker.Label].LogData...)
				insertHistories = append(insertHistories, scanners[ticker.Label].HistoryData...)
				insertOrders = append(insertOrders, scanners[ticker.Label].OrderData...)

				scanners[ticker.Label].LastSync = thisRun
				scanners[ticker.Label].LogData = nil
				scanners[ticker.Label].HistoryData = nil
				scanners[ticker.Label].OrderData = nil
			}

			if (scanners[ticker.Label].LastFailed) {
				failedNum++
				failedTickers = append(failedTickers, ticker.Label)
			}

			scanners[ticker.Label].Mutex.Unlock()
		}

		fmt.Printf("log: db.Insert(Logs: %d, Histories: %d, Orders: %d)", len(insertLogs), len(insertHistories), len(insertOrders))

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
		fmt.Printf("log: Timecheck: %s, update took %s\n", time.Since(tbegin), thisRun.Sub(lastRun))
		SleepAtLeast(thisRun, SyncerSleep)
		kkt("}")
	}

}

