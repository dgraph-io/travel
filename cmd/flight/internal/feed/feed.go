package feed

import (
	"io/ioutil"
	"log"
	"net/http"
)

// Pull extracts the feed from the source.
func Pull(log *log.Logger) error {

	key := "5e85cd97f030026c843fbfe0"
	from := "LOND"
	to := "LAX"

	fromDate := "2020-04-12"
	noAdults := "2"
	noChild := "0"
	noInfant := "1"
	cabinClass := "Economy" // Business, Economy, First, PremiumEconomy
	currency := "USD"

	url := "https://api.flightapi.io/onewaytrip/" + key + "/" + from + "/" + to + "/" + fromDate + "/" + 
	noAdults + "/" + noChild + "/" +noInfant + "/" + cabinClass + "/" + currency 
	
	//https://api.flightapi.io/onewaytrip/5e85cd97f030026c843fbfe0/BLR/KTM/2020-04-20/2/1/1/Economy/INR

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println(string(data))
	return nil
}
