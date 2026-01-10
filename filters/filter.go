package filters

import (
	"fmt"
	"strings"

	"github.com/scylladb-actions/get-version/version"
)

type Filter interface {
	Apply(versions version.Versions) version.Versions
}

type EmptyFilter struct{}

func (f EmptyFilter) Apply(versions version.Versions) version.Versions {
	return versions
}

func wrapErr(err error, f string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	res := fmt.Sprintf(f, args...)
	return fmt.Errorf(res+": %w", err)
}

func ParseFilterString(filter string) (Filter, error) {
	// TBD: add more filters: and, or, grouping
	if filter == "" {
		return EmptyFilter{}, nil
	}
	if strings.Contains(filter, " and ") {
		var filters []Filter
		for _, chunk := range strings.Split(filter, " and ") {
			f, err := parseFilterChunk(chunk)
			if err != nil {
				return nil, err
			}
			filters = append(filters, f)
		}
		return NewAnd(filters...), nil
	}
	if strings.Contains(filter, " or ") {
		var filters []Filter
		for _, chunk := range strings.Split(filter, " or ") {
			f, err := parseFilterChunk(chunk)
			if err != nil {
				return nil, err
			}
			filters = append(filters, f)
		}
		return NewOr(filters...), nil
	}
	return parseFilterChunk(filter)
}

// parseFilterChunk parses a single filter chunk (no "and"/"or" operators)
// Tries GlobalPosition first (if no dots), then Pattern
func parseFilterChunk(chunk string) (Filter, error) {
	// Check if it's a global position filter (no dots, starts with FIRST/LAST)
	if isGlobalPosition(chunk) {
		return NewGlobalPosition(chunk)
	}
	// Fall back to pattern parsing
	return NewPattern(chunk)
}
