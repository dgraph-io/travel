package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/tests"
)

// readiness validates the health check is working.
func readiness(url string) func(t *testing.T) {
	tf := func(t *testing.T) {
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

						err := data.Readiness(ctx, url, test.retryDelay)
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
	return tf
}
