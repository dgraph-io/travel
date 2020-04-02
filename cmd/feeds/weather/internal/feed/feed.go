package feed

import (
	"io/ioutil"
	"log"
	"net/http"
)

// Pull extracts the feed from the source.
func Pull(log *log.Logger) error {
	req, err := http.NewRequest(http.MethodGet, "http://api.openweathermap.org/data/2.5/weather", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("lat", "35")
	q.Add("lon", "139")
	q.Add("appid", "b2302a48062dc1da72430c612557498d")
	req.URL.RawQuery = q.Encode()

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
