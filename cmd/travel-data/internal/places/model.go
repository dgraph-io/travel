package places

// City represents a city and its coordinates.
type City struct {
	Name string
	Lat  float64
	Lng  float64
}

// Search defines parameters that can be used in a Places
// search call.
type Search struct {
	Lat       float64
	Lng       float64
	Keyword   string
	Radius    uint
	pageToken string
}

// Place represents a location that can be found on a Google map.
type Place struct {
	Name             string
	Address          string
	Lat              float64
	Lng              float64
	GooglePlaceID    string
	LocationType     []string
	AvgUserRating    float32
	NumberOfRatings  int
	GmapsURL         string
	PhotoReferenceID string
}
