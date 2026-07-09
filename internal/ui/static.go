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
      --bg:#0d0e11;
      --rail:#101114;
      --surface:#15171b;
      --surface-2:#1a1d22;
      --surface-3:#20242b;
      --text:#f2f2f2;
      --muted:#989da6;
      --soft:#c3c7ce;
      --line:#2a2e35;
      --line-2:#383d46;
      --accent:#8fb4ff;
      --accent-2:#8fb4ff;
      --ok:#85d69a;
      --warn:#e2c16f;
      --bad:#e8798a;
      --shadow:none;
      --radius:14px;
    }
    * { box-sizing:border-box; }
    body {
      margin:0;
      min-height:100vh;
      font-family:Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background:var(--bg);
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
      background:var(--surface-2);
      border-color:var(--line-2);
      box-shadow:none;
      transition:border-color .15s ease, background .15s ease;
      white-space:nowrap;
    }
    button:hover { background:var(--surface-3); border-color:#4a515d; }
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
    input:focus, textarea:focus, select:focus { border-color:var(--accent); box-shadow:0 0 0 3px #8fb4ff18; }
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
      grid-template-columns:220px minmax(0, 1fr);
      min-height:100vh;
    }
    .rail {
      position:sticky;
      top:0;
      height:100vh;
      display:flex;
      flex-direction:column;
      gap:16px;
      padding:18px;
      border-right:1px solid var(--line);
      background:var(--rail);
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
      border:1px solid var(--line);
      border-radius:10px;
      background:var(--surface-2);
      color:var(--text);
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
    .nav-item:hover { color:var(--text); background:var(--surface); border-color:var(--line); }
    .nav-item.active { color:var(--text); background:var(--surface-2); border-color:var(--line-2); }
    .nav-item span:last-child { color:#68707e; font-size:11px; }
    .nav-item.active span:last-child { color:var(--accent); }
    .rail-footer {
      display:none;
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
      padding:20px;
    }
    .topbar {
      display:flex;
      justify-content:space-between;
      align-items:flex-start;
      gap:22px;
      padding:18px;
      border:1px solid var(--line);
      border-radius:20px;
      background:var(--surface);
      box-shadow:none;
    }
    .headline {
      display:flex;
      flex-direction:column;
      gap:7px;
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
      background:#111318;
      cursor:default;
    }
    .project-bar {
      display:flex;
      gap:8px;
      align-items:center;
      flex-wrap:wrap;
      margin-top:2px;
    }
    .project-bar select {
      width:min(420px, 100%);
      min-width:240px;
      background:#101217;
    }
    .project-hint {
      color:var(--muted);
      font-size:12px;
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
      margin:0;
    }
    .metric {
      min-width:0;
      padding:15px;
      border:1px solid var(--line);
      border-radius:16px;
      background:var(--surface);
    }
    .metric span { display:block; color:var(--muted); font-size:12px; }
    .metric strong { display:block; margin-top:7px; font-size:26px; line-height:1; letter-spacing:-.04em; }
    .metric small { display:block; margin-top:8px; min-height:16px; color:var(--soft); font-size:12px; }
    .status-drawer {
      margin-top:14px;
      border:1px solid var(--line);
      border-radius:16px;
      background:var(--surface);
    }
    .status-drawer summary {
      cursor:pointer;
      padding:12px 14px;
      color:var(--soft);
      font-weight:700;
    }
    .status-drawer .metrics { padding:0 14px 14px; }
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
      background:var(--surface);
      box-shadow:none;
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
    .badge.ok { color:#b7efc7; border-color:#2e6040; background:#142018; }
    .badge.warn { color:#ead18b; border-color:#64552f; background:#211d13; }
    .badge.bad { color:#f0b1bd; border-color:#673845; background:#23171b; }
    .dot {
      width:7px;
      height:7px;
      border-radius:50%;
      background:currentColor;
    }
    .command {
      display:grid;
      gap:10px;
      padding:12px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#111318;
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
    .task-list, .agent-list, .version-list, .log-list { display:flex; flex-direction:column; gap:7px; }
    .work-terminal-list {
      display:flex;
      flex-direction:column;
      gap:12px;
    }
    .work-terminal-tabs {
      display:grid;
      grid-template-columns:repeat(auto-fit, minmax(170px, 1fr));
      gap:8px;
      padding-bottom:2px;
    }
    .work-terminal-tab {
      min-width:0;
      display:flex;
      align-items:center;
      gap:8px;
      padding:9px 12px;
      border-radius:999px;
      color:var(--muted);
    }
    .work-terminal-tab strong {
      min-width:0;
      overflow:hidden;
      text-overflow:ellipsis;
      white-space:nowrap;
    }
    .work-terminal-tab.active {
      color:var(--text);
      border-color:var(--line-2);
      background:var(--surface-2);
    }
    .doc-list { display:flex; flex-direction:column; gap:10px; }
    .task-card, .agent-card, .lock-card, .version-card {
      display:grid;
      gap:12px;
      padding:12px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#111318;
    }
    details.card > summary {
      list-style:none;
      cursor:pointer;
    }
    details.card > summary::-webkit-details-marker { display:none; }
    .board-summary {
      margin:-17px -17px 0;
      padding:17px;
      border-radius:var(--radius) var(--radius) 0 0;
    }
    .board-summary:hover { background:#181b20; }
    .board-details:not([open]) .board-summary {
      margin-bottom:-17px;
      border-radius:var(--radius);
    }
    .board-body {
      padding-top:14px;
      border-top:1px solid var(--line);
    }
    .task-card {
      grid-template-columns:70px minmax(0, 1fr) auto;
      align-items:center;
    }
    .task-row {
      border:1px solid var(--line);
      border-radius:12px;
      background:#101217;
    }
    .task-row summary,
    .task-row-line {
      min-width:0;
      display:grid;
      grid-template-columns:72px minmax(0, 1fr) auto;
      gap:10px;
      align-items:center;
      padding:10px 12px;
      list-style:none;
    }
    .task-row summary { cursor:pointer; }
    .task-row summary::-webkit-details-marker { display:none; }
    .task-row.is-done {
      opacity:.72;
      background:#0f1115;
    }
    .task-row.is-done .task-title {
      text-decoration:line-through;
      color:var(--muted);
    }
    .task-row-actions {
      display:flex;
      gap:8px;
      flex-wrap:wrap;
      padding:0 12px 12px 94px;
    }
    .task-row-id {
      color:var(--muted);
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:12px;
    }
    .task-row-tags {
      display:flex;
      gap:6px;
      flex-wrap:wrap;
      justify-content:flex-end;
    }
    .task-title { color:var(--text); line-height:1.35; overflow-wrap:anywhere; }
    .task-meta { display:flex; gap:7px; flex-wrap:wrap; margin-top:7px; }
    .doc-card {
      padding:14px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#111318;
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
    .work-terminal-card {
      min-width:0;
      padding:14px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#111318;
    }
    .work-terminal-card.is-live { border-color:#31573f; }
    .work-terminal-head {
      display:flex;
      justify-content:space-between;
      align-items:flex-start;
      gap:10px;
    }
    .work-terminal-title {
      display:flex;
      gap:10px;
      align-items:center;
      min-width:0;
    }
    .work-terminal-title strong {
      display:block;
      color:var(--text);
      line-height:1.2;
    }
    .work-terminal-title span {
      display:block;
      margin-top:4px;
      color:var(--muted);
      font-size:12px;
      overflow:hidden;
      text-overflow:ellipsis;
      white-space:nowrap;
    }
    .work-terminal-meta {
      display:flex;
      gap:6px;
      flex-wrap:wrap;
      margin-top:12px;
    }
    .work-terminal-actions {
      display:flex;
      gap:8px;
      flex-wrap:wrap;
      margin-top:12px;
    }
    .agent-command-box {
      display:grid;
      gap:8px;
      margin-top:12px;
      padding-top:12px;
      border-top:1px solid var(--line);
    }
    .agent-command-box textarea {
      min-height:74px;
      background:#0f1117;
    }
    .work-terminal-card details {
      margin-top:12px;
      border-top:1px solid var(--line);
      padding-top:10px;
    }
    .work-terminal-card summary {
      cursor:pointer;
      color:var(--soft);
      font-size:12px;
      font-weight:750;
    }
    .work-terminal-output {
      min-height:280px;
      max-height:56vh;
      margin:10px 0 0;
      padding:12px;
      border:1px solid #1d2430;
      border-radius:10px;
      background:#05070a;
      color:#d7deea;
      overflow:auto;
      white-space:pre-wrap;
      overflow-wrap:anywhere;
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:12px;
      line-height:1.45;
    }
    .session-grid {
      display:block;
    }
    .terminal-hub {
      display:grid;
      grid-template-columns:260px minmax(0, 1fr);
      gap:12px;
      align-items:stretch;
    }
    .agent-switcher {
      display:flex;
      flex-direction:column;
      gap:8px;
      min-width:0;
    }
    .agent-switch {
      width:100%;
      display:flex;
      flex-direction:column;
      gap:8px;
      padding:12px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#101217;
      color:var(--muted);
      text-align:left;
      white-space:normal;
      box-shadow:none;
    }
    .agent-switch:hover { background:#141720; border-color:var(--line-2); }
    .agent-switch.active {
      color:var(--text);
      border-color:#31573f;
      background:#111a15;
    }
    .agent-switch-top {
      display:flex;
      justify-content:space-between;
      gap:8px;
      align-items:center;
    }
    .agent-switch-name {
      display:flex;
      gap:8px;
      align-items:center;
      min-width:0;
      font-weight:800;
    }
    .agent-switch-name span:last-child {
      min-width:0;
      overflow:hidden;
      text-overflow:ellipsis;
      white-space:nowrap;
    }
    .agent-switch-meta {
      display:flex;
      gap:6px;
      flex-wrap:wrap;
      font-size:12px;
    }
    .terminal-pane {
      min-width:0;
      display:flex;
      flex-direction:column;
      gap:12px;
      padding:14px;
      border:1px solid var(--line);
      border-radius:16px;
      background:#111318;
    }
    .terminal-pane.is-live { border-color:#31573f; }
    .terminal-pane.is-offline { color:var(--muted); }
    .terminal-toolbar {
      display:flex;
      justify-content:space-between;
      gap:12px;
      align-items:flex-start;
      flex-wrap:wrap;
    }
    .terminal-title {
      display:flex;
      gap:10px;
      align-items:center;
      min-width:0;
    }
    .terminal-title-copy {
      min-width:0;
    }
    .terminal-title-copy strong {
      display:block;
      color:var(--text);
      font-size:16px;
      line-height:1.2;
    }
    .terminal-title-copy span {
      display:block;
      margin-top:3px;
      color:var(--muted);
      font-size:12px;
      white-space:nowrap;
      overflow:hidden;
      text-overflow:ellipsis;
    }
    .terminal-controls {
      display:flex;
      gap:8px;
      align-items:center;
      flex-wrap:wrap;
      justify-content:flex-end;
    }
    .font-controls {
      display:flex;
      gap:4px;
      padding:3px;
      border:1px solid var(--line);
      border-radius:12px;
      background:#0c0f15;
    }
    .font-controls button {
      padding:5px 8px;
      border-radius:8px;
      font-size:12px;
      background:transparent;
    }
    .terminal-meta-grid {
      display:grid;
      grid-template-columns:repeat(4, minmax(0, 1fr));
      gap:8px;
    }
    .terminal-meta {
      min-width:0;
      padding:8px 10px;
      border:1px solid var(--line);
      border-radius:10px;
      background:#0f1117;
    }
    .terminal-meta span {
      display:block;
      color:#7f8898;
      font-size:11px;
    }
    .terminal-meta strong {
      display:block;
      margin-top:3px;
      color:var(--soft);
      font-size:12px;
      overflow:hidden;
      text-overflow:ellipsis;
      white-space:nowrap;
    }
    .provider {
      display:flex;
      gap:10px;
      align-items:center;
      min-width:0;
    }
    .provider-mark {
      display:grid;
      place-items:center;
      flex:0 0 auto;
      width:36px;
      height:36px;
      border:1px solid var(--line-2);
      border-radius:10px;
      background:#0f1013;
      color:var(--text);
      font-weight:850;
      letter-spacing:-.05em;
    }
    .provider-openai .provider-mark { border-radius:50%; }
    .provider-claude .provider-mark { border-radius:11px; }
    .provider-gemini .provider-mark { border-radius:50%; }
    .provider-copy { min-width:0; }
    .provider-copy strong {
      display:block;
      color:var(--text);
      font-size:15px;
      line-height:1.2;
    }
    .provider-copy span {
      display:block;
      margin-top:3px;
      color:var(--muted);
      font-size:12px;
      white-space:nowrap;
      overflow:hidden;
      text-overflow:ellipsis;
    }
    .model-line {
      display:flex;
      gap:8px;
      flex-wrap:wrap;
      align-items:center;
      color:var(--soft);
      font-size:12px;
    }
    .session-state {
      flex:1 1 auto;
      display:flex;
      flex-direction:column;
      gap:7px;
      padding:12px;
      border:1px solid var(--line);
      border-radius:12px;
      background:#101217;
      color:var(--soft);
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:12px;
      line-height:1.45;
    }
    .term-row {
      display:grid;
      grid-template-columns:72px minmax(0, 1fr);
      gap:10px;
    }
    .term-prompt { color:#7f8898; }
    .term-value {
      min-width:0;
      overflow:hidden;
      text-overflow:ellipsis;
      white-space:nowrap;
    }
    .term-muted {
      margin-top:3px;
      color:#8d96a6;
      white-space:normal;
    }
    .mirror-meta {
      display:flex;
      justify-content:space-between;
      gap:10px;
      color:#8d96a6;
      font-family:ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size:11.5px;
    }
    .mirror-output {
      flex:1 1 auto;
      min-height:480px;
      max-height:68vh;
      margin:0;
      padding:14px;
      border:1px solid #1d2430;
      border-radius:10px;
      background:#05070a;
      overflow:auto;
      white-space:pre-wrap;
      overflow-wrap:anywhere;
      color:#d7deea;
      line-height:1.45;
      tab-size:2;
    }
    .mirror-output.is-empty {
      color:#7f8898;
      font-style:italic;
    }
    .mirror-output .ansi-bold { font-weight:800; }
    .mirror-output .ansi-dim { opacity:.68; }
    .mirror-output .ansi-italic { font-style:italic; }
    .mirror-output .ansi-underline { text-decoration:underline; }
    .mirror-output .ansi-inverse {
      color:#05070a;
      background:#d7deea;
    }
    .mirror-output .ansi-fg-30 { color:#7f8898; }
    .mirror-output .ansi-fg-31 { color:#ff8b92; }
    .mirror-output .ansi-fg-32 { color:#95d59b; }
    .mirror-output .ansi-fg-33 { color:#e7c76f; }
    .mirror-output .ansi-fg-34 { color:#8fb4ff; }
    .mirror-output .ansi-fg-35 { color:#d7a5ff; }
    .mirror-output .ansi-fg-36 { color:#7fd7df; }
    .mirror-output .ansi-fg-37 { color:#d7deea; }
    .mirror-output .ansi-fg-90 { color:#9aa4b2; }
    .mirror-output .ansi-fg-91 { color:#ffabb0; }
    .mirror-output .ansi-fg-92 { color:#b7efc7; }
    .mirror-output .ansi-fg-93 { color:#f2da8a; }
    .mirror-output .ansi-fg-94 { color:#b7ccff; }
    .mirror-output .ansi-fg-95 { color:#e5c2ff; }
    .mirror-output .ansi-fg-96 { color:#a7edf2; }
    .mirror-output .ansi-fg-97 { color:#ffffff; }
    .mirror-output .ansi-bg-40 { background:#05070a; }
    .mirror-output .ansi-bg-41 { background:#5a1f28; }
    .mirror-output .ansi-bg-42 { background:#1f4b2b; }
    .mirror-output .ansi-bg-43 { background:#55451d; }
    .mirror-output .ansi-bg-44 { background:#1f315a; }
    .mirror-output .ansi-bg-45 { background:#3b255a; }
    .mirror-output .ansi-bg-46 { background:#1d5055; }
    .mirror-output .ansi-bg-47 { background:#d7deea; color:#05070a; }
    .mirror-output .ansi-bg-100 { background:#202633; }
    .mirror-output .ansi-bg-101 { background:#7a2a36; }
    .mirror-output .ansi-bg-102 { background:#28623a; }
    .mirror-output .ansi-bg-103 { background:#705a25; }
    .mirror-output .ansi-bg-104 { background:#2a4174; }
    .mirror-output .ansi-bg-105 { background:#4f3275; }
    .mirror-output .ansi-bg-106 { background:#27686f; }
    .mirror-output .ansi-bg-107 { background:#ffffff; color:#05070a; }
    .session-foot {
      display:flex;
      justify-content:space-between;
      gap:10px;
      flex-wrap:wrap;
      align-items:center;
    }
    .sync-note {
      margin-top:12px;
      padding:12px;
      border:1px solid var(--line);
      border-radius:14px;
      background:#101217;
      color:var(--muted);
      font-size:13px;
      line-height:1.55;
    }
    .empty {
      padding:16px;
      border:1px dashed var(--line-2);
      border-radius:14px;
      color:var(--muted);
      background:#111318;
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
      background:#17191f;
      color:var(--text);
      box-shadow:none;
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
      .terminal-hub { grid-template-columns:1fr; }
      .agent-switcher {
        display:grid;
        grid-template-columns:repeat(3, minmax(180px, 1fr));
      }
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
      .task-row summary,
      .task-row-line {
        grid-template-columns:1fr;
        gap:7px;
      }
      .task-row-tags { justify-content:flex-start; }
      .task-row-actions { padding:0 12px 12px; }
      .row { flex-direction:column; align-items:stretch; }
      .agent-switcher {
        display:flex;
        flex-direction:row;
        overflow:auto;
        padding-bottom:2px;
      }
      .agent-switch { flex:0 0 220px; }
      .terminal-meta-grid { grid-template-columns:repeat(2, minmax(0, 1fr)); }
      .mirror-output { min-height:360px; max-height:60vh; }
    }
    @media (max-width: 460px) {
      .metrics { grid-template-columns:1fr; }
    }
    button.primary {
      background:var(--text);
      border-color:var(--text);
      color:#111318;
      font-weight:700;
      box-shadow:none;
    }
    button.primary:hover { background:#dfe2e7; border-color:#dfe2e7; }
    .next-step {
      display:flex;
      flex-wrap:wrap;
      justify-content:space-between;
      align-items:center;
      gap:18px;
      padding:18px;
      margin-bottom:16px;
      border:1px solid var(--line-2);
      border-radius:var(--radius);
      background:var(--surface);
      box-shadow:none;
    }
    .next-step.done { background:var(--surface); border-color:#2f5540; }
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
    .check.is-done .check-mark { background:#142018; border-color:#2e6040; color:var(--ok); }
    .lead { color:var(--soft); font-size:15px; line-height:1.62; }
    .concept-row { display:grid; grid-template-columns:repeat(4, minmax(0, 1fr)); gap:12px; margin-top:4px; }
    .concept { padding:13px; border:1px solid var(--line); border-radius:13px; background:#111318; }
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
      border:1px solid var(--line-2);
      background:var(--surface-2);
      color:var(--soft);
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
          <span>Terminal + memory bridge</span>
        </div>
      </div>
      <nav class="rail-nav" aria-label="Dashboard pages">
        <button class="nav-item active" type="button" data-page-target="overview"><span>Setup</span><span>01</span></button>
        <button class="nav-item" type="button" data-page-target="work"><span>Work</span><span>02</span></button>
      </nav>
      <div class="rail-footer">
        Local-only UI. <a href="https://github.com/tianyi-zhang-02/coact" target="_blank" rel="noreferrer">Docs on GitHub</a>.
      </div>
    </aside>

    <div class="main">
      <header class="topbar">
        <div class="headline">
          <div class="eyebrow">Shared project workspace</div>
          <h1>CoAct Control Center</h1>
          <p class="subtitle">Pick a project, start agents in their native terminals, and keep their brief, tasks, locks, and audit trail synced in one local workspace.</p>
          <div class="workspace" id="workspace" title="">Loading workspace…</div>
          <div class="project-bar">
            <select id="projectSelect" title="Active project"></select>
            <button class="small ghost" type="button" data-add-project>Add folder</button>
            <span class="project-hint">Switching projects changes board, locks, messages, launch folder, and agent state.</span>
          </div>
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
                <div class="soft">Full health check from a terminal: <code>coact doctor</code> · docs: <a href="https://github.com/tianyi-zhang-02/coact" target="_blank" rel="noreferrer">GitHub README</a></div>
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
              <p class="hint" style="margin-top:12px">On macOS, Start opens the agent in its native terminal and wires it into the same CoAct bridge. Other platforms can copy the command and run it manually. The control center never exposes arbitrary shell execution.</p>
            </section>
          </div>
          <details class="status-drawer">
            <summary>Workspace status</summary>
            <section class="metrics" aria-label="Workspace summary">
              <div class="metric"><span>Tasks</span><strong id="metricTasks">—</strong><small id="metricTasksSub">waiting</small></div>
              <div class="metric"><span>Live agents</span><strong id="metricAgents">—</strong><small id="metricAgentsSub">heartbeat</small></div>
              <div class="metric"><span>Locks</span><strong id="metricLocks">—</strong><small>conflict gate</small></div>
              <div class="metric"><span>Mode</span><strong id="metricMode">—</strong><small id="metricModeSub">workspace</small></div>
            </section>
          </details>
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
                <p class="lead">CoAct is a local terminal hub and project-memory bridge for coding agents. Claude Code, Codex, and Gemini keep their native CLI behavior, while CoAct gathers their project context into the same brief, task board, inbox, locks, and journal. On macOS it can also launch native terminal sessions from the control center.</p>
                <div class="concept-row">
                  <div class="concept"><strong>Terminal hub</strong><span>Bring agent terminals together without redesigning their UI.</span></div>
                  <div class="concept"><strong>Project memory</strong><span>Brief, tasks, messages, and journal live in one local layer.</span></div>
                  <div class="concept"><strong>File locks</strong><span>Two agents can't edit the same file at once.</span></div>
                  <div class="concept"><strong>Task board</strong><span>Agents claim work, so effort stays coordinated.</span></div>
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
                    <div class="step-item"><span class="step-num">1</span><div class="step-body"><strong>Initialize the repo</strong><span>Once per project, from the Overview or with <code>coact init</code>. It creates the local bridge and agent contracts.</span></div></div>
                    <div class="step-item"><span class="step-num">2</span><div class="step-body"><strong>Create shared project memory</strong><span>Write the brief and add tasks/messages that every agent can read before it starts working.</span></div></div>
                    <div class="step-item"><span class="step-num">3</span><div class="step-body"><strong>Start agents</strong><span>Launch Claude, Codex, or Gemini. On macOS, CoAct opens native terminals; on other platforms, copy the generated command.</span></div></div>
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
                  <summary>Is CoAct redesigning the agent terminal?</summary>
                  <p>No. CoAct keeps each CLI native. It stores shared project memory under <code>.coact</code>, and on macOS it can launch allowlisted agent terminals from the local control center.</p>
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
                  <p>Brief, tasks, messages, locks, presence, and logs are plain local project state under <code>.coact</code>. Agents read and write through CoAct commands and hooks.</p>
                </details>
                <details class="doc-card">
                  <summary>Version manager</summary>
                  <p>Managed versions live under <code>~/.coact/bin/coact-&lt;version&gt;</code>. Switching only repoints the managed <code>~/.coact/coact</code> shim.</p>
                </details>
              </div>
            </section>
          </div>
        </section>

        <section class="page-view" data-page="work">
          <div class="page-grid">
            <section class="card span-7" id="brief-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Context</div>
                  <h2>Brief</h2>
                </div>
                <span class="badge">shared</span>
              </div>
              <div class="stack">
                <textarea id="brief" placeholder="Goal, constraints, decisions, preferred agent split."></textarea>
                <div class="row wrap"><button onclick="saveBrief()">Save brief</button><span class="muted">Saved to <code>.coact/brief.md</code>.</span></div>
              </div>
            </section>

            <section class="card span-5" id="routing-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Routing</div>
                  <h2>Agents work in terminals</h2>
                </div>
                <span class="badge">native CLI</span>
              </div>
              <div class="stack">
                <p class="muted">Talk to Claude, Codex, or Gemini inside their own terminals. CoAct keeps the project memory synced through the brief, task board, locks, and journal.</p>
                <div id="workAgents" class="agent-list"></div>
                <p class="hint">For docs and deeper design notes, use the GitHub README instead of this local screen.</p>
              </div>
            </section>

            <section class="card span-12" id="terminals-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Agents</div>
                  <h2>Agent coordination</h2>
                </div>
                <span class="badge" id="terminalCount">0 agents</span>
              </div>
              <div class="stack">
                <div id="workTerminals" class="work-terminal-list"></div>
                <p class="hint">Steer agents in their native terminals. Use CoAct for shared context, task assignment, inbox notes, locks, and audit history.</p>
              </div>
            </section>

            <details class="card span-12 board-details" id="tasks-card" open>
              <summary class="section-head board-summary">
                <div>
                  <div class="eyebrow">Board</div>
                  <h2>Tasks</h2>
                </div>
                <div class="row wrap"><span class="badge" id="taskCount">0 tasks</span><span class="badge">collapse</span></div>
              </summary>
              <div class="stack board-body" id="taskBoard">
                <div class="row"><input id="taskTitle" placeholder="Add a concrete task for an agent" /><select id="taskOwner"><option value="">Unassigned</option><option value="codex">Codex</option><option value="claude">Claude</option><option value="gemini">Gemini</option></select><button onclick="addTask()">Add task</button></div>
                <div id="tasks" class="task-list"></div>
              </div>
            </details>
          </div>
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
                  <div class="eyebrow">Agent launchpad</div>
                  <h2>Agent launchpad, shared project memory</h2>
                </div>
                <span class="badge warn">macOS beta</span>
              </div>
              <div id="agentSessions" class="session-grid"></div>
              <div class="sync-note">On macOS, CoAct can launch allowlisted agent CLIs in native Terminal sessions and record local transcripts under <code>.coact/terminal</code>. Other platforms should copy the generated command and run it manually. CoAct separately persists shared project memory under <code>.coact</code>: brief, board, inbox messages, locks, presence, and audit journal.</div>
            </section>
          </div>
        </section>

        <section class="page-view" data-page="advanced">
          <div class="page-grid">
            <section class="card span-5" id="versions-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Release</div>
                  <h2>Versions</h2>
                </div>
                <span class="badge warn">advanced</span>
              </div>
              <div id="versions" class="version-list muted">Managed versions appear here after <code>coact update</code>.</div>
            </section>

            <section class="card span-7" id="log-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Audit</div>
                  <h2>Activity log</h2>
                </div>
                <span class="badge" id="logCount">0 events</span>
              </div>
              <div id="log" class="log muted">No events yet.</div>
            </section>
          </div>
        </section>
      </main>
    </div>
  </div>
  <div class="toast" id="toast"></div>

  <script>
    const TOKEN = "__COACT_TOKEN__";
    let lastBrief = null;
    let toastTimer = null;
    let launchCommands = [];
    let lastAgents = [];
    let terminalMirrors = [];
    let fullTerminalMirrors = {};
    let activeMirrorAgent = localStorage.getItem('coactActiveMirrorAgent') || 'codex';
    let terminalFontSize = parseInt(localStorage.getItem('coactTerminalFontSize') || '13', 10);
    const DEFAULT_PAGE = "overview";
    if (Number.isNaN(terminalFontSize)) terminalFontSize = 13;

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
    function providerSpec(id) {
      const specs = {
        codex: {id:'codex', label:'OpenAI / Codex', mark:'O', brand:'openai', binary:'codex'},
        claude: {id:'claude', label:'Claude Code', mark:'C', brand:'claude', binary:'claude'},
        gemini: {id:'gemini', label:'Gemini CLI', mark:'G', brand:'gemini', binary:'gemini'}
      };
      return specs[id] || {id:id, label:id, mark:(id || '?').slice(0, 1).toUpperCase(), brand:'generic', binary:id};
    }
    function commandFor(agent) {
      return launchCommands.find(c => c.agent === agent) || null;
    }
    function agentFor(agents, id) {
      return agents.find(a => a.id === id) || null;
    }
    function mirrorFor(agent) {
      return terminalMirrors.find(m => m.agent === agent) || null;
    }
    function shortPath(path) {
      const parts = String(path || '').split('/').filter(Boolean);
      return parts.slice(-2).join('/') || path || '-';
    }
    function clamp(n, min, max) {
      return Math.max(min, Math.min(max, n));
    }
    function setActiveMirrorAgent(agent) {
      activeMirrorAgent = agent || 'codex';
      localStorage.setItem('coactActiveMirrorAgent', activeMirrorAgent);
      renderAgentSessions(lastAgents);
    }
    function setTerminalFontSize(size) {
      terminalFontSize = clamp(size || 13, 10, 22);
      localStorage.setItem('coactTerminalFontSize', String(terminalFontSize));
      renderAgentSessions(lastAgents);
    }
    function ansiClasses(state) {
      const classes = [];
      if (state.bold) classes.push('ansi-bold');
      if (state.dim) classes.push('ansi-dim');
      if (state.italic) classes.push('ansi-italic');
      if (state.underline) classes.push('ansi-underline');
      if (state.inverse) classes.push('ansi-inverse');
      if (state.fg) classes.push('ansi-fg-' + state.fg);
      if (state.bg) classes.push('ansi-bg-' + state.bg);
      return classes.join(' ');
    }
    function resetAnsiState(state) {
      state.bold = false;
      state.dim = false;
      state.italic = false;
      state.underline = false;
      state.inverse = false;
      state.fg = '';
      state.bg = '';
    }
    function applyAnsiCodes(state, codes) {
      if (!codes.length) codes = [0];
      for (const code of codes) {
        if (code === 0) resetAnsiState(state);
        else if (code === 1) state.bold = true;
        else if (code === 2) state.dim = true;
        else if (code === 3) state.italic = true;
        else if (code === 4) state.underline = true;
        else if (code === 7) state.inverse = true;
        else if (code === 22) { state.bold = false; state.dim = false; }
        else if (code === 23) state.italic = false;
        else if (code === 24) state.underline = false;
        else if (code === 27) state.inverse = false;
        else if (code === 39) state.fg = '';
        else if (code === 49) state.bg = '';
        else if ((code >= 30 && code <= 37) || (code >= 90 && code <= 97)) state.fg = String(code);
        else if ((code >= 40 && code <= 47) || (code >= 100 && code <= 107)) state.bg = String(code);
      }
    }
    function renderAnsi(text) {
      const state = {};
      resetAnsiState(state);
      let out = '';
      let chunk = '';
      const flush = () => {
        if (!chunk) return;
        const classes = ansiClasses(state);
        out += classes ? '<span class="' + classes + '">' + esc(chunk) + '</span>' : esc(chunk);
        chunk = '';
      };
      for (let i = 0; i < text.length; i++) {
        const ch = text[i];
        if (ch !== '\x1b') {
          if (ch === '\r') continue;
          if (ch === '\b') {
            chunk = chunk.slice(0, -1);
            continue;
          }
          if (ch.charCodeAt(0) < 32 && ch !== '\n' && ch !== '\t') continue;
          chunk += ch;
          continue;
        }
        const next = text[i + 1];
        if (next === ']') {
          flush();
          let j = i + 2;
          while (j < text.length && text[j] !== '\x07' && !(text[j] === '\x1b' && text[j + 1] === '\\')) j++;
          i = j < text.length && text[j] === '\x1b' ? j + 1 : j;
          continue;
        }
        if (next === 'P' || next === 'X' || next === '^' || next === '_') {
          flush();
          let j = i + 2;
          while (j < text.length && !(text[j] === '\x1b' && text[j + 1] === '\\')) j++;
          i = j < text.length ? j + 1 : j;
          continue;
        }
        if (next !== '[') {
          flush();
          i += 1;
          continue;
        }
        let j = i + 2;
        while (j < text.length && (text.charCodeAt(j) < 0x40 || text.charCodeAt(j) > 0x7e)) j++;
        if (j >= text.length) break;
        const final = text[j];
        const body = text.slice(i + 2, j);
        flush();
        if (final === 'm') {
          const codes = body.split(';').filter(Boolean).map(n => parseInt(n, 10)).filter(n => !Number.isNaN(n));
          applyAnsiCodes(state, codes);
        }
        i = j;
      }
      flush();
      return out;
    }
    function scrollMirrorsToBottom() {
      document.querySelectorAll('.mirror-output').forEach(el => {
        if (el.dataset.follow === 'true') el.scrollTop = el.scrollHeight;
      });
    }
    function pageFromLocation() {
      const page = (window.location.hash || '').replace(/^#/, '');
      const aliases = {brief:'work', tasks:'work', messages:'work', agents:'work', guide:'overview', advanced:'overview', versions:'overview', log:'overview'};
      const resolved = aliases[page] || page;
      return document.querySelector('[data-page="'+resolved+'"]') ? resolved : DEFAULT_PAGE;
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
        desc = 'Create the local CoAct bridge for this repo — agent contracts, coordination files, and the .coact project memory layer. Nothing else is touched, and coact deinit removes it all.';
        action = '<button class="primary" onclick="initProject()">Initialize repo</button>';
      } else if (!hasBrief) {
        step = 'Step 2 of 4';
        title = 'Write your project brief';
        desc = 'Seed the shared project memory with the goal, constraints, and who does what. It saves to .coact/brief.md for agents to read.';
        action = '<button class="primary" data-goto="work">Write brief</button>';
      } else if (tasks.length === 0) {
        step = 'Step 3 of 4';
        title = 'Add your first task';
        desc = 'Break the work into concrete tasks. Agents claim tasks from the board, so two of them never grab the same thing.';
        action = '<button class="primary" data-goto="work">Add a task</button>';
      } else if (live === 0) {
        step = 'Step 4 of 4';
        title = 'Start an agent';
        desc = "Launch an agent below. This connects it to the same project bridge; macOS can open a native terminal directly.";
        action = '<button class="primary" data-goto="overview">Start agent</button>';
      } else {
        done = true;
        step = "You're set";
        title = live + ' agent' + (live > 1 ? 's' : '') + ' working now';
        desc = 'Agents are live and coordinating through the shared project memory, locks, and board. Track progress in Work.';
        action = '<button data-goto="work">Open work</button>';
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
        lastAgents = agents;
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
        renderProjects(s.projects || [], s.workspace || '');
        renderNextStep(s, tasks, agents);
        renderChecklist(s, tasks, agents);
        loadTerminalMirrors().catch(() => {});
      } catch (e) {
        document.getElementById('initState').className = 'badge bad';
        document.getElementById('initState').textContent = e.message;
      }
    }
    function renderTasks(tasks) {
      const done = tasks.filter(t => t.state === 'done').length;
      document.getElementById('taskCount').textContent = tasks.length + (tasks.length === 1 ? ' task' : ' tasks') + (done ? ' · ' + done + ' done' : '');
      document.getElementById('tasks').innerHTML = tasks.map(t => {
        const stateType = t.state === 'done' ? 'ok' : (t.state === 'doing' ? 'warn' : '');
        const line = '<span class="task-row-id">'+esc(t.id)+'</span><span class="task-title">'+esc(t.title)+'</span><span class="task-row-tags">'+badge(t.state, stateType)+badge(t.owner || 'unassigned')+'</span>';
        if (t.state === 'done') {
          return '<div class="task-row is-done"><div class="task-row-line">'+line+'</div></div>';
        }
        return '<details class="task-row"><summary>'+line+'</summary><div class="task-row-actions"><button class="small ghost" onclick="claimTask(\''+esc(t.id)+'\')">Claim</button><button class="small" onclick="doneTask(\''+esc(t.id)+'\')">Done</button></div></details>';
      }).join('') || '<div class="empty">No tasks yet. Add one above, then assign it to an agent.</div>';
    }
    function renderAgents(agents) {
      const live = agents.filter(a => a.live).length;
      document.getElementById('agentCount').textContent = live + ' live';
      document.getElementById('agents').innerHTML = agents.map(a => '<div class="agent-card"><div><div class="agent-name"><span class="badge '+(a.live?'ok':'bad')+'"><span class="dot"></span>'+(a.live?'live':'dead')+'</span><code>'+esc(a.id)+'</code></div><div class="agent-sub">'+esc(a.status || 'idle')+' · '+esc(a.enforcement || 'policy unknown')+'</div></div><div class="muted">'+esc(a.current_task || '-')+'<br>'+esc(a.beat || '-')+'</div></div>').join('') || '<div class="empty">No agents configured yet.</div>';
      const workAgents = document.getElementById('workAgents');
      if (workAgents) {
        workAgents.innerHTML = agents.map(a => '<div class="agent-card"><div><div class="agent-name"><span class="badge '+(a.live?'ok':'bad')+'">'+(a.live?'live':'offline')+'</span><code>'+esc(a.id)+'</code></div><div class="agent-sub">'+esc(a.status || 'idle')+'</div></div><div class="muted">'+esc(a.current_task || 'no task')+'</div></div>').join('') || '<div class="empty">No agents configured yet.</div>';
      }
      renderWorkTerminals(agents);
    }
    function renderWorkTerminals(agents) {
      const el = document.getElementById('workTerminals');
      if (!el) return;
      const focused = document.activeElement;
      const draft = focused && focused.id && focused.id.startsWith('agentInstruction-')
        ? {id: focused.id, value: focused.value, start: focused.selectionStart, end: focused.selectionEnd}
        : null;
      const ids = ['codex', 'claude', 'gemini'];
      const items = ids.map(id => {
        const spec = providerSpec(id);
        const agent = agentFor(agents, id) || {id:id, adapter:spec.binary, live:false};
        const mirror = mirrorFor(id);
        return {id, spec, agent, mirror};
      }).filter(item => item.agent.live || (item.mirror && item.mirror.exists));
      const count = document.getElementById('terminalCount');
      if (count) count.textContent = items.length + (items.length === 1 ? ' agent' : ' agents');
      if (!items.length) {
        el.innerHTML = '<div class="empty">No agent sessions found yet. Start agents from Setup or run <code>coact claude</code>, <code>coact codex</code>, or <code>coact gemini</code> in native terminals.</div>';
        return;
      }
      if (!items.some(item => item.id === activeMirrorAgent)) {
        activeMirrorAgent = items[0].id;
        localStorage.setItem('coactActiveMirrorAgent', activeMirrorAgent);
      }
      const tabs = items.map(item => {
        const tabLabel = item.id === 'codex' ? 'Codex' : (item.id === 'claude' ? 'Claude' : (item.id === 'gemini' ? 'Gemini' : item.spec.label));
        return '<button class="work-terminal-tab '+(item.id === activeMirrorAgent ? 'active' : '')+'" type="button" data-agent-focus="'+esc(item.id)+'"><span class="provider-mark">'+esc(item.spec.mark)+'</span><strong>'+esc(tabLabel)+'</strong>'+badge(item.agent.live ? 'live' : 'recorded', item.agent.live ? 'ok' : '')+'</button>';
      }).join('');
      const selected = items.find(item => item.id === activeMirrorAgent) || items[0];
      const id = selected.id;
      const spec = selected.spec;
      const agent = selected.agent;
      const mirror = selected.mirror;
      const fullMirror = fullTerminalMirrors[id];
      const hasFull = fullMirror && fullMirror.exists;
      const cmd = commandFor(id);
      const status = agent.status || (agent.live ? 'working' : 'offline');
      const task = agent.current_task || 'no task';
      const mirrorState = mirror && mirror.exists
        ? ((hasFull ? 'full · ' : 'transcript · ') + mirror.size + ' bytes')
        : 'no transcript yet';
      const updatedSource = hasFull ? fullMirror : mirror;
      const updated = updatedSource && updatedSource.updated_at ? updatedSource.updated_at.replace('T', ' ').replace('Z', ' UTC') : '-';
      const transcript = hasFull && fullMirror.tail ? renderAnsi(fullMirror.tail) : '';
      const transcriptBlock = hasFull
        ? '<details open><summary>Show full transcript</summary><pre class="work-terminal-output">'+transcript+'</pre></details>'
        : '<div class="hint" style="margin-top:12px">Full transcript is loaded only when requested, so polling stays light.</div>';
      const fullButton = mirror && mirror.exists
        ? (hasFull ? '<span class="badge ok">full loaded</span>' : '<button class="small" data-full-transcript="'+esc(id)+'">Load full transcript</button>')
        : '';
      const copyButton = cmd ? '<button class="small ghost" data-copy="'+esc(cmd.command)+'">Copy command</button>' : '';
      const composer = '<div class="agent-command-box"><textarea id="agentInstruction-'+esc(id)+'" placeholder="Send an inbox note to '+esc(spec.label)+'. This is async context, not terminal input."></textarea><div class="row wrap"><button class="small primary" data-send-inbox-note="'+esc(id)+'">Send inbox note</button><span class="hint">Writes to <code>.coact/inbox/'+esc(id)+'.md</code>; the agent reads it on its next turn.</span></div></div>';
      const panel = '<div class="work-terminal-card provider-'+esc(spec.brand)+' '+(agent.live?'is-live':'')+'"><div class="work-terminal-head"><div class="work-terminal-title"><div class="provider-mark">'+esc(spec.mark)+'</div><div><strong>'+esc(spec.label)+'</strong><span>'+esc(status)+' · '+esc(task)+'</span></div></div>'+badge(agent.live ? 'live' : 'recorded', agent.live ? 'ok' : '')+'</div><div class="work-terminal-meta">'+badge(mirrorState)+badge('updated '+updated)+'</div><div class="work-terminal-actions">'+fullButton+copyButton+'</div>'+composer+transcriptBlock+'</div>';
      el.innerHTML = '<div class="work-terminal-tabs">'+tabs+'</div>'+panel;
      if (draft) {
        const restored = document.getElementById(draft.id);
        if (restored) {
          restored.value = draft.value;
          restored.focus();
          try { restored.setSelectionRange(draft.start, draft.end); } catch (e) {}
        }
      }
    }
    function renderProjects(projects, workspace) {
      const select = document.getElementById('projectSelect');
      if (!select) return;
      const active = projects.find(p => p.active) || projects.find(p => p.root === workspace) || {root:workspace, name:shortPath(workspace), initialized:true};
      const seen = new Set();
      const opts = [];
      for (const p of projects.concat([active])) {
        if (!p || !p.root || seen.has(p.root)) continue;
        seen.add(p.root);
        opts.push('<option value="'+esc(p.root)+'" '+(p.root === active.root ? 'selected' : '')+'>'+esc((p.name || shortPath(p.root)) + (p.initialized ? '' : ' · not initialized'))+'</option>');
      }
      select.innerHTML = opts.join('');
      select.title = active.root || '';
    }
    function renderAgentSessions(agents) {
      const el = document.getElementById('agentSessions');
      if (!el) return;
      const ids = ['codex', 'claude', 'gemini'];
      if (!ids.includes(activeMirrorAgent)) activeMirrorAgent = ids[0];
      const switches = ids.map(id => {
        const spec = providerSpec(id);
        const agent = agentFor(agents, id) || {id:id, adapter:spec.binary, live:false};
        const cmd = commandFor(id);
        const installed = cmd ? cmd.installed : false;
        const mirror = mirrorFor(id);
        const mirrorState = mirror && mirror.exists
          ? ((mirror.truncated ? 'tail' : 'full') + ' · ' + mirror.size + ' bytes')
          : 'no transcript';
        const status = agent.status || (agent.live ? 'working' : 'offline');
        return '<button class="agent-switch '+(id === activeMirrorAgent ? 'active' : '')+'" type="button" data-agent-focus="'+esc(id)+'"><div class="agent-switch-top"><div class="agent-switch-name"><span class="provider-mark">'+esc(spec.mark)+'</span><span>'+esc(spec.label)+'</span></div>'+badge(agent.live ? 'live' : 'offline', agent.live ? 'ok' : 'bad')+'</div><div class="agent-switch-meta">'+badge(installed ? 'installed' : 'missing', installed ? 'ok' : 'bad')+badge(mirrorState)+'</div><div class="muted">'+esc(status)+' · '+esc(agent.current_task || 'no task')+'</div></button>';
      }).join('');

      const id = activeMirrorAgent;
      const spec = providerSpec(id);
      const agent = agentFor(agents, id) || {id:id, adapter:spec.binary, live:false};
      const cmd = commandFor(id);
      const canLaunch = cmd && cmd.installed && cmd.terminal_supported;
      const installed = cmd ? cmd.installed : false;
      const mirror = mirrorFor(id);
      const mirrorText = mirror && mirror.tail
        ? mirror.tail
        : 'No terminal transcript yet. Start this agent through CoAct to mirror its raw session here.';
      const mirrorHTML = mirror && mirror.tail ? renderAnsi(mirrorText) : esc(mirrorText);
      const mirrorClass = mirror && mirror.tail ? 'mirror-output' : 'mirror-output is-empty';
      const mirrorState = mirror && mirror.exists
        ? ((mirror.truncated ? 'tail · ' : 'full · ') + mirror.size + ' bytes')
        : 'not recording yet';
      const mirrorTime = mirror && mirror.updated_at ? mirror.updated_at.replace('T', ' ').replace('Z', ' UTC') : '-';
      const model = agent.model || 'not reported';
      const status = agent.status || (agent.live ? 'working' : 'offline');
      const task = agent.current_task || 'no active task';
      const startTitle = canLaunch
        ? 'Open ' + spec.label + ' in its native terminal wired into CoAct'
        : (installed ? 'One-click launch is macOS-only right now' : 'Install ' + spec.binary + ' and add it to PATH');
      const startButton = canLaunch
        ? '<button class="small primary" data-launch="'+esc(id)+'" title="'+esc(startTitle)+'">Start</button>'
        : '<button class="small" title="'+esc(startTitle)+'" disabled>Start</button>';
      const copyButton = cmd
        ? '<button class="small ghost" data-copy="'+esc(cmd.command)+'">Copy</button>'
        : '';
      const pane = '<div class="terminal-pane provider-'+esc(spec.brand)+' '+(agent.live?'is-live':'is-offline')+'"><div class="terminal-toolbar"><div class="terminal-title"><div class="provider-mark">'+esc(spec.mark)+'</div><div class="terminal-title-copy"><strong>'+esc(spec.label)+'</strong><span>'+esc(agent.adapter || spec.binary)+' · '+esc(status)+' · '+esc(cmd ? cmd.command : 'coact '+id)+'</span></div></div><div class="terminal-controls"><span class="badge '+(agent.live?'ok':'bad')+'">'+(agent.live?'live':'offline')+'</span><span class="badge '+(installed?'ok':'bad')+'">'+(installed?'installed':'missing')+'</span><div class="font-controls" aria-label="Terminal font size"><button type="button" data-font-delta="-1">A-</button><button type="button" data-font-reset>Reset</button><button type="button" data-font-delta="1">A+</button></div>'+startButton+copyButton+'</div></div><div class="terminal-meta-grid"><div class="terminal-meta"><span>Model</span><strong>'+esc(model)+'</strong></div><div class="terminal-meta"><span>Task</span><strong>'+esc(task)+'</strong></div><div class="terminal-meta"><span>Mirror</span><strong>'+esc(mirrorState)+'</strong></div><div class="terminal-meta"><span>Updated</span><strong>'+esc(mirrorTime)+'</strong></div></div><div class="mirror-meta"><span>Raw terminal projection · output only</span><span>'+esc(terminalFontSize)+'px</span></div><pre class="'+mirrorClass+'" data-follow="true" style="font-size:'+esc(String(terminalFontSize))+'px">'+mirrorHTML+'</pre></div>';
      el.innerHTML = '<div class="terminal-hub"><div class="agent-switcher">'+switches+'</div>'+pane+'</div>';
      scrollMirrorsToBottom();
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
    async function addTask(){ const title=document.getElementById('taskTitle').value; const owner=document.getElementById('taskOwner').value; if(!title.trim()) return; await mutate(owner ? 'Task scheduled' : 'Task added', () => api('/api/tasks',{method:'POST', body:JSON.stringify({title, owner})})); document.getElementById('taskTitle').value=''; }
    async function claimTask(id){ const owner=prompt('Owner agent?', 'claude'); if(!owner) return; mutate('Task claimed', () => api('/api/tasks/'+id+'/claim',{method:'POST', body:JSON.stringify({owner})})); }
    async function doneTask(id){ const owner=prompt('Owner agent?', 'human'); if(!owner) return; mutate('Task marked done', () => api('/api/tasks/'+id+'/done',{method:'POST', body:JSON.stringify({owner})})); }
    async function sendMessage(){ const text=document.getElementById('msgText').value; if(!text.trim()) return; await mutate('Message sent', () => api('/api/messages',{method:'POST', body:JSON.stringify({from:document.getElementById('msgFrom').value,to:document.getElementById('msgTo').value,text})})); document.getElementById('msgText').value=''; }
    async function loadLaunch(){
      const d = await api('/api/launch-commands');
      launchCommands = d.commands || [];
      document.getElementById('launch').innerHTML = launchCommands.map(c => {
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
    async function loadTerminalMirrors(){
      try {
        const d = await api('/api/terminal-mirror');
        terminalMirrors = d.mirrors || [];
        renderWorkTerminals(lastAgents);
        renderAgentSessions(lastAgents);
      } catch (e) {
        terminalMirrors = [];
        renderWorkTerminals(lastAgents);
      }
    }
    async function loadFullTranscript(agent){
      const d = await api('/api/terminal-mirror?agent=' + encodeURIComponent(agent) + '&full=1');
      const mirror = (d.mirrors || [])[0];
      if (!mirror || !mirror.exists) throw new Error('No transcript found for ' + agent);
      fullTerminalMirrors[agent] = mirror;
      renderWorkTerminals(lastAgents);
    }
    async function sendInboxNote(agent){
      const input = document.getElementById('agentInstruction-' + agent);
      const text = input ? input.value.trim() : '';
      if (!text) {
        showToast('Inbox note is empty');
        return;
      }
      await mutate('Inbox note sent to '+agent, () => api('/api/messages', {method:'POST', body:JSON.stringify({from:'human', to:agent, text})}));
      if (input) input.value = '';
    }
    async function switchProject(root){
      if (!root) return;
      await mutate('Project switched', () => api('/api/projects/active',{method:'POST', body:JSON.stringify({root})}));
      fullTerminalMirrors = {};
      loadLaunch().catch(e => showToast(e.message));
      loadTerminalMirrors().catch(() => {});
    }
    async function addProject(){
      const root = prompt('Project folder path?', document.getElementById('workspace').title || '');
      if (!root) return;
      await mutate('Project added', () => api('/api/projects',{method:'POST', body:JSON.stringify({root})}));
      loadLaunch().catch(e => showToast(e.message));
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
      const addProjectButton = ev.target.closest('[data-add-project]');
      if (addProjectButton) {
        await addProject();
        return;
      }
      const focus = ev.target.closest('[data-agent-focus]');
      if (focus) {
        setActiveMirrorAgent(focus.getAttribute('data-agent-focus'));
        return;
      }
      const fontDelta = ev.target.closest('[data-font-delta]');
      if (fontDelta) {
        setTerminalFontSize(terminalFontSize + parseInt(fontDelta.getAttribute('data-font-delta'), 10));
        return;
      }
      const fontReset = ev.target.closest('[data-font-reset]');
      if (fontReset) {
        setTerminalFontSize(13);
        return;
      }
      const fullTranscript = ev.target.closest('[data-full-transcript]');
      if (fullTranscript) {
        const agent = fullTranscript.getAttribute('data-full-transcript');
        try {
          showToast('Loading full transcript…');
          await loadFullTranscript(agent);
          showToast('Full transcript loaded');
        } catch (e) {
          showToast(e.message);
        }
        return;
      }
      const sendInboxNoteButton = ev.target.closest('[data-send-inbox-note]');
      if (sendInboxNoteButton) {
        await sendInboxNote(sendInboxNoteButton.getAttribute('data-send-inbox-note'));
        return;
      }
      const launch = ev.target.closest('[data-launch]');
      if (launch) {
        const agent = launch.getAttribute('data-launch');
        await mutate('Opening '+agent+' in Terminal', () => api('/api/agents/'+agent+'/launch', {method:'POST', body:'{}'}));
        loadTerminalMirrors().catch(() => {});
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
    document.addEventListener('change', ev => {
      const projectSelect = ev.target.closest('#projectSelect');
      if (projectSelect) switchProject(projectSelect.value);
    });
    window.addEventListener('popstate', () => setPage(pageFromLocation(), false));
    setPage(pageFromLocation(), false);
    loadLaunch().catch(e => showToast(e.message));
    loadTerminalMirrors().catch(() => {});
    refresh();
    setInterval(refresh, 1500);
  </script>
</body>
</html>`
