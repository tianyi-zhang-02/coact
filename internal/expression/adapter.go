package expression

// Adapt runs the Chinese expression adapter. It never returns a failed polish;
// on model or validation failure it falls back to raw text.
func Adapt(userMessage, rawText string, cfg Config, polish PolishFunc) Result {
	detect := ShouldPolishChineseOutput(userMessage, rawText, cfg)
	diagnostics := &Diagnostics{
		ChineseRatio: detect.ChineseRatio,
		CodeRatio:    detect.CodeRatio,
		Style:        cfg.Style,
		Locale:       cfg.Locale,
	}
	if !cfg.Diagnostics {
		diagnostics = nil
	}
	if !detect.ShouldRun {
		return Result{Text: rawText, Changed: false, SkippedReason: detect.Reason, Diagnostics: diagnostics}
	}
	if polish == nil {
		return Result{Text: rawText, Changed: false, SkippedReason: "no_polish_model", Diagnostics: diagnostics}
	}
	protectedText, spans := ProtectSpans(rawText, cfg)
	if diagnostics != nil {
		diagnostics.ProtectedSpanCount = len(spans)
	}
	polished, err := polish(BuildPrompt(protectedText, cfg), protectedText)
	if err != nil {
		return Result{Text: rawText, Changed: false, SkippedReason: "polish_failed", Diagnostics: diagnostics}
	}
	if !ValidatePolishedOutput(rawText, protectedText, polished, spans) {
		if diagnostics != nil {
			diagnostics.ValidationFailed = true
		}
		return Result{Text: rawText, Changed: false, SkippedReason: "validation_failed", Diagnostics: diagnostics}
	}
	final := RestoreSpans(polished, spans)
	return Result{Text: final, Changed: final != rawText, Diagnostics: diagnostics}
}
