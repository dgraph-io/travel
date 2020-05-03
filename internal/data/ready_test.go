package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
)

// TestReadiness validates the health check is working.
func TestReadiness(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	apiHost, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	type tableTest struct {
		name       string
		retryDelay time.Duration
		timeout    time.Duration
		success    bool
	}

	tt := []tableTest{
		{"timeout", 500 * time.Millisecond, time.Second, false},
		{"ready", 500 * time.Millisecond, 5 * time.Second, true},
	}

	t.Log("Given the need to be able to validate the database is ready.")
	{
		for testID, test := range tt {
			tf := func(t *testing.T) {
				t.Logf("\tTest %d:\tWhen waiting up to %v for the database to be ready.", testID, test.timeout)
				{
					ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
					defer cancel()

					err := data.Readiness(ctx, apiHost, test.retryDelay)
					switch test.success {
					case true:
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to see Dgraph is ready.", tests.Success, testID)

					case false:
						if err == nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is Not ready.", tests.Failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to see Dgraph is Not ready.", tests.Success, testID)
					}
				}
			}
			t.Run(test.name, tf)
		}
	}
}
