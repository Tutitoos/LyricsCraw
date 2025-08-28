# LyricsCrawl

REST API in Go to fetch song lyrics by scraping Vagalume using Chrome (chromedp). Includes an in-memory cache with TTL to reduce latency on repeated queries.

## Features

- Gin HTTP server with endpoints `/health` and `/v1/lyrics`.
- Scraping with `chromedp` (Chrome runs headless by default).
- In-memory cache with TTL and a background janitor.
- Structured logging with Uber Zap.

## Requirements

- Go (the module declares `go 1.24.4`).
- Google Chrome or Chromium installed on the system (chromedp uses it under the hood).
- macOS, Linux, or Windows.

## Environment variables

- `APP_ENV`: `development` or `production`. Affects logs and Gin mode. Default: `development`.
- `APP_PORT`: server port. Default: `8080`.
- `APP_LYRICS_CACHE_TTL_SECONDS`: cache TTL in seconds. Default: `1800` (30 min).
- `APP_LYRICS_CACHE_MAX_ENTRIES`: maximum cache capacity. Default: `1000`.

You can define them in a `.env` at the repository root; if present, it is loaded on startup.

## Install and run

Clone the repository and fetch dependencies.

Development (manual reload):

```bash
make dev
```

Production (binary in `./bin/app`):

```bash
make build
make start
```

Alternative without Makefile:

```bash
# development
sh scripts/dev.sh

# production
sh scripts/build.sh
sh scripts/start.sh
```

## Endpoints

Health check:

```http
GET /health -> 200 { "status": "ok" }
```

Get lyrics:

```http
GET /v1/lyrics?query=<artist> - <song>
```

curl example:

```bash
curl "http://localhost:8080/v1/lyrics?query=Coldplay%20-%20Yellow"
```

Response:

```json
{
	"data": "<lyrics>",
	"cached": true
}
```

The cache key is the normalized `query` (lowercased, trimmed). If there is a cache hit, `cached` is `true` and the response is immediate; otherwise scraping runs and the result is stored.

## How scraping works (summary)

1) Opens a Chrome context (`headless=true` by default) with a random User-Agent.
2) Navigates to `https://www.vagalume.com.br/search?q=<query>` and takes the first result.
3) Adjusts the URL (removes `-traducao` if applicable) and navigates to the detail page.
4) If an 18+ notice appears, it tries to accept the modal.
5) Extracts the text from `div#lyrics` and removes confirmation messages.

Relevant code:
- `src/scraper/Scraper.go`
- `src/scraper/UserAgentGenerator.go`

## Notes about Chrome

- Each request currently creates a fresh Chrome context and closes it afterwards. This is safe and simple.
- If you need to keep Chrome open persistently across requests, consider creating a reusable global context and managing its lifecycle. (Not implemented by default.)

## In-memory cache

- Implementation in `src/cache/LyricsCache.go`.
- Automatically initialized in `main` by reading environment variables.
- Periodic background cleanup (~TTL/2, minimum 30s).

## Project structure

- `src/main.go`: server bootstrap, loads `.env`, logger, and router.
- `src/api/router/Router.go`: routes and groups.
- `src/api/controller/TokenController.go`: lyrics controller.
- `src/scraper/*`: scraping and user-agent.
- `src/logger/logger.go`: Zap configuration.
- `scripts/*`: helpers for dev/build/start.
- `bruno-http/*`: optional Bruno collection for testing.

## Common issues and fixes

- Chrome not found or slow to start: install stable Google Chrome and keep it updated. In containers, enable flags like `--no-sandbox` (already set in code) as needed.
- Permissions on macOS: if a security dialog appears when launching Chrome, allow the app.
- Broken selectors: Vagalume selectors may change. Check `a.gs-title`, `div#lyrics`, and the 18+ modal if things stop working.

## License

Not specified.
