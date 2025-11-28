# opgl-gateway-service

API Gateway service that orchestrates communication between OPGL microservices.

## Service Overview

- **Port**: 8080 (default)
- **Purpose**: Routes and orchestrates requests between opgl-data-service and opgl-cortex-engine-service
- **Framework**: Go with gorilla/mux router

## Project Structure

```
opgl-gateway-service/
├── main.go                      # Application entry point
├── internal/
│   ├── api/
│   │   ├── router.go            # Route definitions
│   │   ├── handlers.go          # HTTP request handlers
│   │   └── handlers_test.go     # Handler unit tests
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware for preflight requests
│   │   └── logging.go           # Request/response logging middleware
│   ├── models/
│   │   └── models.go            # Shared data models (Summoner, Match, AnalysisResult)
│   └── proxy/
│       ├── interface.go         # ServiceProxyInterface for dependency injection
│       ├── proxy.go             # Service proxy implementation
│       └── proxy_test.go        # Proxy unit tests
├── Makefile                     # Build, test, and run commands
├── Dockerfile                   # Docker containerization
└── .env.example                 # Environment variable template
```

## Endpoints

All endpoints use **POST** method (per project guidelines):

| Endpoint | Description |
|----------|-------------|
| `POST /health` | Health check |
| `POST /api/v1/summoner` | Proxy to opgl-data-service for summoner lookup |
| `POST /api/v1/matches` | Proxy to opgl-data-service for match history |
| `POST /api/v1/analyze` | Orchestrated analysis (data + cortex services) |

## Request Body Format

All endpoints use Riot ID format:

```json
{
  "region": "na",
  "gameName": "Newyenn",
  "tagLine": "GGEZ"
}
```

For matches endpoint, optional `count` parameter (defaults to 20):

```json
{
  "region": "na",
  "gameName": "Newyenn",
  "tagLine": "GGEZ",
  "count": 10
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `OPGL_DATA_URL` | http://localhost:8081 | opgl-data-service URL |
| `OPGL_CORTEX_URL` | http://localhost:8082 | opgl-cortex-engine-service URL |

## Development Commands

```bash
# Run locally
make run

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Build binary
make build

# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Lint code (requires golangci-lint)
make lint
```

## Key Implementation Details

### Handler Pattern
- Handlers receive requests, validate input, call proxy methods, and return JSON responses
- All handlers validate required fields: region, gameName, tagLine
- Error responses use `http.Error()` with appropriate status codes

### Service Proxy Pattern
- `ServiceProxy` handles all HTTP communication with downstream services
- Uses POST requests with JSON bodies for all service calls
- `GetMatchesByPUUID` method exists for internal optimization (avoids redundant lookups)

### Middleware Stack
1. **CORS Middleware** - Handles preflight OPTIONS requests
2. **Logging Middleware** - Logs incoming requests and response status codes using zerolog

### Analysis Flow (POST /api/v1/analyze)
1. Fetch summoner data from opgl-data-service using Riot ID
2. Fetch match history from opgl-data-service using PUUID (efficiency optimization)
3. Send summoner + matches to opgl-cortex-engine-service for analysis
4. Return analysis result to client

## Testing

Tests use interfaces for dependency injection:
- `ServiceProxyInterface` allows mocking proxy calls in handler tests
- Run `make test` for unit tests with race detection

## Dependencies

- `github.com/gorilla/mux` - HTTP router
- `github.com/rs/zerolog` - Structured logging
