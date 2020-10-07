package city

// Info represents a city and its coordinates.
type Info struct {
	ID   string  `json:"id,omitempty"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type addResult struct {
	AddCity struct {
		City []struct {
			ID string `json:"id"`
		} `json:"city"`
	} `json:"addCity"`
}

func (addResult) document() string {
	return `{
		city {
			id
		}
	}`
}
