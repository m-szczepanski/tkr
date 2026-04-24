---
name: database-layer
description: >
  Use this skill for all database-related work in tkr: implementing or
  modifying Repository methods, writing SQL migrations, querying quote history,
  managing the schema, or debugging SQLite behaviour. Triggers include: adding a new
  table or column, writing a new SQL query, implementing any method on the Repository
  interface, writing a migration file, or working with the db package in any way.
  Always read go-conventions/SKILL.md first.
---

# tkr — Database Layer

> Prerequisite: **`.agent/skills/go-conventions/SKILL.md`** must be read before this skill.
> Full schema is in **`.agent/FUNCTIONAL_SPEC.md` §4**. Migration files live in `migrations/`.

---

## 1. The Repository Interface

All database access goes through the `Repository` interface in `internal/db/repository.go`.
**Nothing outside `internal/db/` may import `modernc.org/sqlite` directly.**

```go
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

    // Lifecycle
    Close() error
}
```

---

## 2. Opening the Database

```go
// internal/db/sqlite.go

func Open(path string) (Repository, error) {
    // Expand ~ in path
    expanded, err := expandPath(path)
    if err != nil {
        return nil, fmt.Errorf("db.Open: expand path: %w", err)
    }

    // Ensure parent directory exists
    if err := os.MkdirAll(filepath.Dir(expanded), 0755); err != nil {
        return nil, fmt.Errorf("db.Open: mkdir: %w", err)
    }

    db, err := sqlx.Open("sqlite", expanded)
    if err != nil {
        return nil, fmt.Errorf("db.Open: %w", err)
    }

    // SQLite pragmas — always set these
    if _, err := db.Exec(`
        PRAGMA journal_mode=WAL;
        PRAGMA foreign_keys=ON;
        PRAGMA busy_timeout=5000;
    `); err != nil {
        return nil, fmt.Errorf("db.Open: pragmas: %w", err)
    }

    return &sqliteRepo{db: db}, nil
}
```

**Always set these three PRAGMAs.** WAL mode prevents write-lock contention when the daemon and CLI run simultaneously. `foreign_keys=ON` enforces referential integrity. `busy_timeout` prevents immediate failures under concurrent access.

---

## 3. Writing Repository Methods

### Pattern: named query with sqlx

```go
func (r *sqliteRepo) AddStock(ctx context.Context, s models.Stock) error {
    const q = `
        INSERT INTO stocks (ticker, name, exchange, currency, added_at)
        VALUES (:ticker, :name, :exchange, :currency, :added_at)
        ON CONFLICT(ticker) DO NOTHING
    `
    if _, err := r.db.NamedExecContext(ctx, q, s); err != nil {
        return fmt.Errorf("db.AddStock: %w", err)
    }
    return nil
}
```

### Pattern: query returning a slice

```go
func (r *sqliteRepo) ListStocks(ctx context.Context) ([]models.Stock, error) {
    const q = `SELECT ticker, name, exchange, currency, added_at FROM stocks ORDER BY ticker`

    var stocks []models.Stock
    if err := r.db.SelectContext(ctx, &stocks, q); err != nil {
        return nil, fmt.Errorf("db.ListStocks: %w", err)
    }
    return stocks, nil
}
```

### Pattern: single row with not-found handling

```go
func (r *sqliteRepo) GetStock(ctx context.Context, ticker string) (models.Stock, error) {
    const q = `SELECT ticker, name, exchange, currency, added_at FROM stocks WHERE ticker = ?`

    var s models.Stock
    if err := r.db.GetContext(ctx, &s, q, ticker); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return models.Stock{}, apperrors.ErrTickerNotFound
        }
        return models.Stock{}, fmt.Errorf("db.GetStock: %w", err)
    }
    return s, nil
}
```

**Always map `sql.ErrNoRows` to the appropriate sentinel error.** Never surface it raw.

### Pattern: storing JSON arrays (channel list on AlertRule)

`AlertRule.Channels` is `[]models.ChannelID`. SQLite stores it as a JSON array text column.

```go
// Storing
channelsJSON, err := json.Marshal(rule.Channels)
if err != nil {
    return 0, fmt.Errorf("db.AddAlertRule: marshal channels: %w", err)
}

// Reading — use a scan struct with a string field, then unmarshal
type alertRuleRow struct {
    // ... other fields with db tags ...
    ChannelsJSON string `db:"channels"`
}
var row alertRuleRow
r.db.GetContext(ctx, &row, q, id)
json.Unmarshal([]byte(row.ChannelsJSON), &rule.Channels)
```

---

## 4. Migrations

### File naming

```
migrations/
├── 001_initial_schema.sql
├── 002_add_alert_cooldown.sql
└── 003_add_quote_history_index.sql
```

Numbering is zero-padded to three digits. **Never edit an existing migration file.** Add a new one.

### Migration runner contract (`internal/db/migrate.go`)

The runner:
1. Creates a `schema_migrations` table if it doesn't exist.
2. Reads all `.sql` files from `migrations/` in alphabetical order.
3. For each file, checks if it has already been applied (by filename).
4. If not applied, executes the SQL in a transaction and records the filename.

```go
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename   TEXT PRIMARY KEY,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Writing migration SQL

```sql
-- migrations/002_add_alert_cooldown.sql
-- Always wrap in a transaction.
BEGIN;

ALTER TABLE alert_rules ADD COLUMN cooldown_s INTEGER NOT NULL DEFAULT 3600;

COMMIT;
```

**Rules:**
- Always wrap in `BEGIN; ... COMMIT;`
- Use `IF NOT EXISTS` / `IF EXISTS` where available
- No irreversible destructive statements (no `DROP TABLE`, no `DELETE FROM` without a `WHERE`)
- Test the migration manually with `sqlite3` before committing

---

## 5. Quote History

The `quote_history` table grows unboundedly without pruning. The scheduler calls `PruneHistory` once per day.

```go
func (r *sqliteRepo) PruneHistory(ctx context.Context, olderThan time.Duration) error {
    cutoff := time.Now().UTC().Add(-olderThan)
    const q = `DELETE FROM quote_history WHERE recorded_at < ?`
    if _, err := r.db.ExecContext(ctx, q, cutoff); err != nil {
        return fmt.Errorf("db.PruneHistory: %w", err)
    }
    return nil
}
```

`GetRecentQuotes` is used by the alert evaluator for moving average calculations. Always return in **descending** time order (most recent first):

```go
const q = `
    SELECT ticker, price, volume, change_pct, source, recorded_at
    FROM quote_history
    WHERE ticker = ?
    ORDER BY recorded_at DESC
    LIMIT ?
`
```

---

## 6. Testing the Database Layer

Always use an in-memory SQLite instance in tests. Never use a file path.

```go
func newTestRepo(t *testing.T) db.Repository {
    t.Helper()
    repo, err := db.Open(":memory:")
    require.NoError(t, err)
    t.Cleanup(func() { repo.Close() })
    return repo
}
```

Then run migrations against it:

```go
func TestAddStock(t *testing.T) {
    repo := newTestRepo(t)

    stock := models.Stock{Ticker: "AAPL", Name: "Apple Inc.", Exchange: models.ExchangeNASDAQ, Currency: "USD"}
    err := repo.AddStock(context.Background(), stock)
    require.NoError(t, err)

    got, err := repo.GetStock(context.Background(), "AAPL")
    require.NoError(t, err)
    assert.Equal(t, stock.Ticker, got.Ticker)
}
```

See **`.agent/skills/testing/SKILL.md`** for full patterns.

---

## 7. Definition of Done

- [ ] All new/modified methods have corresponding unit tests using `:memory:`
- [ ] New columns or tables are introduced via a new numbered migration file
- [ ] `sql.ErrNoRows` is mapped to an `apperrors` sentinel at the repository boundary
- [ ] All times persisted as UTC (`time.UTC()`)
- [ ] WAL/foreign_key pragmas remain set in `Open()`
- [ ] `go-conventions` quality gates pass (`make lint test build`)
