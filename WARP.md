# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Repository overview

This repository contains a single-page landing site for the **ronks.ru** auto parts aggregator plus a small Go backend that proxies cross-search requests to an external API. The goal is to provide users with a simple way to search for auto parts and their analogs (кроссы) and to redirect interested users into the main ronks.ru product.

## Architecture and code structure

### Frontend (landing page)

- `index.html`
  - Dark-themed, single-page layout with:
    - Header (`ronks.ru` logo and tagline).
    - Hero section with heading, marketing description, and search form.
    - Results area that renders a table of analogs beneath the search form.
    - Footer with dynamic year (`<span id="year">`).
  - Search UI:
    - Input `#part-number` and button `#search-button` inside a `.search-form`.
    - Below the form:
      - `#error` (used for validation and server error messages).
      - `#tableContainer` (used to render search results).
  - Inline CSS in `<style>` defines the full layout and the results table styling (including `.clickable-row` rows that highlight on hover).
  - Inline `<script>` implements:
    - Year update for the footer.
    - `searchCrosses()` — calls the backend at `/api/analog?n=...`, shows loading state, handles errors, and renders results.
    - `processAnalogData(data, originalPart)` — mirrors the logic from the Python prototype: builds manufacturer and product maps from `manufacturerList.mf` and `productList.p`, filters `analogList.a` by requested part, and returns normalized rows.
    - `renderTable(items)` — renders an HTML `<table>` with original and analog columns and attaches click handlers to each `.clickable-row` that redirect to `https://lk.ronks.ru`.
    - `escapeHtml(text)` — safely escapes values before injecting them into the table.

### Backend (Go proxy server)

- `go.mod`
  - Initializes the Go module for this repository (no external dependencies, only standard library).

- `server.go`
  - Simple HTTP server based on `net/http` that mirrors the behaviour of the original Python `server.py` prototype located in the separate `Cross_Search` folder.
  - Constants:
    - `apiURL = "https://fapi.iisis.ru/fapi/v2/analogList"` — upstream cross-search API.
    - `email`, `ui`, `ver` — same values as in the Python prototype, forwarded as query parameters.
    - `addr = "localhost:8000"` — bind address.
  - Handlers:
    - `handleAnalog` (mounted at `/api/analog`):
      - Accepts only `GET`.
      - Reads required query parameter `n` (part number). If missing, returns JSON `{ "error": "Missing 'n' parameter" }` with HTTP 400.
      - Builds a proxied request to `apiURL` with query parameters `n`, `email`, `ui`, `ver`.
      - Uses `http.Client` with a 10-second timeout to call the upstream API.
      - On success, forwards the upstream response body directly to the client and sets headers:
        - `Content-Type: application/json; charset=utf-8`
        - `Access-Control-Allow-Origin: *`
      - On failure (network/timeout/URL issues), returns JSON `{ "error": "API error: …" }` (or a generic internal error) with HTTP 500.
    - `handleIndex` (mounted at `/` as a catch-all for non-API paths):
      - Serves `index.html` from the repository root using `http.ServeFile`.
      - Adds `Content-Type: text/html; charset=utf-8` and returns 404 if `index.html` is missing.
  - Server setup:
    - Uses `http.NewServeMux()` and `http.Server` with reasonable read/write timeouts.
    - Logs startup messages including the bound address and absolute path to `index.html`.

### External prototype (for reference only)

- Outside this repository, there is a Python prototype in `Cross_Search/server.py` and `cross-search.html`. The Go backend and the JS logic in `index.html` are intentionally aligned with that prototype: same `/api/analog` contract, same upstream API, and compatible response processing.

## Commands and workflows

This project has no Node.js or Python tooling; the only runtime dependency is Go for the backend.

### Run the Go server (development and local usage)

From the repository root (`ronks-ru-landing`):

- Start the server:
  - `go run ./server.go`
- Then open in a browser:
  - `http://localhost:8008`

Behaviour:
- `GET /` (and any non-`/api/analog` path) serves `index.html`.
- `GET /api/analog?n=<partNumber>` proxies the request to the external API and returns JSON, which the frontend in `index.html` consumes.

### Build a local binary

From the repository root:

- `go build -o ronks-server.exe ./server.go`

This produces a self-contained executable (`ronks-server.exe`) that, when run, starts the server on `http://localhost:8008`.

### Tests and linting

There are currently no automated tests or linters configured for this repository. If you introduce them (for example, Go tests or `golangci-lint`), update this file with the canonical commands to run them.

