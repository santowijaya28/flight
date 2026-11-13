# Flight Search API

Flight search and aggregation service for the BookCabin take-home assignment. It loads mock responses from four airline providers, normalizes them into a common format, filters and sorts the combined result set, and exposes the data through an HTTP API.

## Features

- Aggregates AirAsia, Batik Air, Garuda Indonesia, and Lion Air mock data
- Parallel provider queries
- Provider-specific parsers and normalizers
- Search by origin, destination, and departure date
- Filters:
  - price range
  - maximum stops
  - airlines
  - duration
  - departure/arrival time windows
  - cabin class
  - passenger capacity
- Sorting:
  - price
  - duration
  - departure time
  - arrival time
- Multiple sort keys, for example `price_asc,duration_asc`
- Timezone-aware parsing and WIB/WITA/WIT labels
- Normalized baggage, amenities, aircraft, duration, and price
- IDR price formatting, for example `Rp 1.250.000`
- Arrival-after-departure validation

## Project structure

```text
internal/
  cmd/rest/                  HTTP server entry point
  src/config/                INI configuration loader
  src/delivery/              HTTP handlers and query parsing
  src/usecase/               Search orchestration, response building, sorting
  src/repository/            Provider aggregation
    filter/                  Shared flight filter
    provider/                Provider adapters
      airasia/
      batik-air/
      garuda-indonesia/
      lion-air/
  src/model/                 Shared domain models and helpers
pkg/
  log/                       Logging helpers
  response/                  HTTP response helpers
external/                    Mock provider responses
docs/openapi.yaml            OpenAPI 3 specification
```

## Requirements

- Go 1.27 or newer, as specified in `go.mod`
- The service must be run from the repository root because configuration and mock-data paths are relative to the working directory.

## Run

```bash
make run
```

Or directly:

```bash
go run ./internal/cmd/rest
```

The server listens on:

```text
http://localhost:8000
```

## API

### Health check

```http
GET /ping
```

Response:

```text
pong
```

### Search flights

```http
GET /search
```

Example:

```bash
curl --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'cabin_class=economy' \
  --data-urlencode 'sort=price_asc,duration_asc'
```

### Query parameters

| Parameter | Required | Description |
|---|---:|---|
| `origin` | Yes | Origin airport code, e.g. `CGK` |
| `destination` | Yes | Destination airport code, e.g. `DPS` |
| `departure_date` | Yes | Date in `YYYY-MM-DD` format |
| `passengers` | Yes | Passenger count, minimum `1` |
| `cabin_class` | No | Cabin class filter, e.g. `economy` |
| `sort` | No | Sort keys, comma-separated. Defaults to `price_asc` |
| `min_price` | No | Minimum total price |
| `max_price` | No | Maximum total price |
| `max_stops` | No | Maximum number of stops. Use `0` for direct flights |
| `airlines` | No | Comma-separated airline codes or names, e.g. `QZ,JT` |
| `min_duration` | No | Minimum total duration in minutes |
| `max_duration` | No | Maximum total duration in minutes |
| `departure_from` | No | Earliest departure local time, `HH:MM` |
| `departure_to` | No | Latest departure local time, `HH:MM` |
| `arrival_from` | No | Earliest arrival local time, `HH:MM` |
| `arrival_to` | No | Latest arrival local time, `HH:MM` |

Supported sort keys:

```text
price_asc
price_desc
duration_asc
duration_desc
departure_asc
departure_desc
arrival_asc
arrival_desc
```

Multiple sort keys:

```text
sort=price_asc,duration_asc
sort=duration_asc,price_asc
sort=departure_asc,price_asc
```

When all requested keys are equal, flights are ordered by flight ID for deterministic output.

### Example response

```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 13,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 1
  },
  "flights": [
    {
      "id": "QZ7250_AirAsia",
      "provider": "AirAsia",
      "airline": {
        "name": "AirAsia",
        "code": "QZ"
      },
      "flight_number": "QZ7250",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "timezone": "WIB",
        "datetime": "2025-12-15T15:15:00+07:00",
        "timestamp": 1765786500
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "timezone": "WITA",
        "datetime": "2025-12-15T20:35:00+08:00",
        "timestamp": 1765802100
      },
      "duration": {
        "total_minutes": 260,
        "formatted": "4h 20m"
      },
      "stops": 1,
      "price": {
        "amount": 485000,
        "currency": "IDR",
        "formatted": "Rp 485.000"
      },
      "available_seats": 88,
      "cabin_class": "economy",
      "aircraft": null,
      "amenities": [],
      "baggage": {
        "carry_on": "Cabin baggage only",
        "checked": "Additional fee"
      }
    }
  ]
}
```

## OpenAPI

The OpenAPI 3 specification is available at:

```text
docs/openapi.yaml
```

```bash
make openapi
```

You can open the file with Swagger Editor, Redoc, Postman, or another OpenAPI-compatible tool.

## Configuration

Provider mock-data paths are configured in:

```text
internal/files/etc/conf/config.ini
```

Default values:

```ini
[MockDataPath]
AirAsia = "external/airasia_search_response.json"
BatikAir = "external/batik_air_search_response.json"
GarudaIndonesia = "external/garuda_indonesia_search_response.json"
LionAir = "external/lion_air_search_response.json"
```

## Make targets

```bash
make run       # Start the API server
make build     # Build bin/flight-api
make test      # Run Go tests
make vet       # Run go vet
make fmt       # Format all Go files
make tidy      # Run go mod tidy
make clean     # Remove build artifacts
make openapi   # Show OpenAPI specification path
```

## Notes and current limitations

- Mock provider delays and the AirAsia failure simulation are not implemented yet.
- Caching and `cache_hit` metadata are not implemented yet.
- Cross-provider price comparison for the same flight and best-value ranking are not implemented yet.
- The project currently has no automated test files; `make test` compiles and runs the Go test suite.
