package types

import (
	"fmt"

	"github.com/scylladb-actions/get-version/version"
)

const (
	MavenArtifact     = SourceName("maven-artifact")
	GitHubRelease     = SourceName("github-release")
	GitHubTag         = SourceName("github-tag")
	DockerHubImageTag = SourceName("dockerhub-imagetag")
)

type SourceName string

type IgnoredVersion struct {
	Version string
	Reason  error
}

type Source interface {
	GetAllVersions() (out version.Versions, ignored []IgnoredVersion, err error)
}

type SourceBuilder func(Params) (Source, error)

type Sources map[SourceName]SourceBuilder

func (s Sources) Names() []string {
	out := make([]string, 0, len(s))
	for name := range s {
		out = append(out, string(name))
	}
	return out
}

func (s Sources) SourceExists(sourceName SourceName) bool {
	_, ok := s[sourceName]
	return ok
}

func (s Sources) GetSource(p Params) (Source, error) {
	builder := s[p.SourceName]
	if builder == nil {
		return nil, fmt.Errorf("unknown source %q", p.SourceName)
	}
	source, err := builder(p)
	if err != nil {
		return nil, err
	}
	return source, nil
}
