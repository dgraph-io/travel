package weather

// import (
// 	"context"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// )

// // Weather is used to capture only the required weather information
// // from the weathe result.
// // TODO: complete this.
// type Weather struct {
// 	Temperature    float64
// 	MinTemperature float64
// 	MaxTemperature float64
// 	Pressure       int
// 	Humidity       string
// 	WindSpeed      float64
// 	WindDirection  float64
// }

// // Client provides help fetch weather conditions given latitude and longitude.
// // Once initialized the client can be reused to fetch weather details of many locations.
// type Client struct {
// 	request *http.Request
// }

// // NewClient constructs a Client value that is initialized for use with
// // Google places search and Dgraph.
// func NewClient(ctx context.Context, apiKey string) (*Client, error) {
// 	// Construct the places value for use.
// 	req, err := http.NewRequest(http.MethodGet, "http://api.openweathermap.org/data/2.5/weather", nil)
// 	if err != nil {
// 		return err
// 	}

// 	q := req.URL.Query()
// 	// Initialize the API key
// 	q.Add("appid", apiKey)

// 	req.URL.RawQuery = q.Encode()

// 	return &client, nil
// }

// // GetWeather -  Gives you weather information given a latitude and longitude.
// func (client *Client) GetWeather(ctx context.Context, lat float64, lng float64) error {

// 	// Use the weather client to request weather details from given latitude and longitude.
// 	q := client.request.URL.Query()
// 	q.Add("lat", fmt.Sprintf("%f", lat))
// 	q.Add("lon", fmt.Sprintf("%f", lng))

// 	var client http.Client
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	data, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return err
// 	}

// 	// Here is the output https://gist.github.com/hackintoshrao/f55430d644634ecf72ef67a7d847fb8b
// 	// Need to construct a struct to parse and store the result.
// 	log.Println(string(data))
// 	return nil
// }

// // Result use used to unmarshall the weather info from the API.
// // Here is the sample JSON output to parse https://gist.github.com/hackintoshrao/f55430d644634ecf72ef67a7d847fb8b
// // Struct auto-generated using https://mholt.github.io/json-to-go/
// type Result struct {
// 	Coord struct {
// 		Lon int `json:"lon"`
// 		Lat int `json:"lat"`
// 	} `json:"coord"`
// 	Weather []struct {
// 		ID          int    `json:"id"`
// 		Main        string `json:"main"`
// 		Description string `json:"description"`
// 		Icon        string `json:"icon"`
// 	} `json:"weather"`
// 	Base string `json:"base"`
// 	Main struct {
// 		Temp      float64 `json:"temp"`
// 		FeelsLike float64 `json:"feels_like"`
// 		TempMin   float64 `json:"temp_min"`
// 		TempMax   float64 `json:"temp_max"`
// 		Pressure  int     `json:"pressure"`
// 		Humidity  int     `json:"humidity"`
// 	} `json:"main"`
// 	Wind struct {
// 		Speed float64 `json:"speed"`
// 		Deg   float64 `json:"deg"`
// 	} `json:"wind"`
// 	Clouds struct {
// 		All int `json:"all"`
// 	} `json:"clouds"`
// 	Dt  int `json:"dt"`
// 	Sys struct {
// 		Type    int     `json:"type"`
// 		ID      int     `json:"id"`
// 		Message float64 `json:"message"`
// 		Country string  `json:"country"`
// 		Sunrise int     `json:"sunrise"`
// 		Sunset  int     `json:"sunset"`
// 	} `json:"sys"`
// 	Timezone int    `json:"timezone"`
// 	ID       int    `json:"id"`
// 	Name     string `json:"name"`
// 	Cod      int    `json:"cod"`
// }
