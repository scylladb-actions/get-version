package sources

import (
	"github.com/scylladb-actions/get-version/sources/docker"
	"github.com/scylladb-actions/get-version/sources/github"
	"github.com/scylladb-actions/get-version/sources/maven"
	"github.com/scylladb-actions/get-version/types"
)

var AllSources = types.Sources{
	types.GitHubRelease: func(p types.Params) (types.Source, error) {
		return github.NewReleaseSource(p.Repo, p.Prefix), nil
	},
	types.GitHubTag: func(p types.Params) (types.Source, error) {
		return github.NewTagSource(p.Repo, p.Prefix), nil
	},
	types.DockerHubImageTag: func(p types.Params) (types.Source, error) {
		return docker.New(p)
	},
	types.MavenArtifact: func(p types.Params) (types.Source, error) {
		return maven.New(p)
	},
}
