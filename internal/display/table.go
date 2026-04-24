package display

import (
	"errors"
	"io"

	"github.com/yourname/tkr/pkg/models"
)

// RenderQuoteTable renders quote rows in terminal-friendly table format.
func RenderQuoteTable(w io.Writer, quotes []models.Quote) error {
	_, _ = w, quotes
	// TODO(M2-11): implement table rendering.
	return errors.New("not implemented")
}
