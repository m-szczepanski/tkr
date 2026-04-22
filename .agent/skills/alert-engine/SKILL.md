---
name: alert-engine
description: >
  Use this skill for all work involving the alert rule system in tkr:
  implementing or modifying the condition DSL parser, the alert evaluator, moving
  average calculations, cooldown logic, or alert event persistence. Triggers include:
  writing or changing alert condition parsing, implementing any Evaluate function,
  adding a new metric type or operator, debugging why an alert did or did not fire,
  or modifying how the scheduler calls the evaluator. Always read
  go-conventions/SKILL.md first.
---

# tkr — Alert Engine

> Prerequisite: **`.agent/skills/go-conventions/SKILL.md`** must be read before this skill.
> Full spec is in **`.agent/FUNCTIONAL_SPEC.md` §1.3–1.5 and §3.3**.

---

## 1. Conceptual Architecture

The alert system has three independent layers:

```
CLI / config
    ↓
[Parser]          internal/alert/parser.go
    ↓ models.Condition
[Evaluator]       internal/alert/evaluator.go   ← pure functions, no I/O
    ↓ bool
[Scheduler]       internal/scheduler/scheduler.go
    ↓ (on true + cooldown passed)
[Dispatcher]      internal/notifier/dispatcher.go
```

**The evaluator is pure.** It never touches the database, the clock, or the network.
The scheduler owns cooldown enforcement, DB persistence of fired events, and dispatch.

---

## 2. The Condition DSL

### Supported expressions (CLI → stored Condition)

| CLI string | Metric | Operator | Value | Notes |
|---|---|---|---|---|
| `price < 150` | `PRICE` | `LT` | `150` | |
| `price > 200` | `PRICE` | `GT` | `200` | |
| `price <= 99.5` | `PRICE` | `LTE` | `99.5` | |
| `change% > 5` | `CHANGE_PCT` | `GT` | `5` | percent, not decimal |
| `change% < -3` | `CHANGE_PCT` | `LT` | `-3` | negative for drops |
| `change > 2.5` | `CHANGE_ABS` | `GT` | `2.5` | absolute currency change |
| `volume > 5000000` | `VOLUME` | `GT` | `5000000` | |
| `ma20 cross_above ma50` | `MA_CROSS` | `CROSS_ABOVE` | `50` | period=20 |
| `ma20 cross_below ma50` | `MA_CROSS` | `CROSS_BELOW` | `50` | period=20 |

### Parser implementation (`internal/alert/parser.go`)

```go
// ParseCondition parses a condition expression string into a models.Condition.
// Returns apperrors.ErrInvalidCondition wrapping a descriptive message on failure.
func ParseCondition(expr string) (models.Condition, error) {
    // Normalise: lowercase, trim, collapse whitespace
    expr = strings.ToLower(strings.TrimSpace(expr))
    expr = whitespaceRe.ReplaceAllString(expr, " ")

    // Try each pattern in order:
    // 1. MA cross: "ma{N} cross_above|cross_below ma{M}"
    // 2. Price/change/volume: "{metric} {op} {value}"

    if m := maCrossRe.FindStringSubmatch(expr); m != nil {
        return parseMACross(m)
    }
    if m := simpleRe.FindStringSubmatch(expr); m != nil {
        return parseSimple(m)
    }
    return models.Condition{}, fmt.Errorf("%w: %q", apperrors.ErrInvalidCondition, expr)
}
```

**The parser must be tested exhaustively.** See §5 for the table-driven test pattern.

---

## 3. The Evaluator (`internal/alert/evaluator.go`)

### Simple conditions

```go
// Evaluate returns true if the quote satisfies the condition.
// It is a pure function — no I/O, no side effects, no time.Now().
func Evaluate(cond models.Condition, q models.Quote) bool {
    var actual float64
    switch cond.Metric {
    case models.MetricPrice:
        actual = q.Price
    case models.MetricChangeAbs:
        actual = q.Change
    case models.MetricChangePct:
        actual = q.ChangePct
    case models.MetricVolume:
        actual = float64(q.Volume)
    default:
        // MA_CROSS requires history — use EvaluateWithHistory
        return false
    }
    return applyOperator(cond.Operator, actual, cond.Value)
}

func applyOperator(op models.Operator, actual, threshold float64) bool {
    switch op {
    case models.OperatorLT:  return actual < threshold
    case models.OperatorLTE: return actual <= threshold
    case models.OperatorGT:  return actual > threshold
    case models.OperatorGTE: return actual >= threshold
    case models.OperatorEQ:  return math.Abs(actual-threshold) < 0.0001
    default:                 return false
    }
}
```

### MA Cross conditions

```go
// EvaluateWithHistory evaluates MA_CROSS conditions using historical quotes.
// quotes must be ordered most-recent-first (as returned by db.GetRecentQuotes).
// Returns false (not an error) if there is insufficient history.
func EvaluateWithHistory(cond models.Condition, current models.Quote, history []models.Quote) bool {
    if cond.Metric != models.MetricMACross {
        return Evaluate(cond, current)
    }
    if cond.Period == nil {
        return false
    }

    shortPeriod := *cond.Period           // e.g. 20
    longPeriod := int(cond.Value)         // e.g. 50

    // Need at least longPeriod+1 data points to detect a cross
    // (current state + one prior state)
    allQuotes := append([]models.Quote{current}, history...)
    if len(allQuotes) < longPeriod+1 {
        return false  // insufficient history — do not trigger
    }

    currentShortMA := calcMA(allQuotes, 0, shortPeriod)
    currentLongMA  := calcMA(allQuotes, 0, longPeriod)
    prevShortMA    := calcMA(allQuotes, 1, shortPeriod)
    prevLongMA     := calcMA(allQuotes, 1, longPeriod)

    switch cond.Operator {
    case models.OperatorCrossAbove:
        // Short MA crossed above long MA between previous and current bar
        return prevShortMA <= prevLongMA && currentShortMA > currentLongMA
    case models.OperatorCrossBelow:
        return prevShortMA >= prevLongMA && currentShortMA < currentLongMA
    }
    return false
}

// calcMA computes a simple moving average starting at offset in the quotes slice.
// quotes[0] = most recent. offset shifts the window back by N bars.
func calcMA(quotes []models.Quote, offset, period int) float64 {
    if offset+period > len(quotes) {
        return 0
    }
    sum := 0.0
    for i := offset; i < offset+period; i++ {
        sum += quotes[i].Price
    }
    return sum / float64(period)
}
```

---

## 4. Scheduler Integration (Cooldown & Persistence)

The scheduler in `internal/scheduler/scheduler.go` is responsible for:

1. **Cooldown check** before calling the evaluator:
```go
func shouldEvaluate(rule models.AlertRule, now time.Time) bool {
    if !rule.Active {
        return false
    }
    if rule.LastFired == nil {
        return true
    }
    return now.Sub(*rule.LastFired) >= time.Duration(rule.CooldownSeconds)*time.Second
}
```

2. **Calling the evaluator** (pure, no I/O):
```go
triggered := alert.EvaluateWithHistory(rule.Condition, quote, recentHistory)
```

3. **Persisting the event** and updating the rule if triggered:
```go
if triggered {
    event := models.AlertEvent{
        RuleID:      rule.ID,
        Ticker:      rule.Ticker,
        TriggeredAt: now,
        Price:       quote.Price,
        ChangePct:   quote.ChangePct,
        Message:     formatMessage(rule, quote),
    }
    repo.AddAlertEvent(ctx, event)

    rule.LastFired = &now
    if rule.OneShot {
        rule.Active = false
    }
    repo.UpdateAlertRule(ctx, rule)

    dispatcher.Send(ctx, event)
}
```

---

## 5. Testing the Alert Engine

### Parser — table-driven test

```go
func TestParseCondition(t *testing.T) {
    tests := []struct {
        input   string
        want    models.Condition
        wantErr bool
    }{
        {
            input: "price < 150",
            want:  models.Condition{Metric: models.MetricPrice, Operator: models.OperatorLT, Value: 150},
        },
        {
            input: "change% > 5",
            want:  models.Condition{Metric: models.MetricChangePct, Operator: models.OperatorGT, Value: 5},
        },
        {
            input: "ma20 cross_above ma50",
            want:  models.Condition{Metric: models.MetricMACross, Operator: models.OperatorCrossAbove, Value: 50, Period: intPtr(20)},
        },
        {
            input:   "invalid expression",
            wantErr: true,
        },
        {
            input:   "",
            wantErr: true,
        },
    }
    for _, tc := range tests {
        t.Run(tc.input, func(t *testing.T) {
            got, err := alert.ParseCondition(tc.input)
            if tc.wantErr {
                require.Error(t, err)
                assert.ErrorIs(t, err, apperrors.ErrInvalidCondition)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tc.want, got)
        })
    }
}
```

### Evaluator — boundary conditions to always test

```go
// For each metric/operator combination, test:
// 1. Value strictly below threshold (LT should be true, GT false)
// 2. Value exactly at threshold (EQ true; LT false; GT false)
// 3. Value strictly above threshold (GT true; LT false)
// 4. Zero value from bad provider data (should not panic)
// 5. MA_CROSS with insufficient history (should return false, not error)
```

See **`.agent/skills/testing/SKILL.md`** for the full table-driven pattern.

---

## 6. Definition of Done

- [ ] Parser handles all supported expressions from §2
- [ ] Parser returns `apperrors.ErrInvalidCondition` (not a generic error) for invalid input
- [ ] Evaluator is pure (no `time.Now()`, no DB access, no I/O)
- [ ] `EvaluateWithHistory` returns `false` (not an error) when history is insufficient
- [ ] Cooldown logic lives in the scheduler, not the evaluator
- [ ] Table-driven tests cover all metrics, all operators, and edge cases
- [ ] `go-conventions` quality gates pass (`make lint test build`)
