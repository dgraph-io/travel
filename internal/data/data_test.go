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
		Name     string
		Duration time.Duration
		Success  bool
	}{
		{"timeout", time.Second, false},
		{"ready", 5 * time.Second, true},
	}

	t.Log("Given the need to be able to validate the database is ready.")
	{
		for _, test := range tt {
			tf := func(t *testing.T) {
				t.Logf("\tWhen waiting up to %v for the database to be ready.", test.Duration)
				{
					err := data.Readiness(apiHost, test.Duration)

					switch test.Success {
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
			t.Run(test.Name, tf)
		}
	}
}

// TestValidateSchema validates the schema can be validated in Dgraph.
func TestValidateSchema(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate a schema.")
	{
		t.Log("\tWhen handling a city schema.")
		{
			err := data.Readiness(apiHost, 10*time.Second)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to see Dgraph is ready : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to to see Dgraph is ready.", tests.Success)

			data, err := data.New(dbHost)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to connect to Dgraph : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to connect to Dgraph.", tests.Success)

			if err := data.Validate.Schema(context.Background()); err != nil {
				t.Fatalf("\t%s\tShould be able to perform the schema operation : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to perform the schema operation.", tests.Success)
		}
	}
}
