# Leaseweb Assignment API

A Go REST API for listing and filtering a server catalog, with a small
filter UI, built for the Leaseweb technical assessment.

## Live deployment

Base URL: https://leasewebassignment-1.onrender.com

- Filter UI: https://leasewebassignment-1.onrender.com/ui/
- Swagger UI: https://leasewebassignment-1.onrender.com/docs/
- OpenAPI spec: https://leasewebassignment-1.onrender.com/openapi.yaml

If the instance has been idle, the first request may take a few seconds
while Render spins it back up.

## Requirements

- Go 1.25+
- [golangci-lint](https://golangci-lint.run/) v2, optional, only needed for `make lint`

## Setup

```bash
git clone <this repo>
cd leasewebassignment
go mod download
```

The repo ships with `config.yaml` and `data/servers.csv` already in place, so
no further setup is required to run it.

## Running

```bash
make run-server
```

This builds and runs the server with `config.yaml`, loading
`data/servers.csv` on startup and listening on `http://127.0.0.1:8080`.

Alternatively, build a binary directly:

```bash
make build
./bin/leasewebassignment -config=config.yaml
```

Then open `http://localhost:8080/` (it redirects to the filter UI at `/ui/`),
or query the API directly, e.g.:

```bash
curl "http://localhost:8080/v1/servers?ram=16GB&ram=32GB&disk_type=SSD&limit=10"
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

The API is versioned under `/v1` - see [docs/api.md](docs/api.md) for why,
and for the full endpoint reference. An interactive Swagger UI is served at
`/docs/`, backed by the hand-written spec at
[docs/openapi.yaml](docs/openapi.yaml).

The server always boots, even if `app.data_file` is missing or fails to
parse - a bad startup file logs an error but isn't fatal, since
`POST /v1/admin/upload` is already the built-in way to load/replace data at
runtime. `GET /readyz` reports `503` until a load has succeeded at least
once (startup or via that upload), so an orchestrator won't route traffic
to an instance with nothing to serve.

## Verifying the build

```bash
make verify   # go build + go vet + go test -race
```

`make verify` deliberately only needs the Go toolchain - no extra tools to
install to sanity-check the repo. Two additional checks run in CI
(`.github/workflows/ci.yml`) and are available locally if you have the
tooling:

```bash
make lint      # golangci-lint (govet, staticcheck, unused, errcheck, ...)
make bench     # benchmarks the store package's read/write paths
```

Build and vet run before tests in both `make verify` and CI specifically so
a compile break is never masked by an unrelated package's test output.

## Loading a different dataset

Upload via `POST /v1/admin/upload` (see [docs/api.md](docs/api.md)), behind
the configured bearer token - this is the only way to load data after
startup; there's no separate CLI import mode. To change what loads
automatically at startup, edit `app.data_file` in `config.yaml`.

## Docker

```bash
make docker
docker run -p 8080:8080 leasewebassignment:local
```

## Project layout

```
cmd/                              entrypoint
internal/config/                  config file loading + validation
internal/platform/log/            structured (zap) logger construction - generic infra, no app types
internal/platform/shutdown/       signal handling + graceful shutdown - generic infra, no app types
internal/platform/httperr/        generic HTTP error-response scaffolding - no app types either
internal/api/                     translates THIS app's error types (store/service) into httperr - the one non-generic piece
internal/model/                    domain types and CSV field parsing (price, RAM, HDD)
internal/service/                  upload/parse orchestration + logging decorator
internal/store/                    repository (in-memory) + filter evaluation
internal/server/                    HTTP server, routes
internal/server/handlers/            request/response translation per route
internal/server/middleware/          auth, request ID, access log, panic recovery, gzip
internal/testutil/                  shared deterministic fixture used by store/API tests
web/                               static filter UI served at /ui/
docs/api.md                        REST API reference for consumers
docs/openapi.yaml                  hand-written OpenAPI 3 spec, served interactively at /docs/ (see internal/server/swaggerui.go)
postman/                           Postman collection
tests/                             end-to-end HTTP tests, root level not internal/ (see tests/doc.go for why)
```

## Logging & observability

Every request is logged (method, path, status, latency, request ID); every
`5xx` is logged with the underlying error; the startup data load is logged
independently since it runs before the HTTP server is even listening.
`app.logging.level` in config controls verbosity. See
[docs/api.md](docs/api.md#logging--observability) for details.

## Notes on the data

- `data/servers.csv` is what the server actually loads at startup
  (`app.data_file` in `config.yaml`).
- `servers_filters_assignment.csv` at the repo root is the original
  assignment spreadsheet - it's actually an `.xlsx` file with a `.csv`
  extension (confirmed via `file`), is gitignored, and is not read by the
  application. It's kept only as the original source-of-truth reference.
- Every row in the source export carries a handful of trailing spreadsheet
  columns (the assignment's "Filters" reference table, mostly empty past the
  first few rows) - the CSV parser explicitly ignores anything past the
  `Price` column rather than treating it as part of `Location`.
- RAM values carry a memory-technology suffix with no separator (e.g.
  `16GBDDR3`) and disk types carry a generation suffix (e.g. `SATA2`) -
  filtering normalizes both to the family the assignment's own filter spec
  uses (`16GB`, `SATA`). See `internal/model/server.go`.

See [docs/api.md](docs/api.md) for the full API reference.
