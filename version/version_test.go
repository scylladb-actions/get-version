package version_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/scylladb-actions/get-version/version"
)

func TestVersion(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		tcases := []struct {
			value string
			major int
			minor int
			patch int
			extra string
			err   error
		}{
			{
				value: "1.2",
				major: 1,
				minor: 2,
				patch: 0,
				extra: "",
			},
			{
				value: "1.2.3",
				major: 1,
				minor: 2,
				patch: 3,
				extra: "",
			},
			{
				value: "1.2.3-dev",
				major: 1,
				minor: 2,
				patch: 3,
				extra: "-dev",
			},
			{
				value: "1.2.dev",
				major: 1,
				minor: 2,
				patch: 0,
				extra: "-dev",
				err:   fmt.Errorf("patch \"dev\" does not match patch format: [0-9._a-z-]+"),
			},
		}

		for _, tcase := range tcases {
			t.Run(tcase.value, func(t *testing.T) {
				got, err := version.New(tcase.value)
				if err != nil {
					if tcase.err == nil || tcase.err.Error() != err.Error() {
						t.Fatalf("expected error %q, but got %q", tcase.err, err)
					}
				} else {
					if tcase.major != got.Major() {
						t.Errorf("expected major error %d, but got %d", tcase.major, got.Major())
					}
					if tcase.minor != got.Minor() {
						t.Errorf("expected minor error %d, but got %d", tcase.minor, got.Minor())
					}
					if tcase.patch != got.Patch() {
						t.Errorf("expected patch error %d, but got %d", tcase.patch, got.Patch())
					}
					if tcase.extra != got.Extra() {
						t.Errorf("expected extra error %q, but got %q", tcase.extra, got.Extra())
					}
				}
			})
		}
	})

	t.Run("Sort", func(t *testing.T) {
		versions := version.Versions{
			version.NewMust("1.1.2"),
			version.NewMust("1.1.3"),
			version.NewMust("1.2.0"),
			version.NewMust("1.2.2"),
			version.NewMust("1.2.3"),
			version.NewMust("1.3.2"),
			version.NewMust("1.3.0"),
			version.NewMust("1.3.1"),
			version.NewMust("1.3.3"),
			version.NewMust("2.1.2"),
			version.NewMust("2.1.0"),
			version.NewMust("2.1.1"),
			version.NewMust("2.1.3"),
			version.NewMust("2.2"),
			version.NewMust("2.2.3"),
			version.NewMust("2.2.1"),
			version.NewMust("1.2.1"),
			version.NewMust("2.2.2"),
			version.NewMust("2.3"),
			version.NewMust("2.3.3"),
			version.NewMust("2.3.1"),
			version.NewMust("2.3.2"),
			version.NewMust("3.1.0"),
			version.NewMust("3.1.1"),
			version.NewMust("1.1.0"),
			version.NewMust("3.2.1"),
			version.NewMust("3.1.2"),
			version.NewMust("1.1.1"),
			version.NewMust("3.1.3"),
			version.NewMust("3.2.3"),
			version.NewMust("3.2.0"),
			version.NewMust("3.2.2"),
			version.NewMust("3.3.0"),
			version.NewMust("3.3.1"),
			version.NewMust("3.3.2"),
			version.NewMust("3.3.3"),
		}

		expected := version.Versions{
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
			version.NewMust("2.2"),
			version.NewMust("2.2.1"),
			version.NewMust("2.2.2"),
			version.NewMust("2.2.3"),
			version.NewMust("2.3"),
			version.NewMust("2.3.1"),
			version.NewMust("2.3.2"),
			version.NewMust("2.3.3"),
			version.NewMust("3.1.0"),
			version.NewMust("3.1.1"),
			version.NewMust("3.1.2"),
			version.NewMust("3.1.3"),
			version.NewMust("3.2.0"),
			version.NewMust("3.2.1"),
			version.NewMust("3.2.2"),
			version.NewMust("3.2.3"),
			version.NewMust("3.3.0"),
			version.NewMust("3.3.1"),
			version.NewMust("3.3.2"),
			version.NewMust("3.3.3"),
		}

		slices.SortFunc(versions, version.Version.Cmp)
		if !slices.Equal(expected, versions) {
			t.Fatalf("expected to sort versions, but got %v", versions)
		}
	})

	t.Run("Cmp", func(t *testing.T) {
		tcases := []struct {
			value    version.Version
			other    version.Version
			expected int
		}{
			{
				value:    version.NewMust("1.2.3"),
				other:    version.NewMust("1.2.3"),
				expected: 0,
			},
			{
				value:    version.NewMust("1.2.3-dev"),
				other:    version.NewMust("1.2.3"),
				expected: -1,
			},
			{
				value:    version.NewMust("1.2.3"),
				other:    version.NewMust("1.2.3-dev"),
				expected: 1,
			},
			{
				value:    version.NewMust("1.2.3"),
				other:    version.NewMust("1.2.1"),
				expected: 1,
			},
			{
				value:    version.NewMust("1.2.1"),
				other:    version.NewMust("1.2.3"),
				expected: -1,
			},
			{
				value:    version.NewMust("1.2.3"),
				other:    version.NewMust("1.1.4"),
				expected: 1,
			},
			{
				value:    version.NewMust("1.1.4"),
				other:    version.NewMust("1.2.3"),
				expected: -1,
			},
			{
				value:    version.NewMust("2.2.3"),
				other:    version.NewMust("1.4.4"),
				expected: 1,
			},
			{
				value:    version.NewMust("1.4.4"),
				other:    version.NewMust("2.2.3"),
				expected: -1,
			},
		}

		for _, tcase := range tcases {
			t.Run(tcase.value.String()+"-"+tcase.other.String(), func(t *testing.T) {
				got := tcase.value.Cmp(tcase.other)
				if tcase.expected != got {
					t.Fatalf("expected to return %d, but returned %d", tcase.expected, got)
				}
			})
		}
	})
}
