---
name: testing
description: >
  Use this skill whenever writing, reviewing, or improving tests in tkr.
  Triggers include: writing unit tests for any internal package, creating httptest
  mock servers for provider tests, writing integration tests against an in-memory
  SQLite database, testing CLI command output and exit codes, setting up test
  helpers or fixtures, measuring coverage, or debugging a failing test. Always read
  go-conventions/SKILL.md first.
---

# tkr — Testing

> Prerequisite: **`.agent/skills/go-conventions/SKILL.md`** must be read before this skill.
> For domain-specific test patterns, also read the skill for the subsystem under test.

---

## 1. Testing Principles

1. **No real I/O in unit tests.** No real API calls, no real files, no real clock.
2. **Inject dependencies.** If a function calls `time.Now()`, it cannot be tested deterministically — add a `now time.Time` parameter.
3. **Test files alongside source.** `foo.go` → `foo_test.go` in the same package.
4. **Table-driven tests** for anything with multiple input/output cases.
5. **Coverage targets:** >80% in `internal/alert/` and `internal/provider/`. Use `make test` which runs `go test -cover ./...`.

---

## 2. File Structure

```go
package finnhub  // same package as the code under test (white-box)
// OR
package finnhub_test  // external package (black-box, preferred for public API)

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

Use `require` for assertions that should stop the test on failure (setup, first-step errors).
Use `assert` for assertions that can continue accumulating failures.

```go
// require — stops immediately on failure
repo, err := db.Open(":memory:")
require.NoError(t, err)  // no point continuing if DB didn't open

// assert — continues to next assertion
assert.Equal(t, "AAPL", got.Ticker)
assert.Equal(t, 168.42, got.Price)
```

---

## 3. Table-Driven Tests

The standard pattern for any function with multiple cases:

```go
func TestParseCondition(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    models.Condition
        wantErr bool
    }{
        {
            name:  "price less than",
            input: "price < 150",
            want:  models.Condition{Metric: models.MetricPrice, Operator: models.OperatorLT, Value: 150},
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tc := range tests {
        tc := tc  // capture loop variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()  // add when tests are independent and have no shared state

            got, err := alert.ParseCondition(tc.input)

            if tc.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tc.want, got)
        })
    }
}
```

---

## 4. Mocking HTTP Providers with `httptest`

Never call real APIs. Use `httptest.NewServer` to return canned responses.

```go
func TestQuote_Success(t *testing.T) {
    // 1. Define the mock response body
    mockBody := `{
        "c": 168.42,
        "d": -3.21,
        "dp": -1.87,
        "h": 172.00,
        "l": 167.50,
        "o": 171.00,
        "pc": 171.63,
        "t": 1713360000
    }`

    // 2. Start a local HTTP server
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/quote", r.URL.Path)
        assert.Equal(t, "AAPL", r.URL.Query().Get("symbol"))
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, mockBody)
    }))
    defer srv.Close()

    // 3. Point the client at the test server
    client := finnhub.NewWithBaseURL("test-key", srv.URL)

    // 4. Call and assert
    got, err := client.Quote(context.Background(), "AAPL")
    require.NoError(t, err)
    assert.Equal(t, "AAPL", got.Ticker)
    assert.InDelta(t, 168.42, got.Price, 0.001)
    assert.Equal(t, "finnhub", got.Source)
}
```

### Provider must accept a base URL override

Every provider's constructor must have a `NewWithBaseURL(apiKey, baseURL string)` variant
(or accept the URL as a parameter on `New`) so tests can inject the mock server address.

```go
// In client.go:
func New(apiKey string) *Client {
    return NewWithBaseURL(apiKey, "https://finnhub.io/api/v1")
}

func NewWithBaseURL(apiKey, baseURL string) *Client {
    return &Client{
        apiKey: apiKey,
        http:   resty.New().SetBaseURL(baseURL),
    }
}
```

### Required test cases per provider

| Test | Setup | Expected result |
|---|---|---|
| `TestQuote_Success` | HTTP 200 + valid body | Correctly populated `models.Quote` |
| `TestQuote_TickerNotFound` | HTTP 200 + empty/zero body (provider-specific) | `apperrors.ErrTickerNotFound` |
| `TestQuote_RateLimited` | HTTP 429 | `apperrors.ErrRateLimited` |
| `TestQuote_NetworkError` | `srv.Close()` before call | Wrapped error, not panic |
| `TestHistory_Success` | HTTP 200 + multi-row body | Correct `[]models.OHLCV` slice |
| `TestSupports` | No network needed | Boolean per exchange |

---

## 5. Database Tests

Use `:memory:` SQLite — never a file path in tests.

```go
// test helpers — put in internal/db/testhelpers_test.go

func newTestRepo(t *testing.T) db.Repository {
    t.Helper()
    repo, err := db.Open(":memory:")
    require.NoError(t, err)
    t.Cleanup(func() { repo.Close() })
    return repo
}

func seedStock(t *testing.T, repo db.Repository, ticker string) models.Stock {
    t.Helper()
    s := models.Stock{
        Ticker:   ticker,
        Name:     ticker + " Inc.",
        Exchange: models.ExchangeNASDAQ,
        Currency: "USD",
        AddedAt:  time.Now().UTC(),
    }
    require.NoError(t, repo.AddStock(context.Background(), s))
    return s
}
```

Test flow:

```go
func TestGetStock_NotFound(t *testing.T) {
    repo := newTestRepo(t)

    _, err := repo.GetStock(context.Background(), "NONEXISTENT")

    require.Error(t, err)
    assert.ErrorIs(t, err, apperrors.ErrTickerNotFound)
}

func TestAddStock_Idempotent(t *testing.T) {
    repo := newTestRepo(t)
    s := seedStock(t, repo, "AAPL")

    // Adding the same stock again must not error
    err := repo.AddStock(context.Background(), s)
    require.NoError(t, err)

    // And must not duplicate
    stocks, err := repo.ListStocks(context.Background())
    require.NoError(t, err)
    assert.Len(t, stocks, 1)
}
```

---

## 6. Testing with a Fake Clock

Any function that checks time must accept `now time.Time` as a parameter.

```go
// Production call-site
triggered := scheduler.shouldEvaluate(rule, time.Now())

// Test
func TestShouldEvaluate_CooldownNotExpired(t *testing.T) {
    firedAt := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
    rule := models.AlertRule{
        Active:          true,
        LastFired:       &firedAt,
        CooldownSeconds: 3600,
    }

    // 30 minutes later — cooldown not expired
    now := firedAt.Add(30 * time.Minute)
    assert.False(t, scheduler.ShouldEvaluate(rule, now))

    // 61 minutes later — cooldown expired
    now = firedAt.Add(61 * time.Minute)
    assert.True(t, scheduler.ShouldEvaluate(rule, now))
}
```

---

## 7. CLI Command Tests

Test the command's output and exit code by running it against a fake repo and provider.

```go
func TestWatchListCmd_Empty(t *testing.T) {
    // Build a Cobra command tree backed by a test repo
    root := cmd.NewRootWithDeps(fakeRepo, fakeRouter)

    out := &bytes.Buffer{}
    root.SetOut(out)
    root.SetArgs([]string{"watch", "list"})

    err := root.Execute()

    require.NoError(t, err)
    assert.Contains(t, out.String(), "No stocks in watchlist")
}
```

This requires `cmd` to support dependency injection. Wire it up via a constructor:

```go
// cmd/root.go
func NewRootWithDeps(repo db.Repository, router provider.Router) *cobra.Command {
    // ... builds rootCmd with the provided deps
}

// main.go uses the default constructor
cmd.Execute()
```

---

## 8. Test Helpers Shared Across Packages

Put shared test utilities in `internal/testutil/`:

```go
// internal/testutil/fixtures.go

func StockFixture(ticker string) models.Stock { ... }
func QuoteFixture(ticker string, price float64) models.Quote { ... }
func AlertRuleFixture(ticker string, expr string) models.AlertRule { ... }
```

Import only in `_test.go` files.

---

## 9. Running Tests

```bash
# All tests
make test

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Single package
go test ./internal/alert/...

# Single test by name
go test ./internal/alert/... -run TestParseCondition

# Race detector (run before each release)
go test -race ./...
```

---

## 10. Definition of Done (Testing)

- [ ] Every exported function in `internal/` has at least one test
- [ ] Providers: all 6 required cases covered with `httptest`
- [ ] Database methods: all tested with `:memory:` SQLite
- [ ] No real API calls, no real file paths, no `time.Sleep` in tests
- [ ] Table-driven tests used for all multi-case functions
- [ ] `go test -race ./...` passes clean
- [ ] Coverage ≥80% in `internal/alert/` and `internal/provider/`
