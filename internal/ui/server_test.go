package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/versionmgr"
)

func TestStateAndInitAPI(t *testing.T) {
	dir := chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()

	before := getState(t, ts)
	if before.Initialized {
		t.Fatal("fresh temp dir should not be initialized")
	}
	if before.Workspace != dir {
		t.Fatalf("workspace mismatch: got %q want %q", before.Workspace, dir)
	}
	if before.Manifest == nil || !before.Manifest.Supports.UI {
		t.Fatalf("state should include UI-capable version manifest: %#v", before.Manifest)
	}

	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)
	after := getState(t, ts)
	if !after.Initialized {
		t.Fatal("expected initialized state after /api/init")
	}
	if _, err := os.Stat(filepath.Join(dir, ".coact", "board.md")); err != nil {
		t.Fatalf("board not created: %v", err)
	}
}

func TestTaskClaimAndDoneAPI(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	var created taskDTO
	postJSON(t, ts, "/api/tasks", map[string]string{"title": "Implement local UI"}, http.StatusOK, &created)
	if created.ID == "" || created.State != "todo" {
		t.Fatalf("unexpected created task: %#v", created)
	}

	var claimed taskDTO
	postJSON(t, ts, "/api/tasks/"+created.ID+"/claim", map[string]string{"owner": "codex"}, http.StatusOK, &claimed)
	if claimed.Owner != "codex" || claimed.State != "doing" {
		t.Fatalf("unexpected claimed task: %#v", claimed)
	}

	var done taskDTO
	postJSON(t, ts, "/api/tasks/"+created.ID+"/done", map[string]string{"owner": "codex"}, http.StatusOK, &done)
	if done.Owner != "codex" || done.State != "done" {
		t.Fatalf("unexpected done task: %#v", done)
	}

	state := getState(t, ts)
	if !hasTask(state.Tasks, created.ID, "done", "codex") {
		t.Fatalf("state did not include done task: %#v", state.Tasks)
	}
}

func TestMessagesAPIWritesInboxAndJournal(t *testing.T) {
	dir := chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	postJSON(t, ts, "/api/messages", map[string]string{
		"from": "human",
		"to":   "claude",
		"text": "Please review the UI copy.",
	}, http.StatusOK, nil)

	inboxPath := filepath.Join(dir, ".coact", "inbox", "claude.md")
	data, err := os.ReadFile(inboxPath)
	if err != nil {
		t.Fatalf("read inbox: %v", err)
	}
	if !strings.Contains(string(data), "Please review the UI copy.") {
		t.Fatalf("message missing from inbox: %s", data)
	}

	state := getState(t, ts)
	if !hasJournalEvent(state.Log, "msg.send") {
		t.Fatalf("journal did not include msg.send: %#v", state.Log)
	}
}

func TestLaunchCommandsAPI(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/launch-commands")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"agent":"claude"`) || !strings.Contains(string(body), `"agent":"codex"`) {
		t.Fatalf("launch commands missing expected agents: %s", body)
	}
	if !strings.Contains(string(body), `"installed"`) || !strings.Contains(string(body), `"terminal_supported"`) {
		t.Fatalf("launch commands missing status fields: %s", body)
	}
}

func TestAgentLaunchAPIUsesInjectedLauncher(t *testing.T) {
	dir := chdirTemp(t)
	writeFakeBinary(t, dir, "claude")
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+oldPath)

	var gotAgent, gotExe, gotRoot string
	ts := newTestServerWithLauncher(t, func(agent, exe, root string) error {
		gotAgent, gotExe, gotRoot = agent, exe, root
		return nil
	})
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	postJSON(t, ts, "/api/agents/claude/launch", nil, http.StatusOK, nil)
	if gotAgent != "claude" {
		t.Fatalf("launcher agent = %q, want claude", gotAgent)
	}
	if gotExe == "" {
		t.Fatal("launcher should receive coact executable path")
	}
	if gotRoot != dir {
		t.Fatalf("launcher root = %q, want %q", gotRoot, dir)
	}
	state := getState(t, ts)
	if !hasJournalEvent(state.Log, "agent.launch") {
		t.Fatalf("journal did not include agent.launch: %#v", state.Log)
	}
}

func TestAgentLaunchRejectsUnknownAdapter(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	postJSON(t, ts, "/api/agents/not-real/launch", nil, http.StatusBadRequest, nil)
}

func TestVersionSwitchAPIUsesManagedHome(t *testing.T) {
	chdirTemp(t)
	home := t.TempDir()
	writeManagedVersion(t, home, "v0.1.0")
	writeManagedVersion(t, home, "v0.2.0")
	ts := newTestServerWithLauncherAndHome(t, nil, home)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	postJSON(t, ts, "/api/versions/v0.2.0/switch", nil, http.StatusOK, nil)
	state := getState(t, ts)
	if !hasLocalVersion(state.Versions, "v0.2.0", true) {
		t.Fatalf("v0.2.0 should be active after switch: %#v", state.Versions)
	}
	if !hasJournalEvent(state.Log, "version.switch") {
		t.Fatalf("journal did not include version.switch: %#v", state.Log)
	}
}

func TestVersionSwitchRejectsTraversal(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	postJSON(t, ts, "/api/versions/../switch", nil, http.StatusNotFound, nil)
}

func TestGuardRejectsForeignHost(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()

	// Simulate a DNS-rebinding request: connects to loopback but carries the
	// attacker's hostname in the Host header.
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/state", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "attacker.example.com"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("foreign Host should be rejected, got %d", resp.StatusCode)
	}
}

func TestGuardRejectsMutationWithoutToken(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()

	// A cross-origin CSRF POST cannot carry the per-run token.
	resp, err := http.Post(ts.URL+"/api/init", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mutation without token should be rejected, got %d", resp.StatusCode)
	}
}

func TestServeRejectsNonLocalAddress(t *testing.T) {
	err := Serve(Options{Addr: "0.0.0.0", Port: 7331, OpenBrowser: false})
	if err == nil {
		t.Fatal("expected non-local bind rejection")
	}
	if !strings.Contains(err.Error(), "local addresses") {
		t.Fatalf("unexpected error: %v", err)
	}
}

const testToken = "test-token"

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newTestServerWithLauncher(t, nil)
}

func newTestServerWithLauncher(t *testing.T, launcher func(agent, exe, root string) error) *httptest.Server {
	t.Helper()
	return newTestServerWithLauncherAndHome(t, launcher, "")
}

func newTestServerWithLauncherAndHome(t *testing.T, launcher func(agent, exe, root string) error, versionHome string) *httptest.Server {
	t.Helper()
	srv := &Server{token: testToken, launchAgent: launcher, versionHome: versionHome}
	mux := http.NewServeMux()
	srv.routes(mux)
	// Route through guard so tests exercise the real Host + token defenses.
	return httptest.NewServer(srv.guard(mux))
}

func chdirTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(old)
	})
	return cwd
}

func getState(t *testing.T, ts *httptest.Server) stateResponse {
	t.Helper()
	var state stateResponse
	resp, err := http.Get(ts.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("state status %d: %s", resp.StatusCode, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		t.Fatal(err)
	}
	return state
}

func postJSON(t *testing.T, ts *httptest.Server, path string, body any, want int, out any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(http.MethodPost, ts.URL+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Coact-Token", testToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s status %d want %d: %s", path, resp.StatusCode, want, data)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
}

func hasTask(tasks []taskDTO, id, state, owner string) bool {
	for _, task := range tasks {
		if task.ID == id && task.State == state && task.Owner == owner {
			return true
		}
	}
	return false
}

func hasJournalEvent(records []map[string]string, event string) bool {
	for _, record := range records {
		if record["event"] == event {
			return true
		}
	}
	return false
}

func hasLocalVersion(versions []versionmgr.LocalInfo, version string, active bool) bool {
	for _, local := range versions {
		if local.Version == version && local.Active == active {
			return true
		}
	}
	return false
}

func writeFakeBinary(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	data := []byte("#!/bin/sh\nexit 0\n")
	if runtime.GOOS == "windows" {
		path += ".bat"
		data = []byte("@echo off\r\nexit /b 0\r\n")
	}
	if err := os.WriteFile(path, data, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeManagedVersion(t *testing.T, home, version string) {
	t.Helper()
	if err := os.MkdirAll(versionmgr.BinDir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(versionmgr.BinDir(home), versionmgr.BinaryName(version))
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}
