package schema

// UploadFeedRequest is the data required to make a feed/upload request.
type UploadFeedRequest struct {
	CountryCode string  `json:"countrycode"`
	CityName    string  `json:"cityname"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

// UploadFeedResponse is the response from the feed/upload request.
type UploadFeedResponse struct {
	CountryCode string  `json:"country_code"`
	CityName    string  `json:"city_name"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Message     string  `json:"message"`
}
