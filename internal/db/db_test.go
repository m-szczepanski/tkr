package db

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yourname/tkr/pkg/models"
)

func TestMigrateIdempotent(t *testing.T) {
	repo := newTestRepo(t)

	require.NoError(t, Migrate(repo.db))
	require.NoError(t, Migrate(repo.db))

	dir, err := findMigrationsDir()
	require.NoError(t, err)

	files, err := listMigrationFiles(dir)
	require.NoError(t, err)

	var appliedCount int
	require.NoError(t, repo.db.Get(&appliedCount, `SELECT COUNT(1) FROM schema_migrations`))
	require.Equal(t, len(files), appliedCount)

	for _, file := range files {
		filename := filepath.Base(file)

		var count int
		require.NoError(t, repo.db.Get(&count, `SELECT COUNT(1) FROM schema_migrations WHERE filename = ?`, filename))
		require.Equal(t, 1, count)
	}
}

func TestInitialSchemaInsertAndQueryAllTables(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)

	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO stocks (ticker, name, exchange, currency, added_at)
		VALUES (?, ?, ?, ?, ?)
	`, "AAPL", "Apple Inc.", "NASDAQ", "USD", now)
	require.NoError(t, err)

	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO alert_rules (ticker, metric, operator, value, period, channels, active, one_shot, cooldown_s, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "AAPL", "PRICE", "LT", 150.0, nil, `["terminal"]`, 1, 0, 3600, now)
	require.NoError(t, err)

	var ruleID int64
	require.NoError(t, repo.db.Get(&ruleID, `SELECT id FROM alert_rules WHERE ticker = ?`, "AAPL"))
	require.NotZero(t, ruleID)

	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO alert_events (rule_id, ticker, triggered_at, price, change_pct, message)
		VALUES (?, ?, ?, ?, ?, ?)
	`, ruleID, "AAPL", now, 149.5, -0.75, "price below threshold")
	require.NoError(t, err)

	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO quote_history (ticker, price, volume, change_pct, source, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "AAPL", 149.5, 1_000_000, -0.75, "finnhub", now)
	require.NoError(t, err)

	var stockName string
	require.NoError(t, repo.db.Get(&stockName, `SELECT name FROM stocks WHERE ticker = ?`, "AAPL"))
	require.Equal(t, "Apple Inc.", stockName)

	var alertCount int
	require.NoError(t, repo.db.Get(&alertCount, `SELECT COUNT(1) FROM alert_rules WHERE ticker = ?`, "AAPL"))
	require.Equal(t, 1, alertCount)

	var eventCount int
	require.NoError(t, repo.db.Get(&eventCount, `SELECT COUNT(1) FROM alert_events WHERE ticker = ?`, "AAPL"))
	require.Equal(t, 1, eventCount)

	var historyCount int
	require.NoError(t, repo.db.Get(&historyCount, `SELECT COUNT(1) FROM quote_history WHERE ticker = ?`, "AAPL"))
	require.Equal(t, 1, historyCount)

	var indexCount int
	require.NoError(t, repo.db.Get(&indexCount, `SELECT COUNT(1) FROM sqlite_master WHERE type = 'index' AND name = 'idx_qh_ticker_time'`))
	require.Equal(t, 1, indexCount)
}

func TestQuoteHistoryMethods(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	base := time.Now().UTC().Truncate(time.Second)

	quotes := []models.Quote{
		{Ticker: "AAPL", Price: 101.25, Volume: 1000, ChangePct: 0.1, Source: models.ProviderID("finnhub"), Timestamp: base.Add(-3 * time.Hour)},
		{Ticker: "AAPL", Price: 102.50, Volume: 1100, ChangePct: 0.2, Source: models.ProviderID("finnhub"), Timestamp: base.Add(-2 * time.Hour)},
		{Ticker: "AAPL", Price: 103.75, Volume: 1200, ChangePct: 0.3, Source: models.ProviderID("finnhub"), Timestamp: base.Add(-1 * time.Hour)},
	}

	for _, q := range quotes {
		require.NoError(t, repo.SaveQuote(ctx, q))
	}

	recent, err := repo.GetRecentQuotes(ctx, "AAPL", 2)
	require.NoError(t, err)
	require.Len(t, recent, 2)
	require.Equal(t, 103.75, recent[0].Price)
	require.Equal(t, 102.50, recent[1].Price)
	require.Equal(t, "AAPL", recent[0].Ticker)

	empty, err := repo.GetRecentQuotes(ctx, "AAPL", 0)
	require.NoError(t, err)
	require.Empty(t, empty)

	require.NoError(t, repo.PruneHistory(ctx, 90*time.Minute))

	remaining, err := repo.GetRecentQuotes(ctx, "AAPL", 10)
	require.NoError(t, err)
	require.Len(t, remaining, 1)
	require.Equal(t, 103.75, remaining[0].Price)
}

func TestNonM1MethodsReturnSentinelNotImplemented(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.AddStock(ctx, models.Stock{Ticker: "AAPL"})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotImplemented))

	_, err = repo.ListAlertRules(ctx, "AAPL")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotImplemented))
}

func newTestRepo(t *testing.T) *SQLiteRepository {
	t.Helper()

	repository, err := Open(":memory:")
	require.NoError(t, err)

	repo, ok := repository.(*SQLiteRepository)
	require.True(t, ok)

	t.Cleanup(func() {
		require.NoError(t, repo.Close())
	})

	return repo
}
