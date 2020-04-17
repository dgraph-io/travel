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

			db := ready(t, ctx, dbHost, apiHost)

			if err := db.Validate.Schema(ctx); err != nil {
				t.Fatalf("\t%s\tShould be able to perform the schema operation : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to perform the schema operation.", tests.Success)

			schema, err := db.Query.Schema(ctx)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to query for the schema : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to query for the schema.", tests.Success)

			_, goSchema := data.GrapQLSchema()
			if len(schema) != len(goSchema) {
				t.Log("\t\tGot:", len(schema))
				t.Log("\t\tExp:", len(goSchema))
				t.Errorf("\t%s\tShould be able to see %d predicates in the schema : %v", tests.Failed, len(goSchema), err)
			} else {
				t.Logf("\t%s\tShould be able to see %d predicates in the schema.", tests.Success, len(goSchema))
			}

			if diff := cmp.Diff(schema, goSchema); diff != "" {
				t.Fatalf("\t%s\tShould get back the expected schema. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the expected schema.", tests.Success)
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

			addCity(t, ctx, dbHost, apiHost)
		}
	}
}

// ready provides support for making sure the database is ready to be used.
func ready(t *testing.T, ctx context.Context, dbHost string, apiHost string) *data.DB {
	t.Helper()

	err := data.Readiness(ctx, apiHost, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("\t%s\tShould be able to see Dgraph is ready : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould be able to to see Dgraph is ready.", tests.Success)

	db, err := data.NewDB(dbHost, apiHost)
	if err != nil {
		t.Fatalf("\t%s\tShould be able to connect to Dgraph : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould be able to connect to Dgraph.", tests.Success)

	return db
}

// addCity is a support test help function to consolidate the adding of
// a city since so many data tests need this functionality.
func addCity(t *testing.T, ctx context.Context, dbHost string, apiHost string) (*data.DB, string) {
	t.Helper()

	db := ready(t, ctx, dbHost, apiHost)

	if err := db.Validate.Schema(ctx); err != nil {
		t.Fatalf("\t%s\tShould be able to perform the schema operation : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould be able to perform the schema operation.", tests.Success)

	cityAdd := places.City{
		Name: "sydney",
		Lat:  -33.865143,
		Lng:  151.209900,
	}
	cityID, err := db.Validate.City(ctx, cityAdd)
	if err != nil {
		t.Fatalf("\t%s\tShould be able to add a city : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould be able to add a city.", tests.Success)

	city, err := db.Query.City(ctx, cityID)
	if err != nil {
		t.Fatalf("\t%s\tShould be able to query for the city : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould be able to query for the city.", tests.Success)

	if diff := cmp.Diff(cityAdd, city); diff != "" {
		t.Fatalf("\t%s\tShould get back the same city. Diff:\n%s", tests.Failed, diff)
	}
	t.Logf("\t%s\tShould get back the same city.", tests.Success)

	return db, cityID
}
