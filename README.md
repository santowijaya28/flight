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
  - best value
- Multiple sort keys, for example `price_asc,duration_asc`
- Timezone-aware parsing and WIB/WITA/WIT labels
- Normalized baggage, amenities, aircraft, duration, and price
- IDR price formatting, for example `Rp 1.250.000`
- Arrival-after-departure validation
- Simulated provider latency and AirAsia 10% failure rate
- Per-provider timeout and retry with exponential backoff
- Bounded in-memory LRU search cache with a 30-second TTL and `cache_hit` metadata

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
  resilience/                Generic timeout and retry engine
  response/                  HTTP response helpers
  simulation/                Provider latency and failure simulation
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
best_value
best_value_asc
best_value_desc
```

`best_value` ranks flights using normalized values within the current result set:

```text
60% price
30% duration
10% stops
```

Lower scores are better. `best_value` and `best_value_asc` are equivalent.

Multiple sort keys:

```text
sort=price_asc,duration_asc
sort=best_value,price_asc
sort=duration_asc,price_asc
sort=departure_asc,price_asc
```

When all requested keys are equal, flights are ordered by flight ID for deterministic output.

### Sample curl requests

Health check:

```bash
curl -s http://localhost:8000/ping
```

Basic search:

```bash
curl -s --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'cabin_class=economy'
```

Search with sorting:

```bash
curl -s --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'cabin_class=economy' \
  --data-urlencode 'sort=price_asc,duration_asc'
```

Search ranked by best value:

```bash
curl -s --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'cabin_class=economy' \
  --data-urlencode 'sort=best_value'
```

Search with price, stops, airline, and duration filters:

```bash
curl -s --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'cabin_class=economy' \
  --data-urlencode 'min_price=500000' \
  --data-urlencode 'max_price=1000000' \
  --data-urlencode 'max_stops=1' \
  --data-urlencode 'airlines=QZ,JT' \
  --data-urlencode 'min_duration=90' \
  --data-urlencode 'max_duration=240' \
  --data-urlencode 'sort=price_asc'
```

Direct flights only:

```bash
curl -s --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'max_stops=0'
```

Search with departure and arrival time windows:

```bash
curl -s --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'departure_from=06:00' \
  --data-urlencode 'departure_to=18:00' \
  --data-urlencode 'arrival_from=08:00' \
  --data-urlencode 'arrival_to=23:00'
```

Check cache behavior by running the same request twice:

```bash
curl -s --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=2025-12-15' \
  --data-urlencode 'passengers=1' \
  --data-urlencode 'sort=price_asc' \
  | jq '.metadata.cache_hit'
```

The first response returns `false`; an identical request within 30 seconds returns `true`.

Invalid request example:

```bash
curl -i --get 'http://localhost:8000/search' \
  --data-urlencode 'origin=CGK' \
  --data-urlencode 'destination=DPS' \
  --data-urlencode 'departure_date=invalid-date' \
  --data-urlencode 'passengers=1'
```

This returns `400 Bad Request`.

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
    "search_time_ms": 1,
    "cache_hit": false
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

### Error responses

Validation and request errors return a JSON error envelope:

```json
{
  "header": {
    "error": "invalid input: empty origin"
  }
}
```

When every provider fails, the service returns HTTP `500` and joins the provider errors into a single newline-separated message:

```json
{
  "header": {
    "error": "all provider down\nGarudaIndonesia: error oppening files open external/garuda_indonesa_search_response.json: no such file or directory\nAirAsia: error oppening files open external/airasia_search_respose.json: no such file or directory\nLionAir: error opening file: open external/lion_air_search_respons.json: no such file or directory\nBatikAir: error oppening files open external/batik_air_search_resonse.json: no such file or directory"
  }
}
```

Partial provider failures are treated as successful partial results and return HTTP `200`, with failures reflected in `metadata.providers_failed`.

## OpenAPI

The OpenAPI 3 specification is available at:

```text
docs/openapi.yaml
```

```bash
make openapi
```

You can open the file with Swagger Editor, Redoc, Postman, or another OpenAPI-compatible tool.

## Mock provider behavior

Provider calls simulate real-world latency:

| Provider | Delay |
|---|---|
| Garuda Indonesia | 50–100ms |
| Lion Air | 100–200ms |
| Batik Air | 200–400ms |
| AirAsia | 50–150ms |

AirAsia additionally has a **10% simulated failure rate**. A failed AirAsia call is treated as a provider error.

Because responses are cached, repeated identical searches may return immediately with `cache_hit: true`. Use a different search or sort parameter to observe the provider delays again.

## Provider resilience

Each provider is wrapped with configurable timeout and retry settings. Defaults:

- A **1-second per-attempt timeout**
- Up to **3 attempts** total
- Exponential backoff of **50ms → 100ms → 250ms**

Only transient failures are retried:

- Simulated upstream failures
- Context deadline exceeded

Configuration or parsing errors are not retried. If a provider still fails after all attempts, it is treated as a partial provider failure when other providers succeed.

## Caching

Search results are cached in memory using a bounded LRU cache with a 30-second TTL and a maximum of 10,000 entries. The cache is injected into the usecase from the application entry point, so the cache implementation can be replaced without changing the usecase logic.

The cache key includes the full normalized search input, including filters and sort keys.

- First request: `cache_hit: false`
- Repeated request within the TTL: `cache_hit: true`
- When the cache reaches capacity, the least-recently used entry is evicted
- Expired entries are removed on access or by LRU eviction

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

[Cache]
MaxEntries = 10000
TTLSeconds = 30

[Provider]
TimeoutMs = 1000
MaxAttempts = 3
InitialBackoffMs = 50
MaxBackoffMs = 250
```

## Make targets

```bash
make run       # Start the API server
make build     # Build bin/flight-api
make test      # Run Go tests
make vet       # Run go vet
make fmt       # Format all Go files
make tidy      # Run go mod tidy
make mocks     # Generate usecase mocks
make clean     # Remove build artifacts
make openapi   # Show OpenAPI specification path
```

## Notes and current limitations

- Cross-provider price comparison for the same flight is not implemented yet.
- The project currently has no automated test files; `make test` compiles and runs the Go test suite.
