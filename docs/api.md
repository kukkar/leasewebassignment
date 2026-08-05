# Leaseweb Assignment API

## Overview

This service exposes two HTTP endpoints:

- `GET /servers` — query server inventory with filters
- `POST /admin/upload` — upload a CSV file to replace server data

The API returns JSON responses and uses a standard error contract for errors.

## Running the service

Start the server with the repository root config file present:

```bash
make run-server
```

If you do not have a `config.yaml`, create one with the following values:

```yaml
server:
  host: 127.0.0.1
  port: 8080
  timeout: 30

app:
  data_file: data/servers.csv
  upload_dir: data/uploads
  jwt_signing_key: your-secret-key
```

## Base URL

```text
http://localhost:8080
```

## GET /servers

Fetch available servers using optional query filters.

### Query parameters

- `model` — server model name
- `ram` — RAM string, e.g. `32GB`
- `hdd` — HDD string, e.g. `4x1TB`
- `location` — location code or name
- `price_min` — minimum price
- `price_max` — maximum price

### Example request

```http
GET /servers?model=AMD%20EPYC&ram=32GB&location=Amsterdam&price_min=50&price_max=200 HTTP/1.1
Host: localhost:8080
Accept: application/json
```

### Example response

```json
{
  "data": [
    {
      "model": "AMD EPYC 7501",
      "ram": "32GB",
      "hdd": "4x1TB",
      "location": "Amsterdam",
      "price": 105.99
    }
  ]
}
```

## POST /admin/upload

Upload a CSV file to replace the current server inventory.

### Authentication

- Requires `Authorization: Bearer <token>` header
- Use the configured `jwt_signing_key` in your application config

### Request body

- `file` — form-data file upload field

### Example request

```http
POST /admin/upload HTTP/1.1
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

### Error response contract

All errors are returned with the same JSON shape:

```json
{
  "error": {
    "code": "invalid_input",
    "message": "invalid query parameters",
    "details": "..."
  }
}
```

## Postman collection

Import the collection file at `postman/leasewebassignment.postman_collection.json` into Postman.

### Environment variables

- `base_url` — `http://localhost:8080`
- `jwt_token` — your JWT bearer token

## Notes

- The server loads startup data from `app.data_file` on startup.
- The CSV parser accepts price formats like `€294.99`, `$105.99`, and `S$565.99`.
