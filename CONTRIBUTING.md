# Contributing to Quest Studio (`qst-studio`)

Thank you for your interest in contributing! This guide covers the repository layout, local development setup, and checks to run before opening a pull request.

## Prerequisites

- **Go** 1.27+
- **Node.js** 18+ and `npm`
- **Wails CLI v2** (for running the desktop app): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Linux system libraries**: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`

## Architecture Overview

- `app.go` / `main.go`: Wails desktop entry point and thin RPC adapter.
- `internal/`: Pure Go domain packages (`graph`, `quiz`, `practice`, `workspace`) with no Wails UI dependencies.
- `frontend/src/stores/`: Pure TypeScript validation, grading, and type definitions.
- `frontend/src/components/`: Svelte 5 UI components grouped by feature (`layout`, `guide`, `builder`, `practice`).
- `examples/intro-graph/`: Sample Quest knowledge graph for local testing.

## Running Checks Locally

You can validate backend logic and frontend types without launching the desktop window:

```bash
# Run Go unit and integration tests
go test -v ./...

# Run Svelte/TypeScript type checking and production bundle build
cd frontend && npm ci && npm run check && npm run build
```

To run the full desktop app with live reload:

```bash
wails dev
```

## Pull Request Guidelines

1. Keep domain logic in `internal/` (Go) or `frontend/src/stores/` (TypeScript) so it stays unit-testable.
2. Add or update tests when changing parsing, validation, persistence, or grading behavior.
3. Ensure `go test ./...`, `npm run check`, and `npm run build` pass cleanly before submitting.
