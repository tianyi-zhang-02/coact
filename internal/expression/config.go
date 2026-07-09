package expression

// DefaultConfig is enabled by default and conservative about technical spans.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Mode:    "fast",
		Style:   StyleTechnical,
		Locale:  LocaleZhCN,
		Trigger: TriggerConfig{
			UserLanguage:            []string{"zh"},
			MinChineseRatio:         0.10,
			SkipIfShorterThan:       80,
			SkipIfCodeRatioAbove:    0.45,
			SkipForStructuredOutput: true,
		},
		Preserve: PreserveConfig{
			Markdown:       true,
			CodeBlocks:     true,
			InlineCode:     true,
			URLs:           true,
			JSON:           true,
			Commands:       true,
			APINames:       true,
			Paths:          true,
			MarkdownTables: true,
		},
		Glossary: map[string]string{
			"agent":     "智能体",
			"tool call": "工具调用",
			"memory":    "记忆",
			"workflow":  "工作流",
			"pipeline":  "管线",
		},
	}
}

// DisabledConfig lets callers turn the quality layer off explicitly.
func DisabledConfig() Config {
	cfg := DefaultConfig()
	cfg.Enabled = false
	return cfg
}
