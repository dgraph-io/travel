package weather_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgraph-io/travel/business/feeds/weather"
	"github.com/dgraph-io/travel/foundation/tests"
	"github.com/google/go-cmp/cmp"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// TestWeather validates searches can be conducted against api.openweathermap.org.
func TestWeather(t *testing.T) {
	t.Log("Given the need to retreve weather.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single city.", testID)
		{
			server := mockServer()
			t.Cleanup(server.Close)

			ctx := context.Background()
			apiKey := "mocking"
			lat := 33.865143
			lng := 151.209900

			found, err := weather.Search(ctx, apiKey, server.URL, lat, lng)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to search for weather : %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to search for weather.", success, testID)

			weather := weather.Weather{
				CityName:      "",
				Visibility:    "Clear",
				Desc:          "clear sky",
				Temp:          291.69,
				FeelsLike:     289.23,
				MinTemp:       291.69,
				MaxTemp:       291.69,
				Pressure:      1021,
				Humidity:      85,
				WindSpeed:     6.34,
				WindDirection: 168,
				Sunrise:       1588532599,
				Sunset:        1588581628,
			}

			if diff := cmp.Diff(found, weather); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the expected advisory. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the expected advisory.", tests.Success, testID)
		}
	}
}

func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, result)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

var result = `
{
	"coord":{
	   "lon":151.21,
	   "lat":33.87
	},
	"weather":[
	   {
		  "id":800,
		  "main":"Clear",
		  "description":"clear sky",
		  "icon":"01d"
	   }
	],
	"base":"stations",
	"main":{
	   "temp":291.69,
	   "feels_like":289.23,
	   "temp_min":291.69,
	   "temp_max":291.69,
	   "pressure":1021,
	   "humidity":85,
	   "sea_level":1021,
	   "grnd_level":1021
	},
	"wind":{
	   "speed":6.34,
	   "deg":168
	},
	"clouds":{
	   "all":4
	},
	"dt":1588544797,
	"sys":{
	   "sunrise":1588532599,
	   "sunset":1588581628
	},
	"timezone":36000,
	"id":1,
	"name":"",
	"cod":200
 }`
