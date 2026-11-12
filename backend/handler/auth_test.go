package handler

import (
	"bookhub/auth"
	"bookhub/store"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRegisterInvalidJSON(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader("{invalid json"),
	)

	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRegisterInvalidEmail(t *testing.T) {
	h := &Handler{}

	body := `{
        "email": "invalid-email",
        "password": "password123",
        "displayName": "Test User"
    }`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRegisterShortPassword(t *testing.T) {
	h := &Handler{}

	body := `{
        "email": "test@example.com",
        "password": "123",
        "displayName": "Test User"
    }`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRegisterEmptyName(t *testing.T) {
	h := &Handler{}

	body := `{
        "email": "test@example.com",
        "password": "password123",
        "displayName": ""
    }`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader("{invalid json"),
	)

	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}

	h := &Handler{
		Store: &store.Store{Pool: pool},
		Tickets: &auth.Issuer{
			Secret: []byte("this-is-a-secret-key-for-testing-123"),
			TTL:    time.Hour,
		},
	}

	email := "test-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	registerBody := `{
        "email": "` + email + `",
        "password": "password123",
        "displayName": "Test User"
    }`

	registerReq := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(registerBody),
	)

	registerRec := httptest.NewRecorder()

	h.Register(registerRec, registerReq)

	if registerRec.Code != http.StatusCreated {
		t.Fatalf(
			"Register failed: expected 201, got %d: %s",
			registerRec.Code,
			registerRec.Body.String(),
		)
	}

	var member map[string]interface{}

	if err := json.Unmarshal(registerRec.Body.Bytes(), &member); err != nil {
		t.Fatal(err)
	}

	if member["email"] != email {
		t.Error("Registered email does not match")
	}

	if _, exists := member["passwordHash"]; exists {
		t.Error("Password hash should not be exposed")
	}

	loginBody := `{
        "email": "` + email + `",
        "password": "password123"
    }`

	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(loginBody),
	)

	loginRec := httptest.NewRecorder()

	h.Login(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf(
			"Login failed: expected 200, got %d: %s",
			loginRec.Code,
			loginRec.Body.String(),
		)
	}

	var result map[string]interface{}

	if err := json.Unmarshal(loginRec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	token, ok := result["token"].(string)

	if !ok || token == "" {
		t.Error("Login should return a JWT token")
	}

	if result["member"] == nil {
		t.Error("Login should return member information")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	h := &Handler{
		Store: &store.Store{Pool: pool},
		Tickets: &auth.Issuer{
			Secret: []byte("this-is-a-secret-key-for-testing-123"),
			TTL:    time.Hour,
		},
	}

	body := `{
        "email": "nonexistent-test@example.com",
        "password": "wrongpassword"
    }`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rec.Code)
	}
}
