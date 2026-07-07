package cli

import "testing"

func TestSanitizeAgent(t *testing.T) {
	cases := map[string]string{
		"claude":           "claude",
		"CoDeX":            "codex",
		"my-agent_1":       "my-agent_1",
		"../../etc/passwd": "etcpasswd", // traversal chars stripped
		"a b/c":            "abc",       // spaces and separators stripped
		"..":               "agent",     // empty result falls back
		"":                 "agent",
	}
	for in, want := range cases {
		if got := sanitizeAgent(in); got != want {
			t.Errorf("sanitizeAgent(%q) = %q, want %q", in, got, want)
		}
	}
}
