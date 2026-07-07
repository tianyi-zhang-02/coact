package ui

const indexHTML = `<!doctype html>
<html lang="__COACT_LANG__">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>CoAct Control Center</title>
  <style>
    :root {
      color-scheme: dark;
      --bg:#090a0d;
      --rail:#0f1117;
      --surface:#141720;
      --surface-2:#191d28;
      --surface-3:#202636;
      --text:#f4f1ea;
      --muted:#9ba3af;
      --soft:#c8ced8;
      --line:#2a303c;
      --line-2:#394150;
      --accent:#d6b16a;
      --accent-2:#8ab4f8;
      --ok:#8bd9a8;
      --warn:#f3c969;
      --bad:#f07d90;
      --shadow:0 18px 54px #0008;
      --radius:18px;
    }
    * { box-sizing:border-box; }
    body {
      margin:0;
      min-height:100vh;
      font-family:Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background:
        radial-gradient(circle at 78% -10%, #253044 0 18rem, transparent 34rem),
        linear-gradient(135deg, #0a0b0f, #0d1017 48%, #090a0d);
      color:var(--text);
    }
    h1, h2, h3, p { margin:0; }
    h1 { font-size:30px; line-height:1.05; letter-spacing:-.045em; }
    h2 { font-size:17px; line-height:1.2; letter-spacing:-.025em; }
    p { line-height:1.55; }
    a { color:inherit; text-decoration:none; }
    button, input, textarea, select {
      font:inherit;
      color:var(--text);
      border:1px solid var(--line);
      border-radius:12px;
      background:#0e1118;
      padding:10px 12px;
      outline:none;
    }
    button {
      cursor:pointer;
      background:linear-gradient(180deg, #262d3b, #1a1f2b);
      border-color:#3b4558;
      box-shadow:0 8px 22px #0005;
      transition:border-color .15s ease, transform .15s ease, background .15s ease;
      white-space:nowrap;
    }
    button:hover { border-color:#5b6679; transform:translateY(-1px); }
    button:active { transform:translateY(0); }
    button:disabled {
      cursor:not-allowed;
      opacity:.48;
      transform:none;
    }
    button.ghost { background:#11151e; box-shadow:none; color:var(--soft); }
    button.small { padding:7px 10px; border-radius:10px; font-size:12px; }
    input, select { width:100%; }
    textarea { width:100%; min-height:142px; resize:vertical; line-height:1.5; }
    input::placeholder, textarea::placeholder { color:#777f8d; }
    input:focus, textarea:focus, select:focus { border-color:var(--accent); box-shadow:0 0 0 4px #d6b16a20; }
    code {
      display:inline-block;
      max-width:100%;
      overflow:auto;
      vertical-align:middle;
      background:#0a0c11;
      border:1px solid var(--line);
      border-radius:8px;
      color:#e8e2d7;
      padding:2px 7px;
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:.92em;
    }
    .shell {
      display:grid;
      grid-template-columns:260px minmax(0, 1fr);
      min-height:100vh;
    }
    .rail {
      position:sticky;
      top:0;
      height:100vh;
      display:flex;
      flex-direction:column;
      gap:24px;
      padding:22px;
      border-right:1px solid var(--line);
      background:linear-gradient(180deg, #11141c, #0c0e13);
    }
    .brand {
      display:flex;
      align-items:center;
      gap:12px;
      min-width:0;
    }
    .brand-mark {
      display:grid;
      place-items:center;
      width:38px;
      height:38px;
      border-radius:12px;
      background:linear-gradient(145deg, #f0d99b, #8ab4f8);
      color:#090a0d;
      font-weight:900;
      letter-spacing:-.06em;
    }
    .brand-title { min-width:0; }
    .brand-title strong { display:block; font-size:15px; letter-spacing:-.02em; }
    .brand-title span { display:block; color:var(--muted); font-size:12px; margin-top:2px; }
    .rail-nav { display:flex; flex-direction:column; gap:6px; }
    .nav-item {
      display:flex;
      align-items:center;
      justify-content:space-between;
      gap:10px;
      width:100%;
      padding:10px 11px;
      border:1px solid transparent;
      border-radius:12px;
      background:transparent;
      box-shadow:none;
      color:var(--muted);
      font-size:13px;
      text-align:left;
    }
    .nav-item:hover { color:var(--text); background:#171b25; border-color:var(--line); }
    .nav-item.active { color:var(--text); background:#171b25; border-color:var(--line); }
    .nav-item span:last-child { color:#68707e; font-size:11px; }
    .nav-item.active span:last-child { color:var(--accent); }
    .rail-footer {
      margin-top:auto;
      padding:12px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#10141d;
      color:var(--muted);
      font-size:12px;
      line-height:1.45;
    }
    .main {
      min-width:0;
      padding:24px;
    }
    .topbar {
      display:flex;
      justify-content:space-between;
      align-items:flex-start;
      gap:22px;
      padding:22px;
      border:1px solid var(--line);
      border-radius:24px;
      background:linear-gradient(180deg, #171b25d9, #11151de8);
      box-shadow:var(--shadow);
    }
    .headline {
      display:flex;
      flex-direction:column;
      gap:10px;
      min-width:0;
    }
    .eyebrow {
      color:var(--accent);
      font-size:11px;
      font-weight:800;
      letter-spacing:.17em;
      text-transform:uppercase;
    }
    .subtitle {
      max-width:780px;
      color:var(--muted);
      font-size:14px;
    }
    .workspace {
      max-width:100%;
      color:var(--soft);
      white-space:nowrap;
      overflow:hidden;
      text-overflow:ellipsis;
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:12px;
      padding:7px 10px;
      border:1px solid var(--line);
      border-radius:10px;
      background:#0c0f15;
      cursor:default;
    }
    .status-panel {
      display:flex;
      flex-direction:column;
      align-items:flex-end;
      gap:12px;
      min-width:220px;
    }
    .status-line {
      display:flex;
      justify-content:flex-end;
      gap:8px;
      flex-wrap:wrap;
    }
    .metrics {
      display:grid;
      grid-template-columns:repeat(4, minmax(0, 1fr));
      gap:12px;
      margin:18px 0;
    }
    .metric {
      min-width:0;
      padding:15px;
      border:1px solid var(--line);
      border-radius:16px;
      background:linear-gradient(180deg, var(--surface), #10141c);
    }
    .metric span { display:block; color:var(--muted); font-size:12px; }
    .metric strong { display:block; margin-top:7px; font-size:26px; line-height:1; letter-spacing:-.04em; }
    .metric small { display:block; margin-top:8px; min-height:16px; color:var(--soft); font-size:12px; }
    .pages {
      margin-top:18px;
    }
    .page-view {
      display:none;
    }
    .page-view.active {
      display:block;
    }
    .page-grid {
      display:grid;
      grid-template-columns:repeat(12, minmax(0, 1fr));
      gap:16px;
    }
    .card {
      grid-column:span 6;
      min-width:0;
      padding:17px;
      border:1px solid var(--line);
      border-radius:var(--radius);
      background:linear-gradient(180deg, #151922, #11151d);
      box-shadow:0 12px 32px #0005;
    }
    .span-12 { grid-column:span 12; }
    .span-8 { grid-column:span 8; }
    .span-7 { grid-column:span 7; }
    .span-5 { grid-column:span 5; }
    .span-4 { grid-column:span 4; }
    .span-3 { grid-column:span 3; }
    .section-head {
      display:flex;
      justify-content:space-between;
      gap:14px;
      align-items:flex-start;
      margin-bottom:14px;
    }
    .muted { color:var(--muted); }
    .soft { color:var(--soft); }
    .stack { display:flex; flex-direction:column; gap:12px; }
    .row { display:flex; gap:9px; align-items:center; min-width:0; }
    .row.wrap { flex-wrap:wrap; }
    .grid2 { display:grid; grid-template-columns:1fr 1fr; gap:10px; }
    .badge {
      display:inline-flex;
      align-items:center;
      gap:6px;
      width:max-content;
      max-width:100%;
      border-radius:999px;
      padding:5px 9px;
      font-size:12px;
      line-height:1;
      color:var(--soft);
      border:1px solid var(--line-2);
      background:#10141d;
      white-space:nowrap;
    }
    .badge.ok { color:#c8f5d7; border-color:#3d7351; background:#132119; }
    .badge.warn { color:#f7dfa2; border-color:#7f6a36; background:#241e12; }
    .badge.bad { color:#ffc1cb; border-color:#80424f; background:#26151a; }
    .dot {
      width:7px;
      height:7px;
      border-radius:50%;
      background:currentColor;
      box-shadow:0 0 12px currentColor;
    }
    .command {
      display:grid;
      gap:10px;
      padding:12px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#0f131b;
    }
    .cmd-head {
      display:flex;
      align-items:center;
      justify-content:space-between;
      gap:10px;
    }
    .cmd-head strong {
      color:var(--text);
      font-size:14px;
      text-transform:capitalize;
    }
    .hint {
      color:var(--muted);
      font-size:12.5px;
      line-height:1.5;
    }
    .task-list, .agent-list, .version-list, .log-list { display:flex; flex-direction:column; gap:9px; }
    .doc-list { display:flex; flex-direction:column; gap:10px; }
    .task-card, .agent-card, .lock-card, .version-card {
      display:grid;
      gap:12px;
      padding:12px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#0f131b;
    }
    .task-card {
      grid-template-columns:70px minmax(0, 1fr) auto;
      align-items:center;
    }
    .task-title { color:var(--text); line-height:1.35; overflow-wrap:anywhere; }
    .task-meta { display:flex; gap:7px; flex-wrap:wrap; margin-top:7px; }
    .doc-card {
      padding:14px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#0f131b;
    }
    .doc-card h3 {
      font-size:15px;
      margin-bottom:7px;
    }
    .doc-card p, .doc-card li {
      color:var(--muted);
      font-size:14px;
      line-height:1.55;
    }
    .doc-card ul, .doc-card ol {
      margin:8px 0 0;
      padding-left:20px;
    }
    details.doc-card summary {
      cursor:pointer;
      color:var(--text);
      font-weight:700;
    }
    details.doc-card[open] summary { margin-bottom:9px; }
    .agent-card {
      grid-template-columns:minmax(0, 1fr) auto;
      align-items:start;
    }
    .agent-name {
      display:flex;
      gap:8px;
      align-items:center;
      font-weight:750;
    }
    .agent-sub { margin-top:7px; color:var(--muted); font-size:13px; }
    .terminal-preview {
      min-height:170px;
      padding:14px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#07090d;
      color:#d7deea;
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:12px;
      line-height:1.55;
      white-space:pre-wrap;
    }
    .empty {
      padding:16px;
      border:1px dashed var(--line-2);
      border-radius:14px;
      color:var(--muted);
      background:#10141d80;
    }
    .log {
      max-height:330px;
      overflow:auto;
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:12px;
      line-height:1.55;
      white-space:pre-wrap;
    }
    .divider { height:1px; background:var(--line); margin:3px 0; }
    .toast {
      position:fixed;
      right:24px;
      bottom:24px;
      z-index:10;
      max-width:360px;
      padding:12px 14px;
      border:1px solid var(--line-2);
      border-radius:14px;
      background:#151922ee;
      color:var(--text);
      box-shadow:0 18px 50px #0008;
      opacity:0;
      transform:translateY(10px);
      pointer-events:none;
      transition:opacity .18s ease, transform .18s ease;
    }
    .toast.show { opacity:1; transform:translateY(0); }
    @media (max-width: 1120px) {
      .shell { grid-template-columns:1fr; }
      .rail {
        position:static;
        height:auto;
        flex-direction:row;
        align-items:center;
        overflow:auto;
        border-right:0;
        border-bottom:1px solid var(--line);
      }
      .rail-nav { flex-direction:row; }
      .nav-item { width:auto; flex:0 0 auto; }
      .rail-footer { display:none; }
      .span-8, .span-7, .span-5, .span-4, .span-3, .card { grid-column:span 12; }
      .metrics { grid-template-columns:repeat(2, minmax(0, 1fr)); }
      .topbar { flex-direction:column; }
      .status-panel { align-items:flex-start; min-width:0; }
      .status-line { justify-content:flex-start; }
    }
    @media (max-width: 680px) {
      .main { padding:14px; }
      .rail { padding:14px; gap:12px; }
      .rail {
        align-items:flex-start;
        flex-direction:column;
      }
      .rail-nav {
        width:100%;
        overflow:auto;
        padding-bottom:2px;
      }
      .topbar { padding:17px; border-radius:18px; }
      .metrics { grid-template-columns:repeat(2, minmax(0, 1fr)); gap:10px; }
      .metric { padding:12px; }
      .metric strong { font-size:23px; }
      .grid2 { grid-template-columns:1fr; }
      .task-card, .command, .agent-card { grid-template-columns:1fr; }
      .row { flex-direction:column; align-items:stretch; }
    }
    @media (max-width: 460px) {
      .metrics { grid-template-columns:1fr; }
    }
    button.primary {
      background:linear-gradient(180deg, #eccf8d, #cfa860);
      border-color:#eccf8d;
      color:#241a08;
      font-weight:700;
      box-shadow:0 10px 26px #caa85f30;
    }
    button.primary:hover { border-color:#f6dea0; }
    .next-step {
      display:flex;
      flex-wrap:wrap;
      justify-content:space-between;
      align-items:center;
      gap:18px;
      padding:20px 22px;
      margin-bottom:18px;
      border:1px solid var(--line-2);
      border-radius:20px;
      background:linear-gradient(120deg, #1d2436, #14192400);
      box-shadow:var(--shadow);
    }
    .next-step.done { background:linear-gradient(120deg, #15271d, #14192400); border-color:#2f5540; }
    .next-step .ns-body { min-width:min(100%, 340px); flex:1 1 340px; display:flex; flex-direction:column; gap:7px; }
    .next-step .ns-step { color:var(--accent); font-size:11px; font-weight:800; letter-spacing:.16em; text-transform:uppercase; }
    .next-step.done .ns-step { color:var(--ok); }
    .next-step p { color:var(--soft); font-size:14px; max-width:66ch; }
    .next-step .ns-action { flex:0 0 auto; }
    .checklist { display:flex; flex-direction:column; margin:2px 0; }
    .check { display:flex; align-items:center; gap:11px; padding:9px 0; color:var(--muted); font-size:14px; border-bottom:1px solid var(--line); }
    .check:last-child { border-bottom:0; }
    .check-mark {
      display:grid;
      place-items:center;
      flex:0 0 auto;
      width:22px;
      height:22px;
      border-radius:50%;
      border:1px solid var(--line-2);
      font-size:12px;
      color:var(--muted);
    }
    .check.is-done { color:var(--text); }
    .check.is-done .check-mark { background:#14281c; border-color:#3d7351; color:#8bd9a8; }
    .lead { color:var(--soft); font-size:15px; line-height:1.62; }
    .concept-row { display:grid; grid-template-columns:repeat(4, minmax(0, 1fr)); gap:12px; margin-top:4px; }
    .concept { padding:13px; border:1px solid var(--line); border-radius:13px; background:#0f131b; }
    .concept strong { display:block; color:var(--text); font-size:14px; margin-bottom:5px; }
    .concept span { color:var(--muted); font-size:12.5px; line-height:1.5; }
    .steps { display:flex; flex-direction:column; gap:13px; }
    .step-item { display:flex; gap:13px; align-items:flex-start; }
    .step-num {
      flex:0 0 auto;
      display:grid;
      place-items:center;
      width:26px;
      height:26px;
      border-radius:50%;
      background:linear-gradient(145deg, #f0d99b, #8ab4f8);
      color:#0a0b0f;
      font-weight:800;
      font-size:13px;
    }
    .step-item .step-body { min-width:0; }
    .step-item .step-body strong { display:block; color:var(--text); font-size:14px; margin-bottom:3px; }
    .step-item .step-body span { color:var(--muted); font-size:13.5px; line-height:1.5; }
    @media (max-width: 900px) { .concept-row { grid-template-columns:repeat(2, minmax(0, 1fr)); } }
    @media (max-width: 560px) { .concept-row { grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <div class="shell">
    <aside class="rail">
      <div class="brand">
        <div class="brand-mark">Co</div>
        <div class="brand-title">
          <strong>CoAct</strong>
          <span>Local control plane</span>
        </div>
      </div>
      <nav class="rail-nav" aria-label="Dashboard pages">
        <button class="nav-item active" type="button" data-page-target="overview"><span>Overview</span><span>01</span></button>
        <button class="nav-item" type="button" data-page-target="guide"><span>Guide</span><span>02</span></button>
        <button class="nav-item" type="button" data-page-target="brief"><span>Brief</span><span>03</span></button>
        <button class="nav-item" type="button" data-page-target="tasks"><span>Tasks</span><span>04</span></button>
        <button class="nav-item" type="button" data-page-target="agents"><span>Agents</span><span>05</span></button>
        <button class="nav-item" type="button" data-page-target="messages"><span>Messages</span><span>06</span></button>
        <button class="nav-item" type="button" data-page-target="versions"><span>Versions</span><span>07</span></button>
        <button class="nav-item" type="button" data-page-target="log"><span>Audit log</span><span>08</span></button>
      </nav>
      <div class="rail-footer">
        Local-only UI. No arbitrary shell execution. Mutations use CoAct's journaled APIs.
      </div>
    </aside>

    <div class="main">
      <header class="topbar">
        <div class="headline">
          <div class="eyebrow">Multi-agent workspace</div>
          <h1>CoAct Control Center</h1>
          <p class="subtitle">Brief your agents, hand out tasks, and see who's editing what — so Claude, Codex, and Gemini share one repo without copy-paste or collisions.</p>
          <div class="workspace" id="workspace" title="">Loading workspace…</div>
        </div>
        <div class="status-panel">
          <div class="status-line">
            <span class="badge" id="version">coact —</span>
            <span class="badge" id="initBadge">checking</span>
          </div>
          <div class="muted" id="updated">Syncing…</div>
        </div>
      </header>

      <main class="pages">
        <section class="page-view active" data-page="overview">
          <div class="next-step" id="nextStep"></div>
          <section class="metrics" aria-label="Workspace summary">
            <div class="metric"><span>Tasks</span><strong id="metricTasks">—</strong><small id="metricTasksSub">waiting</small></div>
            <div class="metric"><span>Live agents</span><strong id="metricAgents">—</strong><small id="metricAgentsSub">heartbeat</small></div>
            <div class="metric"><span>Locks</span><strong id="metricLocks">—</strong><small>conflict gate</small></div>
            <div class="metric"><span>Mode</span><strong id="metricMode">—</strong><small id="metricModeSub">workspace</small></div>
          </section>
          <div class="page-grid">
            <section class="card span-5" id="setup">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Setup</div>
                  <h2>Getting started</h2>
                </div>
                <span class="badge" id="initState">Checking</span>
              </div>
              <div class="stack">
                <p class="muted">Four steps to a coordinated workspace. Your next action is highlighted above.</p>
                <div class="checklist" id="checklist"></div>
                <div class="soft">Full health check from a terminal: <code>coact doctor</code></div>
              </div>
            </section>

            <section class="card span-7" id="launch-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Launch</div>
                  <h2>Start an agent</h2>
                </div>
                <span class="badge">built-in adapters</span>
              </div>
              <div id="launch" class="stack muted">Loading commands…</div>
              <p class="hint" style="margin-top:12px">Start opens the agent in a new terminal, already wired into CoAct. Only these built-in adapters can be launched — the UI never runs arbitrary shell commands.</p>
            </section>
          </div>
        </section>

        <section class="page-view" data-page="guide">
          <div class="page-grid">
            <section class="card span-12">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Start here</div>
                  <h2>What is CoAct?</h2>
                </div>
                <span class="badge">2-minute read</span>
              </div>
              <div class="stack">
                <p class="lead">CoAct lets several AI coding agents — Claude Code, Codex, Gemini — work in the <strong>same repository at the same time</strong> without stepping on each other. You stay in charge: you write one shared brief and a task list, each agent claims its own tasks, and CoAct's file locks stop two agents from editing the same file. Everything they do is written to a journal you can replay.</p>
                <div class="concept-row">
                  <div class="concept"><strong>Shared brief</strong><span>One source of context every agent reads.</span></div>
                  <div class="concept"><strong>Task board</strong><span>Agents claim tasks, so work never overlaps.</span></div>
                  <div class="concept"><strong>File locks</strong><span>Two agents can't edit the same file at once.</span></div>
                  <div class="concept"><strong>Audit log</strong><span>Every action is recorded for you to review.</span></div>
                </div>
              </div>
            </section>

            <section class="card span-7">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Workflow</div>
                  <h2>Getting a run going</h2>
                </div>
                <span class="badge" id="guideVersionBadge">current</span>
              </div>
              <div class="doc-list">
                <div class="doc-card">
                  <h3>The 3-step workflow</h3>
                  <div class="steps" style="margin-top:12px">
                    <div class="step-item"><span class="step-num">1</span><div class="step-body"><strong>Initialize the repo</strong><span>Once per project, from the Overview or with <code>coact init</code>. It wires the enforcement hook and drops the agent contracts.</span></div></div>
                    <div class="step-item"><span class="step-num">2</span><div class="step-body"><strong>Write the brief, add tasks</strong><span>Set the shared context on the Brief page, then add concrete tasks agents can claim on the Tasks page.</span></div></div>
                    <div class="step-item"><span class="step-num">3</span><div class="step-body"><strong>Start the agents</strong><span>Launch Claude and Codex from the Overview. As they work, coordinate through Messages, Tasks, and Locks.</span></div></div>
                  </div>
                </div>
                <div class="doc-card" id="currentVersionGuide">Loading current version…</div>
              </div>
            </section>

            <section class="card span-5">
              <div class="section-head">
                <div>
                  <div class="eyebrow">History</div>
                  <h2>Versions & switching</h2>
                </div>
                <span class="badge warn">managed only</span>
              </div>
              <div id="guideVersions" class="doc-list muted">Loading versions…</div>
            </section>

            <section class="card span-6">
              <div class="section-head">
                <div>
                  <div class="eyebrow">FAQ</div>
                  <h2>Common questions</h2>
                </div>
              </div>
              <div class="doc-list">
                <details class="doc-card" open>
                  <summary>Do I still need separate terminals?</summary>
                  <p>For this release, yes. The UI can open native local terminals for installed agents. Fully embedded browser terminals require the next PTY/WebSocket layer.</p>
                </details>
                <details class="doc-card">
                  <summary>Does CoAct execute arbitrary commands?</summary>
                  <p>No. UI launch actions are allowlisted through built-in adapters such as Claude, Codex, and Gemini. Generic shell execution is intentionally not exposed.</p>
                </details>
                <details class="doc-card">
                  <summary>What if an agent is missing?</summary>
                  <p>The Launch card marks it as missing and disables Start. Install that agent CLI and make sure it is on PATH, then refresh the UI.</p>
                </details>
                <details class="doc-card">
                  <summary>Can I switch versions here?</summary>
                  <p>Yes, for versions installed in CoAct's managed layout under <code>~/.coact</code>. The UI changes the managed shim only; it does not overwrite system binaries.</p>
                </details>
              </div>
            </section>

            <section class="card span-6">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Technical design</div>
                  <h2>Safety model</h2>
                </div>
              </div>
              <div class="doc-list">
                <details class="doc-card" open>
                  <summary>Local control center</summary>
                  <p>The UI binds to loopback only, checks the Host header, and requires a per-run token for mutations.</p>
                </details>
                <details class="doc-card">
                  <summary>Coordination model</summary>
                  <p>Tasks, messages, locks, presence, and logs are plain local project state under <code>.coact</code>. Agents read and write through CoAct commands and hooks.</p>
                </details>
                <details class="doc-card">
                  <summary>Version manager</summary>
                  <p>Managed versions live under <code>~/.coact/bin/coact-&lt;version&gt;</code>. Switching only repoints the managed <code>~/.coact/coact</code> shim.</p>
                </details>
              </div>
            </section>
          </div>
        </section>

        <section class="page-view" data-page="brief">
          <section class="card span-12" id="brief-card">
            <div class="section-head">
              <div>
                <div class="eyebrow">Context</div>
                <h2>Project brief</h2>
              </div>
              <span class="badge">human-controlled</span>
            </div>
            <div class="stack">
              <textarea id="brief" placeholder="Goal, constraints, decisions, preferred agent split. Saved to .coact/brief.md."></textarea>
              <div class="row wrap"><button onclick="saveBrief()">Save brief</button><span class="muted">Agents can read this shared context; they should not edit it directly.</span></div>
            </div>
          </section>
        </section>

        <section class="page-view" data-page="tasks">
          <section class="card span-12" id="tasks-card">
            <div class="section-head">
              <div>
                <div class="eyebrow">Board</div>
                <h2>Tasks</h2>
              </div>
              <span class="badge" id="taskCount">0 tasks</span>
            </div>
            <div class="stack">
              <div class="row"><input id="taskTitle" placeholder="Add a concrete task for an agent" /><button onclick="addTask()">Add task</button></div>
              <div id="tasks" class="task-list"></div>
            </div>
          </section>
        </section>

        <section class="page-view" data-page="agents">
          <div class="page-grid">
            <section class="card span-8" id="agents-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Presence</div>
                  <h2>Agents</h2>
                </div>
                <span class="badge" id="agentCount">0 live</span>
              </div>
              <div id="agents" class="agent-list"></div>
            </section>

            <section class="card span-4" id="locks-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Safety</div>
                  <h2>Locks</h2>
                </div>
                <span class="badge" id="lockCount">0 active</span>
              </div>
              <div id="locks" class="stack"></div>
            </section>

            <section class="card span-12">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Terminal</div>
                  <h2>Embedded agent terminals</h2>
                </div>
                <span class="badge warn">next layer</span>
              </div>
              <div class="terminal-preview">Current release opens each agent in a native local terminal with CoAct already wired.

Next layer:
  • run Claude and Codex inside managed PTY sessions
  • stream output into this page
  • send keystrokes from the browser to each agent
  • keep all commands allowlisted through CoAct adapters</div>
            </section>
          </div>
        </section>

        <section class="page-view" data-page="messages">
          <section class="card span-12" id="messages-card">
            <div class="section-head">
              <div>
                <div class="eyebrow">Inbox</div>
                <h2>Messages</h2>
              </div>
              <span class="badge">journaled</span>
            </div>
            <div class="stack">
              <div class="grid2"><input id="msgTo" placeholder="to: claude / codex / gemini" /><input id="msgFrom" value="human" /></div>
              <textarea id="msgText" placeholder="Send instructions or handoff context to an agent"></textarea>
              <button onclick="sendMessage()">Send message</button>
            </div>
          </section>
        </section>

        <section class="page-view" data-page="versions">
          <section class="card span-12" id="versions-card">
            <div class="section-head">
              <div>
                <div class="eyebrow">Release</div>
                <h2>Versions</h2>
              </div>
              <span class="badge warn">experimental</span>
            </div>
            <div id="versions" class="version-list muted">Managed versions appear here after <code>coact update</code>.</div>
          </section>
        </section>

        <section class="page-view" data-page="log">
          <section class="card span-12" id="log-card">
            <div class="section-head">
              <div>
                <div class="eyebrow">Audit</div>
                <h2>Activity log</h2>
              </div>
              <span class="badge" id="logCount">0 events</span>
            </div>
            <div id="log" class="log muted">No events yet.</div>
          </section>
        </section>
      </main>
    </div>
  </div>
  <div class="toast" id="toast"></div>

  <script>
    const TOKEN = "__COACT_TOKEN__";
    let lastBrief = null;
    let toastTimer = null;
    const DEFAULT_PAGE = "overview";

    async function api(path, opts) {
      opts = opts || {};
      const headers = Object.assign({'Content-Type':'application/json','X-Coact-Token':TOKEN}, opts.headers || {});
      const res = await fetch(path, Object.assign({}, opts, {headers}));
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || res.statusText);
      return data;
    }
    function esc(s) { return String(s || '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
    function badge(text, type) { return '<span class="badge '+esc(type || '')+'">'+esc(text)+'</span>'; }
    function pageFromLocation() {
      const page = (window.location.hash || '').replace(/^#/, '');
      return document.querySelector('[data-page="'+page+'"]') ? page : DEFAULT_PAGE;
    }
    function setPage(page, updateURL) {
      if (!document.querySelector('[data-page="'+page+'"]')) page = DEFAULT_PAGE;
      document.querySelectorAll('[data-page]').forEach(el => {
        el.classList.toggle('active', el.getAttribute('data-page') === page);
      });
      document.querySelectorAll('[data-page-target]').forEach(el => {
        const active = el.getAttribute('data-page-target') === page;
        el.classList.toggle('active', active);
        if (active) el.setAttribute('aria-current', 'page');
        else el.removeAttribute('aria-current');
      });
      if (updateURL) history.pushState({page}, '', '#'+page);
      window.scrollTo({top: 0, behavior: 'auto'});
    }
    function showToast(message) {
      const el = document.getElementById('toast');
      el.textContent = message;
      el.classList.add('show');
      clearTimeout(toastTimer);
      toastTimer = setTimeout(() => el.classList.remove('show'), 2400);
    }
    async function mutate(label, fn) {
      try {
        await fn();
        showToast(label);
        refresh();
      } catch (e) {
        showToast(e.message);
      }
    }
    function taskStats(tasks) {
      return {
        total: tasks.length,
        doing: tasks.filter(t => t.state === 'doing').length,
        done: tasks.filter(t => t.state === 'done').length,
        todo: tasks.filter(t => t.state === 'todo').length
      };
    }
    function renderNextStep(s, tasks, agents) {
      const el = document.getElementById('nextStep');
      if (!el) return;
      const live = agents.filter(a => a.live).length;
      const hasBrief = (s.brief || '').trim().length > 0;
      let step, title, desc, action, done = false;
      if (!s.initialized) {
        step = 'Step 1 of 4';
        title = 'Initialize this repository';
        desc = 'Wire CoAct into this repo — it adds the enforcement hook, agent contracts, and the .coact workspace. Nothing else is touched, and coact deinit removes it all.';
        action = '<button class="primary" onclick="initProject()">Initialize repo</button>';
      } else if (!hasBrief) {
        step = 'Step 2 of 4';
        title = 'Write your project brief';
        desc = 'Give every agent the same context — the goal, the constraints, and who does what. It saves to .coact/brief.md for agents to read.';
        action = '<button class="primary" data-goto="brief">Write brief</button>';
      } else if (tasks.length === 0) {
        step = 'Step 3 of 4';
        title = 'Add your first task';
        desc = 'Break the work into concrete tasks. Agents claim tasks from the board, so two of them never grab the same thing.';
        action = '<button class="primary" data-goto="tasks">Add a task</button>';
      } else if (live === 0) {
        step = 'Step 4 of 4';
        title = 'Start an agent';
        desc = "Launch Claude or Codex below. Each opens in its own terminal, already wired into CoAct — locks and the board keep them out of each other's way.";
        action = '<button class="primary" data-goto="agents">View agents</button>';
      } else {
        done = true;
        step = "You're set";
        title = live + ' agent' + (live > 1 ? 's' : '') + ' working now';
        desc = 'Agents are live and coordinating through locks and the shared board. Track progress in Tasks and the Audit log; send guidance from Messages.';
        action = '<button data-goto="tasks">Open tasks</button>';
      }
      el.className = 'next-step' + (done ? ' done' : '');
      el.innerHTML = '<div class="ns-body"><div class="ns-step">' + esc(step) + '</div><h2>' + esc(title) + '</h2><p>' + esc(desc) + '</p></div><div class="ns-action">' + action + '</div>';
    }
    function renderChecklist(s, tasks, agents) {
      const el = document.getElementById('checklist');
      if (!el) return;
      const live = agents.filter(a => a.live).length;
      const hasBrief = (s.brief || '').trim().length > 0;
      const steps = [
        {done: s.initialized, label: 'Initialize the repository'},
        {done: s.initialized && hasBrief, label: 'Write the project brief'},
        {done: tasks.length > 0, label: 'Add tasks to the board'},
        {done: live > 0, label: 'Start an agent'}
      ];
      el.innerHTML = steps.map((st, i) => '<div class="check ' + (st.done ? 'is-done' : '') + '"><span class="check-mark">' + (st.done ? '✓' : (i + 1)) + '</span><span>' + esc(st.label) + '</span></div>').join('');
    }
    async function refresh() {
      try {
        const s = await api('/api/state');
        const tasks = s.tasks || [];
        const agents = s.agents || [];
        const locks = s.locks || [];
        const log = s.log || [];
        const stats = taskStats(tasks);
        const liveAgents = agents.filter(a => a.live).length;
        const wsEl = document.getElementById('workspace');
        const ws = s.workspace || '';
        wsEl.textContent = (ws.split('/').filter(Boolean).slice(-2).join('/') || ws) + (s.initialized ? '' : ' · not initialized');
        wsEl.title = ws;
        document.getElementById('version').textContent = 'coact ' + (s.version || 'dev');
        document.getElementById('updated').textContent = 'Last sync ' + new Date().toLocaleTimeString();
        document.getElementById('initBadge').className = 'badge ' + (s.initialized ? 'ok' : 'bad');
        document.getElementById('initBadge').textContent = s.initialized ? 'initialized' : 'not initialized';
        document.getElementById('initState').className = 'badge ' + (s.initialized ? 'ok' : 'bad');
        document.getElementById('initState').textContent = s.initialized ? 'ready' : 'needs init';
        document.getElementById('metricTasks').textContent = stats.total;
        document.getElementById('metricTasksSub').textContent = stats.doing + ' doing · ' + stats.done + ' done';
        document.getElementById('metricAgents').textContent = liveAgents + '/' + agents.length;
        document.getElementById('metricAgentsSub').textContent = liveAgents ? 'active now' : 'no active heartbeat';
        document.getElementById('metricLocks').textContent = locks.length;
        document.getElementById('metricMode').textContent = s.mode || '-';
        document.getElementById('metricModeSub').textContent = s.initialized ? 'initialized repo' : 'unwired repo';
        if (lastBrief === null || document.activeElement.id !== 'brief') {
          document.getElementById('brief').value = s.brief || '';
          lastBrief = s.brief || '';
        }
        renderTasks(tasks);
        renderAgents(agents);
        renderLocks(locks);
        renderLog(log);
        renderVersions(s.versions || [], s.manifest || null);
        renderGuide(s.versions || [], s.manifest || null);
        renderNextStep(s, tasks, agents);
        renderChecklist(s, tasks, agents);
      } catch (e) {
        document.getElementById('initState').className = 'badge bad';
        document.getElementById('initState').textContent = e.message;
      }
    }
    function renderTasks(tasks) {
      document.getElementById('taskCount').textContent = tasks.length + (tasks.length === 1 ? ' task' : ' tasks');
      document.getElementById('tasks').innerHTML = tasks.map(t => {
        const stateType = t.state === 'done' ? 'ok' : (t.state === 'doing' ? 'warn' : '');
        return '<div class="task-card"><div><code>'+esc(t.id)+'</code></div><div><div class="task-title">'+esc(t.title)+'</div><div class="task-meta">'+badge(t.state, stateType)+badge(t.owner || 'unassigned')+'</div></div><div class="row wrap"><button class="small ghost" onclick="claimTask(\''+esc(t.id)+'\')">Claim</button><button class="small" onclick="doneTask(\''+esc(t.id)+'\')">Done</button></div></div>';
      }).join('') || '<div class="empty">No tasks yet. Add one above, then assign it to an agent.</div>';
    }
    function renderAgents(agents) {
      const live = agents.filter(a => a.live).length;
      document.getElementById('agentCount').textContent = live + ' live';
      document.getElementById('agents').innerHTML = agents.map(a => '<div class="agent-card"><div><div class="agent-name"><span class="badge '+(a.live?'ok':'bad')+'"><span class="dot"></span>'+(a.live?'live':'dead')+'</span><code>'+esc(a.id)+'</code></div><div class="agent-sub">'+esc(a.status || 'idle')+' · '+esc(a.enforcement || 'policy unknown')+'</div></div><div class="muted">'+esc(a.current_task || '-')+'<br>'+esc(a.beat || '-')+'</div></div>').join('') || '<div class="empty">No agents configured yet.</div>';
    }
    function renderLocks(locks) {
      document.getElementById('lockCount').textContent = locks.length + ' active';
      document.getElementById('locks').innerHTML = locks.map(l => '<div class="lock-card"><div class="soft">'+esc(l.path)+'</div><div class="row wrap">'+badge(l.owner || 'unknown')+badge((l.ttl_seconds || 0) + 's TTL')+'</div></div>').join('') || '<div class="empty">No active locks.</div>';
    }
    function renderLog(log) {
      document.getElementById('logCount').textContent = log.length + ' events';
      document.getElementById('log').textContent = log.map(r => [r.ts, r.agent, r.event, Object.keys(r).filter(k => !['ts','agent','event'].includes(k)).sort().map(k => k+'='+r[k]).join(' ')].join('  ')).join('\n') || 'No events yet.';
    }
    function renderVersions(versions, manifest) {
      const local = versions.length ? versions.map(v => '<div class="version-card"><div class="row wrap"><code>'+esc(v.version)+'</code> '+(v.active?badge('active','ok'):'')+' '+(!v.active?'<button class="small" data-switch-version="'+esc(v.version)+'">Switch</button>':'')+'</div><div class="muted">'+esc(v.path)+'</div></div>').join('') : '<div class="empty">No managed versions yet. Run <code>coact update</code> from a terminal.</div>';
      const supports = manifest && manifest.supports ? manifest.supports : {};
      const meta = manifest ? '<div class="version-card"><div><strong>'+esc(manifest.version)+'</strong> '+badge(manifest.channel || 'channel')+' '+badge(manifest.stability || 'stability')+(manifest.recommended?' '+badge('recommended','ok'):'')+'</div><div class="muted">'+esc(manifest.summary || '')+'</div><div class="muted">Agents: '+esc((supports.agents || []).join(', ') || '-')+' · realtime: '+esc(supports.realtime || '-')+' · autopilot: '+esc(supports.autopilot || '-')+'</div></div>' : '';
      document.getElementById('versions').innerHTML = meta + local + '<div class="divider"></div><div class="muted">Commands: <code>coact update --channel stable</code> · <code>coact versions</code> · <code>coact switch &lt;version&gt;</code></div>';
    }
    function renderGuide(versions, manifest) {
      const guideBadge = document.getElementById('guideVersionBadge');
      const current = manifest || {};
      guideBadge.textContent = current.version ? 'current · '+current.version : 'current';
      const supports = current.supports || {};
      const notes = (current.notes || []).map(n => '<li>'+esc(n)+'</li>').join('');
      document.getElementById('currentVersionGuide').innerHTML = '<h3>'+esc(current.version || 'Current build')+'</h3><p>'+esc(current.summary || 'Local multi-agent control center for CoAct workspaces.')+'</p><div class="row wrap" style="margin-top:10px">'+badge(current.channel || 'dev')+badge(current.stability || 'experimental')+(current.recommended?badge('recommended','ok'):'')+'</div><ul>'+notes+'</ul><p style="margin-top:10px">Supports: '+esc((supports.agents || []).join(', ') || 'configured agents')+' · UI: '+esc(String(!!supports.ui))+' · realtime: '+esc(supports.realtime || 'not specified')+' · autopilot: '+esc(supports.autopilot || 'not specified')+'</p>';
      const installed = versions.length ? versions.map(v => '<details class="doc-card" '+(v.active?'open':'')+'><summary><code>'+esc(v.version)+'</code> '+(v.active?badge('active','ok'):'')+'</summary><p>'+esc(v.path)+'</p><div class="row wrap" style="margin-top:10px">'+(!v.active?'<button class="small" data-switch-version="'+esc(v.version)+'">Switch to this version</button>':'<span class="badge ok">currently active</span>')+'</div></details>').join('') : '<div class="empty">No managed versions installed yet. Use <code>coact update</code> to install versions into <code>~/.coact</code>.</div>';
      document.getElementById('guideVersions').innerHTML = '<details class="doc-card" open><summary>Current version details</summary><p>'+esc(current.summary || 'Current running build.')+'</p></details>'+installed;
    }
    async function initProject(){ mutate('Repository initialized', () => api('/api/init',{method:'POST', body:'{}'})); }
    async function saveBrief(){ mutate('Brief saved', () => api('/api/brief',{method:'POST', body:JSON.stringify({text:document.getElementById('brief').value})})); }
    async function addTask(){ const title=document.getElementById('taskTitle').value; if(!title.trim()) return; await mutate('Task added', () => api('/api/tasks',{method:'POST', body:JSON.stringify({title})})); document.getElementById('taskTitle').value=''; }
    async function claimTask(id){ const owner=prompt('Owner agent?', 'claude'); if(!owner) return; mutate('Task claimed', () => api('/api/tasks/'+id+'/claim',{method:'POST', body:JSON.stringify({owner})})); }
    async function doneTask(id){ const owner=prompt('Owner agent?', 'human'); if(!owner) return; mutate('Task marked done', () => api('/api/tasks/'+id+'/done',{method:'POST', body:JSON.stringify({owner})})); }
    async function sendMessage(){ const text=document.getElementById('msgText').value; if(!text.trim()) return; await mutate('Message sent', () => api('/api/messages',{method:'POST', body:JSON.stringify({from:document.getElementById('msgFrom').value,to:document.getElementById('msgTo').value,text})})); document.getElementById('msgText').value=''; }
    async function loadLaunch(){
      const d = await api('/api/launch-commands');
      document.getElementById('launch').innerHTML = d.commands.map(c => {
        const canLaunch = c.installed && c.terminal_supported;
        const status = c.installed ? badge('installed','ok') : badge('not on PATH','bad');
        const friendly = 'coact ' + c.agent;
        const title = canLaunch
          ? 'Opens ' + c.agent + ' in a new terminal, already wired into CoAct'
          : (c.installed ? 'One-click start is macOS-only right now — copy the command and run it in a terminal' : 'Install ' + c.agent + ' and add it to PATH, then reload');
        const startBtn = canLaunch
          ? '<button class="small primary" data-launch="'+esc(c.agent)+'" title="'+esc(title)+'">Open in Terminal</button>'
          : '<button class="small" title="'+esc(title)+'" disabled>Open in Terminal</button>';
        return '<div class="command"><div class="cmd-head"><strong>'+esc(c.agent)+'</strong>'+status+'</div><code title="'+esc(c.command)+'">'+esc(friendly)+'</code><div class="row wrap">'+startBtn+'<button class="small ghost" data-copy="'+esc(c.command)+'">Copy command</button></div></div>';
      }).join('');
    }
    document.addEventListener('click', async ev => {
      const nav = ev.target.closest('[data-page-target]');
      if (nav) {
        setPage(nav.getAttribute('data-page-target'), true);
        return;
      }
      const goto = ev.target.closest('[data-goto]');
      if (goto) {
        setPage(goto.getAttribute('data-goto'), true);
        return;
      }
      const launch = ev.target.closest('[data-launch]');
      if (launch) {
        const agent = launch.getAttribute('data-launch');
        await mutate('Opening '+agent+' in Terminal', () => api('/api/agents/'+agent+'/launch', {method:'POST', body:'{}'}));
        return;
      }
      const switchVersion = ev.target.closest('[data-switch-version]');
      if (switchVersion) {
        const version = switchVersion.getAttribute('data-switch-version');
        if (!confirm('Switch the managed coact to ' + version + '?\n\nThis only repoints the ~/.coact/coact shim. System installs are left untouched.')) return;
        await mutate('Switched managed coact to '+version, () => api('/api/versions/'+encodeURIComponent(version)+'/switch', {method:'POST', body:'{}'}));
        return;
      }
      const button = ev.target.closest('[data-copy]');
      if (!button) return;
      try {
        await navigator.clipboard.writeText(button.getAttribute('data-copy'));
        showToast('Command copied');
      } catch (e) {
        showToast('Copy failed');
      }
    });
    window.addEventListener('popstate', () => setPage(pageFromLocation(), false));
    setPage(pageFromLocation(), false);
    loadLaunch().catch(e => showToast(e.message));
    refresh();
    setInterval(refresh, 1500);
  </script>
</body>
</html>`
