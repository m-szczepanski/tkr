package display

import "errors"

// RenderSparkline renders a unicode sparkline for a numeric time-series.
func RenderSparkline(values []float64, width int) (string, error) {
	_, _ = values, width
	// TODO(M2-12): implement sparkline rendering.
	return "", errors.New("not implemented")
}
