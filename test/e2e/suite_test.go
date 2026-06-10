//go:build e2e

// Package e2e contains end-to-end tests that exercise the HTTP API against a
// real running server backed by PostgreSQL, Redis, and Kafka.
//
// Prerequisites (run once before the suite):
//
//	make docker-up
//	make migrate-up
//
// Run with:
//
//	make test-e2e
package e2e_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/golang-api-server/internal/config"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/server"
	"github.com/golang-api-server/pkg/password"
)

const baseURL = "http://localhost:18080"

var testDB *sql.DB

func TestMain(m *testing.M) {
	setTestEnv()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	deps, err := server.NewDependencies(cfg)
	if err != nil {
		log.Fatalf("init dependencies: %v", err)
	}

	app := server.New(deps)
	go func() {
		if err := app.Run(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	if err := waitForServer(20 * time.Second); err != nil {
		log.Fatalf("server not ready: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open test db: %v", err)
	}
	testDB = db

	code := m.Run()

	db.Close()
	// Trigger graceful shutdown.
	syscall.Kill(os.Getpid(), syscall.SIGTERM) //nolint:errcheck
	time.Sleep(300 * time.Millisecond)
	os.Exit(code)
}

// setTestEnv configures env vars for the test server instance.
// Values already set in the environment (e.g. in CI) take precedence.
func setTestEnv() {
	defaults := map[string]string{
		"SERVER_ADDRESS":  ":18080",
		"GRPC_ADDRESS":    ":19090",
		"APP_ENV":         "test",
		"JWT_SECRET":      "e2e-test-secret-for-testing-only!",
		"LOG_LEVEL":       "error",
		"AUTH_RATE_LIMIT": "10000",
		"API_RATE_LIMIT":  "10000",
	}
	for k, v := range defaults {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

// waitForServer polls /health until it returns 200 or the timeout is reached.
func waitForServer(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/health") //nolint:noctx
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("server did not become ready within %s", timeout)
}

// ── Database helpers ──────────────────────────────────────────────────────────

// truncateUsers wipes the users table. Call at the top of each test that creates users.
func truncateUsers(t *testing.T) {
	t.Helper()
	if _, err := testDB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate users: %v", err)
	}
}

// seedAdmin inserts an admin user directly into the DB and returns its access token.
func seedAdmin(t *testing.T) (email, pass, token string) {
	t.Helper()
	email = "admin@e2e.test"
	pass = "Admin1234!"
	hash, err := password.Hash(pass)
	if err != nil {
		t.Fatalf("hash admin password: %v", err)
	}
	if _, err := testDB.Exec(
		`INSERT INTO users (email, password, name, role) VALUES ($1, $2, $3, $4)`,
		email, hash, "E2E Admin", string(model.RoleAdmin),
	); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	auth := mustLogin(t, email, pass)
	return email, pass, auth.AccessToken
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func postJSON(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	return doRequest(t, http.MethodPost, path, "", body)
}

func authedPost(t *testing.T, path, token string, body any) *http.Response {
	t.Helper()
	return doRequest(t, http.MethodPost, path, token, body)
}

func authedGet(t *testing.T, path, token string) *http.Response {
	t.Helper()
	return doRequest(t, http.MethodGet, path, token, nil)
}

func authedPut(t *testing.T, path, token string, body any) *http.Response {
	t.Helper()
	return doRequest(t, http.MethodPut, path, token, body)
}

func authedDelete(t *testing.T, path, token string) *http.Response {
	t.Helper()
	return doRequest(t, http.MethodDelete, path, token, nil)
}

func doRequest(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL+path, r)
	if err != nil {
		t.Fatalf("build request %s %s: %v", method, path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	return resp
}

func mustLogin(t *testing.T, email, pass string) model.AuthResponse {
	t.Helper()
	resp := postJSON(t, "/api/v1/auth/login", map[string]string{"email": email, "password": pass})
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)
	var auth model.AuthResponse
	decodeJSON(t, resp.Body, &auth)
	return auth
}

func decodeJSON(t *testing.T, r io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("want HTTP %d, got %d; body: %s", want, resp.StatusCode, string(body))
	}
}
