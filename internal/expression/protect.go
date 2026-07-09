package expression

import (
	"fmt"
	"regexp"
	"sort"
)

type matchSpan struct {
	start int
	end   int
	kind  string
}

var (
	urlRe           = regexp.MustCompile(`https?://[A-Za-z0-9._~:/?#\[\]@!$&'()*+,;=%-]+`)
	pathRe          = regexp.MustCompile(`(?:[A-Za-z]:\\|/)[A-Za-z0-9._~+@%=/\\:-]+`)
	markdownTableRe = regexp.MustCompile(`(?m)(?:^\|.*\|\s*$\n?){2,}`)
)

// ProtectSpans replaces fragile technical spans with stable placeholders.
func ProtectSpans(input string, cfg Config) (string, []ProtectedSpan) {
	var matches []matchSpan
	addMatches := func(kind string, re *regexp.Regexp) {
		for _, loc := range re.FindAllStringIndex(input, -1) {
			matches = append(matches, matchSpan{start: loc[0], end: loc[1], kind: kind})
		}
	}
	if cfg.Preserve.CodeBlocks {
		addMatches("code_block", codeBlockRe)
	}
	if cfg.Preserve.MarkdownTables {
		addMatches("markdown_table", markdownTableRe)
	}
	if cfg.Preserve.InlineCode {
		addMatches("inline_code", inlineCodeRe)
	}
	if cfg.Preserve.URLs {
		addMatches("url", urlRe)
	}
	if cfg.Preserve.Paths {
		addMatches("path", pathRe)
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].start == matches[j].start {
			return matches[i].end > matches[j].end
		}
		return matches[i].start < matches[j].start
	})

	filtered := matches[:0]
	lastEnd := -1
	for _, m := range matches {
		if m.start < lastEnd {
			continue
		}
		filtered = append(filtered, m)
		lastEnd = m.end
	}

	var spans []ProtectedSpan
	out := make([]byte, 0, len(input))
	cursor := 0
	for i, m := range filtered {
		id := fmt.Sprintf("__COACT_PROTECTED_SPAN_%04d__", i+1)
		out = append(out, input[cursor:m.start]...)
		out = append(out, id...)
		spans = append(spans, ProtectedSpan{ID: id, Kind: m.kind, Value: input[m.start:m.end]})
		cursor = m.end
	}
	out = append(out, input[cursor:]...)
	return string(out), spans
}

// RestoreSpans restores placeholders to their original values.
func RestoreSpans(input string, spans []ProtectedSpan) string {
	out := input
	for _, span := range spans {
		out = regexp.MustCompile(regexp.QuoteMeta(span.ID)).ReplaceAllString(out, span.Value)
	}
	return out
}
