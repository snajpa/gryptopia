package main

import (
	"net/http"
	"time"
)

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

		res.LogData = append(res.LogData, tmpLogData)
		res.HistoryData = append(res.HistoryData, tmpHistoryData...)
		res.OrderData = append(res.OrderData, tmpOrderData...)
		res.LastScan = tstamp
		res.LastFailed = failed

		res.Mutex.Unlock()

		SleepAtLeast(tbegin, ScannerSleep)

	}


}

func ScannersWait(atleast time.Time, scanners map[string]*ScannerItem) {
	var done = false

	for _, v := range scanners {
		//kkt("v.Mutex.RLock() "+v.Label)
		for !done {
			time.Sleep(10 * time.Millisecond)
			v.Mutex.RLock()
			done = v.LastScan.After(atleast)
			v.Mutex.RUnlock()
		}
		//kkt("v.Mutex.RUnlock() "+v.Label)
	}
}
