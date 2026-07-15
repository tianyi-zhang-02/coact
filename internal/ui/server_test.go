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

	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/project"
	"github.com/tianyi-zhang-02/coact/internal/setup"
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

func TestTaskActionsEnforceLifecycle(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	var created taskDTO
	postJSON(t, ts, "/api/tasks", map[string]string{"title": "Audit task lifecycle"}, http.StatusOK, &created)
	postJSON(t, ts, "/api/tasks/"+created.ID+"/done", map[string]string{"owner": "codex"}, http.StatusBadRequest, nil)

	var assigned taskDTO
	postJSON(t, ts, "/api/tasks/"+created.ID+"/assign", map[string]string{"owner": "codex"}, http.StatusOK, &assigned)
	if assigned.State != "claimed" || assigned.Owner != "codex" {
		t.Fatalf("unexpected assigned task: %#v", assigned)
	}
	postJSON(t, ts, "/api/tasks/"+created.ID+"/done", map[string]string{"owner": "codex"}, http.StatusBadRequest, nil)

	var unassigned taskDTO
	postJSON(t, ts, "/api/tasks/"+created.ID+"/unassign", nil, http.StatusOK, &unassigned)
	if unassigned.State != "todo" || unassigned.Owner != "" {
		t.Fatalf("unexpected unassigned task: %#v", unassigned)
	}

	postJSON(t, ts, "/api/tasks/"+created.ID+"/assign", map[string]string{"owner": "claude"}, http.StatusOK, nil)
	postJSON(t, ts, "/api/tasks/"+created.ID+"/claim", map[string]string{"owner": "claude"}, http.StatusOK, nil)
	postJSON(t, ts, "/api/tasks/"+created.ID+"/done", map[string]string{"owner": "claude"}, http.StatusOK, nil)

	var reopened taskDTO
	postJSON(t, ts, "/api/tasks/"+created.ID+"/reopen", nil, http.StatusOK, &reopened)
	if reopened.State != "todo" || reopened.Owner != "" {
		t.Fatalf("unexpected reopened task: %#v", reopened)
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

func TestHandoffAPIReassignsTaskAndJournals(t *testing.T) {
	dir := chdirTemp(t)
	if _, err := setup.Initialize(dir, "human"); err != nil {
		t.Fatal(err)
	}
	sharedBoard, err := board.Load(filepath.Join(dir, ".coact", "board.md"))
	if err != nil {
		t.Fatal(err)
	}
	created := sharedBoard.Add("Waiting implementation")
	if _, err := sharedBoard.Claim(created.ID, "claude", 1800); err != nil {
		t.Fatal(err)
	}
	if err := sharedBoard.Save(); err != nil {
		t.Fatal(err)
	}
	p := &project.Project{Root: dir, CheckoutRoot: dir}
	cfg, err := config.Load(p.ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	manager := lockmgr.New(p, cfg)
	lockResult, err := manager.Acquire("claude", filepath.Join(dir, "internal", "waiting"))
	if err != nil || !lockResult.Acquired {
		t.Fatalf("acquire handoff lock: result=%+v err=%v", lockResult, err)
	}

	srv := &Server{token: testToken, projectHome: t.TempDir()}
	body, _ := json.Marshal(map[string]string{"from": "claude", "to": "codex", "note": "Please continue from the shared brief."})
	req := httptest.NewRequest(http.MethodPost, "/api/handoff", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.handleHandoff(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("handoff status %d: %s", rec.Code, rec.Body.String())
	}
	var result struct {
		Tasks         []string `json:"tasks"`
		ReleasedLocks int      `json:"released_locks"`
		Notified      bool     `json:"notified"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Tasks) != 1 || result.Tasks[0] != created.ID {
		t.Fatalf("unexpected handoff result: %#v", result)
	}
	if result.ReleasedLocks != 1 || !result.Notified {
		t.Fatalf("handoff did not release and notify: %#v", result)
	}
	state, err := srv.state()
	if err != nil {
		t.Fatal(err)
	}
	if !hasTask(state.Tasks, created.ID, "claimed", "codex") {
		t.Fatalf("handoff did not reassign task: %#v", state.Tasks)
	}
	if !hasJournalEvent(state.Log, "handoff") {
		t.Fatalf("handoff missing from journal: %#v", state.Log)
	}
	inboxData, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", "codex.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(inboxData), "Human-approved handoff from claude") {
		t.Fatalf("handoff inbox message missing: %s", inboxData)
	}
}

func TestHandoffAPIRejectsUnknownAgent(t *testing.T) {
	dir := chdirTemp(t)
	if _, err := setup.Initialize(dir, "human"); err != nil {
		t.Fatal(err)
	}
	srv := &Server{token: testToken, projectHome: t.TempDir()}
	body, _ := json.Marshal(map[string]string{"from": "claude", "to": "unknown"})
	req := httptest.NewRequest(http.MethodPost, "/api/handoff", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.handleHandoff(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown agent status %d: %s", rec.Code, rec.Body.String())
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
	if !strings.Contains(string(body), `"agent":"antigravity"`) || strings.Contains(string(body), `"agent":"gemini"`) {
		t.Fatalf("fresh launch commands should prefer Antigravity over legacy Gemini: %s", body)
	}
	if !strings.Contains(string(body), `"installed"`) || !strings.Contains(string(body), `"terminal_supported"`) {
		t.Fatalf("launch commands missing status fields: %s", body)
	}
}

func TestIndexUsesSimplifiedSetupAndWorkPages(t *testing.T) {
	srv := &Server{token: "test-token", lang: "en"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.handleIndex(rec, req)
	body := rec.Body.Bytes()
	page := string(body)
	for _, want := range []string{"data-page-target=\"station\"><span>Station", "data-page-target=\"overview\"><span>Setup", "data-page-target=\"work\"><span>Work", "projectSelect", "taskOwner", "taskPrompt", "planBrief", "planApproval", "approvePlan", "/api/plans", "id=\"taskBoard\"", "data-task-filter=\"open\"", "assignTask", "unassignTask", "reopenTask", "class=\"card span-12 board-details\"", "id=\"workTerminals\"", "id=\"terminalCount\"", "work-terminal-tabs", "data-send-inbox-note", "id=\"coactPixelWorld\"", "id=\"coactPixelBackground\"", "id=\"worldAgentAssist\"", "data-world-assist-handoff", "data-world-quality", "theme-ambient", "ambient-whale", "guard-strip", "id=\"msgText\"", "data-toggle-ambient", "coactAmbientDecorations", "lastDashboardSignature", "terminalDetailsVisible", "refreshInFlight", "/api/handoff", "/world/world.css?v=10", "/world/world.js?v=28"} {
		if !strings.Contains(page, want) {
			t.Fatalf("index missing %q", want)
		}
	}
	for _, unwanted := range []string{"data-page-target=\"agents\"", "data-page-target=\"guide\"", "data-page-target=\"advanced\"", "id=\"messages-card\""} {
		if strings.Contains(page, unwanted) {
			t.Fatalf("index should not expose %q", unwanted)
		}
	}
}

func TestEmbeddedWorldAssetsAreAllowlisted(t *testing.T) {
	srv := &Server{}
	for path, contentType := range map[string]string{
		"/world/world.css":                       "text/css; charset=utf-8",
		"/world/world.js":                        "text/javascript; charset=utf-8",
		"/world/assets/station-orbit.png":        "image/png",
		"/world/assets/station-ocean.png":        "image/png",
		"/world/assets/station-ecodome.png":      "image/png",
		"/world/assets/station-wasteland.png":    "image/png",
		"/world/assets/crew-atlas-v2.png":        "image/png",
		"/world/assets/crew-atlas-ocean.png":     "image/png",
		"/world/assets/crew-atlas-ecodome.png":   "image/png",
		"/world/assets/crew-atlas-wasteland.png": "image/png",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		srv.handleWorldAsset(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != contentType {
			t.Fatalf("%s content type = %q, want %q", path, got, contentType)
		}
	}

	for _, path := range []string{"/world/unknown.js", "/world/../server.go", "/world/world.png", "/world/assets/orbital-station-v3.png", "/world/assets/station-nebula.png", "/world/assets/station-aurora.png", "/world/assets/station-solar.png"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		srv.handleWorldAsset(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", path, rec.Code)
		}
	}
}

func TestWorldHorizontalWalkKeepsOneFacingFrame(t *testing.T) {
	script, err := embeddedWorld.ReadFile("world/world.js")
	if err != nil {
		t.Fatal(err)
	}
	content := string(script)
	if !strings.Contains(content, "actor.direction==='up'?FRAME.up:['left','right'].includes(actor.direction)?FRAME.right:FRAME.down") {
		t.Fatal("walking must use one stable base frame per direction")
	}
	if strings.Contains(content, "walkSide") {
		t.Fatal("frame 8 is a front-running frame, not a side-walk frame")
	}
	for _, want := range []string{"drawWalkingCrew(actor", "entity.segmentKey!==segmentKey", "sourceY+split"} {
		if !strings.Contains(content, want) {
			t.Fatalf("world walking renderer missing %q", want)
		}
	}
	for _, removed := range []string{"drawCrewIdentity", "drawTaskProp"} {
		if strings.Contains(content, removed) {
			t.Fatalf("temporary external decoration should be removed: %s", removed)
		}
	}
}

func TestWorldRendererUsesChromeFriendlyFrameBudget(t *testing.T) {
	script, err := embeddedWorld.ReadFile("world/world.js")
	if err != nil {
		t.Fatal(err)
	}
	content := string(script)
	for _, want := range []string{
		"const MOVING_FRAME_INTERVAL = 16",
		"const CHROMIUM_MOVING_FRAME_INTERVAL = 33",
		"const CHROMIUM_PRESSURE_FRAME_INTERVAL = 40",
		"const CHROMIUM_BACKGROUND_FRAME_INTERVAL = 180",
		"const IDLE_FRAME_INTERVAL = 50",
		"const BACKGROUND_FRAME_INTERVAL = 50",
		"const SLOW_BACKGROUND_FRAME_INTERVAL = 220",
		"desynchronized:true",
		"this.chromium = IS_CHROMIUM",
		"foregroundInterval(entitiesMoving)",
		"backgroundInterval()",
		"this.scheduleFrame(180)",
		"this.averageFrameGap>24",
		"this.qualityPreference==='lite'",
		"if(this.paused){this.frameCount++",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("world performance budget missing %q", want)
		}
	}

	styles, err := embeddedWorld.ReadFile("world/world.css")
	if err != nil {
		t.Fatal(err)
	}
	css := string(styles)
	for _, want := range []string{"aspect-ratio:3/2", "width:min(100%,calc((100vh - 24px)*1.5))", ".pixel-world-shell.is-chromium .pixel-world-viewport canvas { transform:none; }"} {
		if !strings.Contains(css, want) {
			t.Fatalf("world canvas sizing missing %q", want)
		}
	}
}

func TestEmbeddedAssetsAreAllowlisted(t *testing.T) {
	srv := &Server{}
	for _, path := range []string{"/assets/astro-orbit-work.png", "/assets/astro-nova-walk-a.png", "/assets/astro-comet-celebrate.png"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		srv.handleAsset(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "image/png" {
			t.Fatalf("%s content type = %q", path, got)
		}
	}

	for _, path := range []string{"/assets/unknown.png", "/assets/orbital-station-bg.png", "/assets/../server.go"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		srv.handleAsset(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", path, rec.Code)
		}
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

func TestTaskAddCanScheduleOwner(t *testing.T) {
	root := chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	var created taskDTO
	postJSON(t, ts, "/api/tasks", map[string]string{"description": "Run focused UI tests", "prompt": "Run only the UI package tests and report failures.", "owner": "codex"}, http.StatusOK, &created)
	if created.Owner != "codex" || created.State != "claimed" {
		t.Fatalf("scheduled task = %#v, want owner codex claimed", created)
	}
	state := getState(t, ts)
	if !hasTask(state.Tasks, created.ID, "claimed", "codex") {
		t.Fatalf("state missing scheduled task: %#v", state.Tasks)
	}
	if !hasJournalEvent(state.Log, "task.schedule") {
		t.Fatalf("journal did not include task.schedule: %#v", state.Log)
	}
	prompt, err := os.ReadFile(filepath.Join(root, ".coact", "tasks", created.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(prompt), "Run only the UI package tests") {
		t.Fatalf("full prompt was not persisted: %s", prompt)
	}
	inbox, err := os.ReadFile(filepath.Join(root, ".coact", "inbox", "codex.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(inbox), "Full execution prompt") {
		t.Fatalf("assigned agent did not receive prompt: %s", inbox)
	}
}

func TestPlanAPIStartsReviewGatedLeadRun(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	var plan map[string]any
	postJSON(t, ts, "/api/plans", map[string]any{
		"brief": "Design and verify the task workflow", "lead": "codex", "approval_mode": "review", "participants": []string{"codex", "claude"},
	}, http.StatusOK, &plan)
	if plan["lead"] != "codex" || plan["approval_mode"] != "review" || plan["status"] != "pending" {
		t.Fatalf("plan = %#v", plan)
	}
	state := getState(t, ts)
	if state.Plan == nil || state.Plan.Lead != "codex" || state.Plan.ApprovalMode != "review" {
		t.Fatalf("state plan = %#v", state.Plan)
	}
}

func TestPlanHandlerStartsReviewRunWithoutShell(t *testing.T) {
	root := chdirTemp(t)
	if _, err := setup.Initialize(root, "human"); err != nil {
		t.Fatal(err)
	}
	srv := &Server{token: "test-token", activeRoot: root}
	body, _ := json.Marshal(map[string]any{
		"brief": "Plan through the shared service", "lead": "codex", "approval_mode": "review", "participants": []string{"codex", "claude"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/plans", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.handlePlans(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	state, err := srv.state()
	if err != nil {
		t.Fatal(err)
	}
	if state.Plan == nil || state.Plan.Lead != "codex" || state.Plan.ApprovalMode != "review" {
		t.Fatalf("state plan = %#v", state.Plan)
	}
}

func TestProjectsAPIAddsAndSwitchesActiveProject(t *testing.T) {
	dirA := chdirTemp(t)
	dirB := t.TempDir()
	ts := newTestServer(t)
	defer ts.Close()

	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)
	postJSON(t, ts, "/api/projects", map[string]string{"root": dirB}, http.StatusOK, nil)
	state := getState(t, ts)
	if state.Workspace != dirB || state.Initialized {
		t.Fatalf("workspace after adding dirB = %q initialized=%v", state.Workspace, state.Initialized)
	}
	if !hasProject(state.Projects, dirA, false) || !hasProject(state.Projects, dirB, true) {
		t.Fatalf("project list missing expected roots: %#v", state.Projects)
	}

	postJSON(t, ts, "/api/projects/active", map[string]string{"root": dirA}, http.StatusOK, nil)
	state = getState(t, ts)
	if state.Workspace != dirA || !state.Initialized {
		t.Fatalf("workspace after switching back = %q initialized=%v", state.Workspace, state.Initialized)
	}
}

func TestTerminalMirrorAPIReadsTranscriptTail(t *testing.T) {
	dir := chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	logPath := filepath.Join(dir, ".coact", "terminal", "claude.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, []byte(strings.Repeat("x", 30*1024)+"latest output\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var got struct {
		Mirrors []terminalMirrorDTO `json:"mirrors"`
	}
	getJSON(t, ts, "/api/terminal-mirror?agent=claude", http.StatusOK, &got)
	if len(got.Mirrors) != 1 {
		t.Fatalf("expected one mirror, got %#v", got.Mirrors)
	}
	mirror := got.Mirrors[0]
	if mirror.Agent != "claude" || !mirror.Exists {
		t.Fatalf("unexpected mirror: %#v", mirror)
	}
	if !mirror.Truncated {
		t.Fatalf("large transcript should be truncated: %#v", mirror)
	}
	if !strings.Contains(mirror.Tail, "latest output") {
		t.Fatalf("tail missing latest output: %#v", mirror.Tail)
	}
	if !strings.Contains(mirror.Screen, "latest output") {
		t.Fatalf("screen missing latest output: %#v", mirror.Screen)
	}
}

func TestTerminalMirrorAPIFullTranscriptRequiresToken(t *testing.T) {
	dir := chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	logPath := filepath.Join(dir, ".coact", "terminal", "claude.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Repeat("begin ", 6*1024) + "complete output\n"
	if err := os.WriteFile(logPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	getJSON(t, ts, "/api/terminal-mirror?agent=claude&full=1", http.StatusForbidden, nil)

	var got struct {
		Mirrors []terminalMirrorDTO `json:"mirrors"`
	}
	getJSONWithToken(t, ts, "/api/terminal-mirror?agent=claude&full=1", http.StatusOK, &got)
	if len(got.Mirrors) != 1 {
		t.Fatalf("expected one mirror, got %#v", got.Mirrors)
	}
	mirror := got.Mirrors[0]
	if mirror.Truncated {
		t.Fatalf("full transcript should not be truncated: %#v", mirror)
	}
	if mirror.Tail != body {
		t.Fatalf("full transcript mismatch: got %d bytes want %d", len(mirror.Tail), len(body))
	}
}

func TestTerminalMirrorRejectsUnknownAgent(t *testing.T) {
	chdirTemp(t)
	ts := newTestServer(t)
	defer ts.Close()
	postJSON(t, ts, "/api/init", nil, http.StatusOK, nil)

	getJSON(t, ts, "/api/terminal-mirror?agent=not-real", http.StatusBadRequest, nil)
}

func TestAgentTerminalCommandUsesScriptTranscript(t *testing.T) {
	cmd, logPath, err := agentTerminalCommand("claude", "/tmp/coact test", "/tmp/project root")
	if err != nil {
		t.Fatal(err)
	}
	if logPath != filepath.Join("/tmp/project root", ".coact", "terminal", "claude.log") {
		t.Fatalf("logPath = %q", logPath)
	}
	for _, want := range []string{"script -q -F -a", shellQuote(logPath), shellQuote("/tmp/coact test"), shellQuote("claude")} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command %q missing %q", cmd, want)
		}
	}
	if strings.Contains(cmd, "| tee") {
		t.Fatalf("command should preserve PTY with script, not pipe through tee: %q", cmd)
	}
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
	srv := &Server{token: testToken, launchAgent: launcher, versionHome: versionHome, projectHome: t.TempDir()}
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
	getJSON(t, ts, "/api/state", http.StatusOK, &state)
	return state
}

func getJSON(t *testing.T, ts *httptest.Server, path string, want int, out any) {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s status %d want %d: %s", path, resp.StatusCode, want, body)
	}
	if out == nil {
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}

func getJSONWithToken(t *testing.T, ts *httptest.Server, path string, want int, out any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Coact-Token", testToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s status %d want %d: %s", path, resp.StatusCode, want, body)
	}
	if out == nil {
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
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

func hasProject(projects []projectDTO, root string, active bool) bool {
	for _, project := range projects {
		if project.Root == root && project.Active == active {
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
