---
name: provider-integration
description: >
  Use this skill whenever adding, modifying, or debugging a market data provider in
  tkr. Triggers include: implementing a new API integration (Finnhub, Stooq,
  Yahoo Finance, EODHD, Alpha Vantage, Polygon, etc.), changing how a provider handles
  rate limits or errors, updating ticker format mappings, modifying the provider router
  fallback logic, or adding circuit breaker / caching behaviour. Always read
  go-conventions/SKILL.md first.
---

# tkr — Provider Integration

> Prerequisite: **`.agent/skills/go-conventions/SKILL.md`** must be read before this skill.
> For the API details of each provider (endpoints, response shapes, rate limits), see **`.agent/API_PROVIDERS.md`**.

---

## 1. The Provider Interface

Every provider implements exactly this interface, defined in `internal/provider/provider.go`:

```go
type Provider interface {
    ID() string
    Supports(exchange models.ExchangeID) bool
    Quote(ctx context.Context, ticker string) (models.Quote, error)
    History(ctx context.Context, ticker string, days int) ([]models.OHLCV, error)
}
```

No other methods. Do not add provider-specific methods to the interface — keep them unexported on the concrete type.

---

## 2. Checklist for a New Provider

Work through these steps in order. Do not skip steps.

### Step 1 — Create the package

```
internal/provider/{providerid}/
├── client.go        ← Client struct, New(), all interface methods
└── client_test.go   ← httptest-based unit tests
```

`{providerid}` is lowercase, no hyphens (e.g. `finnhub`, `stooq`, `yahoofinance`, `eodhd`).

### Step 2 — Write `client.go`

```go
package finnhub  // package name matches directory

import (
    "context"
    "github.com/go-resty/resty/v2"
    "github.com/yourname/tkr/internal/apperrors"
    "github.com/yourname/tkr/pkg/models"
)

const providerID = "finnhub"

type Client struct {
    apiKey string
    http   *resty.Client
}

func New(apiKey string) *Client {
    return &Client{
        apiKey: apiKey,
        http:   resty.New().SetBaseURL("https://finnhub.io/api/v1"),
    }
}

func (c *Client) ID() string { return providerID }

func (c *Client) Supports(exchange models.ExchangeID) bool {
    supported := map[models.ExchangeID]bool{
        models.ExchangeNYSE:    true,
        models.ExchangeNASDAQ:  true,
        models.ExchangeXETRA:   true,
        models.ExchangeLSE:     true,
        models.ExchangeEuronext: true,
    }
    return supported[exchange]
}
```

### Step 3 — Implement `Quote`

**Required error mapping — use these exactly:**

| HTTP / provider signal | Return |
|---|---|
| HTTP 429 | `apperrors.ErrRateLimited` |
| Ticker not found (varies by provider) | `apperrors.ErrTickerNotFound` |
| Network failure | `fmt.Errorf("finnhub.Quote: %w", err)` |
| Unexpected status | `fmt.Errorf("finnhub.Quote: unexpected status %d", resp.StatusCode())` |

**Map the raw response to `models.Quote`:**

```go
func (c *Client) Quote(ctx context.Context, ticker string) (models.Quote, error) {
    var raw finnhubQuoteResponse

    resp, err := c.http.R().
        SetContext(ctx).
        SetQueryParam("symbol", ticker).
        SetQueryParam("token", c.apiKey).
        SetResult(&raw).
        Get("/quote")

    if err != nil {
        return models.Quote{}, fmt.Errorf("finnhub.Quote: %w", err)
    }
    if resp.StatusCode() == 429 {
        return models.Quote{}, apperrors.ErrRateLimited
    }
    if resp.StatusCode() != 200 {
        return models.Quote{}, fmt.Errorf("finnhub.Quote: status %d", resp.StatusCode())
    }
    // Finnhub returns price=0 for unknown tickers
    if raw.CurrentPrice == 0 {
        return models.Quote{}, apperrors.ErrTickerNotFound
    }

    return models.Quote{
        Ticker:    ticker,
        Price:     raw.CurrentPrice,
        Open:      raw.Open,
        High:      raw.High,
        Low:       raw.Low,
        Close:     raw.PreviousClose,
        Change:    raw.Change,
        ChangePct: raw.ChangePercent,
        Timestamp: time.Unix(raw.Timestamp, 0).UTC(),
        Source:    providerID,
    }, nil
}
```

### Step 4 — Stooq-specific: CSV parsing

Stooq returns CSV, not JSON. Key rules:

```go
// Use encoding/csv — do not split strings manually.
// Headers: Date,Open,High,Low,Close,Volume
// Most recent data is the LAST row — iterate to end.
// Volume may be "N/A" for indices — set to 0, do not error.
// Ticker suffix: GPW = ".PL", e.g. "CDR.PL"
// An empty body (just headers, no rows) means ticker not found.
```

### Step 5 — Yahoo Finance-specific: unofficial endpoint

```go
// URL: https://query1.finance.yahoo.com/v8/finance/chart/{ticker}
// NOT: https://finance.yahoo.com (redirect, adds latency)
//
// Must set User-Agent header:
// "Mozilla/5.0 (compatible; tkr/1.0)"
//
// Price path in response:
// result[0].meta.regularMarketPrice
//
// Rate limit: may return HTTP 429 OR HTTP 200 with empty result array.
// Handle both as ErrRateLimited.
//
// Ticker formats:
//   GPW:      CDR.WA   (NOT .PL — that is Stooq-specific)
//   Frankfurt: VOW3.DE
//   LSE:      BP.L
//   Euronext: AIR.PA
```

### Step 6 — Register in the router

In `internal/provider/router.go`, add the new provider to `DefaultProviders()`:

```go
func DefaultProviders(cfg config.Config) []Provider {
    providers := []Provider{}

    if cfg.Providers.Finnhub.APIKey != "" {
        providers = append(providers, finnhub.New(cfg.Providers.Finnhub.APIKey))
    }
    // ... add your new provider here ...
    // Always append; order determines fallback priority.
    return providers
}
```

### Step 7 — Update ticker mapper

In `internal/provider/ticker_mapper.go`, add mappings for tickers that use different formats across providers:

```go
// Map: canonical ticker → provider-specific ticker
var tickerMap = map[string]map[string]string{
    "CDR.WAR": {
        "stooq":       "CDR.PL",
        "yahoofinance": "CDR.WA",
        "eodhd":       "CDR.WAR",  // EODHD uses canonical format
    },
}

// MapTicker returns the provider-specific ticker for a canonical ticker.
// Falls back to the canonical ticker if no mapping exists.
func MapTicker(canonical, providerID string) string {
    if m, ok := tickerMap[canonical]; ok {
        if mapped, ok := m[providerID]; ok {
            return mapped
        }
    }
    return canonical
}
```

### Step 8 — Add config keys

In `config.example.yaml`, add under `providers:`:

```yaml
providers:
  yourprovider:
    api_key: ""        # required
    enabled: true
    priority: 3        # lower = tried first
```

---

## 3. Router — How Fallback Works

The router (`internal/provider/router.go`) applies this logic for every `Quote` call:

```
1. Determine exchange for ticker (from watchlist DB or ticker suffix).
2. Filter providers to those that Supports(exchange).
3. Sort by priority (ascending).
4. For each provider:
   a. Circuit breaker OPEN? → skip (retry after 5 min)
   b. Rate limited? → skip (exponential backoff: 1→2→4→…→60 min)
   c. Cache hit (TTL = polling_interval / 2)? → return cached quote
   d. Call provider.Quote(ctx, mappedTicker)
      - Success → cache result, reset circuit breaker, return
      - ErrRateLimited → record, skip
      - ErrTickerNotFound → return immediately (no point trying others)
      - Other error → increment circuit breaker failure count, skip
5. All providers failed → return apperrors.ErrNoProvider
```

Circuit breaker threshold: 3 consecutive failures → OPEN for 5 minutes.

---

## 4. Tests Required (minimum)

Every provider must have these test cases in `client_test.go` using `httptest.NewServer`:

| Test | What to verify |
|---|---|
| `TestQuote_Success` | Returns correct `models.Quote` fields from a mocked response |
| `TestQuote_TickerNotFound` | Returns `apperrors.ErrTickerNotFound` |
| `TestQuote_RateLimited` | Returns `apperrors.ErrRateLimited` on HTTP 429 |
| `TestQuote_NetworkError` | Returns a wrapped error on connection failure |
| `TestHistory_Success` | Returns correct slice of `models.OHLCV` |
| `TestSupports_*` | Returns correct bool for supported and unsupported exchanges |

See **`.agent/skills/testing/SKILL.md`** for the `httptest` mock pattern.

---

## 5. Definition of Done

- [ ] `internal/provider/{id}/client.go` implements `Provider` interface
- [ ] All 6 required tests pass
- [ ] Registered in `router.go`
- [ ] Ticker mappings added to `ticker_mapper.go` (if needed)
- [ ] Config keys added to `config.example.yaml`
- [ ] `.agent/API_PROVIDERS.md` updated if endpoint details changed
- [ ] `go-conventions` quality gates pass (`make lint test build`)
