package sources

import (
	"github.com/scylladb-actions/get-version/sources/docker"
	"github.com/scylladb-actions/get-version/sources/github"
	"github.com/scylladb-actions/get-version/sources/maven"
	"github.com/scylladb-actions/get-version/types"
)

var AllSources = types.Sources{
	types.GitHubRelease: func(params types.Params) (types.Source, error) {
		return github.NewReleaseSource(params), nil
	},
	types.GitHubTag: func(params types.Params) (types.Source, error) {
		return github.NewTagSource(params), nil
	},
	types.DockerHubImageTag: func(params types.Params) (types.Source, error) {
		return docker.New(params)
	},
	types.MavenArtifact: func(params types.Params) (types.Source, error) {
		return maven.New(params)
	},
}
