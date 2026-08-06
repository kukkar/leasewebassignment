// Package platform is a grouping directory, not a package itself - it holds
// generic technical infrastructure (log, shutdown) that references none of
// this application's own domain types (no model.Server, no config.Config).
// That's the dividing line from the rest of internal/: platform/* would be
// identical in any Go service, this one or a different one entirely, while
// config/server/service/store are specific to what this application does.
//
// Everything here still lives under internal/ rather than a top-level pkg/,
// deliberately: pkg/ is conventionally for code meant to be imported by a
// different module or a second binary in this repo, and neither exists
// today - there is exactly one cmd/ entrypoint. If a second binary is ever
// added (a worker, a migration tool), platform/ is where code shared
// between them belongs, without the packages needing to move again.
package platform
