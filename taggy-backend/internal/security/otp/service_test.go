package otp

import "testing"

func TestGenerateAndHash(t *testing.T) {
	t.Parallel()

	svc := New(func(plain string) string {
		return "hash:" + plain
	})

	code, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(code.PlainText) != 6 {
		t.Fatalf("expected 6 digits, got %q", code.PlainText)
	}
	for _, r := range code.PlainText {
		if r < '0' || r > '9' {
			t.Fatalf("non-digit in otp: %q", code.PlainText)
		}
	}
	if code.Hash != "hash:"+code.PlainText {
		t.Fatalf("hash mismatch: %q", code.Hash)
	}
	if got := svc.Hash("123456"); got != "hash:123456" {
		t.Fatalf("Hash = %q", got)
	}
}
