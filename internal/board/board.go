// Package board reads and mutates the shared task board (.coact/board.md).
//
// The board is human-first Markdown; each task is a list item with a trailing
// HTML-comment metadata tag, e.g.
//
//	- [~] Refactor auth <!-- coact: id=T-011 state=doing owner=claude ttl=1800 -->
//
// The checkbox glyph mirrors state for humans; the comment is the source of
// truth for machines.
package board

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coactdev/coact/internal/platform"
)

var markerRe = regexp.MustCompile(`<!--\s*coact:\s*(.*?)\s*-->`)
var checkboxRe = regexp.MustCompile(`^\s*[-*]\s*\[.\]\s*`)
var idNumRe = regexp.MustCompile(`(\d+)`)

// Task is one row on the board.
type Task struct {
	ID    string
	Title string
	State string
	Owner string
	Extra map[string]string // ttl, hb, and any other preserved fields
	line  int               // index into Board.lines
}

// Board is a parsed, mutable view of board.md.
type Board struct {
	path  string
	lines []string
}

func nowRFC() string { return time.Now().UTC().Format(time.RFC3339) }

// Load reads and returns the board at path.
func Load(path string) (*Board, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &Board{path: path, lines: strings.Split(string(data), "\n")}, nil
}

func parseFields(s string) map[string]string {
	out := map[string]string{}
	for _, tok := range strings.Fields(s) {
		k, v, _ := strings.Cut(tok, "=")
		out[k] = v
	}
	return out
}

func (b *Board) parse() []*Task {
	var tasks []*Task
	for i, line := range b.lines {
		mm := markerRe.FindStringSubmatch(line)
		if mm == nil {
			continue
		}
		fields := parseFields(mm[1])
		title := line
		if idx := strings.Index(line, "<!--"); idx >= 0 {
			title = line[:idx]
		}
		title = checkboxRe.ReplaceAllString(strings.TrimRight(title, " "), "")
		tasks = append(tasks, &Task{
			ID:    fields["id"],
			Title: strings.TrimSpace(title),
			State: fields["state"],
			Owner: fields["owner"],
			Extra: fields,
			line:  i,
		})
	}
	return tasks
}

// Tasks returns all tasks currently on the board.
func (b *Board) Tasks() []*Task { return b.parse() }

func glyph(state string) string {
	switch state {
	case "done":
		return "x"
	case "doing", "claimed":
		return "~"
	case "blocked", "review":
		return "!"
	default:
		return " "
	}
}

func (t *Task) render() string {
	if t.Extra == nil {
		t.Extra = map[string]string{}
	}
	t.Extra["id"] = t.ID
	t.Extra["state"] = t.State
	t.Extra["owner"] = t.Owner

	ordered := []string{"id", "state", "owner"}
	var rest []string
	for k := range t.Extra {
		if k == "id" || k == "state" || k == "owner" {
			continue
		}
		rest = append(rest, k)
	}
	sort.Strings(rest)
	ordered = append(ordered, rest...)

	parts := make([]string, 0, len(ordered))
	for _, k := range ordered {
		parts = append(parts, k+"="+t.Extra[k])
	}
	return fmt.Sprintf("- [%s] %s <!-- coact: %s -->", glyph(t.State), t.Title, strings.Join(parts, " "))
}

// Save writes the board back atomically.
func (b *Board) Save() error {
	return platform.AtomicWrite(b.path, []byte(strings.Join(b.lines, "\n")), 0o644)
}

func (b *Board) mutate(id string, fn func(*Task) error) (*Task, error) {
	for _, t := range b.parse() {
		if t.ID != id {
			continue
		}
		if err := fn(t); err != nil {
			return nil, err
		}
		b.lines[t.line] = t.render()
		return t, nil
	}
	return nil, fmt.Errorf("no task %q on the board", id)
}

// Claim assigns a todo/unowned task to agent and moves it to doing.
func (b *Board) Claim(id, agent string, ttlSeconds int) (*Task, error) {
	return b.mutate(id, func(t *Task) error {
		if t.Owner != "" && t.Owner != agent && t.State != "todo" {
			return fmt.Errorf("task %s is already owned by %q (%s)", id, t.Owner, t.State)
		}
		if t.State == "done" {
			return fmt.Errorf("task %s is already done", id)
		}
		t.Owner = agent
		t.State = "doing"
		t.Extra["hb"] = nowRFC()
		if ttlSeconds > 0 {
			t.Extra["ttl"] = strconv.Itoa(ttlSeconds)
		}
		return nil
	})
}

// Finish marks a task done. The caller must own it.
func (b *Board) Finish(id, agent string) (*Task, error) {
	return b.mutate(id, func(t *Task) error {
		if t.Owner != "" && t.Owner != agent {
			return fmt.Errorf("task %s is owned by %q, not %q", id, t.Owner, agent)
		}
		t.State = "done"
		delete(t.Extra, "hb")
		delete(t.Extra, "ttl")
		return nil
	})
}

// Add appends a new todo task and returns it.
func (b *Board) Add(title string) *Task {
	id := b.nextID()
	t := &Task{ID: id, Title: title, State: "todo", Owner: "", Extra: map[string]string{}}
	line := t.render()

	insertAt := -1
	for i, l := range b.lines {
		if strings.HasPrefix(strings.TrimSpace(l), "## Backlog") {
			insertAt = i + 1
			break
		}
	}
	if insertAt >= 0 {
		// Skip a single blank line right after the header for readability.
		if insertAt < len(b.lines) && strings.TrimSpace(b.lines[insertAt]) == "" {
			insertAt++
		}
		b.lines = append(b.lines[:insertAt], append([]string{line}, b.lines[insertAt:]...)...)
	} else {
		b.lines = append(b.lines, line)
	}
	return t
}

func (b *Board) nextID() string {
	max := 0
	for _, t := range b.parse() {
		if m := idNumRe.FindString(t.ID); m != "" {
			if n, err := strconv.Atoi(m); err == nil && n > max {
				max = n
			}
		}
	}
	return fmt.Sprintf("T-%03d", max+1)
}
