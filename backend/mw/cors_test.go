package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowedOrigin(t *testing.T) {
	allowed := []string{
		"http://localhost:3000",
		"https://bookhub.vercel.app",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(allowed, next)

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("Allowed origin should receive CORS header")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestCORSBlockedOrigin(t *testing.T) {
	allowed := []string{
		"http://localhost:3000",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(allowed, next)

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Origin", "https://unknown-site.com")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Blocked origin should not receive CORS header")
	}
}

func TestCORSOptionsRequest(t *testing.T) {
	allowed := []string{
		"http://localhost:3000",
	}

	handler := CORS(allowed, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Next handler should not be called for OPTIONS")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/books", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d", rec.Code)
	}
}

func TestCORSAllowedMethods(t *testing.T) {
	handler := CORS(
		[]string{"http://localhost:3000"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	methods := rec.Header().Get("Access-Control-Allow-Methods")

	if methods != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Unexpected allowed methods: %s", methods)
	}
}

func TestCORSAllowedHeaders(t *testing.T) {
	handler := CORS(
		[]string{"http://localhost:3000"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	headers := rec.Header().Get("Access-Control-Allow-Headers")

	if headers != "Authorization, Content-Type" {
		t.Errorf("Unexpected allowed headers: %s", headers)
	}
}
