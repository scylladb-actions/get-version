package github

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scylladb-actions/get-version/types"
	"github.com/scylladb-actions/get-version/version"
)

func TestExecuteQueryWithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-123" {
			t.Errorf("expected Authorization header %q, got %q", "Bearer test-token-123", auth)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"1.0.0","prerelease":false,"draft":false}]`))
	}))
	defer server.Close()

	extractor := func(r *http.Response) (version.Versions, []types.IgnoredVersion, error) {
		return extractVersionsFromRelease(r, "")
	}

	versions, _, _, err := executeQuery(server.Client(), server.URL, "test-token-123", extractor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
}

func TestExecuteQueryWithoutToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("expected no Authorization header, got %q", auth)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"1.0.0","prerelease":false,"draft":false}]`))
	}))
	defer server.Close()

	extractor := func(r *http.Response) (version.Versions, []types.IgnoredVersion, error) {
		return extractVersionsFromRelease(r, "")
	}

	versions, _, _, err := executeQuery(server.Client(), server.URL, "", extractor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
}
