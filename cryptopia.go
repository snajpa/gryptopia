package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)


func CryptopiaGetMarketsData() ([]CryptopiaMarket, error) {
	var parsed CryptopiaMarketsResponse

	resp, reterr := http.Get("https://www.cryptopia.co.nz/api/GetMarkets")

	if (resp.StatusCode != 200) {
		panic("kokot")
	}

	json.NewDecoder(resp.Body).Decode(&parsed)

	return parsed.Data, reterr
}

func CryptopiaGetMarketLogData(ticker string) (CryptopiaMarketLog, error) {
	var parsed CryptopiaMarketResponse

	var ticker_uscore= strings.Replace(ticker, "/", "_", -1)

	resp, reterr := http.Get("https://www.cryptopia.co.nz/api/GetMarket/" + ticker_uscore + "/")

	if (resp.StatusCode != 200) {
		fmt.Printf("resp http pyco <%v>\n", resp)
	}

	json.NewDecoder(resp.Body).Decode(&parsed)

	return parsed.Data, reterr
}

func CryptopiaGetMarketHistoryData(ticker string) ([]CryptopiaMarketHistory, []string, error) {
	var parsed CryptopiaMarketHistoryResponse
	var failedTickers []string
	var ret []CryptopiaMarketHistory

	var ticker_uscore = strings.Replace(ticker, "/", "_", -1)

	resp, reterr := http.Get("https://www.cryptopia.co.nz/api/GetMarketHistory/"+ticker_uscore+"/")

	if (resp.StatusCode != 200) {
		failedTickers = append(failedTickers, ticker)
		fmt.Printf("resp http pyco <%v>\n", resp)
	}

	json.NewDecoder(resp.Body).Decode(&parsed)
	for _, i := range parsed.Data {
		i.Time = time.Unix(i.Timestamp, 0)
		ret = append(ret, i)
	}

	return ret, failedTickers, reterr
}
