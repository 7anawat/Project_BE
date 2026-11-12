package auth

import "testing"

func TestHashPassword(t *testing.T) {
	password := "password123"

	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("Password hash should not be empty")
	}

	if hash == password {
		t.Fatal("Hash should not be the same as original password")
	}
}

func TestCheckPasswordCorrect(t *testing.T) {
	password := "password123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckPassword(hash, password) {
		t.Error("Correct password should return true")
	}
}

func TestCheckPasswordIncorrect(t *testing.T) {
	password := "password123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if CheckPassword(hash, "wrongpassword") {
		t.Error("Incorrect password should return false")
	}
}

func TestCheckPasswordInvalidHash(t *testing.T) {
	if CheckPassword("invalid-hash", "password123") {
		t.Error("Invalid hash should return false")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "password123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Error("Two hashes of the same password should be different")
	}
}
