package places_test

import (
	"context"
	"testing"

	"github.com/dgraph-io/travel/internal/places"
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
		t.Log("\tWhen handling a single city.")
		{
			apiKey := "AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnUek"
			client, err := maps.NewClient(maps.WithAPIKey(apiKey))
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create a map client : %s.", failed, err)
			}
			t.Logf("\t%s\tShould be able to create a map client.", success)

			city := places.City{
				Name: "Sydney",
				Lat:  -33.865143,
				Lng:  151.209900,
			}

			filter := places.Filter{
				Keyword: "hotels",
				Radius:  5000,
			}

			var savePlace string
			for i := 0; i < 2; i++ {
				places, err := city.Search(context.Background(), client, &filter)
				if err != nil {
					t.Fatalf("\t%s\tShould be able to search for places : %s.", failed, err)
				}
				t.Logf("\t%s\tShould be able to search for places.", success)

				exp := 20
				if len(places) != 20 {
					t.Errorf("\t%s\t\tShould get a full page of places : %s.", failed, err)
					t.Log("\t\tGot:", len(places))
					t.Log("\t\tExp:", exp)
				} else {
					t.Logf("\t%s\t\tShould get a full page of places.", success)
				}

				if savePlace == places[0].Name {
					t.Errorf("\t%s\t\tShould get different places per page : %s.", failed, err)
				} else {
					t.Logf("\t%s\t\tShould get different places per page.", success)
				}
				savePlace = places[0].Name
			}
		}
	}
}
