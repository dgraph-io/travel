package places_test

import (
	"context"
	"testing"

	"github.com/dgraph-io/travel/internal/feeds/places"
	"googlemaps.github.io/maps"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// TestSearch validates searches can be conducted against the Google maps API.
func TestSearch(t *testing.T) {
	t.Log("Given the need to retreve places.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single city.", testID)
		{
			apiKey := "AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnUek"
			client, err := maps.NewClient(maps.WithAPIKey(apiKey))
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a map client : %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a map client.", success, testID)

			filter := places.Filter{
				Name:    "Sydney",
				Lat:     -33.865143,
				Lng:     151.209900,
				Keyword: "hotels",
				Radius:  5000,
			}

			var savePlace string
			for i := 0; i < 2; i++ {
				places, err := places.Search(context.Background(), client, filter)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to search for places : %v", failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to search for places.", success, testID)

				exp := 20
				if len(places) != 20 {
					t.Errorf("\t%s\tTest %d:\tShould get a full page of places : %v", failed, testID, err)
					t.Logf("\t\tTest %d:\tGot: %v", testID, len(places))
					t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				} else {
					t.Logf("\t%s\tTest %d:\tShould get a full page of places.", success, testID)
				}

				if savePlace == places[0].Name {
					t.Errorf("\t%s\tTest %d:\tShould get different places per page : %v", failed, testID, err)
				} else {
					t.Logf("\t%s\tTest %d:\tShould get different places per page.", success, testID)
				}
				savePlace = places[0].Name
			}
		}
	}
}
