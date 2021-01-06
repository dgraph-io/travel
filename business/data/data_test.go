package data_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/advisory"
	"github.com/dgraph-io/travel/business/data/auth"
	"github.com/dgraph-io/travel/business/data/city"
	"github.com/dgraph-io/travel/business/data/place"
	"github.com/dgraph-io/travel/business/data/ready"
	"github.com/dgraph-io/travel/business/data/schema"
	"github.com/dgraph-io/travel/business/data/user"
	"github.com/dgraph-io/travel/business/data/weather"
	"github.com/dgraph-io/travel/foundation/tests"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/bcrypt"
)

type TestConfig struct {
	traceID string
	log     *log.Logger
	url     string
	schema  schema.Config
}

// TestData validates all the mutation support in data.
func TestData(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Start up dgraph in a container.
	log, url, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	// Configure everything to run the tests.
	tc := TestConfig{
		traceID: "00000000-0000-0000-0000-000000000000",
		log:     log,
		url:     url,
		schema: schema.Config{
			CustomFunctions: schema.CustomFunctions{
				UploadFeedURL: "http://0.0.0.0:3000/v1/feed/upload",
			},
		},
	}

	t.Run("readiness", readiness(tc.url))
	t.Run("schema", addSchema(tc))
	t.Run("user", addUser(tc))
	t.Run("city", addCity(tc))
	t.Run("place", addPlace(tc))
	t.Run("advisory", replaceAdvisory(tc))
	t.Run("weather", replaceWeather(tc))
	t.Run("auth", performAuth())
}

// waitReady provides support for making sure the database is ready to be used.
func waitReady(t *testing.T, ctx context.Context, testID int, tc TestConfig) *graphql.GraphQL {
	err := ready.Validate(ctx, tc.url, time.Second)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to to see Dgraph is ready.", tests.Success, testID)

	gqlConfig := data.GraphQLConfig{
		URL:            tc.url,
		AuthHeaderName: "X-Travel-Auth",
		AuthToken:      schema.AdminJWT,
	}
	gql := data.NewGraphQL(gqlConfig)

	schema, err := schema.New(gql, tc.schema)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to prepare the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to prepare the schema.", tests.Success, testID)

	// Performing this action here breaks the current version of Dgraph.
	// To see this, uncomment this code and comment lines 96-99.
	// This code used to work on an earlier version of dgraph.
	//
	// if err := schema.DropAll(ctx); err != nil {
	// 	t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data and schema: %v", tests.Failed, testID, err)
	// }
	// t.Logf("\t%s\tTest %d:\tShould be able to drop the data and schema.", tests.Success, testID)

	if err := schema.Create(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to create the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to create the schema.", tests.Success, testID)

	if err := schema.DropData(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data : %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to drop the data.", tests.Success, testID)

	return gql
}

// seedCity is a support test help function to consolidate the seeding of a
// city since so many data tests need this functionality.
func seedCity(t *testing.T, ctx context.Context, testID int, tc TestConfig, newCity city.Info) (*graphql.GraphQL, city.Info) {
	gql := waitReady(t, ctx, testID, tc)
	c := city.New(tc.log, gql)

	cty, err := c.Add(ctx, tc.traceID, newCity)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to add a city: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to add a city.", tests.Success, testID)

	return gql, cty
}

// seedUser is a support test help function to consolidate the seeding of a
// user since so many data tests need this functionality.
func seedUser(t *testing.T, ctx context.Context, testID int, tc TestConfig, newUser user.NewUser, now time.Time) (*graphql.GraphQL, user.Info) {
	gql := waitReady(t, ctx, testID, tc)
	u := user.New(tc.log, gql)

	usr, err := u.Add(ctx, tc.traceID, newUser, now)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to add a user: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to add a user.", tests.Success, testID)

	return gql, usr
}

// readiness validates the health check is working.
func readiness(url string) func(t *testing.T) {
	tf := func(t *testing.T) {
		type tableTest struct {
			name       string
			retryDelay time.Duration
			timeout    time.Duration
			success    bool
		}

		tt := []tableTest{
			{"timeout", 500 * time.Millisecond, time.Second, false},
			{"ready", 500 * time.Millisecond, 20 * time.Second, true},
		}

		t.Log("Given the need to be able to validate the database is ready.")
		{
			for testID, test := range tt {
				tf := func(t *testing.T) {
					t.Logf("\tTest %d:\tWhen waiting up to %v for the database to be ready.", testID, test.timeout)
					{
						ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
						defer cancel()

						err := ready.Validate(ctx, url, test.retryDelay)
						switch test.success {
						case true:
							if err != nil {
								t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
							}
							t.Logf("\t%s\tTest %d:\tShould be able to see Dgraph is ready.", tests.Success, testID)

						case false:
							if err == nil {
								t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is Not ready.", tests.Failed, testID)
							}
							t.Logf("\t%s\tTest %d:\tShould be able to see Dgraph is Not ready.", tests.Success, testID)
						}
					}
				}
				t.Run(test.name, tf)
			}
		}
	}
	return tf
}

// addSchema validates the schema we are storing is what we expect
// for the application.
func addSchema(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate a schema.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling the application schema.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				waitReady(t, ctx, 0, tc)
			}
		}
	}
	return tf
}

// addUser validates a user node can be added to the database.
func addUser(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing a user.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a single user.", testID)
			{
				traceID := "00000000-0000-0000-0000-000000000000"

				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				now := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)

				newUser := user.NewUser{
					Name:            "Bill Kennedy",
					Email:           "bill@ardanlabs.com",
					Role:            "ADMIN",
					Password:        "gophers",
					PasswordConfirm: "gophers",
				}
				gql, addedUser := seedUser(t, ctx, testID, tc, newUser, now)
				u := user.New(tc.log, gql)

				retUser, err := u.QueryByID(ctx, traceID, addedUser.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the user: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the user.", tests.Success, testID)

				if diff := cmp.Diff(addedUser, retUser); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same user.", tests.Success, testID)

				if err := bcrypt.CompareHashAndPassword([]byte(retUser.PasswordHash), []byte(newUser.Password)); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same password hash: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same password hash.", tests.Success, testID)

				userByEmail, err := u.QueryByEmail(ctx, traceID, addedUser.Email)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the user by email: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the user by email.", tests.Success, testID)

				if diff := cmp.Diff(addedUser, userByEmail); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same user by email. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same user by email.", tests.Success, testID)

				_, err = u.Add(ctx, traceID, newUser, now)
				if err == nil {
					t.Fatalf("\t%s\tTest %d:\tShould not be able to add the same user twice.", tests.Failed, testID)
				}
				t.Logf("\t%s\tTest %d:\tShould not be able to add the same user twice: %v", tests.Success, testID, err)

				err = u.Delete(ctx, traceID, addedUser.ID)
				if err != nil {
					t.Logf("\t%s\tTest %d:\tShould be able to delete the user: %v.", tests.Failed, testID, err)
				} else {
					t.Logf("\t%s\tTest %d:\tShould be able to delete the user.", tests.Success, testID)
				}

				_, err = u.QueryByID(ctx, traceID, addedUser.ID)
				if err == nil {
					t.Fatalf("\t%s\tTest %d:\tShould not be able to query for the user.", tests.Failed, testID)
				}
				t.Logf("\t%s\tTest %d:\tShould not be able to query for the user.", tests.Success, testID)
			}
		}
	}
	return tf
}

// addCity validates a city node can be added to the database.
func addCity(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing a city.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a city for Sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				newCity := city.Info{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				c := city.New(tc.log, gql)

				retCity, err := c.QueryByID(ctx, tc.traceID, addedCity.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the city: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the city.", tests.Success, testID)

				if diff := cmp.Diff(addedCity, retCity); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same city. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same city.", tests.Success, testID)

				cityByName, err := c.QueryByName(ctx, tc.traceID, addedCity.Name)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the city by name: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the city by name.", tests.Success, testID)

				if diff := cmp.Diff(addedCity, cityByName); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same city by name. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same city by name.", tests.Success, testID)

				addedCity.ID = ""
				_, err = c.Add(ctx, tc.traceID, addedCity)
				if err == nil {
					t.Fatalf("\t%s\tTest %d:\tShould not be able to add the same city twice.", tests.Failed, testID)
				}
				t.Logf("\t%s\tTest %d:\tShould not be able to add the same city twice: %v", tests.Success, testID, err)

				cities, err := c.QueryNames(ctx, tc.traceID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the list of city names: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the list of city names.", tests.Success, testID)

				if len(cities) != 1 || cities[0] != "sydney" {
					t.Logf("\t\tTest %d:\tgot: %v", testID, cities)
					t.Logf("\t\tTest %d:\texp: %v", testID, []string{"sydney"})
					t.Fatalf("\t%s\tTest %d:\tShould be able to have the correct list: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to have the correct list.", tests.Success, testID)
			}
		}
	}
	return tf
}

// addPlace validates a place can be added to the database.
func addPlace(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing a place.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a place for sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				newCity := city.Info{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				p := place.New(tc.log, gql)

				places := []place.Info{
					{
						PlaceID:          "12345",
						Category:         "test",
						City:             place.City{ID: addedCity.ID},
						CityName:         "sydney",
						Name:             "Bill's SPAM shack",
						Address:          "123 Mocking Bird Lane",
						Lat:              -33.865143,
						Lng:              151.209900,
						LocationType:     []string{"restaurant"},
						AvgUserRating:    5.0,
						NumberOfRatings:  10345,
						GmapsURL:         "",
						PhotoReferenceID: "",
					},
					{
						PlaceID:          "65432",
						Category:         "test",
						City:             place.City{ID: addedCity.ID},
						CityName:         "sydney",
						Name:             "Karthic Coffee",
						Address:          "634 Ventura Blvd",
						Lat:              -33.865198,
						Lng:              151.209945,
						LocationType:     []string{"restaurant"},
						AvgUserRating:    4.0,
						NumberOfRatings:  876,
						GmapsURL:         "",
						PhotoReferenceID: "",
					},
				}

				for i, newPlace := range places {
					addedPlace, err := p.Add(ctx, tc.traceID, newPlace)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to save a place in Dgraph: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to save a place in Dgraph.", tests.Success, testID)

					retPlace, err := p.QueryByID(ctx, tc.traceID, addedPlace.ID)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to query for the place: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to query for the place.", tests.Success, testID)

					if diff := cmp.Diff(addedPlace, retPlace); diff != "" {
						t.Fatalf("\t%s\tTest %d:\tShould get back the same place. Diff:\n%s", tests.Failed, testID, diff)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the same place.", tests.Success, testID)

					places, err := p.QueryByCity(ctx, tc.traceID, addedCity.ID)
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

					if diff := cmp.Diff(places[i], addedPlace); diff != "" {
						t.Fatalf("\t%s\tTest %d:\tShould get back the same place. Diff:\n%s", tests.Failed, testID, diff)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the same place.", tests.Success, testID)

					addedPlace.ID = ""
					_, err = p.Add(ctx, tc.traceID, addedPlace)
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
func replaceAdvisory(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate replacing an advisory.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling an advisory for sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				newCity := city.Info{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				a := advisory.New(tc.log, gql)

				newAdvisory := advisory.Info{
					City:        advisory.City{ID: addedCity.ID},
					Country:     "Australia",
					CountryCode: "AU",
					Continent:   "Australia",
					Score:       4,
					LastUpdated: "today",
					Message:     "feel like teen spirit",
					Source:      "friendly neighborhood community engineers",
				}

				addedAdvisory, err := a.Replace(ctx, tc.traceID, newAdvisory)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace an advisory in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace an advisory in Dgraph.", tests.Success, testID)

				retAdvisory, err := a.QueryByCity(ctx, tc.traceID, addedCity.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the advisory: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the advisory.", tests.Success, testID)

				if diff := cmp.Diff(addedAdvisory, retAdvisory); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same advisory. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same advisory.", tests.Success, testID)

				addedAdvisory.ID = ""
				addedAdvisory.Score = 6
				addedAdvisory, err = a.Replace(ctx, tc.traceID, addedAdvisory)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace an advisory twice in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace an advisory twice in Dgraph.", tests.Success, testID)

				retAdvisory, err = a.QueryByCity(ctx, tc.traceID, addedCity.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the advisory: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the advisory.", tests.Success, testID)

				if diff := cmp.Diff(addedAdvisory, retAdvisory); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same advisory. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same advisory.", tests.Success, testID)
			}
		}
	}
	return tf
}

// replaceWeather validates weather can be stored in the database.
func replaceWeather(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing weather.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling weather for sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				newCity := city.Info{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				w := weather.New(tc.log, gql)

				newWeather := weather.Info{
					City:          weather.City{ID: addedCity.ID},
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

				addedWeather, err := w.Replace(ctx, tc.traceID, newWeather)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace the weather in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace the weather in Dgraph.", tests.Success, testID)

				retWeather, err := w.QueryByCity(ctx, tc.traceID, addedCity.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the weather: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the weather.", tests.Success, testID)

				if diff := cmp.Diff(addedWeather, retWeather); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same weather. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same weather.", tests.Success, testID)

				addedWeather.ID = ""
				addedWeather.Desc = "test replace"
				addedWeather, err = w.Replace(ctx, tc.traceID, addedWeather)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace the weather twice in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace the weather twice in Dgraph.", tests.Success, testID)

				retWeather, err = w.QueryByCity(ctx, tc.traceID, addedCity.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the weather: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the weather.", tests.Success, testID)

				if diff := cmp.Diff(addedWeather, retWeather); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same weather. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same weather.", tests.Success, testID)
			}
		}
	}
	return tf
}

func performAuth() func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to authenticate and authorize access.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a single user.", testID)
			{
				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create a private key: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create a private key.", tests.Success, testID)

				const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
				lookup := func(kid string) (*rsa.PublicKey, error) {
					switch kid {
					case keyID:
						return &privateKey.PublicKey, nil
					}
					return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
				}

				a, err := auth.New("RS256", lookup, auth.Keys{keyID: privateKey})
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", tests.Success, testID)

				claims := auth.Claims{
					StandardClaims: jwt.StandardClaims{
						Issuer:    "travel project",
						Subject:   "0x01",
						ExpiresAt: jwt.At(time.Now().Add(8760 * time.Hour)),
						IssuedAt:  jwt.Now(),
					},
					Auth: auth.StandardClaims{
						Role: "ADMIN",
					},
				}

				token, err := a.GenerateToken(keyID, claims)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT.", tests.Success, testID)

				parsedClaims, err := a.ValidateToken(token)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to parse the claims: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to parse the claims.", tests.Success, testID)

				if exp, got := claims.Auth.Role, parsedClaims.Auth.Role; exp != got {
					t.Logf("\t\tTest %d:\texp: %v", testID, exp)
					t.Logf("\t\tTest %d:\tgot: %v", testID, got)
					t.Fatalf("\t%s\tTest %d:\tShould have the expexted roles: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould have the expexted roles.", tests.Success, testID)
			}
		}
	}
	return tf
}
