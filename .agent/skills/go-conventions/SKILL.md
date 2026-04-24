---
name: go-conventions
description: >
  The foundational coding standard for the tkr Go project. Use this skill
  whenever writing, reviewing, or modifying ANY Go code in this codebase — before
  implementing a feature, adding a dependency, handling an error, or structuring a
  new package. This skill must be read before any other tkr skill is applied.
  It defines the rules that all other skills build on: package layout, error handling,
  logging, naming, and dependency governance.
---

# tkr — Go Conventions

This is the **root skill**. Every other skill in this project inherits from it.
Read this first. Always.

---

## 1. Package Layout — The One Rule

Data flows **upward only**. No sideways imports between sibling packages.

```
pkg/models          ← domain types, no I/O, no imports from internal/
    ↑
internal/db         internal/provider   internal/alert   internal/notifier
    ↑               ↑                   ↑                ↑
internal/scheduler  (orchestrates all of the above)
    ↑
cmd/                (Cobra commands — thin wrappers, no logic)
```

**Forbidden patterns:**
- `internal/alert` importing `internal/db` directly — pass data via function arguments
- `cmd/` containing business logic — delegate everything to `internal/`
- `pkg/models` importing anything from `internal/`

When unsure whether an import is valid, trace the arrow. If it goes sideways or downward, it is wrong.

---

## 2. Error Handling

### Use sentinel errors from `internal/apperrors`

```go
// CORRECT
return fmt.Errorf("quote fetch: %w", apperrors.ErrTickerNotFound)

// WRONG — never invent ad-hoc error strings
return fmt.Errorf("ticker not found")
```

Full list of sentinel errors is in `internal/apperrors/errors.go`. If you need a new one, add it there — never inline.

### Wrap with context, preserve the chain

```go
// CORRECT
if err != nil {
    return fmt.Errorf("router.Quote %s: %w", ticker, err)
}

// WRONG — swallows context
return err
```

### Never panic in production code paths

Panic is only allowed in `init()` for programmer errors (malformed regex literals, etc.).

### Log OR return — never both

```go
// WRONG — double-logging
log.Error().Err(err).Msg("failed")
return fmt.Errorf("...: %w", err)

// CORRECT — log at the boundary where you handle the error
if err != nil {
    return fmt.Errorf("...: %w", err)   // caller logs
}
// OR if this IS the final handler:
log.Error().Err(err).Msg("failed to fetch quote")
// do not return the error further up
```

### Never call `os.Exit` outside `main.go`

Return errors up the stack. Cobra's `RunE` handles the exit.

---

## 3. Logging — zerolog

The root logger is initialised in `cmd/root.go` and propagated via `context.Context`.

```go
import "github.com/rs/zerolog"

// Get logger from context
log := zerolog.Ctx(ctx)

// Correct levels:
log.Debug().Str("ticker", ticker).Msg("fetching quote")         // internal flow
log.Info().Str("provider", p.ID()).Dur("elapsed", e).Msg("ok")  // lifecycle events
log.Warn().Err(err).Str("provider", p.ID()).Msg("fallback")     // recoverable
log.Error().Err(err).Msg("all providers failed")                // unrecoverable
```

**Rules:**
- Always attach structured fields (`.Str`, `.Int`, `.Err`) — never interpolate into the message string.
- Never use `fmt.Println` or `log` (stdlib) for application output.
- CLI display output (tables, sparklines) goes through `internal/display`, not the logger.

---

## 4. Naming Conventions

| Thing | Style | Example |
|---|---|---|
| Files | `snake_case.go` | `ticker_mapper.go` |
| Types & Interfaces | `PascalCase` | `AlertRule`, `Provider` |
| Exported functions | `PascalCase` | `Evaluate(...)` |
| Unexported functions | `camelCase` | `parseCondition(...)` |
| Constants (exported) | `PascalCase` | `ExchangeNYSE` |
| Error variables | `Err` prefix | `ErrTickerNotFound` |
| Interface names | Natural `-er` suffix | `Notifier`, `Provider` |

**Anti-patterns:**
- ❌ `IProvider`, `ProviderInterface` — no Hungarian notation
- ❌ `GetStock` — drop `Get` for simple accessors, prefer `Stock(...)`
- ❌ `data`, `info`, `stuff` as variable names — name the thing

---

## 5. Dependency Governance

### Approved dependencies (do not add others without justification)

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI |
| `github.com/spf13/viper` | Config |
| `modernc.org/sqlite` | SQLite (pure Go, no CGO) |
| `github.com/go-resty/resty/v2` | HTTP client |
| `github.com/robfig/cron/v3` | Scheduler |
| `github.com/gen2brain/beeep` | Desktop notifications |
| `github.com/rs/zerolog` | Logging |
| `github.com/stretchr/testify` | Test assertions |
| `github.com/jmoiron/sqlx` | SQL helpers |

### Adding a new dependency

1. Check if the standard library already covers the use case (`net/http`, `encoding/csv`, `time`, `sync`, etc.).
2. If not, justify it in a comment at the top of the file that uses it.
3. Add to `go.mod` and update this table.

---

## 6. Time Handling

- **Never call `time.Now()` inside `internal/alert/`**. Pass time as a parameter so tests are deterministic.
- All times stored in SQLite use `DATETIME` (UTC). Always call `.UTC()` before persisting.
- Market hours logic reads from `data/market_hours.yaml`, never hardcoded.

---

## 7. Float64 for Prices

- Store prices as `REAL` in SQLite. Be aware of drift.
- Use a tolerance of `0.0001` when comparing prices for equality — never `==`.

```go
func pricesEqual(a, b float64) bool {
    return math.Abs(a-b) < 0.0001
}
```

---

## 8. Concurrency

- Notifier dispatch is concurrent — implementations of `notifier.Notifier` **must** be safe for concurrent calls.
- Use `sync.Mutex` for shared state in the provider router's rate-limit map and circuit breaker.
- Prefer channels for signalling shutdown (daemon stop signal) over shared boolean flags.

---

## 9. Code Quality Gates

Before marking any task done, run:

```bash
make lint    # golangci-lint with project config
make test    # go test ./...
make build   # go build ./...
```

All three must pass clean. Zero new linting warnings are acceptable.

---

## Related Skills

- Adding a provider → read **`.agent/skills/provider-integration/SKILL.md`**
- Adding a CLI command → read **`.agent/skills/cli-command/SKILL.md`**
- Working with the database → read **`.agent/skills/database-layer/SKILL.md`**
- Alert rule logic → read **`.agent/skills/alert-engine/SKILL.md`**
- Writing tests → read **`.agent/skills/testing/SKILL.md`**
