package alert

import (
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// ParseCondition parses condition DSL expressions.
func ParseCondition(expr string) (models.Condition, error) {
	_ = expr
	// TODO(M4-1): implement alert condition parser.
	return models.Condition{}, errors.New("not implemented")
}
