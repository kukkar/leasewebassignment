// Package tests holds end-to-end tests that exercise the full HTTP stack
// (server.NewServer + httptest.Server) rather than a single package in
// isolation. Two deliberate choices, not defaults:
//
//   - Its own package instead of colocating with, say, internal/server:
//     these tests assert on cross-package behavior (routing, middleware
//     chain, handler, service, and store all wired together), so there's no
//     single package they'd naturally belong to. Keeping them separate also
//     means `go test ./internal/server/...` stays a fast, dependency-light
//     unit-test run, while `go test ./...` (which `make verify` runs) still
//     covers the full stack.
//   - Root level instead of internal/tests: this package is a leaf - nothing
//     in the module imports it, so it never needed the import-visibility
//     restriction internal/ exists to enforce. It only depends on the rest
//     of the module, never the other way around, which is exactly the shape
//     a top-level integration/e2e test directory is for in larger Go
//     projects, as opposed to internal/, which is for code other packages
//     depend on and that must stay unimportable from outside the module.
//
// Package-level (non-e2e) tests still live next to the code they exercise -
// see internal/store/*_test.go, internal/service/*_test.go, etc.
package tests
