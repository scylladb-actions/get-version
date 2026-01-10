package filters

import (
	"slices"
	"testing"

	"github.com/scylladb-actions/get-version/version"
)

// TestGlobalPosition tests the GlobalPosition filter which selects a single version
// from the entire version list based on position (LAST, LAST-N, FIRST, FIRST+N).
//
// This filter differs from Pattern filters (e.g., LAST.LAST.LAST) which operate
// on version components. GlobalPosition operates on the entire sorted list,
// making it useful for selecting specific versions regardless of major version changes.
//
// For example: "LAST-1" will always return the second newest version from the list,
// even when major versions change (7.9.1 → 8.0.0).
func TestGlobalPosition(t *testing.T) {
	// Test dataset: 8 versions across 3 major versions
	// This allows testing both forward (FIRST) and backward (LAST) selection
	versions := version.Versions{
		version.NewMust("1.1.0"),
		version.NewMust("1.1.1"),
		version.NewMust("1.2.0"),
		version.NewMust("2.1.0"),
		version.NewMust("2.2.0"),
		version.NewMust("3.1.0"),
		version.NewMust("3.2.0"),
		version.NewMust("3.3.0"), // LAST (index 7)
	}

	tcases := []struct {
		name     string
		filter   string
		expected version.Versions
		wantErr  bool
	}{
		{
			// Basic LAST: Should select the highest version (3.3.0)
			name:   "LAST",
			filter: "LAST",
			expected: version.Versions{
				version.NewMust("3.3.0"),
			},
		},
		{
			// LAST-1: Should select second-to-last version (3.2.0)
			// This is the primary use case for surviving major version changes
			name:   "LAST-1",
			filter: "LAST-1",
			expected: version.Versions{
				version.NewMust("3.2.0"),
			},
		},
		{
			// LAST-3: Test offset selection further back (2.2.0)
			name:   "LAST-3",
			filter: "LAST-3",
			expected: version.Versions{
				version.NewMust("2.2.0"),
			},
		},
		{
			// Basic FIRST: Should select the lowest version (1.1.0)
			name:   "FIRST",
			filter: "FIRST",
			expected: version.Versions{
				version.NewMust("1.1.0"),
			},
		},
		{
			// FIRST+1: Should select second version from start (1.1.1)
			name:   "FIRST+1",
			filter: "FIRST+1",
			expected: version.Versions{
				version.NewMust("1.1.1"),
			},
		},
		{
			// FIRST+3: Test offset selection further forward (2.1.0)
			name:   "FIRST+3",
			filter: "FIRST+3",
			expected: version.Versions{
				version.NewMust("2.1.0"),
			},
		},
		{
			// Out of bounds: LAST-99 exceeds available versions
			// Should return empty slice (graceful handling, not an error)
			name:     "LAST-99 (out of bounds)",
			filter:   "LAST-99",
			expected: version.Versions{},
		},
		{
			// Out of bounds: FIRST+99 exceeds available versions
			// Should return empty slice (graceful handling, not an error)
			name:     "FIRST+99 (out of bounds)",
			filter:   "FIRST+99",
			expected: version.Versions{},
		},
		{
			// Invalid syntax: LAST only accepts '-' operator
			// Using '+' should be a validation error
			name:    "Invalid - LAST+1",
			filter:  "LAST+1",
			wantErr: true,
		},
		{
			// Invalid syntax: FIRST only accepts '+' operator
			// Using '-' should be a validation error
			name:    "Invalid - FIRST-1",
			filter:  "FIRST-1",
			wantErr: true,
		},
		{
			// Invalid: Bare number without FIRST/LAST keyword
			// Should fail validation
			name:    "Invalid - no keyword",
			filter:  "5",
			wantErr: true,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.name, func(t *testing.T) {
			gp, err := NewGlobalPosition(tcase.filter)
			if tcase.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := gp.Apply(versions)
			if !slices.Equal(tcase.expected, got) {
				t.Fatalf("expected %s got %s", tcase.expected, got)
			}
		})
	}
}

// TestGlobalPositionWithChaining tests that GlobalPosition filters work correctly
// when chained with Pattern filters using the 'and' operator.
//
// The 'and' operator chains filters sequentially:
//  1. First filter is applied to reduce the version list
//  2. Second filter is applied to the results of the first filter
//
// This is the key use case for GlobalPosition: filtering versions by pattern,
// then selecting a specific position from the filtered results. This behavior
// survives major version changes because the position is relative to the filtered
// list, not to version components.
//
// Example scenario:
//
//	Versions: [1.1.1, 1.2.1, 2.1.1, 2.2.1, 3.1.1]
//	Filter: "[0-9]+.[0-9]+.LAST and LAST-1"
//	Step 1: [0-9]+.[0-9]+.LAST → [1.2.1, 2.2.1, 3.1.1] (latest patch per major.minor)
//	Step 2: LAST-1 → [2.2.1] (second-to-last from filtered list)
//
//	When 4.0.0 is released:
//	Versions: [1.1.1, 1.2.1, 2.1.1, 2.2.1, 3.1.1, 4.0.0]
//	Filter: "[0-9]+.[0-9]+.LAST and LAST-1"
//	Step 1: [0-9]+.[0-9]+.LAST → [1.2.1, 2.2.1, 3.1.1, 4.0.0]
//	Step 2: LAST-1 → [3.1.1] (still works! Returns second-to-last)
func TestGlobalPositionWithChaining(t *testing.T) {
	// Test dataset: 10 versions with both .0 and .1 patches
	// This allows testing patterns that reduce the list before global position selection
	versions := version.Versions{
		version.NewMust("1.1.0"),
		version.NewMust("1.1.1"),
		version.NewMust("1.2.0"),
		version.NewMust("1.2.1"),
		version.NewMust("2.1.0"),
		version.NewMust("2.1.1"),
		version.NewMust("2.2.0"),
		version.NewMust("2.2.1"),
		version.NewMust("3.1.0"),
		version.NewMust("3.1.1"),
	}

	tcases := []struct {
		name     string
		filter   string
		expected version.Versions
	}{
		{
			// Filter to versions ending in .1, then select the newest
			// *.*.1 → [1.1.1, 1.2.1, 2.1.1, 2.2.1, 3.1.1]
			// LAST → [3.1.1]
			name:   "Pattern then GlobalPosition: *.*.1 and LAST",
			filter: "*.*.1 and LAST",
			expected: version.Versions{
				version.NewMust("3.1.1"),
			},
		},
		{
			// Filter to major version 2, then select the oldest
			// 2.*.* → [2.1.0, 2.1.1, 2.2.0, 2.2.1]
			// FIRST → [2.1.0]
			name:   "Pattern then GlobalPosition: 2.*.* and FIRST",
			filter: "2.*.* and FIRST",
			expected: version.Versions{
				version.NewMust("2.1.0"),
			},
		},
		{
			// Filter to latest major version, then select second-to-last
			// LAST.*.* → [3.1.0, 3.1.1]
			// LAST-1 → [3.1.0]
			name:   "Pattern then GlobalPosition: LAST.*.* and LAST-1",
			filter: "LAST.*.* and LAST-1",
			expected: version.Versions{
				version.NewMust("3.1.0"),
			},
		},
		{
			// THE PRIMARY USE CASE: Get latest patch per major.minor, then select second-to-last
			// This pattern survives major version changes!
			// [0-9]+.[0-9]+.LAST → [1.2.1, 2.2.1, 3.1.1] (latest patch for each major.minor)
			// LAST-1 → [2.2.1] (second-to-last from filtered list)
			//
			// When a new major version (e.g., 4.0.0) is released, this still works:
			// [0-9]+.[0-9]+.LAST → [1.2.1, 2.2.1, 3.1.1, 4.0.0]
			// LAST-1 → [3.1.1] (still returns second-to-last!)
			name:   "Multiple global positions: [0-9]+.[0-9]+.LAST and LAST-1",
			filter: "[0-9]+.[0-9]+.LAST and LAST-1",
			expected: version.Versions{
				version.NewMust("2.2.1"),
			},
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.name, func(t *testing.T) {
			filter, err := ParseFilterString(tcase.filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := filter.Apply(versions)
			if !slices.Equal(tcase.expected, got) {
				t.Fatalf("expected %s got %s", tcase.expected, got)
			}
		})
	}
}

// TestIsGlobalPosition tests the detection function that determines whether
// a filter string should be parsed as a GlobalPosition filter or a Pattern filter.
//
// Detection rules:
//   - Must NOT contain dots (dots indicate a Pattern like "1.2.3" or "*.*.LAST")
//   - Must start with "FIRST" or "LAST" keyword
//
// This distinction is critical for ParseFilterString to route the filter string
// to the correct parser (NewGlobalPosition vs NewPattern).
func TestIsGlobalPosition(t *testing.T) {
	tcases := []struct {
		input    string
		expected bool
	}{
		// Valid global position patterns (no dots, starts with keyword)
		{"LAST", true},
		{"LAST-1", true},
		{"FIRST", true},
		{"FIRST+2", true},

		// Pattern filters (contain dots) - should NOT be detected as global position
		{"1.2.3", false},    // Literal pattern
		{"*.*.LAST", false}, // Pattern with LAST in patch position

		// Invalid inputs - should NOT be detected as global position
		{"", false},       // Empty string
		{"RANDOM", false}, // Wrong keyword (no dots but doesn't start with FIRST/LAST)
		{"5", false},      // Bare number (no keyword)
	}

	for _, tcase := range tcases {
		t.Run(tcase.input, func(t *testing.T) {
			got := isGlobalPosition(tcase.input)
			if got != tcase.expected {
				t.Fatalf("expected %v got %v for input %q", tcase.expected, got, tcase.input)
			}
		})
	}
}
