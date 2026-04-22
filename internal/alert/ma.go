package alert

import (
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// CalculateMA calculates a simple moving average from quote history.
func CalculateMA(history []models.Quote, period int) (float64, error) {
	_, _ = history, period
	// TODO(M5-3): implement moving average calculator.
	return 0, errors.New("not implemented")
}
