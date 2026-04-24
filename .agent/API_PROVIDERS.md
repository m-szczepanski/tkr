# tkr — API Providers Reference

> This document lists every external data source considered for tkr, with coverage details, free-tier limits, authentication method, and our integration priority.

---

## Decision Matrix

| Provider | US | GPW (PL) | Frankfurt | London | Euronext | Free Tier | Real-time | Priority |
|---|:---:|:---:|:---:|:---:|:---:|---|:---:|---|
| **Finnhub** | ✅ | ❌ | ✅ | ✅ | ✅ | 60 req/min | ✅ | P1 – US & EU |
| **Alpha Vantage** | ✅ | ❌ | ✅ | ✅ | ✅ | 25 req/day | 15 min delay (free) | P1 – US fallback |
| **Stooq** | ✅ | ✅ | ✅ | ✅ | ✅ | Unlimited (CSV) | ❌ (EOD) | P1 – GPW primary |
| **Yahoo Finance (unofficial)** | ✅ | ✅ | ✅ | ✅ | ✅ | No key needed | ~15 min | P2 – universal fallback |
| **EODHD** | ✅ | ✅ | ✅ | ✅ | ✅ | 20 req/day | EOD (free) | P2 – EU & PL |
| **Polygon.io** | ✅ | ❌ | ❌ | ❌ | ❌ | 5 req/min (free) | ✅ (paid) | P3 – US premium |
| **IEX Cloud** | ✅ | ❌ | ❌ | ❌ | ❌ | 50k msg credits/mo | ✅ | P3 – US premium |
| **GPW Official API** | ❌ | ✅ | ❌ | ❌ | ❌ | Throttled (web) | 15 min | P3 – PL supplementary |
| **Marketstack** | ✅ | ✅ | ✅ | ✅ | ✅ | 100 req/mo | ❌ (EOD) | P4 – low-priority |

---

## Detailed Provider Profiles

---

### 1. Finnhub

**Website:** https://finnhub.io  
**Base URL:** `https://finnhub.io/api/v1`  
**Auth:** `?token=API_KEY` query param or `X-Finnhub-Token` header  
**Free tier:** 60 requests/minute, US + select international exchanges  
**Paid tiers:** from $0 (free) to $499+/month for premium real-time data

**Endpoints used:**

| Purpose | Endpoint |
|---|---|
| Real-time quote | `GET /quote?symbol={ticker}` |
| Company profile | `GET /stock/profile2?symbol={ticker}` |
| Ticker search | `GET /search?q={query}` |
| Candles (OHLCV) | `GET /stock/candle?symbol={}&resolution={}&from={}&to={}` |
| Exchange list | `GET /stock/exchange` |

**Response — quote:**
```json
{
  "c": 168.42,       // current price
  "d": -3.21,        // change
  "dp": -1.87,       // change percent
  "h": 172.00,       // high
  "l": 167.50,       // low
  "o": 171.00,       // open
  "pc": 171.63,      // previous close
  "t": 1713360000    // unix timestamp
}
```

**Coverage notes:**
- Excellent for US (NYSE, NASDAQ, AMEX).
- Covers major European exchanges (XETRA, LSE, Euronext) on free tier with some symbols.
- Does **not** cover GPW Warsaw.
- WebSocket feed available on paid plan for streaming quotes.

**Rate limiting:** HTTP 429 with `X-RateLimit-Remaining` and `X-RateLimit-Reset` headers.

---

### 2. Alpha Vantage

**Website:** https://www.alphavantage.co  
**Base URL:** `https://www.alphavantage.co/query`  
**Auth:** `&apikey=API_KEY` query param  
**Free tier:** 25 requests/day, 5 requests/minute  
**Paid tiers:** from $50/month (75 req/min) to $600/month (1200 req/min)

**Endpoints used:**

| Purpose | Function |
|---|---|
| Real-time quote | `function=GLOBAL_QUOTE&symbol={ticker}` |
| Time series daily | `function=TIME_SERIES_DAILY&symbol={ticker}` |
| Symbol search | `function=SYMBOL_SEARCH&keywords={query}` |
| FX (currency rates) | `function=CURRENCY_EXCHANGE_RATE` |

**Response — GLOBAL_QUOTE:**
```json
{
  "Global Quote": {
    "01. symbol": "AAPL",
    "02. open": "171.00",
    "03. high": "172.00",
    "04. low": "167.50",
    "05. price": "168.42",
    "06. volume": "55300000",
    "07. latest trading day": "2026-04-17",
    "08. previous close": "171.63",
    "09. change": "-3.21",
    "10. change percent": "-1.87%"
  }
}
```

**Coverage notes:**
- Reliable for US equities.
- International symbols use `{TICKER}.{EXCHANGE_SUFFIX}` format (e.g. `VOW.FRA`).
- Free tier is very limited (25/day) — use only as fallback.
- Does **not** support GPW Warsaw.

---

### 3. Stooq

**Website:** https://stooq.com  
**Base URL:** `https://stooq.com/q/d/l/`  
**Auth:** None (public CSV endpoint)  
**Free tier:** Unlimited (best-effort, no SLA)  
**Paid tier:** None

**Endpoints used:**

| Purpose | URL pattern |
|---|---|
| Historical daily CSV | `GET /q/d/l/?s={ticker}&i=d` |
| Intraday (5-min) CSV | `GET /q/d/l/?s={ticker}&i=5` |

**CSV format:**
```
Date,Open,High,Low,Close,Volume
2026-04-17,168.00,172.00,167.50,168.42,55300000
```

**Ticker format for GPW:**
- GPW tickers end in `.PL`: e.g. `CDR.PL` (CD Projekt Red), `PKN.PL` (PKN Orlen), `KGH.PL` (KGHM)
- WIG20 index: `WIG20.PL`
- Other Polish indices: `WIG.PL`, `MWIG40.PL`, `SWIG80.PL`

**Coverage notes:**
- The **best free source for Polish GPW data**.
- Also covers US, German, UK, and many international markets.
- Data is delayed (typically 15 min during market hours, EOD after close).
- No JSON API — parsing the CSV response is required.
- No authentication makes it the most reliable fallback with zero quota concerns.
- No WebSocket / push support; polling only.

**Important:** Stooq has no official API documentation. The CSV endpoint is stable but unofficial. Treat it as best-effort.

---

### 4. Yahoo Finance (Unofficial)

**Library:** Use `go-quote` or custom HTTP scraping of the v8 endpoint  
**Base URL:** `https://query1.finance.yahoo.com/v8/finance/chart/{ticker}`  
**Auth:** None (but rate limits enforced by IP)  
**Free tier:** Unofficial — no key, throttled by IP (~2000 req/hour observed)

**Endpoints used:**

| Purpose | URL |
|---|---|
| Real-time quote + history | `GET /v8/finance/chart/{ticker}?interval=1d&range=1mo` |
| Quote summary | `GET /v10/finance/quoteSummary/{ticker}?modules=price` |

**Ticker format:**
- GPW: `CDR.WA` (note `.WA` not `.PL` for Yahoo)
- Frankfurt: `VOW3.DE`
- LSE: `BP.L`
- Euronext Paris: `AIR.PA`
- US: no suffix needed

**Coverage notes:**
- Widest exchange coverage of any free source.
- No official API — can break without notice.
- Good for: GPW, LSE, Euronext, XETRA, NASDAQ, NYSE.
- Use as a **universal fallback** when all keyed providers fail.

---

### 5. EODHD (End of Day Historical Data)

**Website:** https://eodhd.com  
**Base URL:** `https://eodhd.com/api`  
**Auth:** `?api_token=API_KEY`  
**Free tier:** 20 API calls/day, EOD data only  
**Paid tiers:** from $19.99/month (unlimited EOD) to $79.99/month (real-time)

**Endpoints used:**

| Purpose | Endpoint |
|---|---|
| Real-time quote | `GET /real-time/{ticker}.{EXCHANGE}?api_token={key}&fmt=json` |
| EOD historical | `GET /eod/{ticker}.{EXCHANGE}?api_token={key}&fmt=json` |
| Ticker search | `GET /search/{query}?api_token={key}` |
| Exchange symbols | `GET /exchange-symbol-list/{EXCHANGE}?api_token={key}&fmt=json` |

**Exchange codes for this project:**
- `US` — NYSE/NASDAQ
- `WAR` — GPW Warsaw  ← **important for Polish stocks**
- `XETRA` — Frankfurt XETRA
- `LSE` — London
- `PA` — Euronext Paris
- `VX` — SIX Swiss Exchange

**Coverage notes:**
- Has **official GPW Warsaw support** (`WAR` exchange code).
- Excellent for historical EOD data across 70+ exchanges.
- Real-time requires a paid plan.
- Recommended as the **primary GPW provider** alongside Stooq, since it has a proper API.

---

### 6. Polygon.io

**Website:** https://polygon.io  
**Base URL:** `https://api.polygon.io/v2`  
**Auth:** `Authorization: Bearer API_KEY`  
**Free tier:** 5 req/min, previous-day data only (no real-time)  
**Paid tiers:** from $29/month (real-time) to $199/month (full WebSocket)

**Endpoints used:**

| Purpose | Endpoint |
|---|---|
| Previous day OHLCV | `GET /aggs/ticker/{ticker}/prev` |
| Snapshot (real-time) | `GET /snapshot/locale/us/markets/stocks/tickers/{ticker}` |
| Ticker details | `GET /reference/tickers/{ticker}` |

**Coverage notes:**
- US markets only (NYSE, NASDAQ, OTC).
- Best-in-class data quality for US equities.
- WebSocket for streaming ticks on paid plans.
- Not useful for international or Polish markets.
- Recommended only if high-quality US real-time data is a priority.

---

### 7. IEX Cloud

**Website:** https://iexcloud.io  
**Base URL:** `https://cloud.iexapis.com/stable`  
**Auth:** `?token=API_KEY`  
**Free tier:** 50,000 "message credits" per month (each endpoint costs different credits)  
**Paid tiers:** from $19/month

**Endpoints used:**

| Purpose | Endpoint | Credits |
|---|---|---|
| Real-time quote | `GET /stock/{ticker}/quote` | 1 credit |
| Historical prices | `GET /stock/{ticker}/chart/1m` | 1 credit/day |
| Company info | `GET /stock/{ticker}/company` | 1 credit |

**Coverage notes:**
- US markets only.
- Clean, well-documented API.
- The credit-based pricing model makes cost prediction tricky.
- Free tier can be exhausted quickly when polling multiple tickers.

---

### 8. GPW Official (Warsaw Stock Exchange)

**Website:** https://www.gpw.pl  
**Data portal:** https://gpwbenchmark.pl (indices), https://gpwcatalyst.pl (bonds)  
**Auth:** Web scraping or unofficial undocumented endpoints  
**Free tier:** Web-level access (scraping)

**Notes:**
- GPW does not offer a public REST API for real-time stock prices.
- The GPW website provides a data download service for institutional clients.
- For free access, **Stooq (.PL)** or **EODHD (WAR exchange)** are strongly preferred.
- GPW Benchmark publishes official WIG index data; consider scraping the CSV download at `https://gpwbenchmark.pl/pub/FIXTOOL/chart-data.json` for index values.

---

### 9. Marketstack

**Website:** https://marketstack.com  
**Base URL:** `http://api.marketstack.com/v1`  
**Auth:** `?access_key=API_KEY`  
**Free tier:** 100 requests/month, EOD data only, HTTP only (no HTTPS on free)  
**Paid tiers:** from $9.99/month

**Endpoints used:**

| Purpose | Endpoint |
|---|---|
| EOD data | `GET /eod?symbols={ticker}&access_key={key}` |
| Ticker search | `GET /tickers?search={query}&access_key={key}` |

**Coverage notes:**
- Wide exchange coverage including Warsaw (`XWAR`).
- Free tier is too limited (100 req/month) for practical use.
- HTTPS only on paid plans — do not use free tier for production.
- Low priority; consider only if EODHD and Stooq are unavailable.

---

## Provider–Exchange Mapping (Summary)

```
NYSE / NASDAQ (US real-time):
  Primary:   Finnhub
  Secondary: Alpha Vantage
  Fallback:  Yahoo Finance (unofficial)
  Premium:   Polygon.io, IEX Cloud

GPW Warsaw (PL):
  Primary:   Stooq (.PL suffix, free CSV)
  Secondary: EODHD (WAR exchange code)
  Fallback:  Yahoo Finance (.WA suffix)

XETRA Frankfurt (DE):
  Primary:   Finnhub (.FRA suffix)
  Secondary: EODHD (XETRA exchange code)
  Fallback:  Yahoo Finance (.DE suffix)

LSE London (GB):
  Primary:   Finnhub (.L suffix)
  Secondary: EODHD (LSE exchange code)
  Fallback:  Yahoo Finance (.L suffix)

Euronext (FR/NL/BE):
  Primary:   Finnhub (.PA / .AMS suffix)
  Secondary: EODHD (PA / AMS exchange code)
  Fallback:  Yahoo Finance (.PA / .AS suffix)
```

---

## Recommended MVP Setup

For the initial version, use only **3 providers** to minimise integration complexity:

1. **Finnhub** — Primary for US and major European exchanges (requires free API key)
2. **Stooq** — Primary for GPW Warsaw (no key needed)
3. **Yahoo Finance (unofficial)** — Universal fallback (no key needed)

This covers 100% of the target exchanges with zero mandatory paid subscriptions.

Add **EODHD** in v0.2 for better GPW historical data and multi-market EOD history.
