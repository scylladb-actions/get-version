package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/scylladb-actions/get-version/types"
	"github.com/scylladb-actions/get-version/version"
)

var (
	githubReleaseURL = "https://api.github.com/repos/%s/releases?per_page=100"
	githubTagURL     = "https://api.github.com/repos/%s/tags?per_page=100"
)

type versionExtractor func(r *http.Response) (version.Versions, []types.IgnoredVersion, error)

func getGitHubReleaseURL(repo string) string {
	return fmt.Sprintf(githubReleaseURL, repo)
}

func getGitHubTagURL(repo string) string {
	return fmt.Sprintf(githubTagURL, repo)
}

func getNextLink(resp *http.Response) string {
	links := resp.Header.Get("link")
	nextLink := ""
	for _, link := range strings.Split(links, ",") {
		chunks := strings.SplitN(link, ";", 2)
		if len(chunks) != 2 {
			continue
		}
		linkInfo, rev := chunks[0], chunks[1]
		if strings.Contains(rev, "rel=\"next\"") {
			nextLink = strings.Trim(linkInfo, "<> ")
			break
		}
	}
	return nextLink
}

func executeQuery(
	url string,
	extractor versionExtractor,
) (out version.Versions, ignored []types.IgnoredVersion, next string, err error) {
	cl := http.DefaultClient
	var rq *http.Request
	rq, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, "", err
	}
	rq.Header.Set("Accept", "application/vnd.github+json")
	resp, err := cl.Do(rq)
	if err != nil {
		return nil, nil, "",
			fmt.Errorf("failed to execute http GET request for url %q: %w", url, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, "",
			fmt.Errorf("failed to execute http GET request for url %q, server replied with %s", url, resp.Status)
	}
	out, ignored, err = extractor(resp)
	if err != nil {
		return nil, nil, "", err
	}
	return out, ignored, getNextLink(resp), nil
}

func extractVersionsFromRelease(resp *http.Response, prefix string) (version.Versions, []types.IgnoredVersion, error) {
	var respBody []struct {
		Name       string
		Prerelease bool
		Draft      bool
	}

	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&respBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse server response: %w", err)
	}

	var out version.Versions
	var ignored []types.IgnoredVersion
	for _, rec := range respBody {
		if rec.Draft {
			continue
		}
		name := rec.Name
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			ignored = append(ignored,
				types.IgnoredVersion{
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

func getVersionsFromGitHub(
	url string,
	extractor versionExtractor,
) (out version.Versions, ignored []types.IgnoredVersion, err error) {
	for url != "" {
		for retry := 0; ; retry++ {
			versions, ignoredVersions, nextURL, err := executeQuery(url, extractor)
			if err != nil {
				if retry > 5 {
					return nil, nil, fmt.Errorf("failed to execute query to %s, last error: %w", url, err)
				}
				continue
			}
			ignored = append(ignored, ignoredVersions...)
			out = append(out, versions...)
			url = nextURL
			break
		}
	}
	return out, ignored, nil
}

type TagSource struct {
	repo   string
	prefix string
}

func (s TagSource) GetAllVersions() (version.Versions, []types.IgnoredVersion, error) {
	return getVersionsFromGitHub(
		getGitHubTagURL(s.repo), func(r *http.Response) (version.Versions, []types.IgnoredVersion, error) {
			return extractVersionsFromRelease(r, s.prefix)
		})
}

func NewTagSource(repo, prefix string) TagSource {
	return TagSource{repo: repo, prefix: prefix}
}

type ReleaseSource struct {
	repo   string
	prefix string
}

func (s ReleaseSource) GetAllVersions() (out version.Versions, ignored []types.IgnoredVersion, err error) {
	return getVersionsFromGitHub(
		getGitHubReleaseURL(s.repo),
		func(r *http.Response) (version.Versions, []types.IgnoredVersion, error) {
			return extractVersionsFromRelease(r, s.prefix)
		})
}

func NewReleaseSource(repo, prefix string) ReleaseSource {
	return ReleaseSource{repo: repo, prefix: prefix}
}
