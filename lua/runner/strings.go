package runner

import (
	"math"
	"strings"
)

// Dedent removes the common leading whitespace from all lines in a string.
// Empty lines (or lines with only whitespace) are ignored when calculating
// the common indent, but are preserved in the output (as empty lines).
func Dedent(s string) string {
	lines := strings.Split(s, "\n")

	// Find the minimum indent among all non-empty lines
	minIndent := math.MaxInt
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			// Skip empty or whitespace-only lines
			continue
		}
		indent := len(line) - len(trimmed)
		if indent < minIndent {
			minIndent = indent
		}
	}

	// If no non-empty lines found, or no common indent, return as-is
	if minIndent == math.MaxInt || minIndent == 0 {
		// Still trim leading/trailing empty lines
		return strings.TrimSpace(s)
	}

	// Remove the common indent from each line
	result := make([]string, len(lines))
	for i, line := range lines {
		if len(line) >= minIndent {
			result[i] = line[minIndent:]
		} else {
			// For lines shorter than minIndent (e.g., empty or whitespace-only),
			// just trim all whitespace
			result[i] = strings.TrimLeft(line, " \t")
		}
	}

	// Join and trim leading/trailing empty lines
	return strings.TrimSpace(strings.Join(result, "\n"))
}
