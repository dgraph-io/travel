// Package places is providing support to query the Google maps places API
// and retrieve places for a specified city.
package places

import (
	"context"
	"io"
	"time"

	"googlemaps.github.io/maps"
)

// Filter defines the specific places to filter out.
type Filter struct {
	Keyword   string
	Radius    uint
	pageToken string
}

// Place contains the place data points captured from the API.
type Place struct {
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

// City represents a city and its coordinates. All fields must be
// populated for a Search to be successful.
type City struct {
	Name string  `json:"city_name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// Search finds places for the specified search criteria.
func (city *City) Search(ctx context.Context, client *maps.Client, filter *Filter) ([]Place, error) {

	// If this call is not looking for page 1, we need to pace
	// the searches out. We are using three seconds.
	if filter.pageToken != "" {
		time.Sleep(3000 * time.Millisecond)
	}

	// We will make three attempts to perform a search. You need to
	// space your paged searches by an undefined amount of time :(.
	// The call may result in an INVALID_REQUEST error if the call
	// is happening at a pace too fast for the API.
	var resp maps.PlacesSearchResponse
	for i := 0; i < 3; i++ {

		// Construct the search request value for the call.
		nsr := maps.NearbySearchRequest{
			Location: &maps.LatLng{
				Lat: city.Lat,
				Lng: city.Lng,
			},
			Keyword:   filter.Keyword,
			PageToken: filter.pageToken,
			Radius:    filter.Radius,
		}

		// Perform the Google Places search.
		var err error
		resp, err = client.NearbySearch(ctx, &nsr)

		// This is the problem. We need to check for the INVALID_REQUEST
		// error. The only way to do that is to compare this string :(
		// If this is the error, then wait for a second before trying again.
		if err != nil {
			if err.Error() == "maps: INVALID_REQUEST - " {
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			return nil, err
		}
		break
	}

	var places []Place
	for _, result := range resp.Results {

		// Validate if a photo even exists for this place.
		var photoReferenceID string
		if len(result.Photos) == 0 {
			photoReferenceID = result.Photos[0].PhotoReference
		}

		// Construct a place value based on search results.
		place := Place{
			PlaceID:          result.PlaceID,
			CityName:         city.Name,
			Name:             result.Name,
			Address:          result.FormattedAddress,
			Lat:              result.Geometry.Location.Lat,
			Lng:              result.Geometry.Location.Lng,
			LocationType:     result.Types,
			AvgUserRating:    result.Rating,
			NumberOfRatings:  result.UserRatingsTotal,
			PhotoReferenceID: photoReferenceID,
		}

		// Save the place in the collection of places.
		places = append(places, place)
	}

	// If the NextPageToken on the result is empty, we have all
	// the results. Send an EOF to confirm that back to the caller.
	filter.pageToken = resp.NextPageToken
	if resp.NextPageToken == "" {
		return places, io.EOF
	}

	return places, nil
}
