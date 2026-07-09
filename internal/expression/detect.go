package expression

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
)

var (
	codeBlockRe  = regexp.MustCompile("(?s)```.*?```")
	inlineCodeRe = regexp.MustCompile("`[^`\n]+`")
)

// ShouldPolishChineseOutput returns whether the adapter should run.
func ShouldPolishChineseOutput(userMessage, output string, cfg Config) DetectResult {
	chineseRatio := ChineseRatio(output)
	codeRatio := CodeRatio(output)
	result := DetectResult{ChineseRatio: chineseRatio, CodeRatio: codeRatio}

	if !cfg.Enabled {
		result.Reason = "disabled"
		return result
	}
	if explicitNoPolish(userMessage) {
		result.Reason = "tool_output"
		return result
	}
	if len([]rune(strings.TrimSpace(output))) < cfg.Trigger.SkipIfShorterThan {
		result.Reason = "too_short"
		return result
	}
	if cfg.Trigger.SkipForStructuredOutput && IsStructuredOutput(output) {
		result.Reason = "structured_output"
		return result
	}
	if chineseRatio < cfg.Trigger.MinChineseRatio {
		result.Reason = "not_chinese"
		return result
	}
	if codeRatio > cfg.Trigger.SkipIfCodeRatioAbove {
		result.Reason = "too_much_code"
		return result
	}
	result.ShouldRun = true
	result.Reason = "enabled"
	return result
}

// ChineseRatio returns the ratio of CJK characters to letters/numbers/CJK chars.
func ChineseRatio(s string) float64 {
	var cjk, total int
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		if unicode.Is(unicode.Han, r) {
			cjk++
			total++
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			total++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(cjk) / float64(total)
}

// CodeRatio approximates how much of the output is code-like protected content.
func CodeRatio(s string) float64 {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0
	}
	codeBytes := 0
	for _, match := range codeBlockRe.FindAllString(trimmed, -1) {
		codeBytes += len(match)
	}
	for _, match := range inlineCodeRe.FindAllString(trimmed, -1) {
		codeBytes += len(match)
	}
	lines := strings.Split(trimmed, "\n")
	for _, line := range lines {
		if looksLikeCommand(strings.TrimSpace(line)) {
			codeBytes += len(line)
		}
	}
	return float64(codeBytes) / float64(len(trimmed))
}

// IsStructuredOutput detects pure JSON, XML, YAML-ish, or tool output blocks.
func IsStructuredOutput(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return false
	}
	var js any
	if json.Unmarshal([]byte(trimmed), &js) == nil {
		return true
	}
	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") && strings.Contains(trimmed, "</") {
		return true
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) >= 2 {
		structured := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "---") || strings.Contains(line, ": ") || strings.HasPrefix(line, "-") {
				structured++
			}
		}
		if structured >= len(lines)*2/3 {
			return true
		}
	}
	return false
}

func explicitNoPolish(s string) bool {
	lower := strings.ToLower(s)
	phrases := []string{"原样输出", "不要润色", "不要改写", "只返回代码", "only code", "verbatim", "raw output"}
	for _, phrase := range phrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func looksLikeCommand(line string) bool {
	if line == "" {
		return false
	}
	prefixes := []string{"go ", "git ", "npm ", "pnpm ", "yarn ", "python ", "python3 ", "curl ", "coact ", "cd ", "mkdir ", "rm ", "chmod ", "docker ", "kubectl "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return strings.HasPrefix(line, "$ ")
}
