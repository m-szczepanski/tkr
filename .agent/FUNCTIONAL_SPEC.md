# tkr — Functional Specification

> Version: 0.1-draft  
> Last updated: 2026-04

This document describes every behaviour, function, and data flow in tkr. It is the authoritative reference for implementation decisions.

---

## 1. Domain Model

### 1.1 Stock

```
Stock {
  ticker       string          // e.g. "AAPL", "CDR.WAR"
  name         string          // e.g. "Apple Inc."
  exchange     ExchangeID      // NYSE | NASDAQ | GPW | XETRA | LSE | EURONEXT
  currency     string          // ISO 4217, e.g. "USD", "PLN", "EUR"
  added_at     time.Time
}
```

### 1.2 Quote

```
Quote {
  ticker       string
  price        float64
  open         float64
  high         float64
  low          float64
  close        float64         // previous close
  volume       int64
  change       float64         // absolute vs. previous close
  change_pct   float64         // percentage vs. previous close
  timestamp    time.Time
  source       ProviderID
}
```

### 1.3 AlertRule

```
AlertRule {
  id           int64
  ticker       string
  condition    Condition       // see §3
  channel      []ChannelID     // terminal | email | webhook
  active       bool
  one_shot     bool            // deactivate after first trigger
  created_at   time.Time
  last_fired   *time.Time
  cooldown     time.Duration   // minimum time between re-fires
}
```

### 1.4 AlertEvent

```
AlertEvent {
  id           int64
  rule_id      int64
  ticker       string
  triggered_at time.Time
  quote        Quote           // snapshot at trigger time
  message      string
}
```

### 1.5 Condition (Alert Rule DSL)

Conditions are stored as structured records, not raw strings.

```
Condition {
  metric    Metric            // PRICE | CHANGE_ABS | CHANGE_PCT | VOLUME | MA_CROSS
  operator  Operator          // LT | LTE | GT | GTE | EQ | CROSS_ABOVE | CROSS_BELOW
  value     float64
  period    *int              // optional, for moving averages (e.g. 20 for MA20)
}
```

Supported condition expressions (CLI syntax → stored Condition):

| CLI expression | Metric | Operator | Value |
|---|---|---|---|
| `price < 150` | PRICE | LT | 150 |
| `price > 200` | PRICE | GT | 200 |
| `change% > 5` | CHANGE_PCT | GT | 5 |
| `change% < -3` | CHANGE_PCT | LT | -3 |
| `volume > 5000000` | VOLUME | GT | 5000000 |
| `ma20 cross_above ma50` | MA_CROSS | CROSS_ABOVE | 50 (period=20) |

---

## 2. CLI Commands

### 2.1 `tkr init`

- Creates `~/.config/tkr/config.yaml` from the bundled template if it does not exist.
- Creates the SQLite database file at the configured path.
- Runs all pending schema migrations.
- Prints a success message with next-step hints.

**Flags:** none

---

### 2.2 `tkr watch`

Sub-commands for managing the watchlist.

#### `watch add <TICKER> [--exchange <ID>]`

- Looks up the ticker via the configured primary provider to confirm it exists.
- If the ticker is ambiguous (same symbol on multiple exchanges), prompts the user to pick one, or accepts `--exchange` flag to disambitate.
- Inserts the stock into the `stocks` table.
- Prints confirmation with the resolved company name and exchange.

**Errors:**
- Ticker not found → `ERR_TICKER_NOT_FOUND`
- Already in watchlist → `WARN_ALREADY_WATCHING` (idempotent, exits 0)

#### `watch remove <TICKER>`

- Removes the stock from the watchlist.
- Does NOT delete associated alert rules by default; deactivates them and prints a warning.
- `--purge-alerts` flag deletes associated rules entirely.

#### `watch list [--format table|json|csv]`

- Fetches current quotes for all watched stocks (may be from cache if last fetch < 1 min ago).
- Renders a table with columns: Ticker | Name | Exchange | Price | Change | Change% | Volume | Last Updated.
- Highlights rows: green for positive change, red for negative, yellow if data is stale (> 15 min).
- `--format json` / `--format csv` suppress colour and output machine-readable data.

---

### 2.3 `tkr quote <TICKER> [TICKER...]`

- Fetches live quotes for one or more tickers (not necessarily in watchlist).
- Displays the same table format as `watch list`.
- `--history <N>` additionally fetches and renders a sparkline of the last N days of closing prices.
- `--json` flag for machine-readable output.

**Sparkline rendering rules:**
- Width: min(terminal_width - 30, 60) characters.
- Uses Unicode block characters: `▁▂▃▄▅▆▇█`.
- Last value is highlighted with a `●` marker.

---

### 2.4 `tkr alert`

#### `alert add <TICKER> --condition "<EXPR>" [--channel <CH>] [--one-shot] [--cooldown <DURATION>]`

- Parses the condition expression (see §1.5).
- Validates that the ticker is in the watchlist (warns but does not block with `--force`).
- Inserts the `AlertRule` into the database.
- Default channel: `terminal`.
- Default cooldown: `1h`.

#### `alert list [--ticker <TICKER>] [--active-only]`

- Displays all alert rules in a table: ID | Ticker | Condition | Channels | Active | Last Fired | Cooldown.

#### `alert remove <ID>`

- Deletes the alert rule by ID.

#### `alert enable <ID>` / `alert disable <ID>`

- Sets `active = true/false` without deleting the rule.

#### `alert history [--ticker <TICKER>] [--limit <N>] [--since <DATE>]`

- Displays past `AlertEvent` records in reverse-chronological order.
- Default limit: 50.

---

### 2.5 `tkr daemon`

#### `daemon start [--foreground]`

- Launches the polling loop.
- By default, forks to the background and writes a PID file to `~/.local/share/tkr/daemon.pid`.
- `--foreground` keeps it in the foreground (useful for systemd / Docker).
- On start, logs the polling interval, number of watched stocks, and active alert rules.

#### `daemon stop`

- Sends SIGTERM to the PID in the PID file.
- Waits up to 5 s for clean shutdown, then SIGKILL.

#### `daemon status`

- Reports whether the daemon is running, the PID, uptime, last poll time, and number of alerts fired since start.

#### `daemon restart`

- Equivalent to `stop` then `start`.

---

### 2.6 `tkr config`

#### `config show`

- Pretty-prints the resolved configuration (masks API keys).

#### `config set <KEY> <VALUE>`

- Updates a single key in the config file.

#### `config validate`

- Checks all configured API keys are reachable and valid.
- Reports provider status: ✓ OK / ✗ FAIL / ⚠ RATE_LIMITED.

---

## 3. Core Subsystems

### 3.1 Provider Router

The router selects which market data provider to use for a given ticker and falls back gracefully.

**Routing logic:**

```
1. Determine the exchange for the ticker.
2. Select the ordered provider list for that exchange from config.
3. Try providers in order:
   a. If the provider is healthy and not rate-limited → use it, cache result.
   b. If rate-limited → record in in-memory rate-limit map, skip.
   c. If network error → log warning, skip.
4. If all providers fail → return ErrNoProvider.
```

**Cache:** Each `Quote` is cached in memory with a TTL equal to half the polling interval. This prevents redundant API calls when both the daemon and a manual `quote` command run close together.

**Provider health tracking:**

- Each provider has a circuit-breaker: after 3 consecutive failures, it is marked `OPEN` for 5 minutes before retrying.
- Rate-limit responses (HTTP 429) back off exponentially: 1 min → 2 min → 4 min … max 60 min.

---

### 3.2 Polling Scheduler

- Implemented with `robfig/cron`.
- One job per polling interval (default `*/5 * * * *`).
- On each tick:
  1. Fetch quotes for all active watched stocks.
  2. Persist quotes to `quote_history` table (keep 90 days by default, configurable).
  3. Run alert evaluation loop (§3.3).
  4. Dispatch any triggered alerts (§3.4).
- Market hours gate: if `market_hours_only: true`, checks whether the relevant exchange is currently open. Skips polling for closed exchanges. Still polls if at least one watched stock's exchange is open.

**Exchange open-hours table (UTC):**

| Exchange | Open (UTC) | Close (UTC) | Days |
|---|---|---|---|
| NYSE / NASDAQ | 14:30 | 21:00 | Mon–Fri |
| GPW Warsaw | 08:00 | 16:50 | Mon–Fri |
| XETRA Frankfurt | 07:00 | 15:30 | Mon–Fri |
| LSE London | 08:00 | 16:30 | Mon–Fri |
| Euronext Paris | 08:00 | 16:30 | Mon–Fri |

Public holidays are sourced from a bundled YAML file and updated with each release.

---

### 3.3 Alert Evaluator

For each active `AlertRule`, after a fresh quote is received:

1. **Cooldown check** — if `now - last_fired < cooldown`, skip.
2. **Metric extraction** — compute the metric value from the `Quote`:
   - `PRICE` → `quote.price`
   - `CHANGE_ABS` → `quote.change`
   - `CHANGE_PCT` → `quote.change_pct`
   - `VOLUME` → `float64(quote.volume)`
   - `MA_CROSS` → query last N close prices from `quote_history`, compute both MAs.
3. **Operator evaluation** — compare extracted value against `condition.value`.
4. If the condition is met:
   - Insert an `AlertEvent` record.
   - Enqueue the alert for dispatch (§3.4).
   - Update `last_fired` on the rule.
   - If `one_shot`, set `active = false`.

---

### 3.4 Notification Dispatcher

Dispatches `AlertEvent` objects to one or more channels concurrently.

**Terminal channel:**
- Writes a formatted line to stdout (if daemon is in foreground) or appends to a log file.
- Fires a native desktop notification via `beeep.Notify` (title = ticker, body = condition + price).

**Email channel:**
- Connects to configured SMTP server.
- Sends a plain-text email with: stock name, current price, condition that triggered, timestamp.
- HTML email template is also included (auto-detected by client).
- Retries up to 3 times with a 10 s delay on failure.

**Webhook channel:**
- Posts a JSON payload to the configured URL.
- Payload schema:
  ```json
  {
    "ticker": "AAPL",
    "name": "Apple Inc.",
    "exchange": "NASDAQ",
    "price": 168.42,
    "change_pct": -3.21,
    "condition": "price < 170",
    "fired_at": "2026-04-17T14:35:00Z"
  }
  ```
- Compatible with Slack Incoming Webhooks and Discord webhooks out of the box.
- Retries up to 3 times with exponential backoff.

---

## 4. Database Schema

```sql
-- Watched stocks
CREATE TABLE stocks (
  ticker      TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  exchange    TEXT NOT NULL,
  currency    TEXT NOT NULL DEFAULT 'USD',
  added_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Alert rules
CREATE TABLE alert_rules (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  ticker      TEXT NOT NULL REFERENCES stocks(ticker),
  metric      TEXT NOT NULL,
  operator    TEXT NOT NULL,
  value       REAL NOT NULL,
  period      INTEGER,              -- for MA_CROSS
  channels    TEXT NOT NULL,        -- JSON array of channel IDs
  active      INTEGER NOT NULL DEFAULT 1,
  one_shot    INTEGER NOT NULL DEFAULT 0,
  cooldown_s  INTEGER NOT NULL DEFAULT 3600,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_fired  DATETIME
);

-- Alert events (history)
CREATE TABLE alert_events (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  rule_id      INTEGER NOT NULL REFERENCES alert_rules(id),
  ticker       TEXT NOT NULL,
  triggered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  price        REAL NOT NULL,
  change_pct   REAL NOT NULL,
  message      TEXT NOT NULL
);

-- Quote history
CREATE TABLE quote_history (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  ticker     TEXT NOT NULL,
  price      REAL NOT NULL,
  volume     INTEGER NOT NULL,
  change_pct REAL NOT NULL,
  source     TEXT NOT NULL,
  recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_qh_ticker_time ON quote_history(ticker, recorded_at DESC);
```

---

## 5. Error Codes

| Code | Meaning |
|---|---|
| `ERR_TICKER_NOT_FOUND` | Provider could not resolve the ticker |
| `ERR_NO_PROVIDER` | All providers failed for this exchange |
| `ERR_RATE_LIMITED` | All providers are rate-limited |
| `ERR_DB` | SQLite read/write failure |
| `ERR_CONFIG` | Config file is missing or malformed |
| `ERR_DAEMON_RUNNING` | Daemon is already running (on `start`) |
| `ERR_DAEMON_NOT_RUNNING` | No daemon PID found (on `stop`/`status`) |
| `ERR_NOTIFY_FAIL` | Notification channel dispatch failed (non-fatal) |
| `WARN_ALREADY_WATCHING` | Ticker is already in the watchlist |
| `WARN_STALE_DATA` | Quote data is older than `stale_threshold` |

---

## 6. Configuration Reference

| Key | Type | Default | Description |
|---|---|---|---|
| `polling_interval` | duration | `5m` | How often to poll for quotes |
| `market_hours_only` | bool | `true` | Skip polling outside exchange hours |
| `stale_threshold` | duration | `20m` | Age at which a quote is marked stale |
| `history_retention` | duration | `2160h` (90d) | How long to keep `quote_history` rows |
| `log_level` | string | `info` | `debug` / `info` / `warn` / `error` |
| `log_file` | string | `~/.local/share/tkr/app.log` | Log file path |
| `database.path` | string | `~/.local/share/tkr/data.db` | SQLite path |
| `providers.<id>.api_key` | string | — | API key for the provider |
| `providers.<id>.enabled` | bool | `true` | Enable/disable a provider |
| `providers.<id>.priority` | int | varies | Lower = tried first |
| `notifications.email.*` | — | — | See §3.4 |
| `notifications.webhook.url` | string | — | Webhook endpoint |

---

## 7. Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General application error |
| 2 | Configuration error |
| 3 | Provider / network error |
| 4 | Database error |
| 5 | Daemon management error |
