package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goschool/internal/api"
	"goschool/internal/db"
	"goschool/internal/env"
	"goschool/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

const (
	testAdminUsername = "admin"
	testAdminPassword = "testpassword"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	env.Init(".env.test")
	logger.Init()

	// Connect to Postgres
	dbClient := db.ConnectPostgres(db.DBConfig{
		Host:        env.Env.PgHost,
		Port:        env.Env.PgPort,
		User:        env.Env.PgUser,
		Password:    env.Env.PgPassword,
		Name:        env.Env.PgDBName,
		SSLMode:     env.Env.PgSSLMode,
		ConnTimeout: env.Env.PgConnTimeout,
	})
	log.Info().Msg("Connect to Postgres successfully")
	defer dbClient.Close()
	// Migrate database
	db.Migrate(dbClient.DB)

	server = httptest.NewServer(api.NewServer(dbClient))
	defer server.Close()

	code := m.Run()

	os.Exit(code)
}

func clearDB(t *testing.T, dbClient *sqlx.DB) {
	t.Helper()
	tables := []string{"users", "user_teachers", "user_students", "subjects", "classes", "teaching_assignments", "tokens"}
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))
	if _, err := dbClient.Exec(query); err != nil {
		t.Fatalf("failed to clear database: %v", err)
	}
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

type reqOptions func(*http.Request)

// requestJSON fires an HTTP request with a JSON body and returns the response.
func requestJSON(t *testing.T, method, path string, body any, reqOpts ...reqOptions) *http.Response {
	t.Helper()

	var reqBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, server.URL+path, &reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, opt := range reqOpts {
		opt(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func withCookies(cookies []*http.Cookie) reqOptions {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}
}

// loginAsAdmin logs in with the seeded admin credentials and returns the
// response cookies (access_token + refresh_token).
func loginAsAdmin(t *testing.T) []*http.Cookie {
	t.Helper()

	resp := requestJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"username": testAdminUsername,
		"password": testAdminPassword,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status %d", resp.StatusCode)
	}
	return resp.Cookies()
}

// decodeJSON decodes the response body into v.
func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
