package filters

import "github.com/scylladb-actions/get-version/version"

type And []Filter

func NewAnd(filters ...Filter) And {
	return filters
}

func (f And) Apply(versions version.Versions) version.Versions {
	out := versions
	for _, filter := range f {
		out = filter.Apply(out)
	}
	return out
}

type Or []Filter

func NewOr(filters ...Filter) Or {
	return filters
}

func (f Or) Apply(versions version.Versions) version.Versions {
	set := map[version.Version]struct{}{}
	for _, filter := range f {
		for _, ver := range filter.Apply(versions) {
			set[ver] = struct{}{}
		}
	}

	var out version.Versions
	for ver := range set {
		out = append(out, ver)
	}
	return out
}
