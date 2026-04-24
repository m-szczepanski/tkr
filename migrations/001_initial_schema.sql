BEGIN;

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

COMMIT;
