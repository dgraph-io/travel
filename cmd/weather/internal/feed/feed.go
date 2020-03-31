package feed

import (
	"log"
	"net/http"
	"io/ioutil"
)

// Pull extracts the feed from the source.
func Pull(log *log.Logger) error {

	var lat, lon, url string

	lat = "35"
	lon = "139"
	url = "http://api.openweathermap.org/data/2.5/weather"
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }

	// compose the query parameters.
	q := req.URL.Query()
    q.Add("lat", lat)
	q.Add("lon", lon)
	q.Add("appid", "b2302a48062dc1da72430c612557498d")
	req.URL.RawQuery = q.Encode()
	
	// create the request.
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// read the response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	log.Println(string(data))
	return nil
}
