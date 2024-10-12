package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/scylladb-actions/get-version/types"
	"github.com/scylladb-actions/get-version/version"
)

var (
	dockerImageTagNamespacedURL = "https://hub.docker.com/v2/namespaces/%s/repositories/%s/tags?page_size=1000"
	dockerImageTagURL           = "https://hub.docker.com/v2/repositories/library/%s/tags?page_size=1000"
)

func getDockerURLFromRepo(repo string) string {
	chunks := strings.SplitN(repo, "/", 2)
	if len(chunks) == 2 {
		return fmt.Sprintf(dockerImageTagNamespacedURL, chunks[0], chunks[1])
	}
	return fmt.Sprintf(dockerImageTagURL, repo)
}

func getDockerImageVersionsOnce(
	url, prefix string,
) (out version.Versions, ignored []types.IgnoredVersion, next string, err error) {
	cl := http.DefaultClient
	var rq *http.Request
	rq, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, "", err
	}
	resp, err := cl.Do(rq)
	if err != nil {
		return nil, nil, "",
			fmt.Errorf("failed to execute http GET request for url %q: %w", url, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, "",
			fmt.Errorf("failed to execute http GET request for url %q, server replied with %s", url, resp.Status)
	}

	type Tag struct {
		Name string
	}

	var body struct {
		Next    string
		Results []Tag
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&body)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to parse server response: %w", err)
	}

	for _, rec := range body.Results {
		if prefix != "" && !strings.HasPrefix(rec.Name, prefix) {
			ignored = append(ignored, types.IgnoredVersion{
				Version: rec.Name,
				Reason:  fmt.Errorf("version %q does not have prefix %q", rec.Name, prefix),
			})
		}
		ver, err := version.New(rec.Name)
		if err != nil {
			ignored = append(ignored, types.IgnoredVersion{
				Version: rec.Name,
				Reason:  err,
			})
			continue
		}
		ver.SetPrefix(prefix)
		out = append(out, ver)
	}
	return out, ignored, body.Next, nil
}

type Source struct {
	params types.Params
}

func (s Source) GetAllVersions() (out version.Versions, ignored []types.IgnoredVersion, err error) {
	url := getDockerURLFromRepo(s.params.Repo)
	for url != "" {
		for retry := 0; ; retry++ {
			versions, ignoredVersions, nextURL, err := getDockerImageVersionsOnce(url, s.params.Prefix)
			if err != nil {
				if retry > 5 {
					return nil, nil, fmt.Errorf("failed to execute query to %s, last error: %w", url, err)
				}
				continue
			}
			out = append(out, versions...)
			ignored = append(ignored, ignoredVersions...)
			url = nextURL
			break
		}
	}

	return out, ignored, nil
}

func New(p types.Params) (Source, error) {
	if p.Repo == "" {
		return Source{}, fmt.Errorf("repo is required")
	}
	return Source{
		params: p,
	}, nil
}
