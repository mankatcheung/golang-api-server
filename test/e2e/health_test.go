//go:build e2e

package e2e_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	resp, err := http.Get(baseURL + "/health") //nolint:noctx
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)
}

func TestReadyz(t *testing.T) {
	resp, err := http.Get(baseURL + "/readyz") //nolint:noctx
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)
}

func TestMetrics(t *testing.T) {
	resp, err := http.Get(baseURL + "/metrics") //nolint:noctx
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "http_requests_total") {
		t.Error("expected Prometheus metric http_requests_total in /metrics output")
	}
}
