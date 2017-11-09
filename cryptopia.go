package main

import (
	"net/http"
	"encoding/json"
	"strings"
	"time"
	"net"
	"fmt"
)


func CryptopiaGetMarketsData() ([]CryptopiaMarket, error) {
	var parsed CryptopiaMarketsResponse
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: HTTPDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: HTTPTLSTimeout,
	}
	var httpClient = &http.Client{
		Timeout: HTTPClientTimeout,
		Transport: netTransport,
	}

	resp, reterr := httpClient.Get("https://cryptopia.co.nz/api/GetMarkets")

	if (resp.StatusCode != 200) {
		panic("kokot")
	}

	json.NewDecoder(resp.Body).Decode(&parsed)

	return parsed.Data, reterr
}

func CryptopiaGetMarketLogData(ticker string) (CryptopiaMarketLog, error) {
	var parsed CryptopiaMarketResponse
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: HTTPDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: HTTPTLSTimeout,
	}
	var httpClient = &http.Client{
		Timeout: HTTPClientTimeout,
		Transport: netTransport,
	}

	var ticker_uscore= strings.Replace(ticker, "/", "_", -1)

	resp, reterr := httpClient.Get("https://cryptopia.co.nz/api/GetMarket/" + ticker_uscore + "/")

	if (reterr != nil) || (resp.StatusCode != 200) {
		fmt.Printf("err: failed http.Get(https://www.cryptopia.co.nz/api/GetMarket/%s/) resp <%v> err <%v>\n", ticker_uscore, resp, reterr)
		return CryptopiaMarketLog{}, reterr
	} else {
		defer resp.Body.Close()
	}

	json.NewDecoder(resp.Body).Decode(&parsed)
	return parsed.Data, reterr
}

func CryptopiaGetMarketHistoryData(ticker string) ([]CryptopiaMarketHistory, error) {
	var parsed CryptopiaMarketHistoryResponse
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: HTTPDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: HTTPTLSTimeout,
	}
	var httpClient = &http.Client{
		Timeout: HTTPClientTimeout,
		Transport: netTransport,
	}

	var ret []CryptopiaMarketHistory

	var ticker_uscore = strings.Replace(ticker, "/", "_", -1)

	resp, reterr := httpClient.Get("https://cryptopia.co.nz/api/GetMarketHistory/"+ticker_uscore+"/")

	if (reterr != nil) || (resp.StatusCode != 200) {
		fmt.Printf("err: failed http.Get(https://cryptopia.co.nz/api/GetMarketHistory/%s/) resp <%v> err <%v>\n", ticker_uscore, resp, reterr)
		return []CryptopiaMarketHistory{}, reterr
	} else {
		defer resp.Body.Close()
	}

	json.NewDecoder(resp.Body).Decode(&parsed)
	for _, i := range parsed.Data {
		i.Time = time.Unix(i.Timestamp, 0)
		ret = append(ret, i)
	}

	return ret, reterr
}
