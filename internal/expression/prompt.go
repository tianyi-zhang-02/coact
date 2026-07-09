package expression

import (
	"sort"
	"strings"
)

var styleInstructions = map[ChinesePolishStyle]string{
	StyleNatural:   "自然、顺口，像中文母语者正常表达。",
	StyleConcise:   "压缩冗余表达，保留必要信息。",
	StyleTechnical: "适合工程和技术场景，准确、清楚、克制。",
	StyleFormal:    "正式、稳重，适合文档和公告。",
	StyleCasual:    "口语化，但不要轻浮。",
	StyleProduct:   "适合产品说明，清楚、有吸引力但不过度营销。",
	StyleSupport:   "适合客服回复，礼貌、清楚、直接解决问题。",
}

// BuildPrompt creates the model instruction for a caller-supplied polish model.
func BuildPrompt(protectedText string, cfg Config) string {
	style := styleInstructions[cfg.Style]
	if style == "" {
		style = styleInstructions[StyleTechnical]
	}
	var glossary []string
	for from, to := range cfg.Glossary {
		glossary = append(glossary, from+" => "+to)
	}
	sort.Strings(glossary)
	return strings.TrimSpace(`你是 CoAct 的中文表达适配器。你的任务是改善模型输出的中文表达，而不是重新回答问题。

必须遵守：
- 不新增事实。
- 不删除重要信息。
- 不改变技术含义。
- 不改变结论。
- 保留 Markdown 结构。
- 保留所有占位符，例如 __COACT_PROTECTED_SPAN_0001__。
- 不解释你的修改过程。
- 不输出额外前言或总结。
- 只输出润色后的正文。

风格要求：
- ` + style + `
- 避免英文直译腔。
- 避免 AI 腔、套话、油腻表达。
- 技术内容要准确、克制。
- 如果原文已经自然，只做很小幅度修改。

地区偏好：` + string(cfg.Locale) + `

术语表：
` + strings.Join(glossary, "\n") + `

原文：
` + protectedText)
}
