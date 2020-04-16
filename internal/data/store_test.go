package data_test

import (
	"testing"

	"github.com/dgraph-io/travel/internal/platform/tests"
)

// TestStore validates all the support that provides data storage.
func TestStore(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("weather", storeWeather)
}

// storeWeather validates a weather node can be stored in the database.
func storeWeather(t *testing.T) {
	t.Helper()

	dbHost, apiHost, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to be able to validate storing weather.")
	{
		t.Log("\tWhen handling weather for sydney.")
		{
		}
	}
}
