package data_test

import (
	"context"
	"testing"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
)

// TestValidateSchema validates the schema can be validated in Dgraph.
func TestValidateSchema(t *testing.T) {
	dbHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate a schema.")
	{
		t.Log("\tWhen handling a city schema.")
		{
			ctx := context.Background()

			// Construct a Data value for working with the database.
			data, err := data.New(dbHost)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to connect to Dgraph : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to connect to Dgraph.", tests.Success)

			// Validate the schema in the database before we start.
			if err := data.Validate.Schema(ctx); err != nil {
				t.Fatalf("\t%s\tShould be able to perform the schema operation : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to perform the schema operation.", tests.Success)
		}
	}
}
