---
name: cli-command
description: >
  Use this skill whenever adding a new top-level CLI command or sub-command to
  tkr, modifying an existing command's flags or behaviour, or wiring
  a command to internal application logic. Triggers include: implementing any
  `tkr <verb>` or `tkr <noun> <verb>` command, adding a new
  flag, changing output formatting for a command, or writing the RunE handler.
  Always read go-conventions/SKILL.md first.
---

# tkr — CLI Command Skill

> Prerequisite: **`.agent/skills/go-conventions/SKILL.md`** must be read before this skill.
> For the full behavioural spec of each command, see **`.agent/FUNCTIONAL_SPEC.md` §2**.

---

## 1. One File per Command

Each top-level command lives in its own file under `cmd/`:

```
cmd/
├── root.go       ← rootCmd, global flags, logger init
├── init.go       ← tkr init
├── watch.go      ← tkr watch [add|remove|list]
├── alert.go      ← tkr alert [add|list|remove|enable|disable|history]
├── quote.go      ← tkr quote
├── daemon.go     ← tkr daemon [start|stop|status|restart]
└── config.go     ← tkr config [show|set|validate]
```

Sub-commands of a parent (e.g. `watch add`) are defined in the same file as their parent.

---

## 2. Anatomy of a Command File

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/yourname/tkr/internal/config"
    "github.com/yourname/tkr/internal/db"
    // ... other internal packages as needed
)

// 1. Declare the top-level command variable.
var watchCmd = &cobra.Command{
    Use:   "watch",
    Short: "Manage your stock watchlist",
    Long: `Add, remove, and list stocks you want to monitor.

tkr will poll prices and evaluate alert rules for watched stocks.`,
}

// 2. Declare sub-commands.
var watchAddCmd = &cobra.Command{
    Use:     "add <TICKER>",
    Short:   "Add a stock to the watchlist",
    Args:    cobra.ExactArgs(1),
    Example: "  tkr watch add AAPL\n  tkr watch add CDR.WAR --exchange GPW",
    RunE:    runWatchAdd,
}

// 3. Declare flag variables at package level (not inside init).
var watchAddExchange string

// 4. Register everything in init().
func init() {
    // Register sub-commands on their parent.
    watchCmd.AddCommand(watchAddCmd)
    watchCmd.AddCommand(watchRemoveCmd)
    watchCmd.AddCommand(watchListCmd)

    // Register the parent on rootCmd.
    rootCmd.AddCommand(watchCmd)

    // Bind flags to their command (not to parent).
    watchAddCmd.Flags().StringVar(&watchAddExchange, "exchange", "", "Exchange ID (e.g. GPW, NYSE)")

    // Persistent flags on the parent are inherited by sub-commands.
    // watchCmd.PersistentFlags().StringVar(...)
}

// 5. The RunE handler — keep it thin.
func runWatchAdd(cmd *cobra.Command, args []string) error {
    ticker := args[0]

    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    repo, err := db.Open(cfg.Database.Path)
    if err != nil {
        return fmt.Errorf("open db: %w", err)
    }
    defer repo.Close()

    router := provider.NewRouter(cfg)

    // Delegate ALL logic to internal packages.
    stock, err := watchlist.Add(cmd.Context(), repo, router, ticker, watchAddExchange)
    if err != nil {
        return err  // errors are already wrapped with context
    }

    display.PrintStock(stock)
    return nil
}
```

---

## 3. Rules for RunE

**Do:**
- Load config with `config.Load()`
- Open DB with `db.Open(cfg.Database.Path)` and `defer repo.Close()`
- Call a function in `internal/` for all logic
- Return errors (Cobra prints them and exits non-zero)
- Use `cmd.Context()` and pass it to all internal calls

**Do not:**
- Put business logic in `RunE` — it belongs in `internal/`
- Call `os.Exit` — Cobra handles exit codes
- Call `fmt.Println` for structured data — use `internal/display`
- Access `os.Args` directly

---

## 4. Args Validation

Use Cobra's built-in validators. Never check `len(args)` manually inside `RunE`:

```go
Args: cobra.ExactArgs(1)          // exactly 1 positional arg
Args: cobra.MinimumNArgs(1)       // at least 1
Args: cobra.RangeArgs(1, 3)       // between 1 and 3
Args: cobra.NoArgs                // forbids positional args
```

For custom validation (e.g. valid ticker format):

```go
Args: func(cmd *cobra.Command, args []string) error {
    if !isValidTicker(args[0]) {
        return fmt.Errorf("invalid ticker format: %q", args[0])
    }
    return nil
},
```

---

## 5. Output Conventions

| Output type | Where it goes |
|---|---|
| Tabular data (quotes, watchlist, alert rules) | `internal/display.PrintTable(...)` |
| Single-item confirmation ("Added AAPL") | `fmt.Fprintf(cmd.OutOrStdout(), ...)` |
| Sparklines | `internal/display.PrintSparkline(...)` |
| Machine-readable JSON | `encoding/json` to `cmd.OutOrStdout()` |
| Errors | Return from `RunE` — never print directly |

**`--format` flag pattern** (used on `watch list`, `quote`):

```go
var outputFormat string
cmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format: table, json, csv")

// In RunE:
switch outputFormat {
case "json":
    return display.PrintJSON(cmd.OutOrStdout(), data)
case "csv":
    return display.PrintCSV(cmd.OutOrStdout(), data)
default:
    display.PrintTable(cmd.OutOrStdout(), data)
}
```

---

## 6. Error Messages for the User

Cobra prints the error returned from `RunE` with the prefix `Error:`. Make messages human-friendly:

```go
// GOOD — tells the user what to do
return fmt.Errorf("ticker %q not found on any configured provider — check your API keys with: tkr config validate", ticker)

// BAD — technical, gives no action
return apperrors.ErrTickerNotFound
```

If an error has a known sentinel type, wrap it with a user-friendly message:

```go
if errors.Is(err, apperrors.ErrTickerNotFound) {
    return fmt.Errorf("ticker %q was not found", ticker)
}
return fmt.Errorf("unexpected error: %w", err)
```

---

## 7. Prompting the User for Input

Only used when a decision is ambiguous (e.g. ticker exists on multiple exchanges). Use `fmt.Fscan` against `cmd.InOrStdin()` so tests can inject input:

```go
func promptExchangeChoice(in io.Reader, out io.Writer, stocks []models.Stock) (models.Stock, error) {
    fmt.Fprintln(out, "Multiple matches found. Choose one:")
    for i, s := range stocks {
        fmt.Fprintf(out, "  [%d] %s — %s (%s)\n", i+1, s.Ticker, s.Name, s.Exchange)
    }
    fmt.Fprint(out, "Enter number: ")

    var choice int
    if _, err := fmt.Fscan(in, &choice); err != nil || choice < 1 || choice > len(stocks) {
        return models.Stock{}, fmt.Errorf("invalid selection")
    }
    return stocks[choice-1], nil
}
```

---

## 8. Definition of Done

- [ ] Command file created in `cmd/`
- [ ] Registered in the correct `init()` (parent or root)
- [ ] All flags declared at package level, bound in `init()`
- [ ] `RunE` is thin — logic delegated to `internal/`
- [ ] `Example` field populated with realistic usage examples
- [ ] Output goes through `internal/display` (not `fmt.Println`)
- [ ] Errors are user-friendly strings, not raw sentinel values
- [ ] `go-conventions` quality gates pass (`make lint test build`)

---

## Related Skills

- If the command fetches market data → **`.agent/skills/provider-integration/SKILL.md`**
- If the command reads/writes the database → **`.agent/skills/database-layer/SKILL.md`**
- If the command involves alert rules → **`.agent/skills/alert-engine/SKILL.md`**
- For writing tests for the command → **`.agent/skills/testing/SKILL.md`**
