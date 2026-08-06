// Package tests holds end-to-end tests that exercise the full HTTP stack
// (server.NewServer + httptest.Server) rather than a single package in
// isolation. It's kept as its own package instead of colocating with, say,
// internal/server - deliberately, not by default:
//
//   - These tests assert on cross-package behavior (routing, middleware
//     chain, handler, service, and store all wired together), so there's no
//     single package they'd naturally belong to.
//   - Keeping them separate means `go test ./internal/server/...` stays a
//     fast, dependency-light unit-test run, while `go test ./...` (which
//     make verify runs) still covers the full stack.
//
// Package-level (non-e2e) tests still live next to the code they exercise -
// see internal/store/*_test.go, internal/service/*_test.go, etc.
package tests
