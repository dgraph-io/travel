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

	_, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	tt := []struct {
		name       string
		retryDelay time.Duration
		timeout    time.Duration
		success    bool
	}{
		{"timeout", 500 * time.Millisecond, time.Second, false},
		{"ready", 500 * time.Millisecond, 5 * time.Second, true},
	}

	t.Log("Given the need to be able to validate the database is ready.")
	{
		for _, test := range tt {
			tf := func(t *testing.T) {
				t.Logf("\tWhen waiting up to %v for the database to be ready.", test.timeout)
				{
					ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
					defer cancel()

					err := data.Readiness(ctx, apiHost, test.retryDelay)
					switch test.success {
					case true:
						if err != nil {
							t.Fatalf("\t%s\tShould be able to see Dgraph is ready : %v", tests.Failed, err)
						}
						t.Logf("\t%s\tShould be able to see Dgraph is ready.", tests.Success)

					case false:
						if err == nil {
							t.Fatalf("\t%s\tShould be able to see Dgraph is Not ready : %v", tests.Failed, err)
						}
						t.Logf("\t%s\tShould be able to see Dgraph is Not ready.", tests.Success)
					}
				}
			}
			t.Run(test.name, tf)
		}
	}
}

// TestSchema validates the schema can be validated in Dgraph.
func TestSchema(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate a schema.")
	{
		t.Log("\tWhen handling a city schema.")
		{
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := data.Readiness(ctx, apiHost, 500*time.Millisecond)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to see Dgraph is ready : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to to see Dgraph is ready.", tests.Success)

			data, err := data.New(dbHost, apiHost)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to connect to Dgraph : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to connect to Dgraph.", tests.Success)

			if err := data.Validate.Schema(ctx); err != nil {
				t.Fatalf("\t%s\tShould be able to perform the schema operation : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to perform the schema operation.", tests.Success)

			schema, err := data.Query.Schema(ctx)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to query for the schema : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to query for the schema.", tests.Success)

			const predicates = 9
			if len(schema) != predicates {
				t.Log("\t\tGot:", len(schema))
				t.Log("\t\tExp:", predicates)
				t.Errorf("\t%s\tShould be able to see %d predicates in the schema : %v", tests.Failed, predicates, err)
			} else {
				t.Logf("\t%s\tShould be able to see %d predicates in the schema.", tests.Success, predicates)
			}
		}
	}
}
