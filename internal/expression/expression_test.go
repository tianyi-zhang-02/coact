package expression

import (
	"strings"
	"testing"
)

func testConfig() Config {
	cfg := DefaultConfig()
	cfg.Trigger.SkipIfShorterThan = 10
	return cfg
}

func TestDetectSkipsDisabledEnglishShortAndStructured(t *testing.T) {
	cfg := testConfig()
	cfg.Enabled = false
	if got := ShouldPolishChineseOutput("", "这是一个足够长的中文句子，需要检测。", cfg); got.Reason != "disabled" {
		t.Fatalf("disabled reason = %s", got.Reason)
	}
	cfg.Enabled = true
	if got := ShouldPolishChineseOutput("", "hello world this is plain english", cfg); got.Reason != "not_chinese" {
		t.Fatalf("english reason = %s", got.Reason)
	}
	if got := ShouldPolishChineseOutput("", "太短", cfg); got.Reason != "too_short" {
		t.Fatalf("short reason = %s", got.Reason)
	}
	if got := ShouldPolishChineseOutput("", `{"ok":true,"message":"中文"}`, cfg); got.Reason != "structured_output" {
		t.Fatalf("json reason = %s", got.Reason)
	}
}

func TestProtectAndRestoreSpans(t *testing.T) {
	cfg := testConfig()
	raw := "运行 `coact plan` 后访问 https://example.com/docs。\n\n```bash\ngo test ./...\n```"
	protected, spans := ProtectSpans(raw, cfg)
	if len(spans) != 3 {
		t.Fatalf("spans = %d, want 3: %#v", len(spans), spans)
	}
	if strings.Contains(protected, "coact plan") || strings.Contains(protected, "https://example.com") || strings.Contains(protected, "go test") {
		t.Fatalf("protected text leaked raw fragile content: %s", protected)
	}
	if restored := RestoreSpans(protected, spans); restored != raw {
		t.Fatalf("restore mismatch\n got: %q\nwant: %q", restored, raw)
	}
}

func TestAdaptPolishesAndPreservesTechnicalSpans(t *testing.T) {
	cfg := testConfig()
	raw := "这是一个强大的方式来处理这个问题。你可以使用下面的命令来启动服务：\n\n```bash\nnpm run dev\n```\n\n然后访问 http://localhost:3000。"
	res := Adapt("请优化中文", raw, cfg, func(prompt, protected string) (string, error) {
		if !strings.Contains(prompt, "中文表达适配器") {
			t.Fatalf("prompt missing adapter instruction")
		}
		return strings.ReplaceAll(protected, "这是一个强大的方式来处理这个问题。你可以使用下面的命令来启动服务", "这种方法可以有效处理这个问题。你可以用下面的命令启动服务"), nil
	})
	if !res.Changed || res.SkippedReason != "" {
		t.Fatalf("unexpected result: %#v", res)
	}
	if !strings.Contains(res.Text, "```bash\nnpm run dev\n```") || !strings.Contains(res.Text, "http://localhost:3000") {
		t.Fatalf("technical spans not preserved: %s", res.Text)
	}
}

func TestAdaptFallbackWhenProtectedSpanIsLost(t *testing.T) {
	cfg := testConfig()
	raw := "请运行 `coact inbox` 查看消息，然后访问 https://example.com 获取更多信息。"
	res := Adapt("", raw, cfg, func(_, _ string) (string, error) {
		return "请运行命令查看消息，然后访问链接获取更多信息。", nil
	})
	if res.Changed || res.Text != raw || res.SkippedReason != "validation_failed" {
		t.Fatalf("expected validation fallback, got %#v", res)
	}
}

func TestAdaptFallbackWhenOutputExpandsTooMuch(t *testing.T) {
	cfg := testConfig()
	raw := "这是一个需要润色的中文段落，包含足够的信息用于触发适配器运行。"
	res := Adapt("", raw, cfg, func(_, protected string) (string, error) {
		return protected + strings.Repeat(" 很长", 200), nil
	})
	if res.Changed || res.SkippedReason != "validation_failed" {
		t.Fatalf("expected expansion fallback, got %#v", res)
	}
}

func TestNoModelFallsBack(t *testing.T) {
	cfg := testConfig()
	raw := "这是一个需要润色的中文段落，包含足够的信息用于触发适配器运行。"
	res := Adapt("", raw, cfg, nil)
	if res.Changed || res.SkippedReason != "no_polish_model" {
		t.Fatalf("expected no model fallback, got %#v", res)
	}
}
