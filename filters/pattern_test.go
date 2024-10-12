package filters

import (
	"slices"
	"testing"

	"github.com/scylladb-actions/get-version/version"
)

func TestPattern(t *testing.T) {
	versions := version.Versions{
		version.NewMust("1.1.0"),
		version.NewMust("1.1.1"),
		version.NewMust("1.1.2"),
		version.NewMust("1.1.3"),
		version.NewMust("1.2.0"),
		version.NewMust("1.2.1"),
		version.NewMust("1.2.2"),
		version.NewMust("1.2.3"),
		version.NewMust("1.3.0"),
		version.NewMust("1.3.1"),
		version.NewMust("1.3.2"),
		version.NewMust("1.3.3"),
		version.NewMust("2.1.0"),
		version.NewMust("2.1.1"),
		version.NewMust("2.1.2"),
		version.NewMust("2.1.3"),
		version.NewMust("2.2.0"),
		version.NewMust("2.2.1"),
		version.NewMust("2.2.2"),
		version.NewMust("2.2.3"),
		version.NewMust("2.3.0"),
		version.NewMust("2.3.1"),
		version.NewMust("2.3.2"),
		version.NewMust("2.3.3"),
		version.NewMust("3.1.0"),
		version.NewMust("3.1.1"),
		version.NewMust("3.1.2"),
		version.NewMust("3.1.3"),
		version.NewMust("3.2.0"),
		version.NewMust("3.2.1"),
		version.NewMust("3.2.3"),
		version.NewMust("3.3.0"),
		version.NewMust("3.3.1"),
		version.NewMust("3.3.2"),
	}

	tcases := []struct {
		pattern  Pattern
		expected version.Versions
	}{
		{
			pattern: NewPatternMust("1.1.*"),
			expected: version.Versions{
				version.NewMust("1.1.0"),
				version.NewMust("1.1.1"),
				version.NewMust("1.1.2"),
				version.NewMust("1.1.3"),
			},
		},
		{
			pattern: NewPatternMust("1.*.1"),
			expected: version.Versions{
				version.NewMust("1.1.1"),
				version.NewMust("1.2.1"),
				version.NewMust("1.3.1"),
			},
		},
		{
			pattern: NewPatternMust("*.1.1"),
			expected: version.Versions{
				version.NewMust("1.1.1"),
				version.NewMust("2.1.1"),
				version.NewMust("3.1.1"),
			},
		},
		{
			pattern: NewPatternMust("*.*.1"),
			expected: version.Versions{
				version.NewMust("1.1.1"),
				version.NewMust("1.2.1"),
				version.NewMust("1.3.1"),
				version.NewMust("2.1.1"),
				version.NewMust("2.2.1"),
				version.NewMust("2.3.1"),
				version.NewMust("3.1.1"),
				version.NewMust("3.2.1"),
				version.NewMust("3.3.1"),
			},
		},
		{
			pattern: NewPatternMust("1.1.LAST"),
			expected: version.Versions{
				version.NewMust("1.1.3"),
			},
		},
		{
			pattern: NewPatternMust("1.1.LAST-1"),
			expected: version.Versions{
				version.NewMust("1.1.2"),
			},
		},
		{
			pattern: NewPatternMust("1.LAST.1"),
			expected: version.Versions{
				version.NewMust("1.3.1"),
			},
		},
		{
			pattern: NewPatternMust("1.LAST-1.1"),
			expected: version.Versions{
				version.NewMust("1.2.1"),
			},
		},
		{
			pattern: NewPatternMust("LAST.1.1"),
			expected: version.Versions{
				version.NewMust("3.1.1"),
			},
		},
		{
			pattern: NewPatternMust("LAST-1.1.1"),
			expected: version.Versions{
				version.NewMust("2.1.1"),
			},
		},
		{
			pattern: NewPatternMust("LAST.LAST.LAST"),
			expected: version.Versions{
				version.NewMust("3.3.2"),
			},
		},
		{
			pattern: NewPatternMust("LAST.*.LAST"),
			expected: version.Versions{
				version.NewMust("3.1.3"),
				version.NewMust("3.2.3"),
				version.NewMust("3.3.2"),
			},
		},
		{
			pattern: NewPatternMust("LAST.*.LAST-1"),
			expected: version.Versions{
				version.NewMust("3.1.2"),
				version.NewMust("3.2.1"),
				version.NewMust("3.3.1"),
			},
		},
		{
			pattern: NewPatternMust("LAST.LAST.*"),
			expected: version.Versions{
				version.NewMust("3.3.0"),
				version.NewMust("3.3.1"),
				version.NewMust("3.3.2"),
			},
		},
		{
			pattern: NewPatternMust("*.LAST.LAST"),
			expected: version.Versions{
				version.NewMust("1.3.3"),
				version.NewMust("2.3.3"),
				version.NewMust("3.3.2"),
			},
		},
		{
			pattern: NewPatternMust("*.LAST-1.LAST-1"),
			expected: version.Versions{
				version.NewMust("1.2.2"),
				version.NewMust("2.2.2"),
				version.NewMust("3.2.1"),
			},
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.pattern.String(), func(t *testing.T) {
			got := tcase.pattern.Apply(versions)
			if !slices.Equal(tcase.expected, got) {
				t.Fatalf("expected %s got %s", tcase.expected, got)
			}
		})
	}
}
