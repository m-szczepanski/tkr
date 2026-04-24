package alert

import "github.com/yourname/tkr/pkg/models"

// Evaluate evaluates non-history-based conditions.
func Evaluate(cond models.Condition, quote models.Quote) bool {
	_, _ = cond, quote
	// TODO(M5-1): implement pure alert evaluator.
	return false
}

// EvaluateWithHistory evaluates conditions requiring historical data.
func EvaluateWithHistory(cond models.Condition, current models.Quote, history []models.Quote) bool {
	_, _, _ = cond, current, history
	// TODO(M5-1): implement historical alert evaluator.
	return false
}
