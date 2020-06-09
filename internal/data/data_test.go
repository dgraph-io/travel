package data_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
)

type TestConfig struct {
	url    string
	schema data.SchemaConfig
}

// TestData validates all the mutation support in data.
func TestData(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Start up dgraph in a container.
	url, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	// Locate the physical location of the schema file.
	_, filename, _, _ := runtime.Caller(0)
	schemaFile := fmt.Sprintf("%s/schema.gql", path.Dir(filename))
	f, err := os.Open(schemaFile)
	if err != nil {
		t.Fatalf("opening schema file: %s  error: %v", schemaFile, err)
	}

	// Read the entire contents of the schema file.
	document, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("reading schema file: %s  error: %v", schemaFile, err)
	}

	// Configure everything to run the tests.
	tc := TestConfig{
		url: url,
		schema: data.SchemaConfig{
			SendEmailURL: "http://0.0.0.0:3000/v1/email",
			Document:     string(document),
			PublicKey:    publicRSAKey,
		},
	}

	t.Run("readiness", readiness(tc.url))
	t.Run("schema", schema(tc))
	t.Run("user", addUser(tc))
	t.Run("city", addCity(tc))
	t.Run("place", addPlace(tc))
	t.Run("advisory", replaceAdvisory(tc))
	t.Run("weather", replaceWeather(tc))
	t.Run("auth", auth())
}

// ready provides support for making sure the database is ready to be used.
func ready(t *testing.T, ctx context.Context, testID int, tc TestConfig) (*data.Schema, *data.DB) {
	err := data.Readiness(ctx, tc.url, time.Second)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to to see Dgraph is ready.", tests.Success, testID)

	dbConfig := data.DBConfig{
		URL: tc.url,
	}
	schema, err := data.NewSchema(dbConfig, tc.schema)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to prepare the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to prepare the schema.", tests.Success, testID)

	db, err := data.NewDB(dbConfig)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to prepare to use the DB: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to prepare to use the DB.", tests.Success, testID)

	return schema, db
}

// seedCity is a support test help function to consolidate the seeding of a
// city since so many data tests need this functionality.
func seedCity(t *testing.T, ctx context.Context, testID int, tc TestConfig, city data.City) (*data.DB, data.City) {
	schema, db := ready(t, ctx, testID, tc)

	if err := schema.DropAll(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data and schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to drop the data and schema.", tests.Success, testID)

	if err := schema.Create(ctx); err != nil {
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

// seedUser is a support test help function to consolidate the seeding of a
// user since so many data tests need this functionality.
func seedUser(t *testing.T, ctx context.Context, testID int, tc TestConfig, newUser data.NewUser, now time.Time) (*data.DB, data.User) {
	schema, db := ready(t, ctx, testID, tc)

	if err := schema.DropAll(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data and schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to drop the data and schema.", tests.Success, testID)

	if err := schema.Create(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to create the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to create the schema.", tests.Success, testID)

	userAdd, err := db.Mutate.AddUser(ctx, newUser, now)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to add a user: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to add a user.", tests.Success, testID)

	return db, userAdd
}
