# Leaseweb Assignment API

## Overview

This service exposes:

- `GET /v1/servers` — query server inventory with filters, paginated
- `POST /v1/admin/upload` — upload a CSV file to replace server data
- `GET /healthz` / `GET /readyz` — ops endpoints (unversioned - see below)

The API returns JSON responses and uses a standard error contract for errors.

An interactive Swagger UI is served at `/docs/` (e.g.
`http://localhost:8080/docs/`), backed by the hand-written OpenAPI 3 spec at
[`openapi.yaml`](openapi.yaml) (also served directly at `/openapi.yaml` for
import into other tools). This document is the prose version of the same
contract.

## Running the service

Start the server with the repository root config file present:

```bash
make run-server
```

If you do not have a `config.yaml`, create one with the following values:

```yaml
server:
  host: 0.0.0.0 # not 127.0.0.1 - see the comment in config.yaml
  port: 8080
  timeout: 30 # required, must be > 0 - applied as both the request read and write timeout

app:
  data_file: data/servers.csv
  upload_dir: data/uploads
  admin_token: your-secret-key
  logging:
    level: info # debug | info | warn | error
```

## Base URL

Local development:

```text
http://localhost:8080
```

Live deployment:

```text
https://leasewebassignment-1.onrender.com
```

| Resource        | Local                                  | Deployed                                                  |
|-----------------|-----------------------------------------|-------------------------------------------------------------|
| Filter UI       | `http://localhost:8080/ui/`             | https://leasewebassignment-1.onrender.com/ui/               |
| Swagger UI      | `http://localhost:8080/docs/`           | https://leasewebassignment-1.onrender.com/docs/              |
| OpenAPI spec    | `http://localhost:8080/openapi.yaml`    | https://leasewebassignment-1.onrender.com/openapi.yaml       |
| API base        | `http://localhost:8080/v1`              | https://leasewebassignment-1.onrender.com/v1                 |

If the deployed instance has been idle, the first request may take a few
seconds while Render spins it back up.

## API versioning

Routes are prefixed with `/v1`. This is a fresh API with no external
consumers yet, so there's no `/servers` compatibility alias - the intent is
that a future breaking change gets a `/v2` prefix rather than mutating `/v1`
underneath existing clients.

`/healthz` and `/readyz` are deliberately **not** versioned - they're
infrastructure contracts an orchestrator/load balancer polls, not part of
the API surface a client integrates against. See the dedicated section
below for how they differ.

## GET /v1/servers

Fetch available servers using optional query filters and pagination.
Filters compose with AND semantics (a result must satisfy every filter that
was supplied); repeated `ram` params compose with OR semantics within that
one filter, mirroring the assignment's spec sheet where RAM is a checkbox
group (select one or many) while the rest are single-value fields.

### Query parameters

| Parameter     | Type                 | Notes                                                                 |
|---------------|----------------------|------------------------------------------------------------------------|
| `model`       | string               | Case-insensitive substring match on model name                        |
| `ram`         | string, repeatable   | Exact match against one of the allowed values; repeat for multi-select, e.g. `ram=16GB&ram=32GB` |
| `location`    | string               | Case-insensitive substring match                                      |
| `disk_type`   | string               | One of `SAS`, `SATA`, `SSD`. Matches against the disk type family parsed from the HDD column — e.g. a drive labeled `SATA2` in the source data matches `disk_type=SATA` |
| `storage_min` | string               | Minimum total storage, accepts `GB`/`TB` units, e.g. `500GB`, `1TB`   |
| `storage_max` | string               | Maximum total storage, accepts `GB`/`TB` units                        |
| `limit`       | integer, optional    | Page size. Default `50`, capped at `200` regardless of what's requested |
| `offset`      | integer, optional    | Zero-based offset into the filtered result set. Default `0`           |

Allowed `ram` and `disk_type` values default to the assignment's spec sheet
(`2GB…96GB` and `SAS`/`SATA`/`SSD`) and can be overridden via
`app.allowed_ram` / `app.allowed_disk_types` in config.

If more than one parameter is invalid, the response reports every problem
in a single `400`, not just the first one encountered — see the error
contract below.

### Caching

Responses set `ETag` (a hash of the response body) and
`Cache-Control: no-cache`. `no-cache` here means "always revalidate with the
server before use," not "don't cache" — send the `ETag` back as
`If-None-Match` on a repeat request and the server returns `304 Not
Modified` with no body if nothing's changed since. This is safe because
the catalog only ever changes on `POST /v1/admin/upload`.

### Example request

```http
GET /v1/servers?ram=16GB&ram=32GB&storage_min=500GB&storage_max=1TB&disk_type=SSD&location=Amsterdam&limit=50&offset=0 HTTP/1.1
Host: localhost:8080
Accept: application/json
```

### Example response

```json
{
  "data": [
    {
      "model": "Dell R210",
      "ram": "16GBDDR3",
      "hdd": "2x2TBSATA2",
      "location": "AmsterdamAMS-01",
      "price": 49.99
    }
  ],
  "meta": {
    "total": 1,
    "limit": 50,
    "offset": 0
  }
}
```

`meta.total` is the full count of servers matching the filter, independent
of how many are in `data` — use it to tell "8 total, 5 shown" apart from
"5 total, 5 shown" when building pagination controls.

## POST /v1/admin/upload

Upload a CSV file to replace the current server inventory. The upload
replaces the in-memory catalog entirely and is not partially applied on
error — a failure to parse leaves the previous catalog untouched.

This is modeled as an admin action rather than a resource
(`POST /v1/admin/upload` instead of, say, `PUT /v1/servers`) deliberately:
it doesn't create or update an addressable resource, it replaces the whole
catalog as a single all-or-nothing operation gated behind a separate
authorization boundary from the read API. If a more resource-oriented shape
is ever wanted, `PUT /v1/catalog` would be the natural evolution.

### Authentication

- Requires an `Authorization: Bearer <token>` header
- `<token>` must equal the configured `app.admin_token`, compared in
  constant time. This is a plain pre-shared secret, not a JWT - there's no
  signature, claims, or expiry to verify, deliberately: this service has no
  user/identity concept, just one administrative action, so a shared secret
  is the right-sized tool rather than standing up token issuance for its own sake.
- An empty `admin_token` always rejects requests (fails closed — there is
  no way to accidentally disable auth by leaving the key blank)

### Request body

CSV columns, in order: `Model,RAM,HDD,Location,Price`. Price accepts `€`,
`$`, and `S$` prefixes. Any columns beyond `Price` are ignored (the source
spreadsheet export carries a "Filters" reference table in trailing columns
on every row - see Notes below).

- `file` — form-data file upload field
- Request body is capped at 20 MiB; a larger upload gets `413` (code
  `request_too_large`) rather than being accepted and left to consume
  unbounded memory or disk. Any part of the upload that spills to a
  temporary file during parsing is deleted once the request completes.

### Example request

```http
POST /v1/admin/upload HTTP/1.1
Host: localhost:8080
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary...

------WebKitFormBoundary...
Content-Disposition: form-data; name="file"; filename="servers.csv"
Content-Type: text/csv

...file contents...
------WebKitFormBoundary...--
```

### Success response

- `204 No Content`

## GET /healthz, GET /readyz

- **`/healthz`** — pure liveness. Always returns `200` once the process is
  serving HTTP at all, regardless of whether any data has loaded.
- **`/readyz`** — readiness. Returns `200` once server data has been
  successfully loaded at least once; until then, `503` with the standard
  error contract:

  ```json
  {
    "error": {
      "code": "service_unavailable",
      "message": "server has no data loaded yet",
      "details": "upload a catalog via POST /v1/admin/upload to recover"
    }
  }
  ```

These are allowed to diverge because the server **always boots**, even if
`app.data_file` is missing or fails to parse at startup - that failure is
logged as an error but is not fatal (see "Notes" below). An orchestrator
that gates traffic on `/readyz` won't route requests to an instance with an
empty catalog; one that only checks `/healthz` will see the process as
alive throughout. `/readyz` flips to `200` the moment data loads
successfully, whether that's the original startup load or a later
`POST /v1/admin/upload` - a successful upload is a valid recovery path,
not just an ongoing-operation endpoint.

## Error response contract

All errors are returned with the same JSON shape:

```json
{
  "error": {
    "code": "invalid_input",
    "message": "invalid query parameters",
    "details": "ram: \"9999GB\" is invalid, must be one of: 2GB, 4GB, ...; disk_type: \"NVME\" is invalid, must be one of: SAS, SATA, SSD"
  }
}
```

Every response also carries an `X-Request-Id` header — include it when
reporting an issue so the matching server log line can be found.

### 404 and 405

- A path that doesn't match any registered route returns `404` (code
  `not_found`) in the standard error shape above — the one exception is the
  literal root path `/`, which redirects to the bundled UI at `/ui/`.
- A known path called with the wrong HTTP method returns `405` (code
  `method_not_allowed`) with an `Allow` header listing the methods that
  path actually supports, rather than silently falling through to a
  different handler or a generic 404.

## CORS

The API does not send CORS headers and does not support cross-origin
requests. This is a deliberate scope decision, not an oversight: the
bundled UI (`web/index.html`) is served by this same process and calls the
API same-origin. If a separately-hosted frontend ever needs to call this
API directly, CORS headers should be added explicitly rather than opened
permissively by default.

## Logging & observability

- Every request is logged (method, path, status, latency, request ID) via
  `internal/server/middleware`'s access-log middleware.
- Every `5xx` is logged server-side with the underlying error - `4xx`s are
  normal traffic (bad client input) and are already visible via the access
  log's status field, so they aren't separately error-logged.
- `UploadServerData` and `LoadServerData` are logged independently via
  `service.NewLoggingService`, since the startup data load runs directly
  from `main()` before the HTTP server is listening - there's no request
  for the access-log middleware to attach to. `GetServers` is deliberately
  *not* separately logged; it's always called through the HTTP handler and
  the access log already covers it, so a second log line would just
  duplicate the first.
- `app.logging.level` (`debug`/`info`/`warn`/`error`) controls the minimum
  level logged; default `info`.
- A panic in any handler is recovered and turned into a structured `500`
  (via the standard error contract above) instead of crashing the
  connection - see `internal/server/middleware/recover.go`.

## Postman collection

Import the collection file at `postman/leasewebassignment.postman_collection.json` into Postman.

### Environment variables

- `base_url` — `http://localhost:8080`
- `admin_token` — your admin bearer token (matches `app.admin_token` in config)

## Performance

- `GET /v1/servers` does zero string parsing per request - total storage,
  RAM family, and disk type family are all parsed once when data is loaded
  (`ReplaceServers`), not once per filter evaluation. See
  `internal/store/store.go`'s `indexedServer` type.
- `go test ./internal/store/... -bench .` (or `make bench`) benchmarks the
  read (`ListServers`) and write (`ReplaceServers`) paths against a
  synthetic 50,000-row catalog.
- Responses are gzip-compressed when the client sends
  `Accept-Encoding: gzip`.

## Notes

- The server loads startup data from `app.data_file` on startup. If that
  load fails (missing file, bad CSV), the failure is logged but the server
  still starts - see "GET /healthz, GET /readyz" above for how to detect
  that state and `POST /v1/admin/upload` to recover from it without a
  redeploy.
- Total storage is computed by parsing the HDD column (e.g. `8x2TBSATA2` → 16TB) and converting everything to GB before comparing against `storage_min`/`storage_max`.
- RAM values in the source data carry a memory-technology suffix with no
  separator (e.g. `16GBDDR3`) - filtering matches against the extracted
  `<digits>GB` family, not the literal string, so `ram=16GB` matches
  `16GBDDR3`.
- `servers_filters_assignment.csv` at the repo root is the original assignment spreadsheet (actually an `.xlsx`, not a CSV — it's gitignored and not read by the application). The data the server actually serves lives at `data/servers.csv`.
