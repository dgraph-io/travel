package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/platform/tests"
)

// schema validates the schema we are storing is what we expect
// for the application.
func schema(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate a schema.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling the application schema.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				schema, _ := ready(t, ctx, 0, tc)

				if err := schema.DropAll(ctx); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data and schema: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to drop the data and schema.", tests.Success, testID)

				if err := schema.Create(ctx); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create the schema: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create the schema.", tests.Success, testID)
			}
		}
	}
	return tf
}
