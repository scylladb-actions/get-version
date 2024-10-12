package types

import "github.com/scylladb-actions/get-version/version"

type OutputName string

const (
	OutputTEXT OutputName = "text"
	OutputJSON OutputName = "json"
	OutputYAML OutputName = "yaml"
)

var knownOutputNames = []OutputName{OutputTEXT, OutputJSON, OutputYAML}

type OutputType interface {
	Write(version.Versions) error
}
