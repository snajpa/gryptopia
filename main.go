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
		/*if ticker.Label != "HUSH/BTC" {
			continue
		}*/
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
			var tkr ScannerItem

			/*if ticker.Label != "HUSH/BTC" {
				continue
			}*/

			tkr = *scanners[ticker.Label]

			tkr.Mutex.Lock()
			if (thisRun.After(tkr.LastScan)) {
				upToDateCtr += len(tkr.LogData)
				insertLogs = append(insertLogs, tkr.LogData...)
				insertHistories = append(insertHistories, tkr.HistoryData...)
				insertOrders = append(insertOrders, tkr.OrderData...)

				tkr.LastSync = thisRun
				tkr.LogData = nil
				tkr.HistoryData = nil
				tkr.OrderData = nil
			}

			if (tkr.LastFailed) {
				failedNum++
				failedTickers = append(failedTickers, ticker.Label)
			}

			tkr.Mutex.Unlock()
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
		fmt.Printf("log: Timecheck: %s, update took %s\n", time.Since(tbegin), thisRun.Sub(lastRun))
		SleepAtLeast(thisRun, SyncerSleep)
		kkt("}")
	}

}

