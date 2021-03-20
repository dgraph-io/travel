package data_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"testing"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/auth"
	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/advisory"
	"github.com/dgraph-io/travel/business/data/city"
	"github.com/dgraph-io/travel/business/data/place"
	"github.com/dgraph-io/travel/business/data/schema"
	"github.com/dgraph-io/travel/business/data/tests"
	"github.com/dgraph-io/travel/business/data/user"
	"github.com/dgraph-io/travel/business/data/weather"
	"github.com/dgraph-io/travel/foundation/keystore"
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
	t.Run("city", upsertCity(tc))
	t.Run("place", addPlace(tc))
	t.Run("advisory", replaceAdvisory(tc))
	t.Run("weather", replaceWeather(tc))
	t.Run("auth", performAuth())
}

// waitReady provides support for making sure the database is ready to be used.
func waitReady(t *testing.T, ctx context.Context, testID int, tc TestConfig) *graphql.GraphQL {
	err := data.Validate(ctx, tc.url, time.Second)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to to see Dgraph is ready.", tests.Success, testID)

	gqlConfig := data.GraphQLConfig{
		URL:            tc.url,
		AuthHeaderName: "X-Travel-Auth",
		AuthToken:      "eyJhbGciOiJSUzI1NiIsImtpZCI6IjU0YmIyMTY1LTcxZTEtNDFhNi1hZjNlLTdkYTRhMGUxZTJjMSIsInR5cCI6IkpXVCJ9.eyJBdXRoIjp7IlJPTEUiOiJBRE1JTiJ9LCJleHAiOjE2MjMzNDI3MTQsImlhdCI6MTU5MTgwNjcxNCwiaXNzIjoidHJhdmVsIHByb2plY3QiLCJzdWIiOiIweDUifQ.dxZsiE9WSXBHB-WenJlSK6zqgXs7ykKpQM3BfrTd_WYvfjIo26FhlPxN-Fr_3dR5-U4aMAw61dTNxMMBNPbD4qs8-CnJ0xfSOl8Xa5Y3p-aKpYvTPL_rPZdjcfqTua2t_sOPmZ3d8_VWkKWmdK-42ab751tmXOCrM6kYXoS1_APQwXKfE_q5eBUlTfrIBR29vtrBfWnpN54wR4i-Uk6DalMOduUmUNuZnYGP9ocIU4Ao1RQ8TsZjo6iIsLGM3r86KYypBWsiRAZPMIZjoZAxqhjRBEOaqNUpq6X3vdhQcRYLgh_36_R1QPlhofAaNKrTMvcZNHkBrBsjOB5pwf6IMQ",
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
func seedCity(t *testing.T, ctx context.Context, testID int, tc TestConfig, newCity city.City) (*graphql.GraphQL, city.City) {
	gql := waitReady(t, ctx, testID, tc)
	store := city.NewStore(tc.log, gql)

	cty, err := store.Upsert(ctx, tc.traceID, newCity)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to add a city: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to add a city.", tests.Success, testID)

	return gql, cty
}

// seedUser is a support test help function to consolidate the seeding of a
// user since so many data tests need this functionality.
func seedUser(t *testing.T, ctx context.Context, testID int, tc TestConfig, newUser user.NewUser, now time.Time) (*graphql.GraphQL, user.User) {
	gql := waitReady(t, ctx, testID, tc)
	store := user.NewStore(tc.log, gql)

	usr, err := store.Add(ctx, tc.traceID, newUser, now)
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

						err := data.Validate(ctx, url, test.retryDelay)
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
				store := user.NewStore(tc.log, gql)

				retUser, err := store.QueryByID(ctx, traceID, addedUser.ID)
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

				userByEmail, err := store.QueryByEmail(ctx, traceID, addedUser.Email)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the user by email: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the user by email.", tests.Success, testID)

				if diff := cmp.Diff(addedUser, userByEmail); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same user by email. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same user by email.", tests.Success, testID)

				_, err = store.Add(ctx, traceID, newUser, now)
				if err == nil {
					t.Fatalf("\t%s\tTest %d:\tShould not be able to add the same user twice.", tests.Failed, testID)
				}
				t.Logf("\t%s\tTest %d:\tShould not be able to add the same user twice: %v", tests.Success, testID, err)

				err = store.Delete(ctx, traceID, addedUser.ID)
				if err != nil {
					t.Logf("\t%s\tTest %d:\tShould be able to delete the user: %v.", tests.Failed, testID, err)
				} else {
					t.Logf("\t%s\tTest %d:\tShould be able to delete the user.", tests.Success, testID)
				}

				_, err = store.QueryByID(ctx, traceID, addedUser.ID)
				if err == nil {
					t.Fatalf("\t%s\tTest %d:\tShould not be able to query for the user.", tests.Failed, testID)
				}
				t.Logf("\t%s\tTest %d:\tShould not be able to query for the user.", tests.Success, testID)
			}
		}
	}
	return tf
}

// upsertCity validates a city node can be added to the database.
func upsertCity(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing a city.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a city for Sydney.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				newCity := city.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				store := city.NewStore(tc.log, gql)

				retCity, err := store.QueryByID(ctx, tc.traceID, addedCity.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the city: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the city.", tests.Success, testID)

				if diff := cmp.Diff(addedCity, retCity); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same city. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same city.", tests.Success, testID)

				cityByName, err := store.QueryByName(ctx, tc.traceID, addedCity.Name)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for the city by name: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for the city by name.", tests.Success, testID)

				if diff := cmp.Diff(addedCity, cityByName); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same city by name. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same city by name.", tests.Success, testID)

				id := addedCity.ID
				addedCity.ID = ""
				upsertCity, err := store.Upsert(ctx, tc.traceID, addedCity)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to upsert the same city twice: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould not be able to upsert the same city twice.", tests.Success, testID)

				if id != upsertCity.ID {
					t.Logf("\t\tTest %d:\tgot: %v", testID, upsertCity.ID)
					t.Logf("\t\tTest %d:\texp: %v", testID, id)
					t.Fatalf("\t%s\tTest %d:\tShould get back the same id for the city.", tests.Failed, testID)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same id for the city.", tests.Success, testID)

				cities, err := store.QueryNames(ctx, tc.traceID)
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

				newCity := city.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				store := place.NewStore(tc.log, gql)

				places := []place.Place{
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
					addedPlace, err := store.Upsert(ctx, tc.traceID, newPlace)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to save a place in Dgraph: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to save a place in Dgraph.", tests.Success, testID)

					retPlace, err := store.QueryByID(ctx, tc.traceID, addedPlace.ID)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to query for the place: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to query for the place.", tests.Success, testID)

					if diff := cmp.Diff(addedPlace, retPlace); diff != "" {
						t.Fatalf("\t%s\tTest %d:\tShould get back the same place. Diff:\n%s", tests.Failed, testID, diff)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the same place.", tests.Success, testID)

					places, err := store.QueryByCity(ctx, tc.traceID, addedCity.ID)
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

					id := addedPlace.ID
					addedPlace.ID = ""
					upsertPlace, err := store.Upsert(ctx, tc.traceID, addedPlace)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to upsert the same place twice: %v", tests.Failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould not be able to upsert the same place twice.", tests.Success, testID)

					if id != upsertPlace.ID {
						t.Logf("\t\tTest %d:\tgot: %v", testID, upsertPlace.ID)
						t.Logf("\t\tTest %d:\texp: %v", testID, id)
						t.Fatalf("\t%s\tTest %d:\tShould get back the same id for the place.", tests.Failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the same id for the place.", tests.Success, testID)
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

				newCity := city.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				store := advisory.NewStore(tc.log, gql)

				newAdvisory := advisory.Advisory{
					City:        advisory.City{ID: addedCity.ID},
					Country:     "Australia",
					CountryCode: "AU",
					Continent:   "Australia",
					Score:       4,
					LastUpdated: "today",
					Message:     "feel like teen spirit",
					Source:      "friendly neighborhood community engineers",
				}

				addedAdvisory, err := store.Replace(ctx, tc.traceID, newAdvisory)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace an advisory in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace an advisory in Dgraph.", tests.Success, testID)

				retAdvisory, err := store.QueryByCity(ctx, tc.traceID, addedCity.ID)
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
				addedAdvisory, err = store.Replace(ctx, tc.traceID, addedAdvisory)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace an advisory twice in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace an advisory twice in Dgraph.", tests.Success, testID)

				retAdvisory, err = store.QueryByCity(ctx, tc.traceID, addedCity.ID)
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

				newCity := city.City{
					Name: "sydney",
					Lat:  -33.865143,
					Lng:  151.209900,
				}
				gql, addedCity := seedCity(t, ctx, testID, tc, newCity)
				store := weather.NewStore(tc.log, gql)

				newWeather := weather.Weather{
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

				addedWeather, err := store.Replace(ctx, tc.traceID, newWeather)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace the weather in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace the weather in Dgraph.", tests.Success, testID)

				retWeather, err := store.QueryByCity(ctx, tc.traceID, addedCity.ID)
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
				addedWeather, err = store.Replace(ctx, tc.traceID, addedWeather)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to replace the weather twice in Dgraph: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to replace the weather twice in Dgraph.", tests.Success, testID)

				retWeather, err = store.QueryByCity(ctx, tc.traceID, addedCity.ID)
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
				keyID := "4754d86b-7a6d-4df5-9c65-224741361492"
				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create a private key: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create a private key.", tests.Success, testID)

				a, err := auth.New("RS256", keystore.NewMap(map[string]*rsa.PrivateKey{keyID: privateKey}))
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
