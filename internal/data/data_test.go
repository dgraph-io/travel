package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/feeds/places"
	"github.com/dgraph-io/travel/internal/platform/tests"
	"github.com/google/go-cmp/cmp"
)

// ready provides support for making sure the database is ready to be used.
func ready(t *testing.T, ctx context.Context, testID int, apiHost string) *data.DB {
	t.Helper()

	err := data.Readiness(ctx, apiHost, time.Second)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to to see Dgraph is ready.", tests.Success, testID)

	db, err := data.NewDB(apiHost)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to connect to Dgraph: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to connect to Dgraph.", tests.Success, testID)

	return db
}

// addCity is a support test help function to consolidate the adding of
// a city since so many data tests need this functionality.
func addCity(t *testing.T, ctx context.Context, testID int, apiHost string) (*data.DB, string) {
	t.Helper()

	db := ready(t, ctx, testID, apiHost)

	if err := db.Schema.Create(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to create the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to create the schema.", tests.Success, testID)

	cityAdd := places.City{
		Name: "sydney",
		Lat:  -33.865143,
		Lng:  151.209900,
	}
	cityID, err := db.Store.City(ctx, cityAdd)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to add a city: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to add a city.", tests.Success, testID)

	city, err := db.Query.City(ctx, cityID)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to query for the city: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to query for the city.", tests.Success, testID)

	if diff := cmp.Diff(cityAdd, city); diff != "" {
		t.Fatalf("\t%s\tTest %d:\tShould get back the same city. Diff:\n%s", tests.Failed, testID, diff)
	}
	t.Logf("\t%s\tTest %d:\tShould get back the same city.", tests.Success, testID)

	return db, cityID
}
