package data_test

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/dgraph-io/travel/internal/data"
// 	"github.com/dgraph-io/travel/internal/platform/tests"
// 	"github.com/google/go-cmp/cmp"
// )

// // TestDelete validates all the support that provides data deleting.
// func TestDelete(t *testing.T) {
// 	if testing.Short() {
// 		t.SkipNow()
// 	}

// 	t.Run("advisory", deleteAdvisory)
// }

// // deleteAdvisory validates an advisory node can be deleted from the database.
// func deleteAdvisory(t *testing.T) {
// 	t.Helper()

// 	apiHost, teardown := tests.NewUnit(t)
// 	defer teardown()

// 	t.Log("Given the need to be able to validate deleting an advisory.")
// 	{
// 		testID := 0
// 		t.Logf("\tTest %d:\tWhen handling an advisory for sydney.", testID)
// 		{
// 			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 			defer cancel()

// 			db, cityID := addCity(t, ctx, 0, apiHost)

// 			addAdvisory := data.Advisory{
// 				Country:     "Australia",
// 				CountryCode: "AU",
// 				Continent:   "Australia",
// 				Score:       4,
// 				LastUpdated: "today",
// 				Message:     "feel like teen spirit",
// 				Source:      "friendly neighborhood community engineers",
// 			}

// 			// addAdvisory, err := db.Store.Advisory(ctx, cityID, addAdvisory)
// 			// if err != nil {
// 			// 	t.Fatalf("\t%s\tTest %d:\tShould be able to save an advisory node in Dgraph: %v", tests.Failed, testID, err)
// 			// }
// 			// t.Logf("\t%s\tTest %d:\tShould be able to save an advisory node in Dgraph.", tests.Success, testID)

// 			advisory, err := db.Query.Advisory(ctx, cityID)
// 			if err != nil {
// 				t.Fatalf("\t%s\tTest %d:\tShould be able to query for the advisory: %v", tests.Failed, testID, err)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould be able to query for the advisory.", tests.Success, testID)

// 			if diff := cmp.Diff(addAdvisory, advisory); diff != "" {
// 				t.Fatalf("\t%s\tTest %d:\tShould get back the same advisory. Diff:\n%s", tests.Failed, testID, diff)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould get back the same advisory.", tests.Success, testID)

// 			if err := db.Delete.Advisory(ctx, cityID); err != nil {
// 				t.Fatalf("\t%s\tTest %d:\tShould be able to delete the advisory: %v", tests.Failed, testID, err)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould be able to delete the advisory.", tests.Success, testID)

// 			if _, err := db.Query.Advisory(ctx, cityID); err == nil {
// 				t.Fatalf("\t%s\tTest %d:\tShould not be able to query for the advisory.", tests.Failed, testID)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould not be able to query for the advisory.", tests.Success, testID)
// 		}
// 	}
// }
