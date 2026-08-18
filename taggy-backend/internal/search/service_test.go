package search

import "testing"

func TestSanitizeQuery(t *testing.T) {
	t.Parallel()
	if got := sanitizeQuery("  web%dev_ "); got != "webdev" {
		t.Fatalf("got %q", got)
	}
}

func TestParseTypes(t *testing.T) {
	t.Parallel()

	all, err := parseTypes(nil)
	if err != nil || !all["skills"] || !all["users"] || !all["communities"] {
		t.Fatalf("nil types: %#v err=%v", all, err)
	}

	subset, err := parseTypes([]string{" Skills ", "users"})
	if err != nil || !subset["skills"] || !subset["users"] || subset["communities"] {
		t.Fatalf("subset: %#v err=%v", subset, err)
	}

	if _, err := parseTypes([]string{"pods"}); err != ErrInvalidType {
		t.Fatalf("want ErrInvalidType, got %v", err)
	}
}

func TestNormalizeLimit(t *testing.T) {
	t.Parallel()
	if got := normalizeLimit(0); got != defaultLimit {
		t.Fatalf("0 → %d", got)
	}
	if got := normalizeLimit(maxLimit + 1); got != maxLimit {
		t.Fatalf("over → %d", got)
	}
}
