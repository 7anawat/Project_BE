package auth

import (
	"testing"
	"time"
)

func TestIssueToken(t *testing.T) {
	issuer := &Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    time.Hour,
	}

	token, err := issuer.Issue(1, "member")

	if err != nil {
		t.Fatalf("Issue token failed: %v", err)
	}

	if token == "" {
		t.Fatal("Token should not be empty")
	}
}

func TestParseToken(t *testing.T) {
	issuer := &Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    time.Hour,
	}

	token, err := issuer.Issue(10, "admin")
	if err != nil {
		t.Fatal(err)
	}

	id, role, err := issuer.Parse(token)

	if err != nil {
		t.Fatalf("Parse token failed: %v", err)
	}

	if id != 10 {
		t.Errorf("Expected ID 10, got %d", id)
	}

	if role != "admin" {
		t.Errorf("Expected role admin, got %s", role)
	}
}

func TestParseExpiredToken(t *testing.T) {
	issuer := &Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    -time.Hour,
	}

	token, err := issuer.Issue(1, "member")
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = issuer.Parse(token)

	if err == nil {
		t.Error("Expired token should return error")
	}
}

func TestParseWrongSecret(t *testing.T) {
	issuer1 := &Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    time.Hour,
	}

	issuer2 := &Issuer{
		Secret: []byte("another-secret-key-for-testing-456"),
		TTL:    time.Hour,
	}

	token, err := issuer1.Issue(1, "member")
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = issuer2.Parse(token)

	if err == nil {
		t.Error("Token signed with another secret should fail")
	}
}

func TestParseEmptyToken(t *testing.T) {
	issuer := &Issuer{
		Secret: []byte("this-is-a-secret-key-for-testing-123"),
		TTL:    time.Hour,
	}

	_, _, err := issuer.Parse("")

	if err == nil {
		t.Error("Empty token should return error")
	}
}
