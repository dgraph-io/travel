package feed

import (
	"io/ioutil"
	"log"
	"net/http"
)

// Pull extracts the feed from the source.
func Pull(log *log.Logger) error {

	
	
	url := "https://api.makcorps.com/free/" 
	city := "sydney"

	url = url + city

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "JWT eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE1ODU4MzE4OTksImlhdCI6MTU4NTgzMDA5OSwibmJmIjoxNTg1ODMwMDk5LCJpZGVudGl0eSI6NDN9.TDuihybgKLEe6EvYrcOSIPnvCp7PAQDg_r_qyAeEU54")

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
