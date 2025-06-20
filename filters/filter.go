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
			f, err := NewPattern(chunk)
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
			f, err := NewPattern(chunk)
			if err != nil {
				return nil, err
			}
			filters = append(filters, f)
		}
		return NewOr(filters...), nil
	}
	return NewPattern(filter)
}
