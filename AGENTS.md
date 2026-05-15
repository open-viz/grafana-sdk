# AGENTS.md

This file provides guidance to coding agents (e.g. Claude Code, claude.ai/code) when working with code in this repository.

## Repository purpose

Go module `go.openviz.dev/grafana-sdk` — a small Grafana HTTP client used by other OpenViz components (`grafana-tools`, the operator/UI server) to push dashboards/datasources to a Grafana instance. Library only; produces no binary despite the `BIN` value in the Makefile.

## Architecture

- `sdk.go` — the entire client surface. `Client` wraps a `resty.Client`, holds an optional `AuthConfig` (Basic or Bearer), and exposes methods such as `SetDashboard`, `DeleteDashboardByUID`, `GetCurrentOrg`, `GetHealth`, `CreateDatasource`, `UpdateDatasource`, `DeleteDatasource`, plus an internal `do(ctx, method, url, body)` helper. Request/response types (`GrafanaDashboard`, `Datasource`, `Org`, `HealthResponse`, `GrafanaResponse`, etc.) are defined here too.
- `sdk_test.go` — integration-style tests that talk to a real Grafana endpoint via env vars and read fixtures from `testdata/`.
- `testdata/dashboard.yaml` — sample dashboard payload used by the tests.
- `hack/`, `Makefile` — standard AppsCode build harness; runs everything inside `ghcr.io/appscode/golang-dev`.
- `vendor/` — checked-in deps.

Single package, single file — when you add a new Grafana endpoint, add another method on `*Client` in `sdk.go` and a matching test in `sdk_test.go`.

## Common commands

All build/test/lint targets run inside the `ghcr.io/appscode/golang-dev` Docker image — Docker must be running.

- `make ci` — CI pipeline: `verify lint build test`. Run before opening a PR.
- `make test` — Go tests. Integration tests against Grafana require these env vars (see `sdk_test.go`):
  - `GRAFANA_URL`, `GRAFANA_USERNAME`, `GRAFANA_PASSWORD` (or `GRAFANA_BEARER_TOKEN`).
  Without them the tests skip the live calls.
- `make lint` — golangci-lint.
- `make build` — builds the (effectively no-op) library.
- `make fmt` — gofmt + goimports.
- `make verify` — `verify-gen verify-modules`; `go mod tidy && go mod vendor` must leave the tree clean.
- `make add-license` / `make check-license` — manage Apache-2.0 license headers.

Run a single Go test locally (requires a Go toolchain):

```
go test . -run TestName -v
```

## Conventions

- Module path is `go.openviz.dev/grafana-sdk` (vanity URL); imports must use that, not the GitHub URL (`open-viz/grafana-sdk`).
- License: Apache-2.0 — every Go file carries the standard AppsCode header (`make add-license` applies it).
- Sign off commits (`git commit -s`); contributions follow the project DCO.
- Vendor directory is checked in; keep `go mod tidy && go mod vendor` clean (enforced by `verify-modules`).
- Package name is `grafana_sdk` (underscore), to keep the directory name `grafana-sdk` consistent with the repo while satisfying Go identifier rules — leave it.
- Keep `Client` methods short and use the internal `do(ctx, method, url, body)` helper rather than calling `resty` directly, so auth wiring and error decoding stay in one place.
