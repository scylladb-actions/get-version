package types

import (
	"flag"
	"fmt"
	"slices"
	"strings"
)

type Params struct {
	SourceName        SourceName
	Repo              string
	FiltersDefinition string
	Prefix            string
	MavenGroup        string
	MavenArtifactID   string
	OutFormat         OutputName
	OutNoPrefix       bool
	OutReverseOrder   bool
	OutAsAction       bool
	SSLVerify         bool
	ShowVersion       bool
	RetryMax          int
	RetryInitialDelay int
	RetryMaxDelay     int
}

func (p *Params) Parse(knownSources Sources) error {
	flag.StringVar((*string)(&p.SourceName), "source", "",
		"Version source, one of: "+strings.Join(knownSources.Names(), ", "))
	flag.StringVar(&p.Repo, "repo", "", "Repository name. "+
		"Examples for dockerhub: ubuntu or alpine/git; for github: golang/go or scylladb/scylla")
	flag.StringVar(&p.FiltersDefinition, "filters", "",
		"Filters to apply to versions. Example: \"LAST.*.*\" ")
	flag.StringVar(&p.Prefix, "prefix", "", "Version prefix")
	flag.StringVar((*string)(&p.OutFormat), "out-format", "text", "Output type: json, yaml, text")
	flag.BoolVar(&p.OutReverseOrder, "out-reverse-order", false, "Reverse order")
	flag.BoolVar(&p.OutNoPrefix, "out-no-prefix", false, "Remove prefix from output")
	flag.StringVar(&p.MavenGroup, "mvn-group", "", "Artifact group to search on the maven")
	flag.StringVar(&p.MavenArtifactID, "mvn-artifact-id", "", "Artifact ID to search on the maven")
	flag.BoolVar(&p.OutAsAction, "out-as-action", false, "Output to a GitHub action output")
	flag.BoolVar(&p.SSLVerify, "ssl-verify", false, "Verify server SSL certificate")
	flag.BoolVar(&p.ShowVersion, "version", false, "Print version and exit")
	flag.IntVar(&p.RetryMax, "retry-max", 5, "Maximum number of retries for rate-limited requests")
	flag.IntVar(&p.RetryInitialDelay, "retry-initial-delay", 1000, "Initial retry delay in milliseconds for exponential backoff")
	flag.IntVar(&p.RetryMaxDelay, "retry-max-delay", 30000, "Maximum retry delay in milliseconds for exponential backoff")

	flag.Parse()

	if p.ShowVersion {
		return nil
	}
	if p.SourceName == "" {
		return fmt.Errorf("--source is empty")
	}
	if !knownSources.SourceExists(p.SourceName) {
		return fmt.Errorf("unknown source %q", p.SourceName)
	}
	if !slices.Contains(knownOutputNames, p.OutFormat) {
		return fmt.Errorf("unknown output format %q", p.OutFormat)
	}
	return nil
}
