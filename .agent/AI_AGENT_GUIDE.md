# tkr — AI Agent Development Guide

> This document is written **for AI coding agents** (Cursor, Claude Code, Copilot Workspace, etc.).  
> It provides the context, conventions, and precise instructions needed to contribute to this project without ambiguity.

---

## 0. Read This First

Before generating any code:

1. Read `FUNCTIONAL_SPEC.md` in full — it is the source of truth for all behaviour.
2. Read `API_PROVIDERS.md` for the data-source integration details.
3. Read `TASKS.md` to understand which task you are working on and what "done" means.
4. Check `go.mod` for the Go version and declared dependencies before importing anything new.

If a requirement in this guide conflicts with `FUNCTIONAL_SPEC.md`, **the spec wins**.

---

## 1. Project Conventions

### 1.1 Language & Version

- Go **1.25+**
- Use the standard library wherever possible. Only add a dependency if it provides significant value (see §1.4 for the approved dependency list).
- All code must pass `go vet ./...` and `golangci-lint run` with the project's `.golangci.yml` config.

### 1.2 Package Layout

```
cmd/              One file per top-level CLI command. Each file registers its
                  cobra.Command in an init() function, no logic beyond flag
                  parsing and calling into internal/.

internal/         All application logic. Not importable by external packages.
  config/         Config loading only. No business logic.
  db/             Database layer only. Returns domain types from pkg/models.
                  No business logic — no alert evaluation here.
  provider/       One sub-package per API provider.
                  Each implements the provider.Provider interface.
  alert/          Alert evaluation engine. Pure functions where possible.
  notifier/       Notification dispatch. One file per channel.
  scheduler/      Cron job setup. Thin wrapper around robfig/cron.
  display/        Terminal rendering. No logic, only formatting.

pkg/
  models/         Domain types only. No methods that perform I/O.
```

**Rule:** No package may import from a sibling package at the same level. Data flows upward: `models` → `db/provider/alert/notifier` → `scheduler` → `cmd`.

### 1.3 Error Handling

- **Never** `panic` in production code paths. Panics are only acceptable in `init()` functions for programmer errors (e.g. invalid regex literals).
- Wrap errors with `fmt.Errorf("context: %w", err)` to preserve the chain.
- Use the sentinel errors defined in `internal/apperrors/errors.go` (e.g. `apperrors.ErrTickerNotFound`). Do not invent new string-based error messages.
- Log errors at the call site with `zerolog`. Do not re-log an already-logged error.

### 1.4 Approved Dependencies

Do not add new dependencies without a compelling reason. The approved set is:

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/viper` | Config loading |
| `modernc.org/sqlite` | SQLite driver (pure Go, no CGO) |
| `github.com/go-resty/resty/v2` | HTTP client with retry |
| `github.com/robfig/cron/v3` | Cron scheduler |
| `github.com/gen2brain/beeep` | Desktop notifications |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/stretchr/testify` | Test assertions |
| `github.com/jmoiron/sqlx` | SQL helper (named queries) |

To add a new dependency: add it to `go.mod`, justify it in the PR description, and update this list.

### 1.5 Naming Conventions

- **Files:** `snake_case.go`
- **Types/Interfaces:** `PascalCase`
- **Functions/Methods:** `PascalCase` (exported), `camelCase` (unexported)
- **Constants:** `PascalCase` for exported, `camelCase` for unexported
- **Error variables:** `Err` prefix, e.g. `ErrTickerNotFound`
- **Interface names:** Use `-er` suffix only when it's natural (e.g. `Notifier`, `Provider`). Avoid `IProvider` or `ProviderInterface`.

### 1.6 Testing

- Every exported function in `internal/` must have a unit test.
- HTTP provider integrations must use `net/http/httptest` stubs — no real API calls in tests.
- Database tests use an in-memory SQLite instance (`:memory:`).
- Test files live alongside source files: `foo.go` → `foo_test.go`.
- Aim for >80% coverage in `internal/alert/` and `internal/provider/`.

---

## 2. Interface Contracts

### 2.1 `provider.Provider`

Every market data provider must implement this interface:

```go
// internal/provider/provider.go

type Provider interface {
    // ID returns the unique identifier for this provider (e.g. "finnhub", "stooq").
    ID() string

    // Supports returns true if this provider can fetch data for the given exchange.
    Supports(exchange models.ExchangeID) bool

    // Quote fetches the latest quote for the given ticker.
    // Returns ErrTickerNotFound if the symbol does not exist on this provider.
    // Returns ErrRateLimited if the API rate limit has been hit.
    Quote(ctx context.Context, ticker string) (models.Quote, error)

    // History fetches daily OHLCV data for the last `days` days.
    History(ctx context.Context, ticker string, days int) ([]models.OHLCV, error)
}
```

### 2.2 `notifier.Notifier`

```go
// internal/notifier/notifier.go

type Notifier interface {
    // ID returns the channel identifier (e.g. "terminal", "email", "webhook").
    ID() string

    // Send dispatches the alert event notification.
    // Implementations must be safe to call concurrently.
    Send(ctx context.Context, event models.AlertEvent) error
}
```

### 2.3 `db.Repository`

```go
// internal/db/repository.go

type Repository interface {
    // Stocks
    AddStock(ctx context.Context, s models.Stock) error
    RemoveStock(ctx context.Context, ticker string) error
    ListStocks(ctx context.Context) ([]models.Stock, error)
    GetStock(ctx context.Context, ticker string) (models.Stock, error)

    // Alert rules
    AddAlertRule(ctx context.Context, rule models.AlertRule) (int64, error)
    RemoveAlertRule(ctx context.Context, id int64) error
    ListAlertRules(ctx context.Context, ticker string) ([]models.AlertRule, error)
    UpdateAlertRule(ctx context.Context, rule models.AlertRule) error

    // Alert events
    AddAlertEvent(ctx context.Context, event models.AlertEvent) error
    ListAlertEvents(ctx context.Context, filter models.AlertEventFilter) ([]models.AlertEvent, error)

    // Quote history
    SaveQuote(ctx context.Context, q models.Quote) error
    GetRecentQuotes(ctx context.Context, ticker string, limit int) ([]models.Quote, error)
    PruneHistory(ctx context.Context, olderThan time.Duration) error
}
```

---

## 3. How to Implement a New Provider

Follow these steps exactly when adding a new market data provider:

1. **Create the package:** `internal/provider/{providerid}/`
2. **Create `client.go`** — holds the `Client` struct and `New(apiKey string) *Client` constructor.
3. **Implement `provider.Provider`** on `*Client`.
4. **Create `client_test.go`** — use `httptest.NewServer` to mock API responses for at least:
   - Successful quote fetch
   - Ticker not found (provider-specific error response)
   - Rate limit response (HTTP 429)
5. **Register the provider** in `internal/provider/router.go`'s `DefaultProviders()` function.
6. **Add config keys** to `config.example.yaml` under `providers.{providerid}`.
7. **Update `API_PROVIDERS.md`** if endpoint details have changed.

**Stooq provider notes (CSV parsing):**

```go
// Stooq returns CSV. Parse like this:
// Date,Open,High,Low,Close,Volume
// Use encoding/csv reader. Trim whitespace from headers.
// The most recent row is the last row; iterate in reverse for latest price.
// If Volume is "N/A", set it to 0 — Stooq omits volume for indices.
```

**Yahoo Finance provider notes:**

```go
// Yahoo Finance endpoint: https://query1.finance.yahoo.com/v8/finance/chart/{ticker}
// No API key required, but set a realistic User-Agent header:
// "Mozilla/5.0 (compatible; tkr/1.0)"
// Response is JSON. The price is at:
// result[0].meta.regularMarketPrice
// rate-limit behaviour: HTTP 429 or HTTP 200 with empty result — handle both.
```

---

## 4. How to Add a New CLI Command

1. Create `cmd/{commandname}.go`.
2. Define a `var {commandname}Cmd = &cobra.Command{...}` variable.
3. Register it in `cmd/root.go`'s `init()` function: `rootCmd.AddCommand({commandname}Cmd)`.
4. All flag definitions go in the command file's own `init()` function.
5. The `RunE` function must:
   - Load config via `config.Load()`
   - Open the DB via `db.Open(cfg.Database.Path)`
   - Call into `internal/` packages for logic
   - Return errors (do not `os.Exit` inside `RunE`)
6. Write a usage example in the command's `Example` field.

---

## 5. Alert Evaluation — Implementation Notes

The evaluator lives in `internal/alert/evaluator.go`. Key rules:

- The `Evaluate(rule models.AlertRule, quote models.Quote) bool` function must be **pure** — no I/O, no DB access.
- Moving average evaluation (`MA_CROSS`) requires historical data; pass it via a separate `EvaluateWithHistory` function that accepts `[]models.Quote`.
- Cooldown enforcement lives in the **scheduler**, not the evaluator. The evaluator only returns true/false.
- Write table-driven tests covering all condition types and edge cases (price == threshold boundary, NaN/zero values from bad provider data).

---

## 6. Database Migrations

- Migration files live in `migrations/` as numbered SQL files: `001_initial_schema.sql`, `002_add_alert_cooldown.sql`, etc.
- Use `golang-migrate` or a hand-rolled runner in `internal/db/migrate.go`.
- Migrations run automatically on `tkr init` and on daemon startup.
- **Never modify existing migration files.** Always add a new file.
- Migration runner must be idempotent (track applied migrations in a `schema_migrations` table).

---

## 7. Logging Guidelines

Use `zerolog` throughout. The root logger is initialised in `cmd/root.go` and passed via context.

```go
// Get logger from context
log := zerolog.Ctx(ctx)

// Correct usage levels:
log.Debug().Str("ticker", ticker).Msg("fetching quote")
log.Info().Str("provider", p.ID()).Dur("elapsed", elapsed).Msg("quote fetched")
log.Warn().Err(err).Str("provider", p.ID()).Msg("provider failed, trying next")
log.Error().Err(err).Msg("all providers failed")
```

- `Debug` — internal flow, per-request details.
- `Info` — daemon lifecycle events, successful operations.
- `Warn` — recoverable errors (provider fallback, stale data).
- `Error` — unrecoverable errors that affect user-visible output.

---

## 8. Common Pitfalls to Avoid

- **Do not store `float64` for currency in the DB.** Always store as `REAL` (SQLite) and be aware of floating point drift when comparing prices. Use a tolerance of 0.0001 when checking equality.
- **Do not call `time.Now()` inside `internal/alert/evaluator.go`.** Inject time as a parameter so tests are deterministic.
- **Do not hard-code exchange market hours.** Load them from `internal/config/market_hours.go` which reads from `data/market_hours.yaml`.
- **Do not use `os.Exit` anywhere except `main.go`.** Return errors up the call stack.
- **Do not log and return an error.** Log OR return, not both (the caller logs at the right level).
- **Yahoo Finance URL must use `query1.finance.yahoo.com`**, not `finance.yahoo.com`. The latter is a redirect and adds latency.
- **Stooq tickers for GPW use `.PL` suffix**, but Yahoo Finance uses `.WA` for the same stocks. The router handles this mapping — see `internal/provider/ticker_mapper.go`.

---

## 9. Definition of Done (Per Task)

A task from `TASKS.md` is complete when:

- [ ] All described behaviour is implemented.
- [ ] Unit tests are written and pass (`go test ./...`).
- [ ] `go vet ./...` passes with zero warnings.
- [ ] No new linting errors introduced.
- [ ] `go build ./...` succeeds.
- [ ] `FUNCTIONAL_SPEC.md` is updated if the implementation deviated from the spec (with a note explaining why).
- [ ] `TASKS.md` task is marked `[x]` done.
