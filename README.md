<div align="center">

# Cherkasyoblenergo API

[![EN](https://img.shields.io/badge/English-0e7837.svg)](README.md) [![UA](https://img.shields.io/badge/Ukrainian-c9c9c9.svg)](README_UA.md)

[![Deploy](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/deploy.yml/badge.svg)](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/deploy.yml)
[![Tests](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/tests.yml/badge.svg)](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/tests.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jeffjunior470/cherkasyoblenergo-api)](https://go.dev/)
[![License](https://img.shields.io/github/license/jeffjunior470/cherkasyoblenergo-api)](LICENSE)
[![Releases](https://img.shields.io/github/release/jeffjunior470/cherkasyoblenergo-api.svg)](https://github.com/jeffjunior470/cherkasyoblenergo-api/releases)
[![Changelog](https://img.shields.io/badge/changelog-md-blue)](CHANGELOG.md)

**Unofficial API** for retrieving power outage schedules from [cherkasyoblenergo.com](https://cherkasyoblenergo.com/)

</div>

This project is a fork of the upstream repository:
[sigmanor/cherkasyoblenergo-api](https://github.com/sigmanor/cherkasyoblenergo-api).

## Features

- Automatic schedule updates (every 5 minutes by default)
- Historical data access
- RESTful API interface
- IP-based rate limiting
- Optional API key authentication

## Installation

### Prerequisites

- Go 1.25 or higher to build from source

SQLite is embedded through a pure-Go driver. A C compiler, a separate SQLite
installation, and CGO are not required. Prebuilt binaries do not require Go.
The driver is `github.com/glebarez/sqlite` v1.11.0, used with GORM.
Windows and Linux need separate binaries for the target architecture.

### Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/jeffjunior470/cherkasyoblenergo-api.git
   cd cherkasyoblenergo-api
   ```

2. **Configure environment variables**

   Create a `.env` file in the working directory; the current configuration loader requires this file, even when settings are supplied through environment variables. You can copy `.env.example` to start. For local HTTP use, set `FORCE_HTTPS=false`; leave `API_KEY=` empty to disable authentication, or replace the example key with your own key. See the table below for all available options:

   | Variable | Required | Default | Description |
   |----------|----------|---------|-------------|
   | `DB_NAME` | No | `cherkasyoblenergo.db` | SQLite database file path |
   | `SERVER_PORT` | No | `8080` | Port for the API server |
   | `NEWS_URL` | No | `https://www.cherkasyoblenergo.com/api/v1/posts/category/news?lang=uk&page=0&size=20` | URL to parse schedules from |
   | `PARSING_INTERVAL_MINUTES` | No | `5` | How often to check for new schedules (minutes) |
   | `RATE_LIMIT_PER_MINUTE` | No | `60` | Max requests per minute per IP |
   | `LOG_LEVEL` | No | `info` | Logging level (`debug`, `info`, `warn`, `error`) |
   | `FORCE_HTTPS` | No | `false` | Redirect HTTP to HTTPS |
   | `API_KEY` | No | - | If set, enables API key authentication |
   | `PROXY_MODE` | No | `none` | Proxy mode for real client IP detection: `cloudflare`, `standard`, or `none` |

3. **Run the application**

   ```bash
   go run ./cmd/server/main.go
   ```

### Building

**Windows (cmd.exe):**

```bat
set CGO_ENABLED=0
go mod download
go build -o cherkasyoblenergo_api.exe ./cmd/server
cherkasyoblenergo_api.exe
```

**Ubuntu / Linux:**

```bash
export CGO_ENABLED=0
go mod download
go build -o cherkasyoblenergo_api ./cmd/server
./cherkasyoblenergo_api
```

Run these commands from the project directory containing `.env`. The database
directory must be writable. After startup, check
`http://localhost:8080/cherkasyoblenergo/api/`. Initial schedule parsing runs in
the background, followed by periodic updates. `today` and `tomorrow` use the
host's local timezone; configure it as `Europe/Kyiv` for Ukrainian dates.

To deploy a built binary, copy it and your `.env` to the target computer. Keep
the SQLite database at the configured `DB_NAME` path. Go and Docker are not
required to run the binary.

### SQLite driver migration

The application now uses a CGO-free SQLite driver instead of
`gorm.io/driver/sqlite` / `github.com/mattn/go-sqlite3`. The database remains a
SQLite file; no export/import is expected for the existing schema. Before
upgrading an existing deployment, stop the application and make a restorable
backup of the database, including any remaining `-wal` and `-shm` files.

Every connection is configured with WAL, a 5000 ms busy timeout, and foreign-key
enforcement. Timestamps are written in SQLite-compatible text format. Stored
offsets are retained on reads rather than normalized through the old
`_loc=Local` option; the represented instant is unchanged. API routes and JSON
field names remain unchanged. Lock retries use SQLite result codes instead of
matching error-message text.

### Docker

The Dockerfile builds with `CGO_ENABLED=0` and does not install GCC or musl-dev.
Docker is optional. The supplied `docker-compose.app.yml` expects an existing
external network named `think-root-network` and a writable database mount for
UID/GID `1000:1000`; adapt these settings for a standalone deployment.

## API Documentation

### Base URL

```
/cherkasyoblenergo/api
```

### Get Power Outage Schedule

```
GET /blackout-schedule
```

Retrieve scheduling records based on filter options.

#### Query Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `option` | Yes | `all`, `latest_n`, `by_date`, or `by_schedule_date` |
| `formatted` | No | `1` for pretty-printed JSON with two-space indentation; `0` or omitted for compact JSON |
| `format` | No | `json` (default) for JSON output, or `html` for an HTML page with schedules. HTML output uses `static/schedule.css`, which can be edited before building the Docker image |
| `date` | For `by_date`, `by_schedule_date` | `YYYY-MM-DD`, `today`, or `tomorrow` |
| `limit` | For `latest_n` | Integer > 0 |
| `queue` | No | Comma-separated queue identifiers (e.g., `3_2` or `4_1,3_1`) |

#### Filter Options

- `all` - Retrieves all schedule records
- `latest_n` - Gets limited number of recent records (requires `limit`)
- `by_date` - Gets records for specific publication date (requires `date`)
- `by_schedule_date` - Gets records for specific schedule date (requires `date`, optional `limit`)

#### Example Requests

**Get all schedules:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=all"
```

**Get latest 5 schedules:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=latest_n&limit=5"
```

**Get today's schedule for queue 3_2:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=by_schedule_date&date=today&queue=3_2"
```

**Get schedule with multiple queues:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=latest_n&limit=1&queue=4_1,3_1,2_2"
```

#### Response Examples

**Full response (without queue filter):**
```json
[
  {
    "id": 1234,
    "news_id": 100,
    "title": "Schedule for November 14",
    "date": "2024-03-20T10:30:00Z",
    "schedule_date": "2024-11-14",
    "1_1": "08:00-10:00",
    "1_2": "10:00-12:00",
    "2_1": "12:00-14:00",
    "2_2": "14:00-16:00",
    "3_1": "09:00-11:00",
    "3_2": "11:00-13:00",
    "4_1": "13:00-15:00",
    "4_2": "15:00-17:00",
    "5_1": "07:00-09:00",
    "5_2": "09:00-11:00",
    "6_1": "11:00-13:00",
    "6_2": "13:00-15:00"
  }
]
```

**Filtered response (with queue filter):**
```json
[
  {
    "id": 1234,
    "news_id": 100,
    "title": "Schedule for November 14",
    "date": "2024-03-20T10:30:00Z",
    "schedule_date": "2024-11-14",
    "3_2": "00:30 - 02:30, 06:00 - 09:00"
  }
]
```

### Rate Limiting

- Default: 60 requests per minute per IP
- Configurable via `RATE_LIMIT_PER_MINUTE` environment variable
- Response headers:
  - `X-RateLimit-Limit` - Maximum requests allowed
  - `X-RateLimit-Remaining` - Remaining requests in current window

### Authentication (Optional)

By default, the API is public and requires no authentication. For private instances, you can enable API key authentication:

1. Set the `API_KEY` environment variable:
   ```properties
   API_KEY=your-secret-key
   ```

2. Include the key in all requests:
   ```bash
   curl -H "X-API-Key: your-secret-key" "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=latest_n&limit=1"
   ```

If `API_KEY` is not set or empty, the API remains public.

Health check is available at `GET /health` and does not require an API key. It returns the service and database status in JSON format. Docker Compose uses this endpoint for its container health check.

### Error Responses

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 400 | Bad Request (invalid parameters) |
| 401 | Unauthorized (invalid or missing API key, if authentication is enabled) |
| 429 | Too Many Requests (rate limit exceeded) |
| 500 | Internal Server Error |

## Running Tests

Windows (cmd.exe):

```bat
set CGO_ENABLED=0
go test ./...
```

Ubuntu / Linux:

```bash
CGO_ENABLED=0 go test ./...
```

CI also runs with CGO disabled. Database tests cover connection PRAGMAs, legacy
UTC timestamps, new timestamps with offsets, reopening a file database, schema
migration, and lock-error detection. Fetch tests use a local HTTP test server.

## License

This project is licensed under the BSD 2-Clause License. See the [LICENSE](LICENSE) file for details.
