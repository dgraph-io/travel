package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
	"github.com/google/go-cmp/cmp"
)

// addCity validates a city node can be added to the database.
func addCity(apiHost string) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing a city.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a city for Sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				cityAdd := data.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				db, cityAdd := seedCity(t, ctx, testID, apiHost, cityAdd)

				city, err := db.Query.City(ctx, cityAdd.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the city: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the city.", tests.Success, testID)

				if diff := cmp.Diff(cityAdd, city); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same city. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same city.", tests.Success, testID)

				cityByName, err := db.Query.CityByName(ctx, cityAdd.Name)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the city by name: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the city by name.", tests.Success, testID)

				if diff := cmp.Diff(cityAdd, cityByName); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same city by name. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same city by name.", tests.Success, testID)

				cityAdd.ID = ""
				_, err = db.Mutate.AddCity(ctx, cityAdd)
				if err == nil {
					t.Fatalf("\t%s\tTest %d:\tShould not be able to add the same city twice.", tests.Failed, testID)
				}
				t.Logf("\t%s\tTest %d:\tShould not be able to add the same city twice: %v", tests.Success, testID, err)
			}
		}
	}
	return tf
}

// addPlace validates a place can be added to the database.
func addPlace(apiHost string) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing a place.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a place for sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				cityAdd := data.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				db, cityAdd := seedCity(t, ctx, testID, apiHost, cityAdd)

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

				for i, place := range places {
					addPlace, err := db.Mutate.AddPlace(ctx, cityAdd.ID, place)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to save a place in Dgraph: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to save a place in Dgraph.", tests.Success, testID)

					qryPlace, err := db.Query.Place(ctx, addPlace.ID)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to query for the place: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to query for the place.", tests.Success, testID)

					if diff := cmp.Diff(addPlace, qryPlace); diff != "" {
						t.Fatalf("\t%s\tTest %d:\tShould get back the same place. Diff:\n%s", tests.Failed, testID, diff)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the same place.", tests.Success, testID)

					places, err := db.Query.Places(ctx, cityAdd.ID)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to query city places: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to query city places.", tests.Success, testID)

					if len(places) != i+1 {
						t.Errorf("\t\t\tGot: %v", len(places))
						t.Errorf("\t\t\tExp: %v", i+1)
						t.Fatalf("\t%s\tTest %d:\tShould be able to get back %d places: %v", tests.Failed, testID, i+1, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to get back %d places.", tests.Success, testID, i+1)

					if diff := cmp.Diff(places[i], addPlace); diff != "" {
						t.Fatalf("\t%s\tTest %d:\tShould get back the same place. Diff:\n%s", tests.Failed, testID, diff)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the same place.", tests.Success, testID)

					addPlace.ID = ""
					_, err = db.Mutate.AddPlace(ctx, cityAdd.ID, addPlace)
					if err == nil {
						t.Fatalf("\t%s\tTest %d:\tShould not be able to add the same place twice.", tests.Failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould not be able to add the same place twice: %v", tests.Success, testID, err)
				}
			}
		}
	}
	return tf
}

// replaceAdvisory validates an advisory can be stored in the database.
func replaceAdvisory(apiHost string) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate replacing an advisory.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling an advisory for sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				cityAdd := data.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				db, cityAdd := seedCity(t, ctx, testID, apiHost, cityAdd)

				addAdvisory := data.Advisory{
					Country:     "Australia",
					CountryCode: "AU",
					Continent:   "Australia",
					Score:       4,
					LastUpdated: "today",
					Message:     "feel like teen spirit",
					Source:      "friendly neighborhood community engineers",
				}

				addAdvisory, err := db.Mutate.ReplaceAdvisory(ctx, cityAdd.ID, addAdvisory)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace an advisory in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace an advisory in Dgraph.", tests.Success, testID)

				advisory, err := db.Query.Advisory(ctx, cityAdd.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the advisory: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the advisory.", tests.Success, testID)

				if diff := cmp.Diff(addAdvisory, advisory); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same advisory. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same advisory.", tests.Success, testID)

				addAdvisory.ID = ""
				addAdvisory.Score = 6
				addAdvisory, err = db.Mutate.ReplaceAdvisory(ctx, cityAdd.ID, addAdvisory)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace an advisory twice in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace an advisory twice in Dgraph.", tests.Success, testID)

				advisory, err = db.Query.Advisory(ctx, cityAdd.ID)
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
	return tf
}

// replaceWeather validates weather can be stored in the database.
func replaceWeather(apiHost string) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing weather.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling weather for sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				cityAdd := data.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				db, cityAdd := seedCity(t, ctx, testID, apiHost, cityAdd)

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

				addWeather, err := db.Mutate.ReplaceWeather(ctx, cityAdd.ID, addWeather)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace the weather in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace the weather in Dgraph.", tests.Success, testID)

				weather, err := db.Query.Weather(ctx, cityAdd.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the weather: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the weather.", tests.Success, testID)

				if diff := cmp.Diff(addWeather, weather); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same weather. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same weather.", tests.Success, testID)

				addWeather.ID = ""
				addWeather.Desc = "test replace"
				addWeather, err = db.Mutate.ReplaceWeather(ctx, cityAdd.ID, addWeather)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace the weather twice in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace the weather twice in Dgraph.", tests.Success, testID)

				weather, err = db.Query.Weather(ctx, cityAdd.ID)
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
	return tf
}
