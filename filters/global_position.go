package filters

import (
	"fmt"
	"strings"

	"github.com/scylladb-actions/get-version/version"
)

// GlobalPosition filter selects a single version from the entire version list
// based on position: LAST, LAST-N, FIRST, FIRST+N
type GlobalPosition struct {
	keyword string // "LAST" or "FIRST"
	offset  int    // The N in LAST-N or FIRST+N (0 for bare LAST/FIRST)
}

// NewGlobalPosition creates a GlobalPosition filter from a string like "LAST", "LAST-1", "FIRST", "FIRST+2"
func NewGlobalPosition(value string) (GlobalPosition, error) {
	gp := GlobalPosition{}

	// Determine if it's FIRST or LAST
	switch {
	case len(value) >= 5 && value[:5] == "FIRST":
		gp.keyword = "FIRST"
		offset, err := getIdx(value, "FIRST", '+')
		if err != nil {
			return GlobalPosition{}, fmt.Errorf("invalid global position filter %q: %w", value, err)
		}
		gp.offset = offset
	case len(value) >= 4 && value[:4] == "LAST":
		gp.keyword = "LAST"
		offset, err := getIdx(value, "LAST", '-')
		if err != nil {
			return GlobalPosition{}, fmt.Errorf("invalid global position filter %q: %w", value, err)
		}
		gp.offset = offset
	default:
		return GlobalPosition{}, fmt.Errorf("invalid global position filter %q: must start with FIRST or LAST", value)
	}

	return gp, nil
}

// Apply implements the Filter interface
// Returns a single version (or empty slice if offset is out of bounds)
func (f GlobalPosition) Apply(versions version.Versions) version.Versions {
	if len(versions) == 0 {
		return versions
	}

	// Sort versions (ascending order)
	sorted := versions.Order(false)

	var index int
	if f.keyword == "FIRST" {
		index = f.offset
	} else { // "LAST"
		index = len(sorted) - 1 - f.offset
	}

	// Check bounds
	if index < 0 || index >= len(sorted) {
		return version.Versions{}
	}

	// Return single version
	return version.Versions{sorted[index]}
}

// String returns a string representation of the filter
func (f GlobalPosition) String() string {
	if f.offset == 0 {
		return f.keyword
	}
	if f.keyword == "FIRST" {
		return fmt.Sprintf("FIRST+%d", f.offset)
	}
	return fmt.Sprintf("LAST-%d", f.offset)
}

// isGlobalPosition checks if a filter string matches global position syntax
// (no dots and starts with FIRST or LAST)
func isGlobalPosition(filter string) bool {
	// Must not contain dots (Pattern filters always have dots)
	if len(filter) == 0 {
		return false
	}

	// Check if it starts with FIRST or LAST
	hasFirst := len(filter) >= 5 && filter[:5] == "FIRST"
	hasLast := len(filter) >= 4 && filter[:4] == "LAST"

	return !strings.Contains(filter, ".") && (hasFirst || hasLast)
}
