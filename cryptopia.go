package main

import (
	"net/http"
	"encoding/json"
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
