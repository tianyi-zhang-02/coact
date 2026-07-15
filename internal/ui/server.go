// Package ui implements CoAct's local-only control center.
package ui

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/adapter"
	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/buildinfo"
	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/metalock"
	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
	"github.com/tianyi-zhang-02/coact/internal/setup"
	"github.com/tianyi-zhang-02/coact/internal/versionmgr"
)

// Options configure the local UI server.
type Options struct {
	Addr        string
	Port        int
	Lang        string
	OpenBrowser bool
}

// Serve starts the local UI server and blocks.
func Serve(opts Options) error {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1"
	}
	if opts.Addr != "127.0.0.1" && opts.Addr != "localhost" {
		return fmt.Errorf("ui only supports local addresses; got %q", opts.Addr)
	}
	if opts.Port == 0 {
		opts.Port = 7331
	}
	if opts.Lang == "" {
		opts.Lang = "en"
	}
	if opts.Lang != "zh" {
		opts.Lang = "en"
	}

	token, err := randomToken()
	if err != nil {
		return err
	}
	srv := &Server{lang: opts.Lang, token: token}
	mux := http.NewServeMux()
	srv.routes(mux)

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opts.Addr, opts.Port))
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String()
	fmt.Fprintf(os.Stderr, "coact ui: %s\n", url)
	if opts.OpenBrowser {
		if err := openBrowser(url); err != nil {
			fmt.Fprintf(os.Stderr, "coact ui: could not open browser: %v\n", err)
		}
	}
	httpSrv := &http.Server{Handler: srv.guard(mux), ReadHeaderTimeout: 10 * time.Second}
	return httpSrv.Serve(ln)
}

// Server serves HTTP API handlers. It keeps only local UI session state: the
// per-run CSRF token and the active project selected in the browser.
type Server struct {
	lang        string
	token       string
	launchAgent func(agent, exe, root string) error
	versionHome string
	projectHome string
	mu          sync.RWMutex
	activeRoot  string
}

// guard enforces the two protections that make a browser-reachable control
// center safe even though it binds to loopback:
//
//   - a Host-header allowlist, which defeats DNS-rebinding (a malicious page can
//     resolve its own hostname to 127.0.0.1, but the rebound request still
//     carries the attacker's Host, not a loopback name); and
//   - a per-run token on every mutating request, which defeats cross-origin
//     CSRF (a foreign page cannot read the token out of our same-origin HTML, so
//     it cannot forge the header — and setting a custom header would anyway trip
//     a CORS preflight we never approve).
//
// Read-only GETs are exempt from the token: the same-origin policy already keeps
// their responses from a foreign page, and they mutate nothing.
func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			http.Error(w, "coact ui: forbidden host", http.StatusForbidden)
			return
		}
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if !s.validToken(r) {
				http.Error(w, "coact ui: missing or invalid token", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) validToken(r *http.Request) bool {
	return s.token != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Coact-Token")), []byte(s.token)) == 1
}

// isLoopbackHost reports whether the request's Host header names this machine.
func isLoopbackHost(host string) bool {
	h := host
	if hostOnly, _, err := net.SplitHostPort(host); err == nil {
		h = hostOnly
	}
	h = strings.TrimSuffix(strings.TrimPrefix(h, "["), "]")
	if h == "localhost" {
		return true
	}
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("/assets/", s.handleAsset)
	mux.HandleFunc("/world/", s.handleWorldAsset)
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/init", s.handleInit)
	mux.HandleFunc("/api/brief", s.handleBrief)
	mux.HandleFunc("/api/agents/", s.handleAgentAction)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskAction)
	mux.HandleFunc("/api/handoff", s.handleHandoff)
	mux.HandleFunc("/api/messages", s.handleMessages)
	mux.HandleFunc("/api/launch-commands", s.handleLaunchCommands)
	mux.HandleFunc("/api/projects", s.handleProjects)
	mux.HandleFunc("/api/projects/active", s.handleProjectActive)
	mux.HandleFunc("/api/terminal-mirror", s.handleTerminalMirror)
	mux.HandleFunc("/api/versions/", s.handleVersionAction)
}

func (s *Server) handleWorldAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/world/")
	contentType := ""
	assetPath := "world/" + name
	switch name {
	case "world.css":
		contentType = "text/css; charset=utf-8"
	case "world.js":
		contentType = "text/javascript; charset=utf-8"
	case "assets/station-orbit.png", "assets/station-ocean.png", "assets/station-ecodome.png", "assets/station-wasteland.png", "assets/crew-atlas-v2.png", "assets/crew-atlas-ocean.png", "assets/crew-atlas-ecodome.png", "assets/crew-atlas-wasteland.png":
		contentType = "image/png"
	default:
		http.NotFound(w, r)
		return
	}
	data, err := embeddedWorld.ReadFile(assetPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	if r.Method == http.MethodGet {
		_, _ = w.Write(data)
	}
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/assets/")
	switch name {
	case "astro-orbit-idle.png", "astro-orbit-work.png", "astro-orbit-plan.png", "astro-orbit-offline.png", "astro-orbit-walk-a.png", "astro-orbit-walk-b.png", "astro-orbit-celebrate.png",
		"astro-nova-idle.png", "astro-nova-work.png", "astro-nova-plan.png", "astro-nova-offline.png", "astro-nova-walk-a.png", "astro-nova-walk-b.png", "astro-nova-celebrate.png",
		"astro-comet-idle.png", "astro-comet-work.png", "astro-comet-plan.png", "astro-comet-offline.png", "astro-comet-walk-a.png", "astro-comet-walk-b.png", "astro-comet-celebrate.png":
	default:
		http.NotFound(w, r)
		return
	}
	data, err := embeddedAssets.ReadFile("assets/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	if r.Method == http.MethodGet {
		_, _ = w.Write(data)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	page := strings.ReplaceAll(indexHTML, "__COACT_TOKEN__", s.token)
	page = strings.ReplaceAll(page, "__COACT_LANG__", s.lang)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(page))
}

type stateResponse struct {
	Workspace   string                 `json:"workspace"`
	Initialized bool                   `json:"initialized"`
	Mode        string                 `json:"mode"`
	Version     string                 `json:"version"`
	Commit      string                 `json:"commit"`
	Brief       string                 `json:"brief"`
	Tasks       []taskDTO              `json:"tasks"`
	Agents      []agentDTO             `json:"agents"`
	Locks       []lockmgr.Lock         `json:"locks"`
	Log         []map[string]string    `json:"log"`
	Versions    []versionmgr.LocalInfo `json:"versions"`
	Manifest    *versionmgr.Manifest   `json:"manifest,omitempty"`
	Projects    []projectDTO           `json:"projects"`
}

type taskDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	State string `json:"state"`
	Owner string `json:"owner"`
}

type agentDTO struct {
	ID          string `json:"id"`
	Adapter     string `json:"adapter"`
	Enforcement string `json:"enforcement"`
	Live        bool   `json:"live"`
	Status      string `json:"status"`
	CurrentTask string `json:"current_task"`
	Beat        string `json:"beat"`
}

type projectDTO struct {
	Name        string `json:"name"`
	Root        string `json:"root"`
	Initialized bool   `json:"initialized"`
	Active      bool   `json:"active"`
}

type projectRegistry struct {
	Projects []projectRecord `json:"projects"`
}

type projectRecord struct {
	Root     string `json:"root"`
	Name     string `json:"name,omitempty"`
	LastSeen string `json:"last_seen,omitempty"`
}

type terminalMirrorDTO struct {
	Agent     string `json:"agent"`
	Exists    bool   `json:"exists"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Tail      string `json:"tail"`
	Screen    string `json:"screen,omitempty"`
	Truncated bool   `json:"truncated"`
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	st, err := s.state()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *Server) state() (*stateResponse, error) {
	p, initialized, err := s.currentProject()
	if err != nil {
		return nil, err
	}

	st := &stateResponse{
		Workspace:   p.Root,
		Initialized: initialized,
		Version:     buildinfo.Version,
		Commit:      buildinfo.Commit,
		Versions:    versionmgr.LocalVersions(s.managedHome()),
		Manifest:    versionmgr.BundledManifest(),
		Projects:    s.projectList(p.Root),
	}
	if st.Manifest != nil && st.Manifest.Version == "dev" {
		st.Manifest.Version = buildinfo.Version
	}
	cfg := config.Default()
	if initialized {
		if loaded, err := config.Load(p.ConfigPath()); err == nil {
			cfg = loaded
		}
		st.Mode = cfg.Mode
		st.Brief = readString(p.BriefPath())
		st.Tasks = readTasks(p.BoardPath())
		st.Agents = readAgents(p, cfg)
		m := lockmgr.New(p, cfg)
		if locks, err := m.List(); err == nil {
			st.Locks = locks
		}
		if recs, err := journal.ReadRecent(p.JournalDir(), 50); err == nil {
			st.Log = recs
		}
	} else {
		st.Mode = cfg.Mode
		for _, ad := range adapter.Defaults() {
			st.Agents = append(st.Agents, agentDTO{
				ID:          ad.ID,
				Adapter:     ad.Binary,
				Enforcement: ad.Enforcement(),
			})
		}
	}
	return st, nil
}

func readTasks(path string) []taskDTO {
	b, err := board.Load(path)
	if err != nil {
		return nil
	}
	var tasks []taskDTO
	for _, t := range b.Tasks() {
		tasks = append(tasks, taskDTO{ID: t.ID, Title: t.Title, State: t.State, Owner: t.Owner})
	}
	return tasks
}

func readAgents(p *project.Project, cfg *config.Config) []agentDTO {
	sessions, _ := presence.List(p.SessionDir())
	byID := map[string]*presence.Session{}
	for _, ss := range sessions {
		byID[ss.Agent] = ss
	}

	var out []agentDTO
	for _, ac := range cfg.Agents {
		ad, _ := adapter.Get(ac.ID)
		dto := agentDTO{ID: ac.ID, Adapter: ac.Adapter, Enforcement: ad.Enforcement()}
		if dto.Adapter == "" {
			dto.Adapter = ad.Binary
		}
		if ss := byID[ac.ID]; ss != nil {
			dto.Live = presence.IsLive(p.SessionDir(), ss.Agent, cfg.Presence.TTLSeconds)
			dto.Status = ss.Status
			dto.CurrentTask = ss.CurrentTask
			if age, ok := ss.Age(); ok {
				dto.Beat = shortDuration(age) + " ago"
			}
		}
		out = append(out, dto)
	}
	return out
}

func configuredAdapters(p *project.Project) []adapter.Adapter {
	cfg := config.Default()
	if p != nil {
		if loaded, err := config.Load(p.ConfigPath()); err == nil {
			cfg = loaded
		}
	}
	var out []adapter.Adapter
	for _, agent := range cfg.Agents {
		if ad, ok := adapter.Get(agent.ID); ok {
			out = append(out, ad)
		}
	}
	return out
}

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	p, _, err := s.currentProject()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	res, err := setup.Initialize(p.Root, "human")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.setActiveRoot(res.Root)
	_ = s.rememberProject(res.Root)
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleBrief(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	p, err := s.requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := platform.AtomicWrite(p.BriefPath(), []byte(req.Text), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = journal.Append(p.JournalDir(), "human", "brief.save", nil)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	p, err := s.requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		Title string `json:"title"`
		Owner string `json:"owner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	if err := board.ValidateTitle(title); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var out *board.Task
	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		out = b.Add(title)
		owner := sanitizeAgent(req.Owner)
		if owner != "" {
			if assigned, err := b.Assign(out.ID, owner); err == nil {
				out = assigned
			} else {
				return err
			}
		}
		if err := b.Save(); err != nil {
			return err
		}
		event := "task.add"
		agent := "human"
		meta := map[string]string{"id": out.ID}
		if out.Owner != "" {
			event = "task.schedule"
			meta["owner"] = out.Owner
		}
		_ = journal.Append(p.JournalDir(), agent, event, meta)
		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, taskDTO{ID: out.ID, Title: out.Title, State: out.State, Owner: out.Owner})
}

func (s *Server) handleTaskAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	p, err := s.requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	id, action, ok := parseTaskAction(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	var req struct {
		Owner string `json:"owner"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	owner := sanitizeAgent(req.Owner)
	if (action == "assign" || action == "claim" || action == "done") && owner == "" {
		writeError(w, http.StatusBadRequest, errors.New("owner is required"))
		return
	}

	var out *board.Task
	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		switch action {
		case "assign":
			out, err = b.Assign(id, owner)
		case "claim":
			out, err = b.Claim(id, owner, 1800)
		case "done":
			out, err = b.Finish(id, owner)
		case "unassign":
			out, err = b.Unassign(id)
		case "reopen":
			out, err = b.Reopen(id)
		default:
			return fmt.Errorf("unknown task action %q", action)
		}
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			return err
		}
		eventActor := owner
		if eventActor == "" {
			eventActor = "human"
		}
		event := "task." + action
		if action == "done" {
			event = "task.finish"
		}
		meta := map[string]string{"id": out.ID}
		if owner != "" {
			meta["owner"] = owner
		}
		_ = journal.Append(p.JournalDir(), eventActor, event, meta)
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, taskDTO{ID: out.ID, Title: out.Title, State: out.State, Owner: out.Owner})
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	p, err := s.requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	from := sanitizeAgent(req.From)
	to := sanitizeAgent(req.To)
	text := strings.TrimSpace(req.Text)
	if from == "" {
		from = "human"
	}
	if to == "" || text == "" {
		writeError(w, http.StatusBadRequest, errors.New("to and text are required"))
		return
	}
	if err := inbox.Send(p.InboxDir(), from, to, text); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = journal.Append(p.JournalDir(), from, "msg.send", map[string]string{"to": to, "via": "ui"})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleHandoff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	p, err := s.requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	from := sanitizeAgent(req.From)
	to := sanitizeAgent(req.To)
	if from == "" || to == "" || from == to {
		writeError(w, http.StatusBadRequest, errors.New("handoff requires two different agents"))
		return
	}
	cfg, err := config.Load(p.ConfigPath())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	known := map[string]bool{}
	for _, agent := range cfg.Agents {
		known[agent.ID] = true
	}
	if !known[from] || !known[to] {
		writeError(w, http.StatusBadRequest, errors.New("handoff agents must be configured in this workspace"))
		return
	}

	var moved []string
	var released int
	manager := lockmgr.New(p, cfg)
	err = withBoardLock(p, func() error {
		originalBoard, err := os.ReadFile(p.BoardPath())
		if err != nil {
			return err
		}
		sharedBoard, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		moved = sharedBoard.Reassign(from, to)
		if err := sharedBoard.Save(); err != nil {
			return err
		}
		released, err = manager.ReleaseAll(from)
		if err != nil {
			if rollbackErr := platform.AtomicWrite(p.BoardPath(), originalBoard, 0o644); rollbackErr != nil {
				return fmt.Errorf("releasing locks: %v; rolling back board: %v", err, rollbackErr)
			}
			return fmt.Errorf("releasing locks: %w (board rolled back)", err)
		}
		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	note := strings.TrimSpace(req.Note)
	message := fmt.Sprintf("Human-approved handoff from %s.", from)
	if len(moved) > 0 {
		message += " Tasks now yours: " + strings.Join(moved, ", ") + "."
	}
	if note != "" {
		message += " Note: " + note
	}
	notifyErr := inbox.Send(p.InboxDir(), "human", to, message)
	notified := notifyErr == nil
	_ = journal.Append(p.JournalDir(), "human", "handoff", map[string]string{
		"from": from, "to": to, "tasks": strings.Join(moved, ","), "released_locks": fmt.Sprintf("%d", released), "notified": fmt.Sprintf("%t", notified), "via": "ui",
	})
	response := map[string]any{"ok": true, "tasks": moved, "released_locks": released, "notified": notified}
	if notifyErr != nil {
		response["warning"] = "tasks moved, but recipient notification could not be written"
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleAgentAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	agent, action, ok := parseAgentAction(r.URL.Path)
	if !ok || action != "launch" {
		http.NotFound(w, r)
		return
	}
	ad, ok := adapter.Get(agent)
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Errorf("unknown adapter %q", agent))
		return
	}
	p, err := s.requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := exec.LookPath(ad.Binary); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("%q is not installed or not on PATH", ad.Binary))
		return
	}
	exe := "coact"
	if e, err := os.Executable(); err == nil {
		exe = e
	}
	launcher := s.launchAgent
	if launcher == nil {
		launcher = launchAgentTerminal
	}
	if err := launcher(agent, exe, p.Root); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = journal.Append(p.JournalDir(), "human", "agent.launch", map[string]string{"agent": agent, "via": "ui"})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "agent": agent})
}

type launchCommandDTO struct {
	Agent             string `json:"agent"`
	Command           string `json:"command"`
	Installed         bool   `json:"installed"`
	BinaryPath        string `json:"binary_path,omitempty"`
	TerminalSupported bool   `json:"terminal_supported"`
}

func (s *Server) handleLaunchCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	exe := "coact"
	if e, err := os.Executable(); err == nil {
		exe = e
	}
	p, _, err := s.currentProject()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var commands []launchCommandDTO
	for _, ad := range configuredAdapters(p) {
		binPath, err := exec.LookPath(ad.Binary)
		commands = append(commands, launchCommandDTO{
			Agent:             ad.ID,
			Command:           exe + " " + ad.ID,
			Installed:         err == nil,
			BinaryPath:        binPath,
			TerminalSupported: terminalLaunchSupported(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"commands": commands})
}

func (s *Server) handleTerminalMirror(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	p, err := s.requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	agent := sanitizeAgent(r.URL.Query().Get("agent"))
	full := r.URL.Query().Get("full") == "1" || strings.EqualFold(r.URL.Query().Get("full"), "true")
	if full && !s.validToken(r) {
		writeError(w, http.StatusForbidden, errors.New("full transcript requires UI token"))
		return
	}
	if full && agent == "" {
		writeError(w, http.StatusBadRequest, errors.New("full transcript requires an agent"))
		return
	}
	var agents []adapter.Adapter
	if agent != "" {
		ad, ok := adapter.Get(agent)
		if !ok {
			writeError(w, http.StatusBadRequest, fmt.Errorf("unknown adapter %q", agent))
			return
		}
		agents = []adapter.Adapter{ad}
	} else {
		agents = configuredAdapters(p)
	}
	limit := int64(24 * 1024)
	if full {
		limit = 0
	}
	var mirrors []terminalMirrorDTO
	for _, ad := range agents {
		mirrors = append(mirrors, readTerminalMirror(p, ad.ID, limit))
	}
	writeJSON(w, http.StatusOK, map[string]any{"mirrors": mirrors})
}

func (s *Server) handleVersionAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	version, action, ok := parseVersionAction(r.URL.Path)
	if !ok || action != "switch" {
		http.NotFound(w, r)
		return
	}
	if err := versionmgr.Switch(s.managedHome(), version); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if p, err := s.requireProject(); err == nil {
		_ = journal.Append(p.JournalDir(), "human", "version.switch", map[string]string{"version": version, "via": "ui"})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": version})
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		p, _, err := s.currentProject()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"projects": s.projectList(p.Root), "active": p.Root})
	case http.MethodPost:
		var req struct {
			Root string `json:"root"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		root := strings.TrimSpace(req.Root)
		if root == "" {
			located, err := project.Locate()
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			root = located
		}
		p, _, err := s.projectFromRoot(root)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		s.setActiveRoot(p.Root)
		if err := s.rememberProject(p.Root); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "root": p.Root, "projects": s.projectList(p.Root)})
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleProjectActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req struct {
		Root string `json:"root"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	p, _, err := s.projectFromRoot(req.Root)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.setActiveRoot(p.Root)
	if err := s.rememberProject(p.Root); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "root": p.Root, "projects": s.projectList(p.Root)})
}

func (s *Server) managedHome() string {
	if s.versionHome != "" {
		return s.versionHome
	}
	return versionmgr.DefaultHome()
}

func (s *Server) projectsHome() string {
	if s.projectHome != "" {
		return s.projectHome
	}
	return versionmgr.DefaultHome()
}

func (s *Server) projectRegistryPath() string {
	return filepath.Join(s.projectsHome(), "projects.json")
}

func (s *Server) currentProject() (*project.Project, bool, error) {
	if root := s.activeRootValue(); root != "" {
		p, initialized, err := s.projectFromRoot(root)
		if err == nil {
			_ = s.rememberProject(p.Root)
			return p, initialized, nil
		}
		s.setActiveRoot("")
	}
	if p, err := project.Resolve(); err == nil {
		s.setActiveRoot(p.Root)
		_ = s.rememberProject(p.Root)
		return p, true, nil
	}
	root, err := project.Locate()
	if err != nil {
		return nil, false, err
	}
	p, initialized, err := s.projectFromRoot(root)
	if err != nil {
		return nil, false, err
	}
	s.setActiveRoot(p.Root)
	_ = s.rememberProject(p.Root)
	return p, initialized, nil
}

func (s *Server) requireProject() (*project.Project, error) {
	p, initialized, err := s.currentProject()
	if err != nil || !initialized {
		return nil, errors.New("coact is not initialized here")
	}
	return p, nil
}

func (s *Server) projectFromRoot(root string) (*project.Project, bool, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, false, errors.New("project root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, false, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, false, err
	}
	if !info.IsDir() {
		return nil, false, fmt.Errorf("%q is not a directory", abs)
	}
	if p, err := project.ResolveFrom(abs); err == nil {
		return p, true, nil
	}
	return &project.Project{Root: abs, CheckoutRoot: abs}, false, nil
}

func (s *Server) activeRootValue() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeRoot
}

func (s *Server) setActiveRoot(root string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeRoot = root
}

func (s *Server) projectList(activeRoot string) []projectDTO {
	reg := s.loadProjectRegistry()
	seen := map[string]bool{}
	var out []projectDTO
	for _, rec := range reg.Projects {
		p, initialized, err := s.projectFromRoot(rec.Root)
		if err != nil || seen[p.Root] {
			continue
		}
		seen[p.Root] = true
		name := strings.TrimSpace(rec.Name)
		if name == "" {
			name = filepath.Base(p.Root)
		}
		out = append(out, projectDTO{Name: name, Root: p.Root, Initialized: initialized, Active: p.Root == activeRoot})
	}
	if activeRoot != "" && !seen[activeRoot] {
		if p, initialized, err := s.projectFromRoot(activeRoot); err == nil {
			out = append(out, projectDTO{Name: filepath.Base(p.Root), Root: p.Root, Initialized: initialized, Active: true})
		}
	}
	return out
}

func (s *Server) rememberProject(root string) error {
	p, _, err := s.projectFromRoot(root)
	if err != nil {
		return err
	}
	reg := s.loadProjectRegistry()
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range reg.Projects {
		if reg.Projects[i].Root == p.Root {
			if reg.Projects[i].Name == "" || reg.Projects[i].LastSeen == "" {
				reg.Projects[i].Name = filepath.Base(p.Root)
				reg.Projects[i].LastSeen = now
				return s.saveProjectRegistry(reg)
			}
			return nil
		}
	}
	reg.Projects = append(reg.Projects, projectRecord{Root: p.Root, Name: filepath.Base(p.Root), LastSeen: now})
	return s.saveProjectRegistry(reg)
}

func (s *Server) loadProjectRegistry() projectRegistry {
	var reg projectRegistry
	data, err := os.ReadFile(s.projectRegistryPath())
	if err != nil {
		return reg
	}
	_ = json.Unmarshal(data, &reg)
	return reg
}

func (s *Server) saveProjectRegistry(reg projectRegistry) error {
	if err := os.MkdirAll(s.projectsHome(), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return platform.AtomicWrite(s.projectRegistryPath(), data, 0o600)
}

func withBoardLock(p *project.Project, fn func() error) error {
	lockPath := filepath.Join(p.CoactDir(), "board.lock")
	if err := metalock.Acquire(lockPath, 5*time.Second, 10*time.Second); err != nil {
		return err
	}
	defer metalock.Release(lockPath)
	return fn()
}

func parseTaskAction(path string) (id, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/api/tasks/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func parseAgentAction(path string) (agent, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/api/agents/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return sanitizeAgent(parts[0]), parts[1], true
}

func parseVersionAction(path string) (version, action string, ok bool) {
	rest := strings.TrimPrefix(path, "/api/versions/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || parts[1] == "" {
		return "", "", false
	}
	version = strings.TrimSpace(parts[0])
	if strings.ContainsAny(version, `/\`) || strings.Contains(version, "..") {
		return "", "", false
	}
	return version, parts[1], true
}

func sanitizeAgent(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	var b strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func readString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func terminalDir(root string) string {
	return filepath.Join(root, ".coact", "terminal")
}

func terminalLogPath(root, agent string) string {
	return filepath.Join(terminalDir(root), sanitizeAgent(agent)+".log")
}

func readTerminalMirror(p *project.Project, agent string, limit int64) terminalMirrorDTO {
	path := terminalLogPath(p.Root, agent)
	out := terminalMirrorDTO{Agent: agent, Path: path}
	info, err := os.Stat(path)
	if err != nil {
		return out
	}
	out.Exists = true
	out.Size = info.Size()
	out.UpdatedAt = info.ModTime().UTC().Format(time.RFC3339)

	file, err := os.Open(path)
	if err != nil {
		return out
	}
	defer file.Close()

	start := int64(0)
	if limit > 0 && out.Size > limit {
		start = out.Size - limit
		out.Truncated = true
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return out
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return out
	}
	out.Tail = string(data)
	out.Screen = terminalScreenSnapshot(data)
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
}

func shortDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func terminalLaunchSupported() bool {
	return runtime.GOOS == "darwin"
}

func launchAgentTerminal(agent, exe, root string) error {
	if !terminalLaunchSupported() {
		return fmt.Errorf("one-click terminal launch is currently supported on macOS only; run %s manually", shellQuote(exe+" "+agent))
	}
	cmd, logPath, err := agentTerminalCommand(agent, exe, root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return err
	}
	script := `tell application "Terminal" to do script "` + appleScriptQuote(cmd) + `"`
	return exec.Command("osascript", "-e", script).Start()
}

func agentTerminalCommand(agent, exe, root string) (string, string, error) {
	if agent = sanitizeAgent(agent); agent == "" {
		return "", "", errors.New("agent is required")
	}
	logPath := terminalLogPath(root, agent)
	cmd := "mkdir -p " + shellQuote(filepath.Dir(logPath)) +
		" && cd " + shellQuote(root) +
		" && script -q -F -a " + shellQuote(logPath) +
		" " + shellQuote(exe) + " " + shellQuote(agent)
	return cmd, logPath, nil
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func appleScriptQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}
