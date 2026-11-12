package mw

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookhub/auth"
)

// TestRequireLoginNoToken
// ทดสอบกรณีไม่มี Token ต้องได้ HTTP 401
func TestRequireLoginNoToken(t *testing.T) {
	issuer := &auth.Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    time.Hour,
	}

	handler := RequireLogin(issuer, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rec.Code)
	}
}

// TestRequireLoginInvalidToken
// ทดสอบกรณี Token ไม่ถูกต้อง ต้องได้ HTTP 401
func TestRequireLoginInvalidToken(t *testing.T) {
	issuer := &auth.Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    time.Hour,
	}

	handler := RequireLogin(issuer, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rec.Code)
	}
}

// TestRequireLoginValidToken
// ทดสอบกรณี Token ถูกต้อง ต้องเข้าใช้งานได้
func TestRequireLoginValidToken(t *testing.T) {
	issuer := &auth.Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    time.Hour,
	}

	token, err := issuer.Issue(10, "member")
	if err != nil {
		t.Fatal(err)
	}

	handler := RequireLogin(issuer, func(w http.ResponseWriter, r *http.Request) {
		user := WhoFrom(r.Context())

		if user.ID != 10 {
			t.Errorf("Expected user ID 10, got %d", user.ID)
		}

		if user.Role != "member" {
			t.Errorf("Expected role member, got %s", user.Role)
		}

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

// TestRequireRoleMemberDenied
// ทดสอบ Member ไม่สามารถเข้า Admin ได้ ต้องได้ HTTP 403
func TestRequireRoleMemberDenied(t *testing.T) {
	nextCalled := false

	handler := RequireRole("admin", func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)

	// จำลองข้อมูลผู้ใช้ที่ Login แล้ว
	req = req.WithContext(
		context.WithValue(req.Context(), ctxKey{}, Who{
			ID:   1,
			Role: "member",
		}),
	)

	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", rec.Code)
	}

	if nextCalled {
		t.Error("Member should not access admin handler")
	}
}

// TestRequireRoleAdminAllowed
// ทดสอบ Admin สามารถเข้าใช้งานหน้า Admin ได้
func TestRequireRoleAdminAllowed(t *testing.T) {
	handler := RequireRole("admin", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)

	// จำลองข้อมูลผู้ใช้ Admin
	req = req.WithContext(
		context.WithValue(req.Context(), ctxKey{}, Who{
			ID:   1,
			Role: "admin",
		}),
	)

	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}
