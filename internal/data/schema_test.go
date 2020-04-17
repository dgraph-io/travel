package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
	"github.com/google/go-cmp/cmp"
)

// TestSchema validates the schema we are storing is what we expect
// for the application.
func TestSchema(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate a schema.")
	{
		t.Log("\tWhen handling the application schema.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			db := ready(t, ctx, dbHost, apiHost)

			if err := db.Schema.Create(ctx); err != nil {
				t.Fatalf("\t%s\tShould be able to perform the schema operation : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to perform the schema operation.", tests.Success)

			schema, err := db.Schema.Retrieve(ctx)
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
