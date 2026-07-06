package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/coactdev/coact/internal/config"
	"github.com/coactdev/coact/internal/lockmgr"
	"github.com/coactdev/coact/internal/presence"
	"github.com/coactdev/coact/internal/project"
)

func cmdLock(args []string) int {
	fs := flag.NewFlagSet("lock", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	check := fs.Bool("check", false, "report whether the lock could be acquired, without acquiring")
	positionals, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(positionals) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact lock [--agent id] [--check] <path>")
		return 2
	}
	path := positionals[0]

	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	agent := agentID(*agentFlag)
	m := lockmgr.New(p, cfg)

	var res *lockmgr.Result
	if *check {
		res, err = m.Check(agent, path)
	} else {
		res, err = m.Acquire(agent, path)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}

	if res.Acquired {
		if *check {
			fmt.Printf("available: %s\n", res.Path)
		} else {
			fmt.Printf("locked %s (%s) by %s\n", res.Path, res.Reason, agent)
			_ = presence.Touch(p.SessionDir(), agent, "working")
		}
		return 0
	}

	c := res.Conflict
	fmt.Fprintf(os.Stderr, "denied: %s is held by %q since %s\n", res.Path, c.Owner, c.AcquiredAt)
	return 3
}

func cmdUnlock(args []string) int {
	fs := flag.NewFlagSet("unlock", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "participant id")
	positionals, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	if len(positionals) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact unlock [--agent id] <path>")
		return 2
	}
	path := positionals[0]

	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	agent := agentID(*agentFlag)
	m := lockmgr.New(p, cfg)
	if err := m.Release(agent, path); err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	fmt.Printf("unlocked %s\n", path)
	return 0
}

// loadProject finds the project and loads its config, printing a helpful error
// if the workspace is not initialized.
func loadProject() (*project.Project, *config.Config, bool) {
	p, err := project.Find()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return nil, nil, false
	}
	cfg, err := config.Load(p.ConfigPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: reading config: %v\n", err)
		return nil, nil, false
	}
	return p, cfg, true
}
