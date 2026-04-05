//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

const (
	// baseURL points at the test stack started by `make test/up`
	baseURL           = "http://localhost:8081"
	testAdminUsername = "admin"
	testAdminPassword = "testpassword"
)

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

// doJSON fires an HTTP request with a JSON body and returns the response.
func doJSON(t *testing.T, method, path string, body any, cookies []*http.Cookie) *http.Response {
	t.Helper()

	var reqBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, baseURL+path, &reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

// loginAsAdmin logs in with the seeded admin credentials and returns the
// response cookies (access_token + refresh_token).
func loginAsAdmin(t *testing.T) []*http.Cookie {
	t.Helper()

	resp := doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"username": testAdminUsername,
		"password": testAdminPassword,
	}, nil)
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
