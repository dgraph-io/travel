package tests

import (
	"testing"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// Configuration for running tests.
const (
	dbImage = "dgraph/standalone:v20.03.1"
)

// NewUnit creates a test value with necessary application state to run
// database tests. It will return the host to use to connect to the database.
func NewUnit(t *testing.T) (apiHost string, teardown func()) {

	// Start a DB container instance with dgraph running.
	c := startDBContainer(t, dbImage)

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown = func() {
		t.Helper()
		t.Log("tearing down test ...")
		stopContainer(t, c.ID)
	}

	return c.APIHost, teardown
}
