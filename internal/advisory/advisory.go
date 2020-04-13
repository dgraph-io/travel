package advisory

import (
	"io/ioutil"
	"log"
	"net/http"
)

// Retrieve finds advisories for the specified location.
func Retrieve(log *log.Logger) error {
	req, err := http.NewRequest(http.MethodGet, "https://www.travel-advisory.info/api", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("countrycode", "AU")

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
