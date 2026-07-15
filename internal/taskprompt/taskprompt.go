// Package taskprompt stores the full agent prompt separately from the compact
// task description shown on the shared board.
package taskprompt

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/platform"
)

const maxPromptBytes = 256 * 1024

var taskIDPattern = regexp.MustCompile(`^T-[0-9]+$`)

// Detail is the private, local execution context for one board task.
type Detail struct {
	ID          string
	Description string
	Prompt      string
}

// ValidatePrompt rejects empty, excessively large, or NUL-containing prompts.
func ValidatePrompt(prompt string) error {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return fmt.Errorf("task prompt is required")
	}
	if len(prompt) > maxPromptBytes {
		return fmt.Errorf("task prompt is too long (maximum %d bytes)", maxPromptBytes)
	}
	if strings.ContainsRune(prompt, '\x00') {
		return fmt.Errorf("task prompt cannot contain NUL bytes")
	}
	return nil
}

// Write persists a human-readable prompt file under .coact/tasks.
func Write(dir string, detail Detail) error {
	if !taskIDPattern.MatchString(detail.ID) {
		return fmt.Errorf("invalid task id %q", detail.ID)
	}
	description := strings.TrimSpace(detail.Description)
	prompt := strings.TrimSpace(detail.Prompt)
	if description == "" {
		return fmt.Errorf("task description is required")
	}
	if err := ValidatePrompt(prompt); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	content := fmt.Sprintf("# %s · %s\n\n## Prompt\n\n%s\n", detail.ID, description, prompt)
	return platform.AtomicWrite(filepath.Join(dir, detail.ID+".md"), []byte(content), 0o600)
}

// Read loads one task's full execution prompt.
func Read(dir, id string) (*Detail, error) {
	if !taskIDPattern.MatchString(id) {
		return nil, fmt.Errorf("invalid task id %q", id)
	}
	data, err := os.ReadFile(filepath.Join(dir, id+".md"))
	if err != nil {
		return nil, err
	}
	content := string(data)
	firstLine, rest, ok := strings.Cut(content, "\n")
	if !ok || !strings.HasPrefix(firstLine, "# "+id+" · ") {
		return nil, fmt.Errorf("task prompt %s has an invalid header", id)
	}
	marker := "\n## Prompt\n\n"
	_, prompt, ok := strings.Cut(rest, marker)
	if !ok {
		return nil, fmt.Errorf("task prompt %s is missing its Prompt section", id)
	}
	detail := &Detail{
		ID:          id,
		Description: strings.TrimSpace(strings.TrimPrefix(firstLine, "# "+id+" · ")),
		Prompt:      strings.TrimSpace(prompt),
	}
	if err := ValidatePrompt(detail.Prompt); err != nil {
		return nil, err
	}
	return detail, nil
}
