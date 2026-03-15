package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSWildcardAllowsAnyOrigin(t *testing.T) {
	h := CORS("*")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/accounts", nil)
	req.Header.Set("Origin", "https://anything.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want \"*\"", got)
	}
}

func TestCORSRestrictedOriginAllowsConfiguredOriginOnly(t *testing.T) {
	h := CORS("https://app.example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/accounts", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed origin got Access-Control-Allow-Origin = %q, want empty", got)
	}

	req2 := httptest.NewRequest("GET", "/accounts", nil)
	req2.Header.Set("Origin", "https://app.example.com")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if got := rec2.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("allowed origin got Access-Control-Allow-Origin = %q, want https://app.example.com", got)
	}
}

func TestCORSRestrictedOriginSupportsCommaSeparatedList(t *testing.T) {
	h := CORS("https://app.example.com, https://staging.example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/accounts", nil)
	req.Header.Set("Origin", "https://staging.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://staging.example.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want https://staging.example.com", got)
	}
}
