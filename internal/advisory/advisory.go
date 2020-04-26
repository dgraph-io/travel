package advisory

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// Advisory contains the travel advisory result captured for a city. When
// the advisory score is below 4 out of 5, it's not considered safe to travel.
type Advisory struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Continent   string  `json:"continent"`
	Score       float64 `json:"score"`
	LastUpdated string  `json:"last_updated"`
	Message     string  `json:"message"`
	Source      string  `json:"source"`
}

// Search can locate weather for a given latitude and longitude.
func Search(ctx context.Context, countryCode string) (Advisory, error) {

	// Construct a request to perform the advisory search.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.travel-advisory.info/api", nil)
	if err != nil {
		return Advisory{}, errors.Wrap(err, "new request")
	}

	// Apply the country code to the request.
	q := req.URL.Query()
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
	// use our own tags for JSON marshaling.
	advisory := Advisory{
		Country:     res.Data.AU.Name,
		CountryCode: res.Data.AU.IsoAlpha2,
		Continent:   res.Data.AU.Continent,
		Score:       res.Data.AU.Advisory.Score,
		LastUpdated: res.Data.AU.Advisory.Updated,
		Message:     res.Data.AU.Advisory.Message,
		Source:      res.Data.AU.Advisory.Source,
	}

	return advisory, nil
}

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
