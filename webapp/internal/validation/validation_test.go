package validation

import "testing"

func TestUsername(t *testing.T) {
	good := []string{"user1", "admin", "a.b-c_d"}
	bad := []string{"", "ab", "has space", "x'; DROP TABLE users;--", "<script>"}
	for _, s := range good {
		if err := Username(s); err != nil {
			t.Errorf("Username(%q) should be valid: %v", s, err)
		}
	}
	for _, s := range bad {
		if err := Username(s); err == nil {
			t.Errorf("Username(%q) should be rejected", s)
		}
	}
}

func TestEmail(t *testing.T) {
	if got, err := Email("  User1@Lab.Local "); err != nil || got != "user1@lab.local" {
		t.Errorf("email should be normalized, got %q, %v", got, err)
	}
	for _, s := range []string{"", "not-an-email", "a@b@c", "Name <a@b.com>"} {
		if _, err := Email(s); err == nil {
			t.Errorf("Email(%q) should be rejected", s)
		}
	}
}

func TestPassword(t *testing.T) {
	if err := Password("Secret12345"); err != nil {
		t.Errorf("valid password rejected: %v", err)
	}
	bad := []string{"", "short1", "onlyletterslong", "12345678901", string(make([]byte, 73))}
	for _, s := range bad {
		if err := Password(s); err == nil {
			t.Errorf("Password(%q) should be rejected", s)
		}
	}
}