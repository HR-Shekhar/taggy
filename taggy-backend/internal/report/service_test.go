package report

import (
	"testing"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
)

func TestParseTargetType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		want    sqlc.ReportTargetType
		wantErr bool
	}{
		{in: "user", want: sqlc.ReportTargetTypeUSER},
		{in: "POD", want: sqlc.ReportTargetTypePOD},
		{in: " AUDIO_ROOM ", want: sqlc.ReportTargetTypeAUDIOROOM},
		{in: "COMMUNITY_CHANNEL", want: sqlc.ReportTargetTypeCOMMUNITYCHANNEL},
		{in: "nope", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := parseTargetType(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeLimit(t *testing.T) {
	t.Parallel()

	if got := normalizeLimit(0); got != defaultLimit {
		t.Fatalf("0 → %d", got)
	}
	if got := normalizeLimit(-1); got != defaultLimit {
		t.Fatalf("-1 → %d", got)
	}
	if got := normalizeLimit(10); got != 10 {
		t.Fatalf("10 → %d", got)
	}
	if got := normalizeLimit(maxLimit + 50); got != maxLimit {
		t.Fatalf("over max → %d", got)
	}
}
