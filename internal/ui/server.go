// Package ui implements CoAct's local-only control center.
package ui

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// Server serves HTTP API handlers. It is stateless apart from the per-run CSRF
// token minted at startup.
type Server struct {
	lang        string
	token       string
	launchAgent func(agent, exe, root string) error
	versionHome string
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
			if s.token == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Coact-Token")), []byte(s.token)) != 1 {
				http.Error(w, "coact ui: missing or invalid token", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
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
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/init", s.handleInit)
	mux.HandleFunc("/api/brief", s.handleBrief)
	mux.HandleFunc("/api/agents/", s.handleAgentAction)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskAction)
	mux.HandleFunc("/api/messages", s.handleMessages)
	mux.HandleFunc("/api/launch-commands", s.handleLaunchCommands)
	mux.HandleFunc("/api/versions/", s.handleVersionAction)
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
	p, err := project.Resolve()
	initialized := err == nil
	if !initialized {
		root, locateErr := project.Locate()
		if locateErr != nil {
			return nil, locateErr
		}
		p = &project.Project{Root: root, CheckoutRoot: root}
	}

	st := &stateResponse{
		Workspace:   p.Root,
		Initialized: initialized,
		Version:     buildinfo.Version,
		Commit:      buildinfo.Commit,
		Versions:    versionmgr.LocalVersions(s.managedHome()),
		Manifest:    versionmgr.BundledManifest(),
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
		for _, ad := range adapter.All() {
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

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	root, err := project.Locate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	res, err := setup.Initialize(root, "human")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleBrief(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	p, err := requireProject()
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
	p, err := requireProject()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		Title string `json:"title"`
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
	var out *board.Task
	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		out = b.Add(title)
		if err := b.Save(); err != nil {
			return err
		}
		_ = journal.Append(p.JournalDir(), "human", "task.add", map[string]string{"id": out.ID})
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
	p, err := requireProject()
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
	if owner == "" {
		owner = "human"
	}

	var out *board.Task
	err = withBoardLock(p, func() error {
		b, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		switch action {
		case "claim":
			out, err = b.Claim(id, owner, 1800)
		case "done":
			out, err = b.Finish(id, owner)
		default:
			return fmt.Errorf("unknown task action %q", action)
		}
		if err != nil {
			return err
		}
		if err := b.Save(); err != nil {
			return err
		}
		event := "task." + action
		if action == "done" {
			event = "task.finish"
		}
		_ = journal.Append(p.JournalDir(), owner, event, map[string]string{"id": out.ID})
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
	p, err := requireProject()
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
	p, err := requireProject()
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
	var commands []launchCommandDTO
	for _, ad := range adapter.All() {
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
	if p, err := requireProject(); err == nil {
		_ = journal.Append(p.JournalDir(), "human", "version.switch", map[string]string{"version": version, "via": "ui"})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": version})
}

func (s *Server) managedHome() string {
	if s.versionHome != "" {
		return s.versionHome
	}
	return versionmgr.DefaultHome()
}

func requireProject() (*project.Project, error) {
	p, err := project.Resolve()
	if err != nil {
		return nil, errors.New("coact is not initialized here")
	}
	return p, nil
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
	cmd := "cd " + shellQuote(root) + " && " + shellQuote(exe) + " " + shellQuote(agent)
	script := `tell application "Terminal" to do script "` + appleScriptQuote(cmd) + `"`
	return exec.Command("osascript", "-e", script).Start()
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
