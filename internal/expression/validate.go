package expression

import "strings"

// ValidatePolishedOutput checks that polishing did not corrupt protected content.
func ValidatePolishedOutput(raw, protectedText, polishedProtected string, spans []ProtectedSpan) bool {
	if strings.TrimSpace(polishedProtected) == "" {
		return false
	}
	for _, span := range spans {
		if !strings.Contains(polishedProtected, span.ID) {
			return false
		}
	}
	if len([]rune(polishedProtected)) > int(float64(len([]rune(protectedText)))*1.8) {
		return false
	}
	restored := RestoreSpans(polishedProtected, spans)
	if strings.Count(restored, "```") != strings.Count(raw, "```") {
		return false
	}
	if len(urlRe.FindAllString(restored, -1)) != len(urlRe.FindAllString(raw, -1)) {
		return false
	}
	if len(markdownTableRe.FindAllString(restored, -1)) != len(markdownTableRe.FindAllString(raw, -1)) {
		return false
	}
	return true
}
