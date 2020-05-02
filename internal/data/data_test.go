package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
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

// seedCity is a support test help function to consolidate the seeding of a
// city since so many data tests need this functionality.
func seedCity(t *testing.T, ctx context.Context, testID int, apiHost string, city data.City) (*data.DB, data.City) {
	t.Helper()

	db := ready(t, ctx, testID, apiHost)

	if err := db.Schema.Create(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to create the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to create the schema.", tests.Success, testID)

	cityAdd, err := db.Mutate.AddCity(ctx, city)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to add a city: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to add a city.", tests.Success, testID)

	return db, cityAdd
}
