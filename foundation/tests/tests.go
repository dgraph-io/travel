// Package tests contains supporting code for running tests.
package tests

import (
	"fmt"
	"log"
	"os"
	"testing"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// Configuration for running tests.
const (
	dbImage = "dgraph/standalone:master"
	dbPort  = "8080"
)

// NewUnit creates a test value with necessary application state to run
// database tests. It will return the host to use to connect to the database.
func NewUnit(t *testing.T) (*log.Logger, string, func()) {
	c := startContainer(t, dbImage, dbPort)

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		t.Log("tearing down test ...")
		stopContainer(t, c.ID)
	}

	url := fmt.Sprintf("http://%s", c.Host)
	log := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	return log, url, teardown
}
