package proxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// New creates a reverse proxy that forwards requests to the Python API.
// The targetURL should be the internal URL of the Python service
// (e.g., "http://api:8000" in Docker, or the Cloud Run internal URL).
//
// If the target is a Cloud Run URL (*.run.app), the proxy will
// automatically fetch and attach a Google Cloud ID token for
// service-to-service authentication.
func New(targetURL string) http.Handler {
	target, err := url.Parse(targetURL)
	if err != nil {
		slog.Error("invalid proxy target URL", "url", targetURL, "error", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "proxy misconfigured", http.StatusBadGateway)
		})
	}

	needsAuth := strings.HasSuffix(target.Host, ".run.app")

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host

		if needsAuth {
			if token, err := fetchIDToken(req.Context(), targetURL); err == nil {
				req.Header.Set("Authorization", "Bearer "+token)
			} else {
				slog.Error("failed to fetch ID token for proxy", "error", err)
			}
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("proxy error", "path", r.URL.Path, "error", err)
		http.Error(w, `{"detail":"Upstream service unavailable"}`, http.StatusBadGateway)
	}

	return proxy
}

// fetchIDToken retrieves a Google Cloud ID token from the metadata server.
// This only works when running on GCP (Cloud Run, GCE, etc.).
func fetchIDToken(ctx context.Context, audience string) (string, error) {
	metadataURL := fmt.Sprintf(
		"http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=%s",
		audience,
	)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
