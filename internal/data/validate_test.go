package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/places"
	"github.com/dgraph-io/travel/internal/platform/tests"
	"github.com/google/go-cmp/cmp"
)

// TestValidate validates all the support that provides data validation.
func TestValidate(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("schema", validateSchema)
	t.Run("city", validateCity)
}

// validateSchema validates the schema can be validated in Dgraph.
func validateSchema(t *testing.T) {
	t.Helper()

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate a schema.")
	{
		t.Log("\tWhen handling a city schema.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := data.Readiness(ctx, apiHost, 500*time.Millisecond)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to see Dgraph is ready : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to to see Dgraph is ready.", tests.Success)

			data, err := data.New(dbHost, apiHost)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to connect to Dgraph : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to connect to Dgraph.", tests.Success)

			if err := data.Validate.Schema(ctx); err != nil {
				t.Fatalf("\t%s\tShould be able to perform the schema operation : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to perform the schema operation.", tests.Success)

			schema, err := data.Query.Schema(ctx)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to query for the schema : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to query for the schema.", tests.Success)

			const predicates = 9
			if len(schema) != predicates {
				t.Log("\t\tGot:", len(schema))
				t.Log("\t\tExp:", predicates)
				t.Errorf("\t%s\tShould be able to see %d predicates in the schema : %v", tests.Failed, predicates, err)
			} else {
				t.Logf("\t%s\tShould be able to see %d predicates in the schema.", tests.Success, predicates)
			}
		}
	}
}

// validateCity validates the health check is working.
func validateCity(t *testing.T) {
	t.Helper()

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate a city.")
	{
		t.Log("\tWhen handling a city like Sydney.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := data.Readiness(ctx, apiHost, 500*time.Millisecond)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to see Dgraph is ready : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to to see Dgraph is ready.", tests.Success)

			data, err := data.New(dbHost, apiHost)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to connect to Dgraph : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to connect to Dgraph.", tests.Success)

			cityAdd := places.City{
				Name: "sydney",
				Lat:  -33.865143,
				Lng:  151.209900,
			}
			cityID, err := data.Validate.City(ctx, cityAdd)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to add a city : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to add a city.", tests.Success)

			city, err := data.Query.CityByID(ctx, cityID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to query for the city : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to query for the city.", tests.Success)

			if diff := cmp.Diff(cityAdd, city); diff != "" {
				t.Fatalf("\t%s\tShould get back the same city. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same city.", tests.Success)
		}
	}
}
