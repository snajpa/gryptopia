package main

import (
	"net/http"
	"encoding/json"
	"strings"
	"time"
	"fmt"
)

func CryptopiaGetMarketsData(httpClient *http.Client) ([]CryptopiaMarket, error) {
	var parsed CryptopiaMarketsResponse

	resp, reterr := httpClient.Get("https://cryptopia.co.nz/api/GetMarkets")

	if resp != nil {
		defer resp.Body.Close()
	}

	if (resp.StatusCode != 200) {
		panic("kokot")
	}

	json.NewDecoder(resp.Body).Decode(&parsed)

	return parsed.Data, reterr
}

func CryptopiaGetMarketLogData(httpClient *http.Client, ticker string) (CryptopiaMarketLog, error) {
	var parsed CryptopiaMarketResponse


	var ticker_uscore= strings.Replace(ticker, "/", "_", -1)

	resp, reterr := httpClient.Get("https://cryptopia.co.nz/api/GetMarket/" + ticker_uscore + "/1")

	if resp != nil {
		defer resp.Body.Close()
	}

	if (reterr != nil) || (resp.StatusCode != 200) {
		fmt.Printf("err: failed http.Get(https://www.cryptopia.co.nz/api/GetMarket/%s/) resp <%v> err <%v>\n", ticker_uscore, resp, reterr)
		return CryptopiaMarketLog{}, reterr
	}

	json.NewDecoder(resp.Body).Decode(&parsed)

	return parsed.Data, reterr
}

func CryptopiaGetMarketHistoryData(httpClient *http.Client, ticker string) ([]CryptopiaMarketHistory, error) {
	var parsed CryptopiaMarketHistoryResponse

	var ret []CryptopiaMarketHistory

	var ticker_uscore = strings.Replace(ticker, "/", "_", -1)

	resp, reterr := httpClient.Get("https://cryptopia.co.nz/api/GetMarketHistory/"+ticker_uscore+"/1")

	if resp != nil {
		defer resp.Body.Close()
	}

	if (reterr != nil) || (resp.StatusCode != 200) {
		fmt.Printf("err: failed http.Get(https://cryptopia.co.nz/api/GetMarketHistory/%s/) resp <%v> err <%v>\n", ticker_uscore, resp, reterr)
		return []CryptopiaMarketHistory{}, reterr
	}

	json.NewDecoder(resp.Body).Decode(&parsed)

	for _, i := range parsed.Data {
		i.Time = time.Unix(i.Timestamp, 0)
		i.IsSell = (i.Type == "Sell")
		ret = append(ret, i)
	}

	return ret, reterr
}

func CryptopiaGetMarketOrdersData(httpClient *http.Client, ticker string) ([]CryptopiaMarketOrder, error) {
	var parsed CryptopiaMarketOrdersResponse

	var ret []CryptopiaMarketOrder

	var ticker_uscore = strings.Replace(ticker, "/", "_", -1)

	resp, reterr := httpClient.Get("https://cryptopia.co.nz/api/GetMarketOrders/"+ticker_uscore)

	if resp != nil {
		defer resp.Body.Close()
	}

	if (reterr != nil) || (resp.StatusCode != 200) {
		fmt.Printf("err: failed http.Get(https://cryptopia.co.nz/api/GetMarketOrders/%s) resp <%v> err <%v>\n", ticker_uscore, resp, reterr)
		return []CryptopiaMarketOrder{}, reterr
	}

	json.NewDecoder(resp.Body).Decode(&parsed)

	for _, i := range parsed.Data.Buy {
		i.Type = "Buy"
		i.IsSell = false
		ret = append(ret, i)
	}

	for _, i := range parsed.Data.Sell {
		i.Type = "Sell"
		i.IsSell = true
		ret = append(ret, i)
	}

	return ret, reterr
}
