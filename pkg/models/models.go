package models

import "time"

type ExchangeID string
type ProviderID string
type ChannelID string
type Metric string
type Operator string

const (
	ExchangeNYSE     ExchangeID = "NYSE"
	ExchangeNASDAQ   ExchangeID = "NASDAQ"
	ExchangeGPW      ExchangeID = "GPW"
	ExchangeXETRA    ExchangeID = "XETRA"
	ExchangeLSE      ExchangeID = "LSE"
	ExchangeEuronext ExchangeID = "EURONEXT"
)

const (
	MetricPrice     Metric = "PRICE"
	MetricChangeAbs Metric = "CHANGE_ABS"
	MetricChangePct Metric = "CHANGE_PCT"
	MetricVolume    Metric = "VOLUME"
	MetricMACross   Metric = "MA_CROSS"
)

const (
	OperatorLT         Operator = "LT"
	OperatorLTE        Operator = "LTE"
	OperatorGT         Operator = "GT"
	OperatorGTE        Operator = "GTE"
	OperatorEQ         Operator = "EQ"
	OperatorCrossAbove Operator = "CROSS_ABOVE"
	OperatorCrossBelow Operator = "CROSS_BELOW"
)

const (
	ChannelTerminal ChannelID = "terminal"
	ChannelEmail    ChannelID = "email"
	ChannelWebhook  ChannelID = "webhook"
)

type Stock struct {
	Ticker   string     `db:"ticker" json:"ticker"`
	Name     string     `db:"name" json:"name"`
	Exchange ExchangeID `db:"exchange" json:"exchange"`
	Currency string     `db:"currency" json:"currency"`
	AddedAt  time.Time  `db:"added_at" json:"added_at"`
}

type Quote struct {
	Ticker    string    `db:"ticker" json:"ticker"`
	Price     float64   `db:"price" json:"price"`
	Open      float64   `db:"open" json:"open"`
	High      float64   `db:"high" json:"high"`
	Low       float64   `db:"low" json:"low"`
	Close     float64   `db:"close" json:"close"`
	Volume    int64     `db:"volume" json:"volume"`
	Change    float64   `db:"change" json:"change"`
	ChangePct float64   `db:"change_pct" json:"change_pct"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	Source    string    `db:"source" json:"source"`
}

type OHLCV struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

type Condition struct {
	Metric   Metric
	Operator Operator
	Value    float64
	Period   *int // only for MA_CROSS
}

type AlertRule struct {
	ID              int64
	Ticker          string
	Condition       Condition
	Channels        []ChannelID
	Active          bool
	OneShot         bool
	CooldownSeconds int
	CreatedAt       time.Time
	LastFired       *time.Time
}

type AlertEvent struct {
	ID          int64
	RuleID      int64
	Ticker      string
	TriggeredAt time.Time
	Price       float64
	ChangePct   float64
	Message     string
}

type AlertEventFilter struct {
	Ticker string
	Since  *time.Time
	Limit  int
}
