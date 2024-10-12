package version

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var patchReg = regexp.MustCompile("^([0-9]+)([a-z0-9._-]*)")

func NewPatch(value string) (Patch, error) {
	patchMatch := patchReg.FindStringSubmatch(value)
	if len(patchMatch) != 3 {
		return Patch{}, fmt.Errorf("patch %q does not match patch format: [0-9._a-z-]+", value)
	}
	patch, err := strconv.Atoi(patchMatch[1])
	if err != nil {
		return Patch{}, fmt.Errorf("can't convert patch %q to int", patchMatch[1])
	}
	extra := patchMatch[2]
	return Patch{
		patch:    patch,
		patchStr: value,
		extra:    extra,
	}, nil
}

func NewPatch2(patch int, extra string) Patch {
	if patch == math.MinInt {
		return emptyPatch
	}
	return Patch{
		patch: patch,
		extra: extra,
	}
}

type Patch struct {
	patch    int
	patchStr string
	extra    string
}

func (p Patch) IsDev() bool {
	return strings.Contains(p.extra, "dev")
}

func (p Patch) IsPre() bool {
	return strings.Contains(p.extra, "pre")
}

func (p Patch) IsRC() bool {
	return strings.Contains(p.extra, "rc")
}

func (p Patch) IsPROD() bool {
	return p.extra == ""
}

func (p Patch) AsInt() int {
	if p.IsEmpty() {
		return 0
	}
	return p.patch
}

func (p Patch) IsEmpty() bool {
	return p.patch == math.MinInt
}

func (p Patch) String() string {
	if p.IsEmpty() {
		return p.extra
	}
	if p.patchStr != "" {
		return p.patchStr
	}
	return strconv.Itoa(p.patch) + p.extra
}

func (p Patch) Extra() string {
	return p.extra
}

func (p Patch) Cmp(o Patch) int {
	if p.AsInt() != o.AsInt() {
		return sign(p.AsInt() - o.AsInt())
	}
	prodVal := 0
	if p.IsPROD() {
		prodVal = 1
	}
	otherProdVal := 0
	if o.IsPROD() {
		otherProdVal = 1
	}
	return prodVal - otherProdVal
}

func (p Patch) Equal(o Patch) bool {
	return p == o
}

type Version struct {
	major    int
	majorStr string
	minor    int
	minorStr string
	patch    Patch
	prefix   string
}

func (v *Version) SetPrefix(prefix string) {
	v.prefix = prefix
}

func (v Version) Major() int {
	return v.major
}

func (v Version) MajorStr() string {
	if v.majorStr != "" {
		return v.majorStr
	}
	return strconv.Itoa(v.major)
}

func (v Version) Minor() int {
	return v.minor
}

func (v Version) MinorStr() string {
	if v.minorStr != "" {
		return v.minorStr
	}
	return strconv.Itoa(v.minor)
}

func (v Version) PatchRaw() Patch {
	return v.patch
}

func (v Version) Patch() int {
	return v.patch.AsInt()
}

func (v Version) Extra() string {
	return v.patch.Extra()
}

func (v Version) PatchStr() string {
	return v.patch.String()
}

func (v Version) IsDev() bool {
	return v.patch.IsDev()
}

func (v Version) IsPre() bool {
	return v.patch.IsPre()
}

func (v Version) IsRC() bool {
	return v.patch.IsRC()
}

func (v Version) IsPROD() bool {
	return v.patch.IsPROD()
}

func (v Version) String() string {
	patchStr := v.patch.String()
	if patchStr == "" {
		return fmt.Sprintf("%s%s.%s", v.prefix, v.MajorStr(), v.MinorStr())
	}
	return fmt.Sprintf("%s%s.%s.%s", v.prefix, v.MajorStr(), v.MinorStr(), patchStr)
}

func (v Version) NoPrefixString() string {
	patchStr := v.patch.String()
	if patchStr == "" {
		return fmt.Sprintf("%s.%s", v.MajorStr(), v.MinorStr())
	}
	return fmt.Sprintf("%s.%s.%s", v.MajorStr(), v.MinorStr(), patchStr)
}

func (v Version) Equal(o Version) bool {
	return v.major == o.major && v.minor == o.minor && v.patch == o.patch
}

func sign(val int) int {
	if val < 0 {
		return -1
	}
	if val == 0 {
		return 0
	}
	return 1
}

func (v Version) Cmp(o Version) int {
	if v.major != o.major {
		return sign(v.major - o.major)
	}
	if v.minor != o.minor {
		return sign(v.minor - o.minor)
	}
	if v.patch != o.patch {
		return v.patch.Cmp(o.patch)
	}

	return 0
}

var emptyPatch = Patch{
	patch: math.MinInt,
}

func New(value string) (Version, error) {
	chunk := strings.SplitN(value, ".", 3)
	if len(chunk) < 2 || len(chunk) > 3 {
		return Version{}, fmt.Errorf("can't convert major %q to int", value)
	}
	majorStr := chunk[0]
	minorStr := chunk[1]
	var patchStr string
	if len(chunk) == 3 {
		patchStr = chunk[2]
	}

	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return Version{}, fmt.Errorf("can't convert minor %q to int", value)
	}
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return Version{}, fmt.Errorf("can't convert patch %q to int", value)
	}

	patch := emptyPatch
	if patchStr != "" {
		patch, err = NewPatch(patchStr)
		if err != nil {
			return Version{}, err
		}
	}

	return Version{
		major:    major,
		majorStr: majorStr,
		minor:    minor,
		minorStr: minorStr,
		patch:    patch,
	}, nil
}

func New2(major, minor, patch int, extra string) Version {
	return Version{
		major: major,
		minor: minor,
		patch: NewPatch2(patch, extra),
	}
}

func NewMust(value string) Version {
	version, err := New(value)
	if err != nil {
		panic(err.Error())
	}
	return version
}

type Versions []Version

func (v Versions) GetAllPeaces(fn func(Version) int) []int {
	var ids []int
	for _, ver := range v {
		value := fn(ver)
		if !slices.Contains(ids, value) {
			ids = append(ids, value)
		}
	}
	return ids
}

func (v Versions) UniqueMajors() []string {
	ids := v.GetAllPeaces(Version.Major)
	slices.Sort(ids)
	return intoToStr(ids)
}

func (v Versions) UniqueMinors() []string {
	ids := v.GetAllPeaces(Version.Minor)
	slices.Sort(ids)
	return intoToStr(ids)
}

func (v Versions) UniquePatches() []string {
	var patches []Patch
	for _, ver := range v {
		patch := ver.PatchRaw()
		if !slices.ContainsFunc(patches, patch.Equal) {
			patches = append(patches, patch)
		}
	}

	slices.SortFunc(patches, Patch.Cmp)
	out := make([]string, len(patches))
	for i, patch := range patches {
		out[i] = patch.String()
	}
	return out
}

func (v Versions) GroupAndFilter(get func(version Version) string, filter func(Versions) Versions) Versions {
	if len(v) == 0 {
		return v
	}
	var out Versions
	var tmp Versions
	prior := get(v[0])
	for _, ver := range v {
		key := get(ver)
		if key == prior {
			tmp = append(tmp, ver)
			continue
		}
		out = append(out, filter(tmp)...)
		tmp = tmp[0:0]
		prior = key
	}
	if len(tmp) > 0 {
		out = append(out, filter(tmp)...)
	}
	return out
}

func (v Versions) AsStringSlice(prefix bool) []string {
	out := make([]string, len(v))
	for i, ver := range v {
		if prefix {
			out[i] = ver.String()
		} else {
			out[i] = ver.NoPrefixString()
		}
	}
	return out
}

func (v Versions) Order(reverse bool) Versions {
	if reverse {
		slices.SortFunc(v, func(a, b Version) int {
			return b.Cmp(a)
		})
	} else {
		slices.SortFunc(v, Version.Cmp)
	}
	return v
}

func intoToStr(values []int) []string {
	out := make([]string, len(values))
	for i, val := range values {
		out[i] = strconv.Itoa(val)
	}
	return out
}
