package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	cliconfig "github.com/docker/cli/cli/config"
)

const (
	dockerHubAuthConfigKey = "https://index.docker.io/v1/"
	dockerEnvConfigKey     = "DOCKER_AUTH_CONFIG"
)

var dockerHubAuthTokenURL = "https://hub.docker.com/v2/auth/token"

type envAuthConfig struct {
	Auth string `json:"auth"`
}

type dockerEnvConfig struct {
	AuthConfigs map[string]envAuthConfig `json:"auths"`
}

type dockerBasicAuth struct {
	Username string
	Password string
}

func getDockerHubAuthToken(cl *http.Client) (string, error) {
	cfg := cliconfig.LoadDefaultConfigFile(io.Discard)
	envAuthConfigs, envErr := parseDockerAuthConfigFromEnv()
	if envErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to create credential store from DOCKER_AUTH_CONFIG: ", envErr)
	}
	var firstErr error

	if envAuth, ok := envAuthConfigs[dockerHubAuthConfigKey]; ok {
		return createDockerHubAccessToken(cl, envAuth.Username, envAuth.Password)
	}

	authCfg, err := cfg.GetAuthConfig(dockerHubAuthConfigKey)
	if err != nil {
		firstErr = err
	}
	isEmptyAuthConfig := authCfg.Username == "" &&
		authCfg.Password == "" &&
		authCfg.Auth == "" &&
		authCfg.IdentityToken == "" &&
		authCfg.RegistryToken == ""
	if isEmptyAuthConfig {
		if firstErr != nil {
			return "", firstErr
		}
		return "", nil
	}

	if authCfg.RegistryToken != "" {
		return authCfg.RegistryToken, nil
	}
	if authCfg.IdentityToken != "" {
		return authCfg.IdentityToken, nil
	}
	if authCfg.Username == "" && authCfg.Password == "" && authCfg.Auth != "" {
		username, password, decodeErr := decodeDockerAuth(authCfg.Auth)
		if decodeErr != nil {
			return "", decodeErr
		}
		authCfg.Username = username
		authCfg.Password = password
	}
	if authCfg.Username == "" || authCfg.Password == "" {
		return "", fmt.Errorf("docker credentials for Docker Hub are missing username or password")
	}
	return createDockerHubAccessToken(cl, authCfg.Username, authCfg.Password)
}

func parseDockerAuthConfigFromEnv() (map[string]dockerBasicAuth, error) {
	envConfig := os.Getenv(dockerEnvConfigKey)
	if envConfig == "" {
		return nil, nil
	}

	var parsed dockerEnvConfig
	decoder := json.NewDecoder(strings.NewReader(envConfig))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to parse %s: %w", dockerEnvConfigKey, err)
	}
	if decoder.More() {
		return nil, fmt.Errorf("%s does not support more than one JSON object", dockerEnvConfigKey)
	}

	authConfigs := make(map[string]dockerBasicAuth, len(parsed.AuthConfigs))
	for addr, authCfg := range parsed.AuthConfigs {
		if authCfg.Auth == "" {
			return nil, fmt.Errorf("%s is missing key `auth` for %s", dockerEnvConfigKey, addr)
		}
		username, password, err := decodeDockerAuth(authCfg.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to decode %s auth for %s: %w", dockerEnvConfigKey, addr, err)
		}
		authConfigs[addr] = dockerBasicAuth{Username: username, Password: password}
	}

	return authConfigs, nil
}

func decodeDockerAuth(authValue string) (string, string, error) {
	if authValue == "" {
		return "", "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(authValue)
	if err != nil {
		return "", "", err
	}
	username, password, ok := strings.Cut(string(decoded), ":")
	if !ok || username == "" {
		return "", "", fmt.Errorf("invalid auth configuration")
	}
	return username, strings.Trim(password, "\x00"), nil
}

func createDockerHubAccessToken(cl *http.Client, username, secret string) (string, error) {
	body, err := json.Marshal(struct {
		Identifier string `json:"identifier"`
		Secret     string `json:"secret"`
	}{
		Identifier: username,
		Secret:     secret,
	})
	if err != nil {
		return "", err
	}
	rq, err := http.NewRequest(http.MethodPost, dockerHubAuthTokenURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	rq.Header.Set("Content-Type", "application/json")
	resp, err := cl.Do(rq)
	if err != nil {
		return "", fmt.Errorf("failed to request Docker Hub auth token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"failed to request Docker Hub auth token, server replied with %s: %s",
			resp.Status,
			string(bytes.TrimSpace(respBody)),
		)
	}
	var responseBody struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to parse Docker Hub auth token response: %w", err)
	}
	if responseBody.AccessToken == "" {
		return "", fmt.Errorf("failed to request Docker Hub auth token: empty access_token in response")
	}
	return responseBody.AccessToken, nil
}
