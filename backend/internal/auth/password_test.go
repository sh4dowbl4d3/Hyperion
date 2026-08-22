package auth

import "testing"

func TestHashAndVerifyRoundTrip(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if len(hash) == 0 || hash == password {
		t.Fatal("hash must be non-empty and differ from the plaintext password")
	}
	if err := VerifyPassword(hash, password); err != nil {
		t.Errorf("VerifyPassword(correct) = %v, want nil", err)
	}
	if err := VerifyPassword(hash, "wrong password"); err == nil {
		t.Error("VerifyPassword(wrong) = nil, want mismatch error")
	}
}

func TestHashesAreSaltedPerCall(t *testing.T) {
	password := "same-input-password"

	first, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first hash: %v", err)
	}
	second, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second hash: %v", err)
	}
	if first == second {
		t.Error("two hashes of the same password are identical; per-user salt is missing")
	}
	if err := VerifyPassword(first, password); err != nil {
		t.Errorf("verify first hash: %v", err)
	}
	if err := VerifyPassword(second, password); err != nil {
		t.Errorf("verify second hash: %v", err)
	}
}

func TestHashPasswordRejectsOversizedInput(t *testing.T) {
	tooLong := make([]byte, 73)
	for i := range tooLong {
		tooLong[i] = 'a'
	}
	if _, err := HashPassword(string(tooLong)); err == nil {
		t.Error("expected rejection of 73-byte password (bcrypt would truncate silently)")
	}
	if _, err := HashPassword(string(tooLong[:72])); err != nil {
		t.Errorf("72-byte password should be accepted: %v", err)
	}
}

func TestVerifyPasswordRejectsInvalidHashFormat(t *testing.T) {
	if err := VerifyPassword("not-a-bcrypt-hash", "whatever"); err == nil {
		t.Error("expected error for malformed stored hash")
	}
}
