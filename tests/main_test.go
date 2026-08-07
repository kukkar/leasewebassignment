package tests

import (
	"os"
	"testing"
)

// TestMain chdirs into the repo root before running any test in this
// package. server.routes() serves a handful of paths straight off disk
// relative to the process's working directory (web/, docs/openapi.yaml) -
// exactly like the compiled binary is expected to run (make run-server,
// the Dockerfile's WORKDIR, ./bin/leasewebassignment invoked from the repo
// root). `go test` instead starts every package's test binary with its
// working directory set to that package's own source directory - tests/ in
// this case - so without this, any test that actually fetches one of those
// disk-served paths would fail not because of a real bug, but because the
// test process's cwd doesn't match how the server is meant to be run.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
