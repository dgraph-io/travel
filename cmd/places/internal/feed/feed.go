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
	mc, err := maps.NewClient(maps.WithAPIKey("AIzaSyAA6GLbxGfMf_8E7VeiwCqB_ukJtCXN5p4"))
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
		PageToken: "pg1",
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
