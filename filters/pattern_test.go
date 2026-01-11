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

func TestPatternWithRegexAndLAST(t *testing.T) {
	// This test reproduces the issue where chaining a regex filter with *.*.LAST
	// returns incorrect results when versions span multiple major versions
	versions := version.Versions{
		version.NewMust("7.9.0"),
		version.NewMust("7.9.1"),
		version.NewMust("7.9.2"),
		version.NewMust("7.9.3"),
		version.NewMust("7.9.4"),
		version.NewMust("7.9.5"),
		version.NewMust("7.9.0-1-ubi8"),
		version.NewMust("7.9.1-1-ubi8"),
		version.NewMust("8.0.0"),
		version.NewMust("8.0.1"),
		version.NewMust("8.0.2"),
		version.NewMust("8.0.3"),
		version.NewMust("8.0.0-1-ubi9"),
		version.NewMust("8.0.1-1-ubi9"),
		version.NewMust("8.1.0"),
		version.NewMust("8.1.1"),
		version.NewMust("8.1.0-1-ubi9"),
		version.NewMust("8.1.1-1-ubi9"),
	}

	// Step 1: Filter to only versions with pure numeric patches (no suffixes)
	regexFilter := NewPatternMust("*.*.^[0-9]+$")
	filteredStep1 := regexFilter.Apply(versions)
	expectedStep1 := version.Versions{
		version.NewMust("7.9.0"),
		version.NewMust("7.9.1"),
		version.NewMust("7.9.2"),
		version.NewMust("7.9.3"),
		version.NewMust("7.9.4"),
		version.NewMust("7.9.5"),
		version.NewMust("8.0.0"),
		version.NewMust("8.0.1"),
		version.NewMust("8.0.2"),
		version.NewMust("8.0.3"),
		version.NewMust("8.1.0"),
		version.NewMust("8.1.1"),
	}
	if !slices.Equal(expectedStep1, filteredStep1) {
		t.Fatalf("Step 1 failed: expected %s got %s", expectedStep1, filteredStep1)
	}

	// Step 2: Get LAST patch for each major.minor combination
	lastFilter := NewPatternMust("*.*.LAST")
	filteredStep2 := lastFilter.Apply(filteredStep1)
	expectedStep2 := version.Versions{
		version.NewMust("7.9.5"), // Latest patch for 7.9
		version.NewMust("8.0.3"), // Latest patch for 8.0
		version.NewMust("8.1.1"), // Latest patch for 8.1
	}
	if !slices.Equal(expectedStep2, filteredStep2) {
		t.Fatalf("Step 2 failed: expected %s got %s", expectedStep2, filteredStep2)
	}

	// Step 3: Get the absolute LAST version from the list
	globalLastFilter, err := NewGlobalPosition("LAST")
	if err != nil {
		t.Fatal(err)
	}
	filteredStep3 := globalLastFilter.Apply(filteredStep2)
	expectedStep3 := version.Versions{
		version.NewMust("8.1.1"), // Should be 8.1.1, the newest version
	}
	if !slices.Equal(expectedStep3, filteredStep3) {
		t.Fatalf("Step 3 failed: expected %s got %s", expectedStep3, filteredStep3)
	}

	// Full chain test
	filter, err := ParseFilterString("*.*.^[0-9]+$ and *.*.LAST and LAST")
	if err != nil {
		t.Fatal(err)
	}
	result := filter.Apply(versions)
	if !slices.Equal(expectedStep3, result) {
		t.Fatalf("Full chain failed: expected %s got %s", expectedStep3, result)
	}
}
