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
			thisRun,
			thisRun,
			thisRun,
			false,
			0,
			0,
			sync.RWMutex{},
			ticker.Label,
			http.Transport{
				MaxIdleConns:        16000,
				MaxIdleConnsPerHost: 8000,
				Dial: (&net.Dialer{
					Timeout: HTTPDialTimeout,
				}).Dial,
				TLSHandshakeTimeout: HTTPTLSTimeout,
			},
			CryptopiaMarketLog{},
			0,
			0,
		}

		time.Sleep(10 * time.Millisecond)
		go Scanner(scanners[ticker.Label])
	}

	kkt("ScannersWait()")
	thisRun = time.Now()
	ScannersWait(thisRun, scanners)

	for {
		var tickerCnt = len(scanners)

		var failCnt, upToDateCtr, okCnt int
		var scanTot, scanAvg, syncTot, syncAvg time.Duration
		var thisRun time.Time

		failCnt = 0
		upToDateCtr = 0
		okCnt = 0
		thisRun = time.Now()

		for _, t := range uniqMarkets {
			tickerCnt++

			scanners[t.Label].Mutex.RLock()
			lastScanTook 	:= scanners[t.Label].LastScanTook
			lastSyncTook 	:= scanners[t.Label].LastSyncTook
			lastFinish 		:= scanners[t.Label].LastFinish
			lastOK 			:= scanners[t.Label].LastOK
			scanners[t.Label].Mutex.RUnlock()

			if (lastOK) {
				scanTot += lastScanTook
				syncTot += lastSyncTook
				okCnt++
			} else {
				failCnt++
			}

			if lastFinish.After(thisRun.Add(-ScannerSleep)){
				upToDateCtr++
			}
		}

		scanAvg = time.Duration(int64(scanTot / time.Nanosecond) / int64(okCnt)) * time.Nanosecond
		syncAvg = time.Duration(int64(syncTot / time.Nanosecond) / int64(okCnt)) * time.Nanosecond

		fmt.Printf("log: failCnt: %d\n", failCnt)
		fmt.Printf("log: upToDateCtr: %d\n", upToDateCtr)
		fmt.Printf("log: scanAvg: %s\n", scanAvg)
		fmt.Printf("log: syncAvg: %s\n", syncAvg)

		fmt.Printf("log: Timecheck: %s\n", time.Since(tbegin))
		SleepAtLeast(thisRun, StatusSleep)
		kkt("}")
	}

}

