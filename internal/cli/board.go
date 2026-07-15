package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/metalock"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
	"github.com/tianyi-zhang-02/coact/internal/taskprompt"
)

// withBoardLock serializes board mutations under a dedicated meta-lock so two
// agents never claim the same task at once (SPEC §3.3).
func withBoardLock(p *project.Project, fn func() error) error {
	if err := os.MkdirAll(p.CoactDir(), 0o755); err != nil {
		return err
	}
	lockPath := filepath.Join(p.CoactDir(), "board.lock")
	if err := metalock.Acquire(lockPath, 5*time.Second, 10*time.Second); err != nil {
		return err
	}
	defer metalock.Release(lockPath)
	return fn()
}

func cmdBoard(args []string) int {
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	b, err := board.Load(p.BoardPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: reading board: %v\n", err)
		return 1
	}
	tasks := b.Tasks()
	if len(tasks) == 0 {
		fmt.Println("board is empty — add one with: coact task add \"<title>\"")
		return 0
	}
	for _, t := range tasks {
		owner := t.Owner
		if owner == "" {
			owner = "-"
		}
		fmt.Printf("  %-7s %-9s %-10s %s\n", t.ID, t.State, owner, t.Title)
	}
	return 0
}

func cmdClaim(args []string) int {
	fs := flag.NewFlagSet("claim", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	ttl := fs.Int("ttl", 1800, "claim ttl in seconds")
	ids, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(ids) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact claim [--agent id] [--ttl secs] <task-id>")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	agent := agentID(*agentFlag)

	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		t, err := b.Claim(ids[0], agent, *ttl)
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			return err
		}
		_ = presence.Beat(p.SessionDir(), agent, "working", t.ID)
		_ = journal.Append(p.JournalDir(), agent, "task.claim", map[string]string{"id": t.ID})
		fmt.Printf("claimed %s (%s) by %s\n", t.ID, t.Title, agent)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	return 0
}

func cmdDone(args []string) int {
	fs := flag.NewFlagSet("done", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	ids, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(ids) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact done [--agent id] <task-id>")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	agent := agentID(*agentFlag)

	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		t, err := b.Finish(ids[0], agent)
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			return err
		}
		_ = journal.Append(p.JournalDir(), agent, "task.finish", map[string]string{"id": t.ID})
		fmt.Printf("done %s (%s)\n", t.ID, t.Title)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	return 0
}

func cmdTask(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: coact task <add|show|assign|unassign|reopen> ...")
		return 2
	}
	switch args[0] {
	case "add":
		return cmdTaskAdd(args[1:])
	case "show":
		return cmdTaskShow(args[1:])
	case "assign":
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: coact task assign <task-id> <agent>")
			return 2
		}
		owner := agentID(args[2])
		if owner == "" {
			fmt.Fprintln(os.Stderr, "coact task assign: agent must contain a-z, 0-9, _ or -")
			return 2
		}
		code := mutateTask(args[1], "human", "task.assign", func(b *board.Board) (*board.Task, error) {
			return b.Assign(args[1], owner)
		})
		if code == 0 {
			p, _, ok := loadProject()
			if ok {
				notifyTaskAssigned(p, owner, args[1], "human")
			}
		}
		return code
	case "unassign":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: coact task unassign <task-id>")
			return 2
		}
		return mutateTask(args[1], "human", "task.unassign", func(b *board.Board) (*board.Task, error) {
			return b.Unassign(args[1])
		})
	case "reopen":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: coact task reopen <task-id>")
			return 2
		}
		return mutateTask(args[1], "human", "task.reopen", func(b *board.Board) (*board.Task, error) {
			return b.Reopen(args[1])
		})
	default:
		fmt.Fprintln(os.Stderr, "usage: coact task <add|show|assign|unassign|reopen> ...")
		return 2
	}
}

func cmdTaskAdd(args []string) int {
	fs := flag.NewFlagSet("task add", flag.ContinueOnError)
	descriptionFlag := fs.String("description", "", "short dashboard description")
	promptFlag := fs.String("prompt", "", "full execution prompt")
	promptFileFlag := fs.String("prompt-file", "", "read the full execution prompt from a file")
	ownerFlag := fs.String("owner", "", "assign the task without starting it")
	positionals, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if *descriptionFlag != "" && len(positionals) > 0 {
		fmt.Fprintln(os.Stderr, "coact task add: use either --description or a positional description, not both")
		return 2
	}
	description := strings.TrimSpace(*descriptionFlag)
	if description == "" {
		description = strings.TrimSpace(strings.Join(positionals, " "))
	}
	if description == "" {
		fmt.Fprintln(os.Stderr, "usage: coact task add [--prompt text|--prompt-file path] [--owner agent] \"<short description>\"")
		return 2
	}
	if err := board.ValidateTitle(description); err != nil {
		fmt.Fprintf(os.Stderr, "coact task add: %v\n", err)
		return 2
	}
	prompt, err := taskPromptInput(description, *promptFlag, *promptFileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact task add: %v\n", err)
		return 2
	}
	owner := sanitizeAgent(*ownerFlag)
	if *ownerFlag != "" && owner != strings.ToLower(strings.TrimSpace(*ownerFlag)) {
		fmt.Fprintln(os.Stderr, "coact task add: owner must contain a-z, 0-9, _ or -")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	var created *board.Task
	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		created, err = addTaskWithPrompt(p, b, description, prompt, owner)
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			_ = os.Remove(filepath.Join(p.TasksDir(), created.ID+".md"))
			return err
		}
		event := "task.add"
		meta := map[string]string{"id": created.ID}
		if created.Owner != "" {
			event = "task.schedule"
			meta["owner"] = created.Owner
		}
		_ = journal.Append(p.JournalDir(), agentID(""), event, meta)
		fmt.Printf("added %s: %s\n", created.ID, created.Title)
		fmt.Printf("prompt: .coact/tasks/%s.md\n", created.ID)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	if created.Owner != "" {
		notifyTaskAssigned(p, created.Owner, created.ID, "human")
	}
	return 0
}

func cmdTaskShow(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact task show <task-id>")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	detail, err := taskprompt.Read(p.TasksDir(), strings.ToUpper(strings.TrimSpace(args[0])))
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact task show: %v\n", err)
		return 1
	}
	fmt.Printf("%s  %s\n\n%s\n", detail.ID, detail.Description, detail.Prompt)
	return 0
}

func taskPromptInput(description, prompt, promptFile string) (string, error) {
	if strings.TrimSpace(prompt) != "" && strings.TrimSpace(promptFile) != "" {
		return "", fmt.Errorf("use either --prompt or --prompt-file, not both")
	}
	if strings.TrimSpace(promptFile) != "" {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			return "", err
		}
		prompt = string(data)
	}
	if strings.TrimSpace(prompt) == "" {
		prompt = description
	}
	prompt = strings.TrimSpace(prompt)
	if err := taskprompt.ValidatePrompt(prompt); err != nil {
		return "", err
	}
	return prompt, nil
}

func addTaskWithPrompt(p *project.Project, b *board.Board, description, prompt, owner string) (*board.Task, error) {
	task := b.Add(description)
	if owner != "" {
		assigned, err := b.Assign(task.ID, owner)
		if err != nil {
			return nil, err
		}
		task = assigned
	}
	if err := taskprompt.Write(p.TasksDir(), taskprompt.Detail{ID: task.ID, Description: description, Prompt: prompt}); err != nil {
		return nil, err
	}
	return task, nil
}

func notifyTaskAssigned(p *project.Project, owner, id, from string) {
	detail, err := taskprompt.Read(p.TasksDir(), id)
	if err != nil {
		return
	}
	message := fmt.Sprintf("Assigned task %s: %s\n\nFull execution prompt:\n%s\n\nClaim it with `coact claim %s` before editing.", detail.ID, detail.Description, detail.Prompt, detail.ID)
	_ = inbox.Send(p.InboxDir(), from, owner, message)
}

func mutateTask(id, actor, event string, mutate func(*board.Board) (*board.Task, error)) int {
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	err := withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		task, err := mutate(b)
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			return err
		}
		_ = journal.Append(p.JournalDir(), actor, event, map[string]string{"id": task.ID, "owner": task.Owner})
		fmt.Printf("%s %s (%s)\n", strings.TrimPrefix(event, "task."), task.ID, task.Title)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	return 0
}
