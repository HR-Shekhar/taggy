package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	t.Parallel()

	svc := New(Config{Cost: 4}) // low cost for tests
	hash, err := svc.Hash("secret-pass")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if err := svc.Verify("secret-pass", hash); err != nil {
		t.Fatalf("Verify good: %v", err)
	}
	if err := svc.Verify("wrong", hash); err != ErrInvalidPassword {
		t.Fatalf("Verify bad: got %v", err)
	}
}
