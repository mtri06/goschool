package integration

import (
	"bytes"
	"encoding/json"
	"goschool/internal/db"
	"goschool/internal/env"
	"goschool/internal/server"
	"goschool/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

var (
	svr      *httptest.Server
	dbClient *sqlx.DB
)

func TestMain(m *testing.M) {
	env.Init("../.env.test")
	logger.Init()

	// Connect to Postgres
	dbClient = db.ConnectPostgres(db.DBConfig{
		Host:        env.Env.PgHost,
		Port:        env.Env.PgPort,
		User:        env.Env.PgUser,
		Password:    env.Env.PgPassword,
		Name:        env.Env.PgDBName,
		SSLMode:     env.Env.PgSSLMode,
		ConnTimeout: env.Env.PgConnTimeout,
	})
	log.Info().Msg("Connect to Postgres successfully")
	// Migrate database
	db.Migrate(dbClient.DB)

	svr = httptest.NewServer(server.New(dbClient))
	defer svr.Close()

	code := m.Run()

	os.Exit(code)
}

func clearDB(t *testing.T) {
	t.Helper()
	tables := []string{"user_teachers", "user_students", "subjects", "classes", "teaching_assignments", "tokens"}
	for _, tbl := range tables {
		if _, err := dbClient.Exec("DELETE FROM " + tbl); err != nil {
			panic(err)
		}
	}
	if _, err := dbClient.Exec("DELETE FROM users WHERE username != $1", env.Env.AdminUsername); err != nil {
		panic(err)
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

	req, err := http.NewRequest(method, svr.URL+path, &reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, opt := range reqOpts {
		if opt != nil {
			opt(req)
		}
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
		"username": env.Env.AdminUsername,
		"password": env.Env.AdminPassword,
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
	require.NoError(t, json.NewDecoder(resp.Body).Decode(v), "decode response body")
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
