package advisory

// Advisory contains the travel advisory result captured for a city.
type Advisory struct {
	ID          string  `json:"id,omitempty"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Continent   string  `json:"continent"`
	Score       float64 `json:"score"`
	LastUpdated string  `json:"last_updated"`
	Message     string  `json:"message"`
	Source      string  `json:"source"`
}

type addResult struct {
	AddAdvisory struct {
		Advisory []struct {
			ID string `json:"id"`
		} `json:"advisory"`
	} `json:"addAdvisory"`
}

func (addResult) document() string {
	return `{
		advisory {
			id
		}
	}`
}

type updateCityResult struct {
	UpdateCity struct {
		City []struct {
			ID string `json:"id"`
		} `json:"city"`
	} `json:"updateCity"`
}

func (updateCityResult) document() string {
	return `{
		city {
			id
		}
	}`
}

type deleteResult struct {
	DeleteAdvisory struct {
		Msg     string
		NumUids int
	} `json:"deleteAdvisory"`
}

func (deleteResult) document() string {
	return `{
		msg,
		numUids,
	}`
}
