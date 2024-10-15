package filters

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/scylladb-actions/get-version/version"
)

type StringPattern string

func (p StringPattern) asString() string {
	return string(p)
}

func (p StringPattern) isSpecial() bool {
	return p.isLast() || p.isFirst()
}

func (p StringPattern) isAny() bool {
	return p == "*"
}

func (p StringPattern) isRegexp() bool {
	return strings.Contains(p.asString(), "(") || strings.Contains(p.asString(), "[")
}

func (p StringPattern) isFirst() bool {
	return strings.HasPrefix(p.asString(), "FIRST")
}

func (p StringPattern) isLast() bool {
	return strings.HasPrefix(p.asString(), "LAST")
}

func (p StringPattern) Apply(
	sorted version.Versions,
	versionPeaceGetter func(version.Version) string,
	versionPeaceSorter func(version.Versions) []string,
) version.Versions {
	if p.isAny() {
		return sorted
	}
	var filter func(v version.Version) bool
	switch {
	case p.isFirst():
		idx := getIdxMust(p.asString(), "FIRST", '+')
		sortedPeaces := versionPeaceSorter(sorted)
		if len(sortedPeaces) <= idx+1 {
			return nil
		}
		value := sortedPeaces[idx]
		filter = func(v version.Version) bool {
			return versionPeaceGetter(v) != value
		}
	case p.isLast():
		idx := getIdxMust(p.asString(), "LAST", '-')
		sortedPeaces := versionPeaceSorter(sorted)
		if len(sortedPeaces) <= idx {
			return nil
		}
		value := sortedPeaces[len(sortedPeaces)-1-idx]
		filter = func(v version.Version) bool {
			return versionPeaceGetter(v) != value
		}
	case p.isRegexp():
		re := regexp.MustCompile(p.asString())
		filter = func(v version.Version) bool {
			return !re.MatchString(versionPeaceGetter(v))
		}
	case p.isAny():
		return sorted
	default:
		value := p.asString()
		filter = func(v version.Version) bool {
			return versionPeaceGetter(v) != value
		}
	}
	return slices.DeleteFunc(sorted, filter)
}

func getIdx(value, prefix string, symbol uint8) (int, error) {
	if value == prefix {
		return 0, nil
	}
	if value[len(prefix)] != symbol {
		return 0, fmt.Errorf("only allowed symbol after %s is %s", prefix, string(symbol))
	}
	value = strings.TrimPrefix(value, prefix+string(symbol))
	idx, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s should be followed by a number", prefix+string(symbol))
	}
	return idx, nil
}

func getIdxMust(value, prefix string, symbol uint8) int {
	idx, err := getIdx(value, prefix, symbol)
	if err != nil {
		panic(err)
	}
	return idx
}

func (p StringPattern) Validate() error {
	switch {
	case p.isFirst():
		_, err := getIdx(p.asString(), "FIRST", '+')
		return err
	case p.isLast():
		_, err := getIdx(p.asString(), "LAST", '-')
		return err
	case p.isRegexp():
		_, err := regexp.Compile(p.asString())
		if err != nil {
			return fmt.Errorf("wrong regexp format: %w", err)
		}
		return nil
	case p.asString() == "":
		return fmt.Errorf("empty pattern")
	}
	return nil
}

type Pattern struct {
	major StringPattern
	minor StringPattern
	patch StringPattern
}

func (f Pattern) Apply(versions version.Versions) version.Versions {
	filtered := slices.Clone(versions)
	if !f.minor.isSpecial() && !f.patch.isSpecial() {
		filtered = f.major.Apply(filtered, version.Version.MajorStr, version.Versions.UniqueMajors)
		filtered = f.minor.Apply(filtered, version.Version.MinorStr, version.Versions.UniqueMinors)
		return f.patch.Apply(filtered, version.Version.PatchStr, version.Versions.UniquePatches)
	}

	filtered = f.major.Apply(filtered, version.Version.MajorStr, version.Versions.UniqueMajors)
	if !f.patch.isSpecial() {
		return filtered.GroupAndFilter(version.Version.MajorStr, func(versions version.Versions) version.Versions {
			res := f.minor.Apply(versions, version.Version.MinorStr, version.Versions.UniqueMinors)
			return f.patch.Apply(res, version.Version.PatchStr, version.Versions.UniquePatches)
		})
	}

	return filtered.GroupAndFilter(version.Version.MajorStr, func(versions version.Versions) version.Versions {
		res := f.minor.Apply(versions, version.Version.MinorStr, version.Versions.UniqueMinors)
		return res.GroupAndFilter(version.Version.MinorStr, func(versions version.Versions) version.Versions {
			return f.patch.Apply(versions, version.Version.PatchStr, version.Versions.UniquePatches)
		})
	})
}

func (f Pattern) Validate() error {
	return errors.Join(
		wrapErr(f.major.Validate(), "failed to validate major pattern"),
		wrapErr(f.minor.Validate(), "failed to validate minor pattern"),
		wrapErr(f.patch.Validate(), "failed to validate patch pattern"),
	)
}

func (f Pattern) String() string {
	return fmt.Sprintf("%s.%s.%s", f.major, f.minor, f.patch)
}

func NewPattern(value string) (Pattern, error) {
	chunks := strings.SplitN(value, ".", 3)
	if len(chunks) != 3 {
		return Pattern{}, fmt.Errorf("can't convert %q to version pattern", value)
	}
	filter := Pattern{
		major: StringPattern(chunks[0]),
		minor: StringPattern(chunks[1]),
		patch: StringPattern(chunks[2]),
	}

	err := filter.Validate()
	if err != nil {
		return Pattern{}, err
	}
	return filter, nil
}

func NewPatternMust(value string) Pattern {
	p, err := NewPattern(value)
	if err != nil {
		panic(err)
	}
	return p
}
