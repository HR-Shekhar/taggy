package auth

import (
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "ok", in: "taggy_tester", wantErr: false},
		{name: "dots_ok", in: "a.b_c", wantErr: false},
		{name: "too_short", in: "ab", wantErr: true},
		{name: "too_long", in: strings.Repeat("a", 31), wantErr: true},
		{name: "spaces", in: "bad name", wantErr: true},
		{name: "hyphen", in: "bad-name", wantErr: true},
		{name: "empty", in: "", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateUsername(tt.in)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.in)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.in, err)
			}
		})
	}
}
