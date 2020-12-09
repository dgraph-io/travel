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

// =============================================================================

type id struct {
	Resp struct {
		Entities []struct {
			ID string `json:"id"`
		} `json:"entities"`
	} `json:"resp"`
}

func (id) document() string {
	return `{
		entities: advisory {
			id
		}
	}`
}

type result struct {
	Resp struct {
		Msg     string
		NumUids int
	} `json:"resp"`
}

func (result) document() string {
	return `{
		msg,
		numUids,
	}`
}
