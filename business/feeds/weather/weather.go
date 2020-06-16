// Package weather is providing support to query the Open Weather API
// and retrieve weather for a specified city.
// https://openweathermap.org/api
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// Weather contains the weather data points captured from the API.
type Weather struct {
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

// Search can locate weather for a given latitude and longitude.
func Search(ctx context.Context, apiKey string, url string, lat float64, lng float64) (Weather, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Weather{}, errors.Wrap(err, "new request")
	}

	q := req.URL.Query()
	q.Add("appid", apiKey)
	q.Add("lat", fmt.Sprintf("%f", lat))
	q.Add("lon", fmt.Sprintf("%f", lng))
	req.URL.RawQuery = q.Encode()

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return Weather{}, errors.Wrap(err, "client do")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Weather{}, errors.Wrap(err, "readall")
	}

	var res result
	if err := json.Unmarshal(data, &res); err != nil {
		return Weather{}, errors.Wrapf(err, "unmarshal[%s]", string(data))
	}

	if res.ID == 0 {
		return Weather{}, errors.New("invalid API key")
	}

	var visibility string
	var description string
	if len(res.Sky) > 0 {
		visibility = res.Sky[0].Visibility
		description = res.Sky[0].Description
	}

	weather := Weather{
		CityName:      res.Name,
		Visibility:    visibility,
		Desc:          description,
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
