package httpclient

import (
	"crypto/tls"
	"net/http"

	"github.com/scylladb-actions/get-version/types"
)

func New(p types.Params) *http.Client {
	if p.SSLVerify {
		return http.DefaultClient
	}
	httpClient := *http.DefaultClient
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &httpClient
}
