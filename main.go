package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scylladb-actions/get-version/filters"
	"github.com/scylladb-actions/get-version/output"
	"github.com/scylladb-actions/get-version/sources"
	"github.com/scylladb-actions/get-version/types"
)

func main() {
	p := types.Params{}
	err := p.Parse(sources.AllSources)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		flag.Usage()
		os.Exit(1)
	}
	if p.ShowVersion {
		fmt.Fprintln(os.Stdout, buildVersion)
		return
	}

	source, err := sources.AllSources.GetSource(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	filter, err := filters.ParseFilterString(p.FiltersDefinition)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	o, err := output.NewOutput(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	allVersions, _, err := source.GetAllVersions()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	filteredVersions := filter.Apply(allVersions)

	err = o.Write(filteredVersions)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
