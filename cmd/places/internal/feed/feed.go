package feed

import (
	"context"
	"encoding/json"
	"log"

	"googlemaps.github.io/maps"
)

// Pull extracts the feed from the source.
func Pull(log *log.Logger) error {

	// Construct a new client for API access.
	mc, err := maps.NewClient(maps.WithAPIKey("AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnU"))
	if err != nil {
		return err
	}

	latLng := maps.LatLng{
		Lat: -33.865143,
		Lng: 151.209900,
	}
	nsr := maps.NearbySearchRequest{
		Location:  &latLng,
		Keyword:   "Sydney",
		PageToken: "",
		Radius: 5000,
	}
	resp, err := mc.NearbySearch(context.TODO(), &nsr)
	if err != nil {
		return err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	log.Println(string(data))

	return nil
}
