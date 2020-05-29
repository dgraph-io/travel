package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
)

// TestData validates all the mutation support in data.
func TestData(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	url, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	t.Run("readiness", readiness(url))
	t.Run("schema", schema(url))
	t.Run("city", addCity(url))
	t.Run("place", addPlace(url))
	t.Run("advisory", replaceAdvisory(url))
	t.Run("weather", replaceWeather(url))
}

// ready provides support for making sure the database is ready to be used.
func ready(t *testing.T, ctx context.Context, testID int, url string) *data.DB {
	err := data.Readiness(ctx, url, time.Second)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to to see Dgraph is ready.", tests.Success, testID)

	dgraph := data.Dgraph{
		URL: url,
	}
	db, err := data.NewDB(dgraph)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to connect to Dgraph: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to connect to Dgraph.", tests.Success, testID)

	return db
}

// seedCity is a support test help function to consolidate the seeding of a
// city since so many data tests need this functionality.
func seedCity(t *testing.T, ctx context.Context, testID int, url string, city data.City) (*data.DB, data.City) {
	db := ready(t, ctx, testID, url)

	if err := db.Schema.DropAll(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data and schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to drop the data and schema.", tests.Success, testID)

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
