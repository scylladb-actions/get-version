package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/scylladb-actions/get-version/types"
	"github.com/scylladb-actions/get-version/version"
)

func NewText(params types.Params, output io.Writer) Text {
	return Text{params: params, output: output}
}

type Text struct {
	params types.Params
	output io.Writer
}

func (o Text) Write(versions version.Versions) error {
	for _, version := range versions.
		Order(o.params.OutReverseOrder).
		AsStringSlice(!o.params.OutNoPrefix) {
		fmt.Fprintln(o.output, version)
	}
	return nil
}

func NewJSON(params types.Params, output io.Writer) JSON {
	return JSON{params: params, output: output}
}

type JSON struct {
	params types.Params
	output io.Writer
}

func (o JSON) Write(versions version.Versions) error {
	return json.NewEncoder(o.output).Encode(
		versions.
			Order(o.params.OutReverseOrder).
			AsStringSlice(!o.params.OutNoPrefix),
	)
}

func NewYAML(params types.Params, output io.Writer) JSON {
	return JSON{params: params, output: output}
}

type YAML struct {
	params types.Params
	output io.Writer
}

func (o YAML) Write(versions version.Versions) error {
	return yaml.NewEncoder(o.output).Encode(
		versions.
			Order(o.params.OutReverseOrder).
			AsStringSlice(!o.params.OutNoPrefix),
	)
}

func NewOutput(params types.Params) (types.OutputType, error) {
	output, err := getOutputDestination(params)
	if err != nil {
		return nil, err
	}

	switch params.OutFormat {
	case types.OutputJSON:
		return NewJSON(params, output), nil
	case types.OutputYAML:
		return NewYAML(params, output), nil
	default:
		return NewText(params, output), nil
	}
}

func getOutputDestination(p types.Params) (*os.File, error) {
	if p.OutAsAction {
		// echo "versions=$(output)" >>"$GITHUB_OUTPUT"
		gitHubActionOutput := os.Getenv("GITHUB_OUTPUT")
		if gitHubActionOutput == "" {
			return nil, fmt.Errorf("GITHUB_OUTPUT is not set")
		}
		writer, err := os.OpenFile(os.Getenv("GITHUB_OUTPUT"), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %q: %w", gitHubActionOutput, err)
		}
		_, err = writer.Write([]byte("versions="))
		if err != nil {
			return nil, fmt.Errorf("failed to write to file %q: %w", gitHubActionOutput, err)
		}
		return writer, nil
	}
	return os.Stdout, nil
}
