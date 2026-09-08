package axi

const DefaultHintLimit = 3

// LimitHints removes empty and duplicate suggestions while preserving order.
func LimitHints(limit int, hints ...string) []string {
	if limit <= 0 {
		limit = DefaultHintLimit
	}

	result := make([]string, 0, limit)
	seen := make(map[string]struct{}, len(hints))
	for _, hint := range hints {
		if hint == "" {
			continue
		}
		if _, exists := seen[hint]; exists {
			continue
		}
		seen[hint] = struct{}{}
		result = append(result, hint)
		if len(result) == limit {
			break
		}
	}
	return result
}
