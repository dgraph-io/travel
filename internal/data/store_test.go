package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/advisory"
	"github.com/dgraph-io/travel/internal/places"
	"github.com/dgraph-io/travel/internal/platform/tests"
	"github.com/dgraph-io/travel/internal/weather"
	"github.com/google/go-cmp/cmp"
)

// TestStore validates all the support that provides data storage.
func TestStore(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("advisory", storeAdvisory)
	t.Run("weather", storeWeather)
	t.Run("place", storePlace)
}

// storeAdvisory validates an advisory node can be stored in the database.
func storeAdvisory(t *testing.T) {
	t.Helper()

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate storing an advisory.")
	{
		t.Log("\tWhen handling an advisory for sydney.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			data, cityID := addCity(t, ctx, dbHost, apiHost)

			addAdvisory := advisory.Advisory{
				Country:     "Australia",
				CountryCode: "AU",
				Continent:   "Australia",
				Score:       4,
				LastUpdated: "today",
				Message:     "feel like teen spirit",
				Source:      "friendly neighborhood community engineers",
			}

			if err := data.Store.Advisory(ctx, cityID, addAdvisory); err != nil {
				t.Fatalf("\t%s\tShould be able to save an advisory node in Dgraph : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to save an advisory node in Dgraph.", tests.Success)

			advisory, err := data.Query.Advisory(ctx, cityID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to query for the advisory : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to query for the advisory.", tests.Success)

			if diff := cmp.Diff(addAdvisory, advisory); diff != "" {
				t.Fatalf("\t%s\tShould get back the same advisory. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same advisory.", tests.Success)
		}
	}
}

// storeWeather validates a weather node can be stored in the database.
func storeWeather(t *testing.T) {
	t.Helper()

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate storing weather.")
	{
		t.Log("\tWhen handling weather for sydney.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			data, cityID := addCity(t, ctx, dbHost, apiHost)

			addWeather := weather.Weather{
				ID:            1001,
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

			if err := data.Store.Weather(ctx, cityID, addWeather); err != nil {
				t.Fatalf("\t%s\tShould be able to save a weather node in Dgraph : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to save a weather node in Dgraph.", tests.Success)

			weather, err := data.Query.Weather(ctx, cityID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to query for the weather : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to query for the weather.", tests.Success)

			if diff := cmp.Diff(addWeather, weather); diff != "" {
				t.Fatalf("\t%s\tShould get back the same weather. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same weather.", tests.Success)
		}
	}
}

// storePlace validates a place node can be stored in the database.
func storePlace(t *testing.T) {
	t.Helper()

	dbHost, apiHost, _ := tests.NewUnit(t)
	// defer teardown()

	t.Log("Given the need to be able to validate storing a place.")
	{
		t.Log("\tWhen handling a place for sydney.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			data, cityID := addCity(t, ctx, dbHost, apiHost)

			/*
					BUG!!!
				 	****** IN *****>  {"place_id":"65432","city_name":"sydney","name":"Karthic Coffee","address":"634 Ventura Blvd","lat":-33.865198,"lng":151.209945,"location_type":["resturant"],"avg_user_rating":4.5,"no_user_rating":876,"gmaps_url":"","photo_id":""}
				 	****** OUT *****> {"place_id":"65432","city_name":"sydney","name":"Karthic Coffee","address":"634 Ventura Blvd","lat":-33.865198,"lng":151.209945,"location_type":["resturant"],"avg_user_rating":4,"no_user_rating":876,"gmaps_url":"","photo_id":""}
			*/

			addPlaces := []places.Place{
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

			for _, place := range addPlaces {
				if err := data.Store.Place(ctx, cityID, place); err != nil {
					t.Fatalf("\t%s\tShould be able to save place %q node in Dgraph : %v", tests.Failed, place.Name, err)
				}
				t.Logf("\t%s\tShould be able to save place %q node in Dgraph.", tests.Success, place.Name)
			}

			places, err := data.Query.Places(ctx, cityID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to query for the places : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to query for the places.", tests.Success)

			for i, place := range addPlaces {
				if diff := cmp.Diff(places[i], place); diff != "" {
					t.Fatalf("\t%s\tShould get back the same place for %q. Diff:\n%s", tests.Failed, place.Name, diff)
				}
				t.Logf("\t%s\tShould get back the same place for %q.", tests.Success, place.Name)
			}
		}
	}
}
