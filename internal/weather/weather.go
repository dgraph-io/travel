package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Weather contains the weather data points captured from the API.
type Weather struct {
	Name  string `json:"name"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lon"`
	} `json:"coord"`
	Sky []struct {
		Visibility string `json:"main"`
		Desc       string `json:"description"`
	} `json:"weather"`
	Points struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		MinTemp   float64 `json:"temp_min"`
		MaxTemp   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed     float64 `json:"speed"`
		Direction int     `json:"deg"`
	} `json:"wind"`
	RiseSet struct {
		Sunrise int `json:"sunrise"`
		Sunset  int `json:"sunset"`
	} `json:"sys"`
}

// Search can locate weather for a given latitude and longitude.
// Here is the output https://gist.github.com/hackintoshrao/f55430d644634ecf72ef67a7d847fb8b
func Search(ctx context.Context, apiKey string, lat float64, lng float64) (*Weather, error) {

	// Construct a request.
	req, err := http.NewRequest(http.MethodGet, "http://api.openweathermap.org/data/2.5/weather", nil)
	if err != nil {
		return nil, err
	}

	// Apply the apiKey, lat and lng to the request.
	q := req.URL.Query()
	q.Add("appid", apiKey)
	q.Add("lat", fmt.Sprintf("%f", lat))
	q.Add("lon", fmt.Sprintf("%f", lng))
	req.URL.RawQuery = q.Encode()

	// Execute the request.
	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the entire JSON response into memory.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into a Weather value.
	var weather Weather
	if err := json.Unmarshal(data, &weather); err != nil {
		return nil, err
	}

	return &weather, nil
}
