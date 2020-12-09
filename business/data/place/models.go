package place

// Info contains the place data points captured from the API.
type Info struct {
	ID               string   `json:"id,omitempty"`
	Category         string   `json:"category"`
	City             City     `json:"city"`
	PlaceID          string   `json:"place_id"`
	CityName         string   `json:"city_name"`
	Name             string   `json:"name"`
	Address          string   `json:"address"`
	Lat              float64  `json:"lat"`
	Lng              float64  `json:"lng"`
	LocationType     []string `json:"location_type"`
	AvgUserRating    float32  `json:"avg_user_rating"`
	NumberOfRatings  int      `json:"no_user_rating"`
	GmapsURL         string   `json:"gmaps_url"`
	PhotoReferenceID string   `json:"photo_id"`
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
		entities: place {
			id
		}
	}`
}
