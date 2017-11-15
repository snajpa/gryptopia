package main

import (
	"net/http"
	"time"
	"github.com/go-pg/pg"
)

func Scanner(res *ScannerItem) {

	res.Mutex.Lock()
	var httpClient = &http.Client{Timeout: HTTPClientTimeout, Transport: &res.NetTransport}
	var label = res.Label
	res.Mutex.Unlock()

	db := pg.Connect(&pg.Options{
		Addr:                  "172.16.0.9:5432",
		User:                  "postgres",
		Password:              "",
		Database:              "test",
	})

	for {
		var tbegin = time.Now()

		var tmpLogData, err1 = CryptopiaGetMarketLogData(httpClient, label)
		var tmpHistoryData, err2 = CryptopiaGetMarketHistoryData(httpClient, label)
		var tmpOrderData, err3 = CryptopiaGetMarketOrdersData(httpClient, label)
		var failed bool

		if (err1 != nil) || (err2 != nil)  || (err3 != nil) {
			failed = true
		} else {
			failed = false
		}

		tmpLogData.Time = tbegin
		for i, _ := range tmpOrderData {
			tmpOrderData[i].Time = tbegin
			tmpOrderData[i].CryptopiaMarketId = tmpLogData.CryptopiaMarketId
		}

		tsync := time.Now()

		db.Insert(&tmpLogData)
		db.Model(&tmpHistoryData).
			OnConflict("DO NOTHING").
			Insert()

		db.Model(&tmpOrderData).
			OnConflict("DO NOTHING").
			Insert()

		res.Mutex.Lock()

		res.LogDataLen = tmpLogData
		res.HistoryDataLen = len(tmpHistoryData)
		res.OrderDataLen = len(tmpOrderData)

		res.LastOK = !failed

		res.LastScan = tbegin
		res.LastSync = tsync
		res.LastFinish = time.Now()

		res.LastScanTook = tsync.Sub(tbegin)
		res.LastSyncTook = res.LastFinish.Sub(tsync)
		res.Mutex.Unlock()

		SleepAtLeast(tbegin, ScannerSleep)

	}


}

func ScannersWait(atleast time.Time, scanners map[string]*ScannerItem) {
	var done = false

	for _, v := range scanners {
		for !done {
			time.Sleep(10 * time.Millisecond)
			v.Mutex.RLock()
			done = v.LastFinish.After(atleast)
			v.Mutex.RUnlock()
		}
	}
}
