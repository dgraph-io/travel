package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
	"github.com/google/go-cmp/cmp"
)

// TestStore validates all the support that provides data storage.
func TestStore(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("city", storeCity)
	t.Run("advisory", storeAdvisory)
	t.Run("weather", storeWeather)
	t.Run("places", storePlaces)
}

// storeCity validates a city node can be stored in the database.
func storeCity(t *testing.T) {
	t.Helper()

	apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate storing a city.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a city for Sydney.", testID)
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			addCity(t, ctx, testID, apiHost)
		}
	}
}

// storeAdvisory validates an advisory node can be stored in the database.
func storeAdvisory(t *testing.T) {
	t.Helper()

	apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate storing an advisory.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling an advisory for sydney.", testID)
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			db, cityID := addCity(t, ctx, 0, apiHost)

			addAdvisory := data.Advisory{
				Country:     "Australia",
				CountryCode: "AU",
				Continent:   "Australia",
				Score:       4,
				LastUpdated: "today",
				Message:     "feel like teen spirit",
				Source:      "friendly neighborhood community engineers",
			}

			if err := db.Store.Advisory(ctx, cityID, addAdvisory); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to save an advisory node in Dgraph: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to save an advisory node in Dgraph.", tests.Success, testID)

			advisory, err := db.Query.Advisory(ctx, cityID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query for the advisory: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query for the advisory.", tests.Success, testID)

			if diff := cmp.Diff(addAdvisory, advisory); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same advisory. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same advisory.", tests.Success, testID)
		}
	}
}

// storeWeather validates a weather node can be stored in the database.
func storeWeather(t *testing.T) {
	t.Helper()

	apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate storing weather.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling weather for sydney.", testID)
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			db, cityID := addCity(t, ctx, 0, apiHost)

			addWeather := data.Weather{
				CityName:      "Sydney",
				Visibility:    "clear",
				Desc:          "going to be a great day",
				Temp:          98.6,
				FeelsLike:     100.2,
				MinTemp:       92.2,
				MaxTemp:       99.3,
				Pressure:      701,
				Humidity:      80,
				WindSpeed:     14.2,
				WindDirection: 345,
				Sunrise:       1009923,
				Sunset:        10009945,
			}

			if err := db.Store.Weather(ctx, cityID, addWeather); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to save a weather node in Dgraph: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to save a weather node in Dgraph.", tests.Success, testID)

			weather, err := db.Query.Weather(ctx, cityID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query for the weather: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query for the weather.", tests.Success, testID)

			if diff := cmp.Diff(addWeather, weather); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same weather. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same weather.", tests.Success, testID)
		}
	}
}

// storePlaces validates a place node can be stored in the database.
func storePlaces(t *testing.T) {
	t.Helper()

	apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate storing a place.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a place for sydney.", testID)
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			db, cityID := addCity(t, ctx, 0, apiHost)

			places := []data.Place{
				{
					PlaceID:          "12345",
					CityName:         "sydney",
					Name:             "Bill's SPAM shack",
					Address:          "123 Mocking Bird Lane",
					Lat:              -33.865143,
					Lng:              151.209900,
					LocationType:     []string{"resturant"},
					AvgUserRating:    5.0,
					NumberOfRatings:  10345,
					GmapsURL:         "",
					PhotoReferenceID: "",
				},
				{
					PlaceID:          "65432",
					CityName:         "sydney",
					Name:             "Karthic Coffee",
					Address:          "634 Ventura Blvd",
					Lat:              -33.865198,
					Lng:              151.209945,
					LocationType:     []string{"resturant"},
					AvgUserRating:    4.0,
					NumberOfRatings:  876,
					GmapsURL:         "",
					PhotoReferenceID: "",
				},
			}

			if err := db.Store.Places(ctx, cityID, places); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to save places in Dgraph: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to save places in Dgraph.", tests.Success, testID)

			places, err := db.Query.Places(ctx, cityID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query for the places: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query for the places.", tests.Success, testID)

			for i, place := range places {
				if diff := cmp.Diff(places[i], place); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same place for %q. Diff:\n%s", tests.Failed, testID, place.Name, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same place for %q.", tests.Success, testID, place.Name)
			}
		}
	}
}
