package pod

import (
	"regexp"
	"strings"
	"unicode"
)

// Pod slugs: lowercase letters, digits, hyphens; 3–60 chars; no leading/trailing hyphen.
var podSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func normalizePodSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

func validatePodSlug(slug string) error {
	slug = normalizePodSlug(slug)
	if len(slug) < 3 || len(slug) > 60 || !podSlugPattern.MatchString(slug) {
		return ErrInvalidPodSlug
	}
	return nil
}

// slugifyPodName turns a human pod name into a URL-friendly slug.
// Example: "Night Owls Study Crew!" → "night-owls-study-crew"
func slugifyPodName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	b.Grow(len(name))
	lastHyphen := false
	for _, r := range name {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastHyphen = false
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if b.Len() > 0 && !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 60 {
		slug = strings.Trim(slug[:60], "-")
	}
	if len(slug) < 3 {
		slug = "pod-" + slug
		slug = strings.Trim(slug, "-")
		if len(slug) < 3 {
			slug = "pod-group"
		}
	}
	return slug
}
