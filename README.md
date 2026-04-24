# tkr

> A lightweight, terminal-native stock monitoring and alerting system written in Go.

tkr lets you track stocks from multiple global exchanges — including the US markets, Polish GPW/WIG, and major international indices — right from your command line. Configure price thresholds, percentage change alerts, and volume spikes, then let the daemon watch the markets for you and fire off notifications when your conditions are met.

## Development Approach
This project was developed using **Agentic AI workflows** to accelerate the development cycle and explore modern AI implementation patterns. 

Key focus areas included:
* **Agentic Frameworks:** Utilizing tools like `Roo Code` for autonomous task execution.
* **Architecture:** Implementing sub-agents and specialized "Skills" to handle complex logic.
* **Learning Objective:** Practical hands-on experience with multi-agent orchestration and prompt engineering in a production-like environment.

## Features

- **Multi-exchange support** — US (NYSE, NASDAQ), Polish GPW/WIG, Frankfurt, London, and more
- **Flexible alert rules** — price above/below, % change, volume threshold, moving average cross
- **Local persistent storage** — SQLite-backed watchlist and alert history, no cloud required
- **Pure CLI interface** — scriptable, tmux-friendly, no GUI dependencies
- **Background daemon mode** — runs silently, polls on a configurable schedule
- **Notification channels** — terminal desktop notification, email, and webhook (Slack/Discord)
- **Inline sparklines** — ASCII charts of recent price history directly in the terminal
- **Multi-provider API routing** — falls back to secondary providers when primary is rate-limited
- **Timezone-aware** — market hours respected per exchange; alerts only fire when the market is open (configurable)

## Tech Stack

| Layer | Choice |
| --- | --- |
| Language | Go 1.25+ |
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| Config | [Viper](https://github.com/spf13/viper) (YAML/TOML/env) |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| HTTP client | `net/http` + `resty` |
| Scheduler | `robfig/cron` |
| Notifications | `gen2brain/beeep` (desktop) + SMTP + webhooks |
| Testing | `testify` |
| Logging | `zerolog` |

## Quick Start

```bash

# Install

go install github.com/yourname/tkr@latest

# Initialise config and local database

tkr init

# Add stocks to your watchlist

tkr watch add AAPL # US stock

tkr watch add CDR.WAR # CD Projekt Red on GPW

tkr watch add VOW3.FRA # VW on Frankfurt

# List your watchlist

tkr watch list

# Set an alert: notify when AAPL drops below $170

tkr alert add AAPL --condition "price < 170" --channel terminal

# Set a % change alert

tkr alert add CDR.WAR --condition "change% > 5" --channel email

# Check current prices manually

tkr quote AAPL CDR.WAR VOW3.FRA

# Start the background daemon (polls every 5 minutes)

tkr daemon start

# View recent alert history

tkr alert history

```

## Configuration

tkr looks for its config at `~/.config/tkr/config.yaml`:

```yaml

polling_interval: 5m

market_hours_only: true

log_level: info

providers:

alphavantage:

api_key: YOUR_KEY

finnhub:

api_key: YOUR_KEY

stooq:

enabled: true # free, no key needed

notifications:

email:

smtp_host: smtp.gmail.com

smtp_port: 587

from: you@gmail.com

password: YOUR_APP_PASSWORD

to:

- you@gmail.com

webhook:

url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL

database:

path: ~/.local/share/tkr/data.db
```

## Supported Exchanges

| Exchange | Ticker Format | Primary Provider | Notes |
| --- | --- | --- | --- |
| NYSE / NASDAQ (US) | `AAPL`, `TSLA` | Finnhub / Alpha Vantage | Real-time with paid tier |
| GPW Warsaw (PL) | `CDR.WAR`, `PKN.WAR` | Stooq / GPW API | 15 min delay on free tier |
| Frankfurt (XETRA) | `VOW3.FRA` | Yahoo Finance / EODHD | |
| London Stock Exchange | `BP.LON` | Yahoo Finance / EODHD | |
| Euronext | `AIR.EPA` (Paris) | Yahoo Finance / EODHD | |
| Global indices | `^GSPC`, `^WIG20` | Alpha Vantage / Stooq | Read-only, no alerts on indices |

## Project Structure

```
tkr/

├── cmd/ # Cobra command definitions

│ ├── root.go

│ ├── watch.go

│ ├── alert.go

│ ├── quote.go

│ ├── daemon.go

│ └── init.go

├── internal/

│ ├── config/ # Viper config loading

│ ├── db/ # SQLite repository layer

│ ├── provider/ # API provider implementations + router

│ │ ├── alphavantage/

│ │ ├── finnhub/

│ │ ├── stooq/

│ │ ├── yahoofinance/

│ │ └── eodhd/

│ ├── alert/ # Alert engine and rule evaluator

│ ├── notifier/ # Notification channel implementations

│ ├── scheduler/ # Cron-based polling scheduler

│ └── display/ # Terminal output, sparklines, tables

├── pkg/

│ └── models/ # Shared domain types (Quote, Alert, Stock…)

├── migrations/ # SQL schema migrations

├── config.example.yaml

├── go.mod

└── README.md

```

## License

MIT — see [LICENSE](LICENSE).
