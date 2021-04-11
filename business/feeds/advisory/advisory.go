// Package advisory is providing support to query the Travel Advisory API
// and retrieve an advisory for a specified city.
// www.travel-advisory.info
package advisory

import (
	"context"
	"encoding/json"
	"io"
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
func Search(ctx context.Context, url string, countryCode string) (Advisory, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Advisory{}, errors.Wrap(err, "new request")
	}

	q := req.URL.Query()
	q.Add("countrycode", countryCode)
	req.URL.RawQuery = q.Encode()

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return Advisory{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Advisory{}, err
	}

	var res result
	if err := json.Unmarshal(data, &res); err != nil {
		return Advisory{}, err
	}

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
