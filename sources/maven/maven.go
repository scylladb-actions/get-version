package maven

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/scylladb-actions/get-version/httpclient"
	"github.com/scylladb-actions/get-version/types"
	"github.com/scylladb-actions/get-version/version"
)

type versionExtractor func(r *http.Response) (version.Versions, []types.IgnoredVersion, error)

func getURL(group, artifactID string) string {
	return "https://search.maven.org/solrsearch/select?q=g:" +
		group + "%20AND%20a:" + artifactID + "&core=gav&rows=1000&wt=json"
}

func executeQuery(
	cl *http.Client,
	url string,
	extractor versionExtractor,
) (out version.Versions, ignored []types.IgnoredVersion, err error) {
	var rq *http.Request
	rq, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	rq.Header.Set("Accept", "application/json")
	resp, err := cl.Do(rq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute http GET request for url %q: %w", url, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("failed to execute http GET request for url %q, server replied with %s", url, resp.Status)
	}
	out, ignored, err = extractor(resp)
	if err != nil {
		return nil, nil, err
	}
	return out, ignored, nil
}

func extractVersions(resp *http.Response, prefix string) (version.Versions, []types.IgnoredVersion, error) {
	type VersionRecord struct {
		Version string `json:"v"`
	}

	var respBody struct {
		Response struct {
			Docs []VersionRecord `json:"docs"`
		}
	}

	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&respBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse server response: %w", err)
	}

	var out version.Versions
	var ignored []types.IgnoredVersion
	for _, rec := range respBody.Response.Docs {
		name := rec.Version
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			ignored = append(ignored, types.IgnoredVersion{
				Version: name,
				Reason:  fmt.Errorf("version %q does not have prefix %q", name, prefix),
			})
			continue
		}
		name = strings.TrimPrefix(name, prefix)
		ver, err := version.New(name)
		if err != nil {
			ignored = append(ignored, types.IgnoredVersion{Version: name, Reason: err})
			continue
		}
		ver.SetPrefix(prefix)
		out = append(out, ver)
	}
	return out, ignored, nil
}

func getVersionsFromMVN(
	cl *http.Client,
	url string,
	extractor versionExtractor,
) (version.Versions, []types.IgnoredVersion, error) {
	for retry := 0; ; retry++ {
		versions, ignoredVersions, err := executeQuery(cl, url, extractor)
		if err == nil {
			return versions, ignoredVersions, nil
		}
		if retry > 5 {
			return nil, nil, fmt.Errorf("failed to execute query to %s, last error: %w", url, err)
		}
	}
}

type Source struct {
	params types.Params
}

func (s Source) GetAllVersions() (version.Versions, []types.IgnoredVersion, error) {
	return getVersionsFromMVN(
		httpclient.New(s.params),
		getURL(s.params.MavenGroup, s.params.MavenArtifactID),
		func(r *http.Response) (version.Versions, []types.IgnoredVersion, error) {
			return extractVersions(r, s.params.Prefix)
		})
}

func New(p types.Params) (Source, error) {
	if p.MavenArtifactID == "" {
		return Source{}, fmt.Errorf("maven artifact id is empty")
	}
	if p.MavenGroup == "" {
		return Source{}, fmt.Errorf("maven group is empty")
	}
	return Source{params: p}, nil
}
