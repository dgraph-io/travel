package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Search can locate weather for a given latitude and longitude.
// Here is the output https://gist.github.com/hackintoshrao/f55430d644634ecf72ef67a7d847fb8b
func Search(ctx context.Context, apiKey string, lat float64, lng float64) (Weather, error) {

	// Construct a request.
	req, err := http.NewRequest(http.MethodGet, "http://api.openweathermap.org/data/2.5/weather", nil)
	if err != nil {
		return Weather{}, err
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
		return Weather{}, err
	}
	defer resp.Body.Close()

	// Read the entire JSON response into memory.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Weather{}, err
	}

	// Unmarshal the JSON into a Weather value.
	var res result
	if err := json.Unmarshal(data, &res); err != nil {
		return Weather{}, err
	}

	// Convert the result to a Weather value so we can
	// use our own tags for JSON marshaling.
	weather := Weather{
		ID:            res.ID,
		CityName:      res.Name,
		Visibility:    res.Sky[0].Visibility,
		Desc:          res.Sky[0].Description,
		Temp:          res.Points.Temp,
		FeelsLike:     res.Points.FeelsLike,
		MinTemp:       res.Points.MinTemp,
		MaxTemp:       res.Points.MaxTemp,
		Pressure:      res.Points.Pressure,
		Humidity:      res.Points.Humidity,
		WindSpeed:     res.Wind.Speed,
		WindDirection: res.Wind.Direction,
		Sunrise:       res.RiseSet.Sunrise,
		Sunset:        res.RiseSet.Sunset,
	}

	return weather, nil
}

// Weather contains the weather data points captured from the API.
type Weather struct {
	ID            int     `json:"weather_id"`
	CityName      string  `json:"city_name"`
	Visibility    string  `json:"visibility"`
	Desc          string  `json:"description"`
	Temp          float64 `json:"temp"`
	FeelsLike     float64 `json:"feels_like"`
	MinTemp       float64 `json:"temp_min"`
	MaxTemp       float64 `json:"temp_max"`
	Pressure      int     `json:"pressure"`
	Humidity      int     `json:"humidity"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection int     `json:"wind_direction"`
	Sunrise       int     `json:"sunrise"`
	Sunset        int     `json:"sunset"`
}

// result represents the result of the weather query.
type result struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lon"`
	} `json:"coord"`
	Sky []struct {
		Visibility  string `json:"main"`
		Description string `json:"description"`
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
