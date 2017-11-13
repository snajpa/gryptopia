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

func sleepAtLeast(torig time.Time, d time.Duration) {
	elapsed := time.Since(torig)

	if (elapsed > d) {
		return
	}

	time.Sleep(d - elapsed)
}

func Scanner(res *ScannerItem) {

	res.Mutex.RLock()
	var httpClient = &http.Client{Timeout: HTTPClientTimeout, Transport: &res.netTransport}
	var label = res.Label
	res.Mutex.RUnlock()

	for {
		var tbegin = time.Now()
		var tstamp = time.Now()

		var tmpLogData, err1 = CryptopiaGetMarketLogData(httpClient, label)
		var tmpHistoryData, err2 = CryptopiaGetMarketHistoryData(httpClient, label)
		var tmpOrderData, err3 = CryptopiaGetMarketOrdersData(httpClient, label)
		var failed bool

		if (err1 != nil) || (err2 != nil)  || (err3 != nil) {
			failed = true
		} else {
			failed = false
		}

		tmpLogData.Time = tstamp
		for i, _ := range tmpOrderData {
			tmpOrderData[i].Time = tstamp
		}

		res.Mutex.Lock()

		res.LogData = tmpLogData
		res.HistoryData = tmpHistoryData
		res.OrderData = tmpOrderData
		res.LastRun = tstamp
		res.LastFailed = failed

		res.Mutex.Unlock()

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
			thisRun,
			false,
			sync.RWMutex{},
			ticker.Label,
			http.Transport{
				MaxIdleConns: 16000,
				MaxIdleConnsPerHost: 8000,
				Dial: (&net.Dialer{
					Timeout: HTTPDialTimeout,
				}).Dial,
				TLSHandshakeTimeout: HTTPTLSTimeout,
			},
			CryptopiaMarketLog{},
			[]CryptopiaMarketHistory{},
			[]CryptopiaMarketOrder{},
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
			if (thisRun.After(tkr.LastRun)) {
				upToDateCtr++
				insertLogs = append(insertLogs, tkr.LogData)
				insertHistories = append(insertHistories, tkr.HistoryData...)
				insertOrders = append(insertOrders, tkr.OrderData...)
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
		sleepAtLeast(thisRun, MainSleep)
		kkt("}")
	}

}

