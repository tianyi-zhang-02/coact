package expression

// DefaultConfig is conservative: disabled and only runs on substantial Chinese text.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Mode:    "fast",
		Style:   StyleTechnical,
		Locale:  LocaleZhCN,
		Trigger: TriggerConfig{
			UserLanguage:            []string{"zh"},
			MinChineseRatio:         0.25,
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

// DisabledConfig matches product default: the quality layer is opt-in.
func DisabledConfig() Config {
	cfg := DefaultConfig()
	cfg.Enabled = false
	return cfg
}
