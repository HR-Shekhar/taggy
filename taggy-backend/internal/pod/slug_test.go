package pod

import "testing"

func TestValidatePodSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		wantErr bool
	}{
		{in: "daily-grind", wantErr: false},
		{in: "abc", wantErr: false},
		{in: "AB", wantErr: true},
		{in: "bad_slug", wantErr: true},
		{in: "-leading", wantErr: true},
		{in: "trailing-", wantErr: true},
		{in: "a", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			err := validatePodSlug(tt.in)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.in)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.in, err)
			}
		})
	}
}

func TestNormalizePodSlug(t *testing.T) {
	t.Parallel()
	if got := normalizePodSlug("  Hello-World  "); got != "hello-world" {
		t.Fatalf("got %q", got)
	}
}

func TestSlugifyPodName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{in: "Night Owls Study Crew!", want: "night-owls-study-crew"},
		{in: "  Web Dev 101  ", want: "web-dev-101"},
		{in: "AB", want: "pod-ab"},
	}
	for _, tt := range tests {
		got := slugifyPodName(tt.in)
		if got != tt.want {
			t.Fatalf("slugifyPodName(%q)=%q want %q", tt.in, got, tt.want)
		}
		if err := validatePodSlug(got); err != nil {
			t.Fatalf("slugify result %q invalid: %v", got, err)
		}
	}
}
