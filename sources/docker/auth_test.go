package docker

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	cliconfig "github.com/docker/cli/cli/config"
)

func TestParseDockerAuthConfigFromEnv(t *testing.T) {
	auth := base64.StdEncoding.EncodeToString([]byte("user:pass"))
	t.Setenv(dockerEnvConfigKey, `{"auths":{"`+dockerHubAuthConfigKey+`":{"auth":"`+auth+`"}}}`)

	authConfigs, err := parseDockerAuthConfigFromEnv()
	if err != nil {
		t.Fatalf("parseDockerAuthConfigFromEnv failed: %v", err)
	}

	authConfig, ok := authConfigs[dockerHubAuthConfigKey]
	if !ok {
		t.Fatalf("expected auth config for %s", dockerHubAuthConfigKey)
	}
	if authConfig.Username != "user" {
		t.Fatalf("unexpected username %q", authConfig.Username)
	}
	if authConfig.Password != "pass" {
		t.Fatalf("unexpected password %q", authConfig.Password)
	}
}

func TestParseDockerAuthConfigFromEnv_Invalid(t *testing.T) {
	t.Setenv(dockerEnvConfigKey, `{"auths":{"`+dockerHubAuthConfigKey+`":{"username":"user"}}}`)

	_, err := parseDockerAuthConfigFromEnv()
	if err == nil {
		t.Fatalf("expected parseDockerAuthConfigFromEnv to fail for invalid input")
	}
}

func TestGetDockerHubAuthToken_UsesDockerAuthConfigEnv(t *testing.T) {
	configDir := t.TempDir()
	originalConfigDir := cliconfig.Dir()
	cliconfig.SetDir(configDir)
	t.Cleanup(func() {
		cliconfig.SetDir(originalConfigDir)
	})

	writeDockerConfigFile(t, configDir, `{"auths":{}}`)

	auth := base64.StdEncoding.EncodeToString([]byte("docker-user:docker-pass"))
	t.Setenv(dockerEnvConfigKey, `{"auths":{"`+dockerHubAuthConfigKey+`":{"auth":"`+auth+`"}}}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected method %s, got %s", http.MethodPost, r.Method)
		}
		var payload struct {
			Identifier string `json:"identifier"`
			Secret     string `json:"secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode auth payload: %v", err)
		}
		if payload.Identifier != "docker-user" {
			t.Fatalf("unexpected identifier %q", payload.Identifier)
		}
		if payload.Secret != "docker-pass" {
			t.Fatalf("unexpected secret %q", payload.Secret)
		}
		_, _ = w.Write([]byte(`{"access_token":"env-token"}`))
	}))
	defer server.Close()

	originalTokenURL := dockerHubAuthTokenURL
	dockerHubAuthTokenURL = server.URL
	t.Cleanup(func() {
		dockerHubAuthTokenURL = originalTokenURL
	})

	token, err := getDockerHubAuthToken(server.Client())
	if err != nil {
		t.Fatalf("getDockerHubAuthToken failed: %v", err)
	}
	if token != "env-token" {
		t.Fatalf("unexpected token %q", token)
	}
}

func TestGetDockerHubAuthToken_UsesRegistryTokenFromConfig(t *testing.T) {
	configDir := t.TempDir()
	originalConfigDir := cliconfig.Dir()
	cliconfig.SetDir(configDir)
	t.Cleanup(func() {
		cliconfig.SetDir(originalConfigDir)
	})
	t.Setenv(dockerEnvConfigKey, "")

	writeDockerConfigFile(t, configDir, `{"auths":{"`+dockerHubAuthConfigKey+`":{"registrytoken":"registry-token"}}}`)

	token, err := getDockerHubAuthToken(http.DefaultClient)
	if err != nil {
		t.Fatalf("getDockerHubAuthToken failed: %v", err)
	}
	if token != "registry-token" {
		t.Fatalf("unexpected token %q", token)
	}
}

func writeDockerConfigFile(t *testing.T, configDir, data string) {
	t.Helper()
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("failed to write docker config file: %v", err)
	}
}
