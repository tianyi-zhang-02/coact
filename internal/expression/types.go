// Package expression provides model-agnostic response adapters.
package expression

// ChinesePolishStyle controls the target Chinese expression style.
type ChinesePolishStyle string

const (
	StyleNatural   ChinesePolishStyle = "natural"
	StyleConcise   ChinesePolishStyle = "concise"
	StyleTechnical ChinesePolishStyle = "technical"
	StyleFormal    ChinesePolishStyle = "formal"
	StyleCasual    ChinesePolishStyle = "casual"
	StyleProduct   ChinesePolishStyle = "product"
	StyleSupport   ChinesePolishStyle = "support"
)

// ChineseLocale controls regional wording preferences.
type ChineseLocale string

const (
	LocaleZhCN ChineseLocale = "zh-CN"
	LocaleZhTW ChineseLocale = "zh-TW"
	LocaleAuto ChineseLocale = "auto"
)

// Config controls Chinese expression polishing.
type Config struct {
	Enabled     bool
	Mode        string
	Style       ChinesePolishStyle
	Locale      ChineseLocale
	Trigger     TriggerConfig
	Preserve    PreserveConfig
	Glossary    map[string]string
	Diagnostics bool
}

// TriggerConfig decides when the adapter should run.
type TriggerConfig struct {
	UserLanguage            []string
	MinChineseRatio         float64
	SkipIfShorterThan       int
	SkipIfCodeRatioAbove    float64
	SkipForStructuredOutput bool
}

// PreserveConfig controls spans that must never be rewritten.
type PreserveConfig struct {
	Markdown       bool
	CodeBlocks     bool
	InlineCode     bool
	URLs           bool
	JSON           bool
	Commands       bool
	APINames       bool
	Paths          bool
	MarkdownTables bool
}

// ProtectedSpan is text that is replaced with a placeholder before polishing.
type ProtectedSpan struct {
	ID    string
	Kind  string
	Value string
}

// DetectResult explains whether the adapter should run.
type DetectResult struct {
	ShouldRun    bool
	Reason       string
	ChineseRatio float64
	CodeRatio    float64
}

// Result is returned by the adapter.
type Result struct {
	Text          string
	Changed       bool
	SkippedReason string
	Diagnostics   *Diagnostics
}

// Diagnostics is safe telemetry: it contains measurements, not raw content.
type Diagnostics struct {
	ChineseRatio       float64
	CodeRatio          float64
	ProtectedSpanCount int
	Style              ChinesePolishStyle
	Locale             ChineseLocale
	ValidationFailed   bool
}

// PolishFunc is supplied by the caller's model pipeline. Tests can inject a fake.
type PolishFunc func(prompt string, protectedText string) (string, error)
