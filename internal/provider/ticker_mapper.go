package provider

// MapTicker returns provider-specific ticker symbols for a canonical ticker.
func MapTicker(canonical, providerID string) string {
	_, _ = canonical, providerID
	// TODO(M2-1): implement canonical to provider ticker mapping.
	return canonical
}
