package city

// Info represents a city and its coordinates.
type Info struct {
	ID   string  `json:"id,omitempty"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
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
		entities: city {
			id
		}
	}`
}
