# tkr — Project Tasks

> Track progress here. Mark tasks `[x]` when done.  
> Each task references the relevant section of `FUNCTIONAL_SPEC.md` and `AI_AGENT_GUIDE.md`.

---

## How to Use This File

- Tasks are grouped into milestones.
- Each milestone produces a working, testable increment of the application.
- Within a milestone, tasks should be done roughly in order (dependencies noted where relevant).
- Estimated effort is marked: 🟢 small (< 2h) · 🟡 medium (2–6h) · 🔴 large (> 6h)

---

## Milestone 0 — Project Skeleton

> Goal: `go build ./...` succeeds. All packages exist. No logic yet.

- [x] 🟢 **M0-1** Initialise `go.mod` with module path `github.com/yourname/tkr`, Go 1.22
- [x] 🟢 **M0-2** Add all approved dependencies to `go.mod` and `go.sum` (`cobra`, `viper`, `resty`, `cron`, `zerolog`, `testify`, `sqlx`, `sqlite`, `beeep`)
- [x] 🟢 **M0-3** Create directory tree: `cmd/`, `internal/config`, `internal/db`, `internal/provider`, `internal/alert`, `internal/notifier`, `internal/scheduler`, `internal/display`, `internal/apperrors`, `pkg/models`, `migrations/`, `data/`
- [x] 🟢 **M0-4** Create `main.go` that calls `cmd.Execute()`
- [x] 🟢 **M0-5** Create `cmd/root.go` with root Cobra command, zerolog initialisation, and global `--config` / `--log-level` flags
- [x] 🟢 **M0-6** Create `pkg/models/` with all domain types: `Stock`, `Quote`, `OHLCV`, `AlertRule`, `AlertEvent`, `AlertEventFilter`, `ExchangeID`, `ProviderID`, `ChannelID`, `Condition`, `Metric`, `Operator` (spec §1)
- [x] 🟢 **M0-7** Create `internal/apperrors/errors.go` with all sentinel error variables (spec §5)
- [x] 🟢 **M0-8** Create stub files for each interface: `internal/provider/provider.go`, `internal/notifier/notifier.go`, `internal/db/repository.go` (spec §2 of AI guide)
- [x] 🟢 **M0-9** Create `.golangci.yml` with linting rules (enable `errcheck`, `govet`, `staticcheck`, `unused`)
- [x] 🟢 **M0-10** Create `Makefile` with targets: `build`, `test`, `lint`, `run`, `install`

---

## Milestone 1 — Config & Database

> Goal: `tkr init` works. Config and DB are created on disk.

- [ ] 🟡 **M1-1** Implement `internal/config/config.go`: Viper-based loader, maps to a `Config` struct covering all keys in spec §6, handles `~` expansion in paths
- [ ] 🟢 **M1-2** Create `config.example.yaml` with all keys documented inline
- [ ] 🟢 **M1-3** Create `data/market_hours.yaml` with open/close UTC times for all 5 exchanges (spec §3.2)
- [ ] 🟡 **M1-4** Implement `internal/db/migrate.go`: reads `*.sql` files from `migrations/`, tracks applied migrations in `schema_migrations` table, idempotent
- [ ] 🟡 **M1-5** Write `migrations/001_initial_schema.sql` with all tables from spec §4
- [ ] 🟡 **M1-6** Implement `internal/db/sqlite.go`: SQLite connection, implements `Repository` interface (all methods — can return `ErrNotImplemented` stubs initially)
- [ ] 🟡 **M1-7** Implement `cmd/init.go`: creates config file from example template, creates DB directory, runs migrations, prints next-step hints (spec §2.1)
- [ ] 🟢 **M1-8** Write unit tests for `internal/config/` (load from file, env override, path expansion)
- [ ] 🟢 **M1-9** Write unit tests for `internal/db/` using `:memory:` SQLite (insert + query for each table)

---

## Milestone 2 — Provider Integrations

> Goal: `tkr quote AAPL CDR.WAR` returns live prices.

- [ ] 🟢 **M2-1** Create `internal/provider/ticker_mapper.go`: maps canonical `{TICKER}.{EXCHANGE}` format to provider-specific ticker strings (e.g. `CDR.WAR` → `CDR.PL` for Stooq, `CDR.WA` for Yahoo)
- [ ] 🟡 **M2-2** Implement `internal/provider/finnhub/client.go`: `Quote` and `History` methods, HTTP 429 → `ErrRateLimited`, 404-equivalent → `ErrTickerNotFound` (spec §1 of API_PROVIDERS.md)
- [ ] 🟡 **M2-3** Write tests for Finnhub provider using `httptest.NewServer` (success, not found, rate limited)
- [ ] 🟡 **M2-4** Implement `internal/provider/stooq/client.go`: CSV parsing, GPW support, handle `N/A` volume, return last row as current quote (spec §3 of API_PROVIDERS.md)
- [ ] 🟡 **M2-5** Write tests for Stooq provider (valid CSV, empty response, malformed CSV)
- [ ] 🟡 **M2-6** Implement `internal/provider/yahoofinance/client.go`: JSON parsing of v8 chart endpoint, User-Agent header, handle HTTP 429 and empty result (spec §4 of API_PROVIDERS.md)
- [ ] 🟡 **M2-7** Write tests for Yahoo Finance provider
- [ ] 🔴 **M2-8** Implement `internal/provider/router.go`: `Router` struct, `DefaultProviders()`, routing logic with fallback, in-memory quote cache (TTL = half polling interval), circuit breaker, rate-limit backoff (spec §3.1 of FUNCTIONAL_SPEC.md)
- [ ] 🟡 **M2-9** Write tests for the router (fallback behaviour, cache hit, circuit breaker open)
- [ ] 🟡 **M2-10** Implement `cmd/quote.go`: fetch quotes via router, render table with colour, `--history N` sparkline, `--json` flag (spec §2.3)
- [ ] 🟢 **M2-11** Implement `internal/display/table.go`: coloured table renderer for quotes (green/red/yellow rows)
- [ ] 🟢 **M2-12** Implement `internal/display/sparkline.go`: ASCII sparkline using Unicode block chars, terminal-width aware (spec §2.3)

---

## Milestone 3 — Watchlist Management

> Goal: `tkr watch add/remove/list` works end-to-end.

- [ ] 🟡 **M3-1** Implement `db.Repository` methods fully: `AddStock`, `RemoveStock`, `ListStocks`, `GetStock` (replace stubs from M1-6)
- [ ] 🟡 **M3-2** Implement `cmd/watch.go` with sub-commands `add`, `remove`, `list` (spec §2.2)
  - `add`: resolve ticker via provider, insert to DB, print confirmation
  - `remove`: remove from DB, deactivate alerts, warn user; `--purge-alerts` flag
  - `list`: fetch quotes for all watched stocks, render table, `--format` flag
- [ ] 🟢 **M3-3** Handle exchange disambiguation: when `--exchange` flag is missing and ticker is ambiguous, prompt the user with a numbered list
- [ ] 🟢 **M3-4** Write integration tests for `watch add` / `watch list` using in-memory DB and mock provider

---

## Milestone 4 — Alert Rules

> Goal: Users can add, list, enable, disable, and remove alert rules via CLI.

- [ ] 🟡 **M4-1** Implement `internal/alert/parser.go`: parses condition expression strings into `models.Condition` structs (spec §1.5), returns `ErrInvalidCondition` with a clear message on bad syntax
- [ ] 🟢 **M4-2** Write table-driven tests for the condition parser covering all supported expressions and error cases
- [ ] 🟡 **M4-3** Implement `db.Repository` alert methods: `AddAlertRule`, `RemoveAlertRule`, `ListAlertRules`, `UpdateAlertRule`, `AddAlertEvent`, `ListAlertEvents`
- [ ] 🟡 **M4-4** Implement `cmd/alert.go` with sub-commands: `add`, `list`, `remove`, `enable`, `disable`, `history` (spec §2.4)
- [ ] 🟢 **M4-5** Write unit tests for alert DB methods

---

## Milestone 5 — Alert Evaluation Engine

> Goal: Given a quote, the engine correctly evaluates all condition types.

- [ ] 🟡 **M5-1** Implement `internal/alert/evaluator.go`: pure `Evaluate(rule, quote)` and `EvaluateWithHistory(rule, quote, history)` functions (spec §3.3, AI guide §5)
- [ ] 🔴 **M5-2** Write comprehensive table-driven tests for the evaluator:
  - All 5 metric types (PRICE, CHANGE_ABS, CHANGE_PCT, VOLUME, MA_CROSS)
  - All operators (LT, LTE, GT, GTE, EQ, CROSS_ABOVE, CROSS_BELOW)
  - Edge cases: value at exact threshold, zero values, insufficient history for MA
- [ ] 🟡 **M5-3** Implement moving average calculation helper in `internal/alert/ma.go`
- [ ] 🟢 **M5-4** Write tests for the MA calculator

---

## Milestone 6 — Notification Channels

> Goal: All three notification channels can dispatch an `AlertEvent`.

- [ ] 🟡 **M6-1** Implement `internal/notifier/terminal.go`: writes to stdout + fires desktop notification via `beeep` (spec §3.4)
- [ ] 🟡 **M6-2** Implement `internal/notifier/email.go`: SMTP with TLS, plain-text + HTML body, retry logic (spec §3.4)
- [ ] 🟢 **M6-3** Create `internal/notifier/templates/alert_email.html` — simple HTML email template
- [ ] 🟡 **M6-4** Implement `internal/notifier/webhook.go`: POST JSON payload, Slack/Discord compatible, exponential backoff retry (spec §3.4)
- [ ] 🟢 **M6-5** Implement `internal/notifier/dispatcher.go`: `Dispatcher` struct that holds multiple `Notifier` instances, dispatches concurrently, collects errors
- [ ] 🟢 **M6-6** Write tests for webhook notifier using `httptest.NewServer`
- [ ] 🟢 **M6-7** Write tests for dispatcher (partial failure handling)

---

## Milestone 7 — Daemon & Scheduler

> Goal: `tkr daemon start` runs in the background and fires alerts.

- [ ] 🔴 **M7-1** Implement `internal/scheduler/scheduler.go`: cron job, fetches quotes for all watched stocks, calls evaluator for each active rule, dispatches triggered alerts, saves quote history, prunes old history (spec §3.2)
- [ ] 🟡 **M7-2** Implement market-hours gate in scheduler: load exchange hours from config, skip polling if `market_hours_only` is true and all watched exchanges are closed
- [ ] 🔴 **M7-3** Implement `cmd/daemon.go`: `start` (fork to background, write PID file), `stop` (SIGTERM + SIGKILL fallback), `status`, `restart` (spec §2.5)
- [ ] 🟢 **M7-4** Write unit tests for market-hours gate logic (inject fake `time.Now`)
- [ ] 🟢 **M7-5** Write integration test: full scheduler tick with mock provider + in-memory DB verifying that an alert event is saved when a condition is met

---

## Milestone 8 — Config Management Commands

> Goal: `tkr config show/set/validate` works.

- [ ] 🟢 **M8-1** Implement `cmd/config.go` with sub-commands `show`, `set`, `validate` (spec §2.6)
- [ ] 🟡 **M8-2** Implement `config validate`: pings each enabled provider with a test request, reports per-provider status with colour coding
- [ ] 🟢 **M8-3** Ensure API keys are masked in `config show` output (replace with `****`)

---

## Milestone 9 — Polish & Release Prep

> Goal: Production-ready v0.1.0.

- [ ] 🟡 **M9-1** Add EODHD provider (`internal/provider/eodhd/client.go`) as secondary GPW source (API_PROVIDERS.md §5)
- [ ] 🟢 **M9-2** Write `migrations/002_add_indexes.sql` for performance (already included in §4 of spec, verify and add any missing indexes)
- [ ] 🟡 **M9-3** Add `--quiet` and `--no-colour` global flags; respect `NO_COLOR` env var
- [ ] 🟢 **M9-4** Add shell completion support via Cobra: `tkr completion bash/zsh/fish`
- [ ] 🟢 **M9-5** Write `CONTRIBUTING.md`
- [ ] 🟢 **M9-6** Write `CHANGELOG.md` for v0.1.0
- [ ] 🟢 **M9-7** Create `.github/workflows/ci.yml`: run `go test ./...` + `golangci-lint` on push and PR
- [ ] 🟢 **M9-8** Create `Dockerfile` (multi-stage build, scratch base image, runs as non-root)
- [ ] 🟡 **M9-9** End-to-end smoke test: script that runs `init`, `watch add`, `alert add`, `quote`, `daemon start`, waits, `daemon stop`, `alert history` — verifies exit codes and output patterns
- [ ] 🟢 **M9-10** Tag `v0.1.0` and publish release

---

## Backlog (Future Milestones)

These are not part of v0.1.0 but should be tracked for later.

- [ ] **BL-1** Alpha Vantage provider integration
- [ ] **BL-2** IEX Cloud provider integration
- [ ] **BL-3** Polygon.io provider integration
- [ ] **BL-4** `alert add` support for `ma_cross` conditions via CLI
- [ ] **BL-5** Telegram notification channel
- [ ] **BL-6** Pushover notification channel
- [ ] **BL-7** `tkr report` command: weekly portfolio summary email
- [ ] **BL-8** Import watchlist from CSV
- [ ] **BL-9** `tkr screen` command: scan all watched stocks against a condition without saving a rule
- [ ] **BL-10** systemd unit file generator (`tkr daemon install-systemd`)
- [ ] **BL-11** Support for ETFs and currency pairs (EUR/PLN, etc.)
- [ ] **BL-12** HTTPS certificate pinning for all provider requests

---

## Task Summary

| Milestone | Tasks | Estimated Effort |
|---|---|---|
| M0 — Skeleton | 10 | ~4h |
| M1 — Config & DB | 9 | ~8h |
| M2 — Providers | 12 | ~14h |
| M3 — Watchlist | 4 | ~6h |
| M4 — Alert Rules | 5 | ~7h |
| M5 — Evaluator | 4 | ~8h |
| M6 — Notifications | 7 | ~8h |
| M7 — Daemon | 5 | ~12h |
| M8 — Config cmds | 3 | ~3h |
| M9 — Polish | 10 | ~8h |
| **Total** | **69** | **~78h** |
