package advisory

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// result represents the result of the weather query.
type result struct {
	APIStatus struct {
		Request struct {
			Item string `json:"item"`
		} `json:"request"`
		Reply struct {
			Cache  string `json:"cache"`
			Code   int    `json:"code"`
			Status string `json:"status"`
			Note   string `json:"note"`
			Count  int    `json:"count"`
		} `json:"reply"`
	} `json:"api_status"`
	Data struct {
		// Hardcoded for Australia.
		// TODO: Need to make it generic if this has to work for other countries.
		AU struct {
			IsoAlpha2 string `json:"iso_alpha2"`
			Name      string `json:"name"`
			Continent string `json:"continent"`
			Advisory  struct {
				Score         float64 `json:"score"`
				SourcesActive int     `json:"sources_active"`
				Message       string  `json:"message"`
				Updated       string  `json:"updated"`
				Source        string  `json:"source"`
			} `json:"advisory"`
		} `json:"au"`
	} `json:"data"`
}

// Advisory contains the travel advisory result captured from the feed.
// Here is the sample response from the advisory feed
// https://gist.github.com/hackintoshrao/e07b9f742edf4606f61dce877aa72392.
type Advisory struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Continent   string `json:"continent"`
	// Advisory score out of 5
	// Scores below 4 is considered not so safe to travel
	Score       float64 `json:"advisory_score"`
	LastUpdated string  `json:"advisory_last_updated"`
	Message     string  `json:"advisory_message"`
	Source      string  `json:"temp_min"`
}

// Search can locate weather for a given latitude and longitude.
// Here is the output https://gist.github.com/hackintoshrao/f55430d644634ecf72ef67a7d847fb8b
func Search(ctx context.Context, countryCode string) (Advisory, error) {

	req, err := http.NewRequest(http.MethodGet, "https://www.travel-advisory.info/api", nil)
	if err != nil {
		return Advisory{}, err
	}

	q := req.URL.Query()
	// Setting the country code.
	q.Add("countrycode", countryCode)

	req.URL.RawQuery = q.Encode()

	// Execute the request.
	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return Advisory{}, err
	}
	defer resp.Body.Close()

	// Read the entire JSON response into memory.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Advisory{}, err
	}

	// Unmarshal the JSON into a Advisory value.
	var res result
	if err := json.Unmarshal(data, &res); err != nil {
		return Advisory{}, err
	}

	// Convert the result to a Advisory value so we can
	// use out own names for JSON marshaling.
	advisory := Advisory{
		Country:     res.Data.AU.Name,
		CountryCode: res.Data.AU.IsoAlpha2,
		Continent:   res.Data.AU.Continent,

		// Advisory score out of 5
		// Scores below 4 is considered not so safe to travel
		Score:       res.Data.AU.Advisory.Score,
		LastUpdated: res.Data.AU.Advisory.Updated,
		Message:     res.Data.AU.Advisory.Message,
		Source:      res.Data.AU.Advisory.Source,
	}

	return advisory, nil
}
