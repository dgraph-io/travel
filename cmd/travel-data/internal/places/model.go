package places

type Place struct {
	Name string
	Address string
	Lat float64
	Lng float64
	GooglePlaceID string
	LocationType []string
	AvgUserRating float32
	NumberOfRatings int
	GmapsURL string
	PhotoReferenceID string
}

type City struct {
	Name string
	Lat float64
	Lng float64
}

// Location represents a geo-location on a map for Google location search
type PlacesSearchRequest struct {
	Lat       float64
	Lng       float64
	Keyword   string
	Radius    uint
	pageToken string
}
