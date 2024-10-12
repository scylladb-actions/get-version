package filters

import "github.com/scylladb-actions/get-version/version"

type And []Filter

func (f And) Apply(versions version.Versions) version.Versions {
	result := versions
	for _, filter := range f {
		result = filter.Apply(result)
	}
	return versions
}

type Or []Filter

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
