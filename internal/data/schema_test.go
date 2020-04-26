package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/platform/tests"
)

// TestSchema validates the schema we are storing is what we expect
// for the application.
func TestSchema(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate a schema.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling the application schema.", testID)
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			db := ready(t, ctx, 0, apiHost)

			if err := db.Schema.Create(ctx); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to perform the schema operation: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to perform the schema operation.", tests.Success, testID)
		}
	}
}
