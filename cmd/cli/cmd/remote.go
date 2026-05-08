package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const defaultDaemonURL = "http://127.0.0.1:8081"

func normalizeRemoteEndpoint(rawURL string, endpoint string) (string, error) {
	if rawURL == "" {
		rawURL = defaultDaemonURL
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsedURL.Host == "" {
		return "", fmt.Errorf("remote URL must include a host")
	}

	endpoint = "/" + strings.TrimPrefix(endpoint, "/")
	cleanPath := strings.TrimRight(parsedURL.Path, "/")
	if cleanPath == "" {
		parsedURL.Path = endpoint
	} else if cleanPath != endpoint {
		parsedURL.Path = cleanPath + endpoint
	}

	return parsedURL.String(), nil
}

func addToken(req *http.Request, token string) {
	if token == "" {
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Pin-Token", token)
}
