package advisory

// Info contains the travel advisory result captured for a city.
type Info struct {
	ID          string  `json:"id,omitempty"`
	City        City    `json:"city"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Continent   string  `json:"continent"`
	Score       float64 `json:"score"`
	LastUpdated string  `json:"last_updated"`
	Message     string  `json:"message"`
	Source      string  `json:"source"`
}

// City is used to capture the city id in relationships.
type City struct {
	ID string `json:"id"`
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
