package ui

const indexHTML = `<!doctype html>
<html lang="__COACT_LANG__">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>CoAct Control Center</title>
  <link rel="stylesheet" href="/world/world.css?v=10" />
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
      transition:background .35s ease, color .35s ease;
    }
    body[data-coact-theme="ocean"] { --bg:#091317; --rail:#0d171b; --surface:#111d21; --surface-2:#17262b; --surface-3:#1d3036; --line:#26383d; --line-2:#385158; --accent:#63d8d1; }
    body[data-coact-theme="ecodome"] { --bg:#0d1410; --rail:#111914; --surface:#162019; --surface-2:#1c2a20; --surface-3:#243429; --line:#2b3b30; --line-2:#3d5544; --accent:#8bd58b; }
    body[data-coact-theme="wasteland"] { --bg:#17110d; --rail:#1b1410; --surface:#211914; --surface-2:#2a2019; --surface-3:#34271e; --line:#3d2f26; --line-2:#584337; --accent:#e5a262; }
    .theme-ambient { position:fixed; inset:0; z-index:4; overflow:hidden; pointer-events:none; opacity:.62; transition:opacity .25s ease; }
    body[data-active-page="station"] .theme-ambient { opacity:0; }
    body.ambient-off .theme-ambient { display:none; }
    .ambient-scene { display:none; position:absolute; inset:0; }
    body[data-coact-theme="orbit"] .ambient-orbit,
    body[data-coact-theme="ocean"] .ambient-ocean,
    body[data-coact-theme="ecodome"] .ambient-ecodome,
    body[data-coact-theme="wasteland"] .ambient-wasteland { display:block; }
    .ambient-creature { position:absolute; filter:drop-shadow(0 4px 8px #0008); image-rendering:pixelated; }
    .ambient-star { width:4px; height:4px; background:#b9d5ff; box-shadow:0 0 10px #8fb4ff; animation:ambient-pulse 2.8s steps(4,end) infinite; }
    .ambient-star.one { left:11px; top:22%; }
    .ambient-star.two { right:14px; top:61%; animation-delay:-1.2s; }
    .ambient-comet { right:-20px; top:34%; width:74px; height:3px; background:linear-gradient(90deg,transparent,#8fb4ff66,#eaf3ff); transform:rotate(-18deg); animation:ambient-drift 7s ease-in-out infinite; }
    .ambient-fish { width:29px; height:13px; border-radius:55% 45% 45% 55%; background:#6be4d6; animation:ambient-swim 5.5s ease-in-out infinite; }
    .ambient-fish::before { content:""; position:absolute; right:-9px; top:2px; border-top:5px solid transparent; border-bottom:5px solid transparent; border-left:10px solid #48aaa5; }
    .ambient-fish::after { content:""; position:absolute; left:6px; top:4px; width:3px; height:3px; background:#10282d; box-shadow:9px 5px 0 -1px #b9fff4; }
    .ambient-fish.one { left:7px; top:24%; }
    .ambient-fish.two { right:8px; top:72%; transform:scaleX(-1) scale(.72); animation-delay:-2s; }
    .ambient-hammer { left:-14px; top:66%; width:88px; height:18px; border-radius:55% 70% 65% 45%; background:#5d929b; animation:ambient-drift 6.5s ease-in-out infinite; }
    .ambient-hammer::before { content:""; position:absolute; left:2px; top:-6px; width:20px; height:30px; border-radius:4px; background:#729faa; }
    .ambient-hammer::after { content:""; position:absolute; right:-13px; top:3px; border-top:7px solid transparent; border-bottom:7px solid transparent; border-left:15px solid #4d7f88; }
    .ambient-whale { right:-26px; top:26%; width:108px; height:38px; border-radius:70% 35% 55% 65%; background:#426e83; animation:ambient-float 7.5s ease-in-out infinite; }
    .ambient-whale::before { content:""; position:absolute; right:-16px; top:3px; width:25px; height:25px; background:#426e83; clip-path:polygon(0 45%,100% 0,76% 48%,100% 100%); }
    .ambient-whale::after { content:""; position:absolute; left:24px; top:10px; width:4px; height:4px; border-radius:50%; background:#d9ffff; box-shadow:-8px 25px 0 5px #5c8799; }
    .ambient-leaf { width:19px; height:11px; border-radius:100% 0 100% 0; background:#83c96f; animation:ambient-fall 6s linear infinite; }
    .ambient-leaf.one { left:12px; top:18%; }
    .ambient-leaf.two { right:12px; top:45%; animation-delay:-3s; background:#c4d66f; }
    .ambient-butterfly { right:8px; top:70%; width:8px; height:8px; background:#ffd979; box-shadow:-7px -2px 0 2px #8dd88d,7px -2px 0 2px #8dd88d; animation:ambient-float 4s ease-in-out infinite; }
    .ambient-vine { left:0; top:38%; width:7px; height:150px; background:linear-gradient(#315f3f,#79a963,#315f3f); box-shadow:7px 18px 0 -2px #568b54,9px 64px 0 -2px #80aa5e; opacity:.7; }
    .ambient-dust { right:0; top:26%; width:74px; height:2px; background:#e5a26266; box-shadow:-20px 19px 0 #c7804460,-6px 43px 0 #efc18155; animation:ambient-drift 5s ease-in-out infinite; }
    .ambient-tumbleweed { left:-10px; top:72%; width:42px; height:42px; border:4px dotted #b87842; border-radius:50%; animation:ambient-roll 7s linear infinite; }
    .ambient-cactus { right:4px; top:58%; width:11px; height:58px; border-radius:5px 5px 0 0; background:#65794d; box-shadow:-9px 19px 0 -3px #65794d,9px 28px 0 -3px #65794d; }
    body[data-active-page="work"] .ambient-whale { top:58%; transform:scale(.82); }
    body[data-active-page="work"] .ambient-hammer { top:31%; }
    body[data-active-page="work"] .ambient-vine { top:19%; }
    body[data-active-page="work"] .ambient-tumbleweed { top:42%; }
    @keyframes ambient-float { 0%,100% { translate:0 0; } 50% { translate:0 -16px; } }
    @keyframes ambient-drift { 0%,100% { translate:0 0; } 50% { translate:18px 8px; } }
    @keyframes ambient-swim { 0%,100% { translate:0 0; } 50% { translate:16px -9px; } }
    @keyframes ambient-pulse { 0%,100% { opacity:.35; scale:.8; } 50% { opacity:1; scale:1.25; } }
    @keyframes ambient-fall { 0% { translate:0 -30px; rotate:0deg; } 100% { translate:22px 110vh; rotate:320deg; } }
    @keyframes ambient-roll { 0% { translate:0 0; rotate:0deg; } 50% { translate:45px -7px; rotate:180deg; } 100% { translate:0 0; rotate:360deg; } }
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
      position:relative;
      z-index:1;
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
      align-items:center;
      gap:16px;
      padding:12px 14px;
      border:1px solid var(--line);
      border-radius:16px;
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
      color:var(--muted);
      white-space:nowrap;
      overflow:hidden;
      text-overflow:ellipsis;
      font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size:12px;
      padding:0;
      border:0;
      background:transparent;
      cursor:default;
    }
    .project-bar {
      display:flex;
      gap:8px;
      align-items:center;
      flex-wrap:wrap;
      margin-top:0;
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
      gap:6px;
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
    .task-toolbar {
      display:flex;
      justify-content:space-between;
      gap:10px;
      align-items:center;
      flex-wrap:wrap;
    }
    .task-create { display:grid; grid-template-columns:minmax(220px,.8fr) minmax(320px,1.7fr) auto auto; gap:9px; align-items:start; }
    .task-create textarea { min-height:76px; resize:vertical; }
    .task-create input, .task-create select, .task-create button { min-height:44px; }
    .plan-create { display:grid; grid-template-columns:minmax(280px,2fr) minmax(140px,.6fr) minmax(180px,.8fr) auto; gap:9px; align-items:start; }
    .plan-create textarea { min-height:92px; }
    .plan-participants { display:flex; gap:14px; flex-wrap:wrap; color:var(--muted); font-size:13px; }
    .plan-participants label { display:flex; gap:6px; align-items:center; }
    .plan-participants input { width:auto; }
    .plan-status { padding:12px 14px; border:1px solid var(--line); border-radius:var(--radius-sm); background:var(--surface-2); }
    .task-filters { display:flex; gap:6px; flex-wrap:wrap; }
    .task-filter.active {
      color:var(--text);
      border-color:var(--line-2);
      background:var(--surface-2);
    }
    .work-terminal-list {
      display:flex;
      flex-direction:column;
      gap:12px;
    }
    .guard-strip {
      display:flex;
      gap:8px;
      flex-wrap:wrap;
    }
    .guard-strip span {
      display:inline-flex;
      align-items:center;
      gap:7px;
      padding:7px 10px;
      border:1px solid var(--line);
      border-radius:999px;
      color:var(--soft);
      background:#0e1317aa;
      font-size:12px;
    }
    .guard-strip span::before { content:"✓"; color:var(--ok); font-weight:900; }
    .crew-strip {
      display:grid;
      grid-template-columns:repeat(3,minmax(0,1fr));
      gap:8px;
    }
    .crew-strip .agent-card { min-height:74px; padding:10px; }
    .quick-message {
      display:grid;
      grid-template-columns:140px minmax(0,1fr) auto;
      gap:8px;
      align-items:center;
    }
    .quick-message select,.quick-message input { min-width:0; }
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
    .task-row.is-doing { border-color:#5f512d; }
    .task-row.is-claimed { border-color:#304b66; }
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
    .provider-antigravity .provider-mark { border-radius:50%; }
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
      .task-create { grid-template-columns:1fr; }
      .plan-create { grid-template-columns:1fr; }
      .agent-switcher {
        display:flex;
        flex-direction:row;
        overflow:auto;
        padding-bottom:2px;
      }
      .agent-switch { flex:0 0 220px; }
      .terminal-meta-grid { grid-template-columns:repeat(2, minmax(0, 1fr)); }
      .mirror-output { min-height:360px; max-height:60vh; }
      .crew-strip,.quick-message { grid-template-columns:1fr; }
      .theme-ambient { opacity:.28; }
    }
    @media (max-width: 460px) {
      .metrics { grid-template-columns:1fr; }
    }
    @media (prefers-reduced-motion: reduce) {
      .ambient-creature { animation:none !important; }
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
    .station-wrap {
      position:relative;
      min-height:calc(100vh - 205px);
      overflow:hidden;
      border:1px solid #25314a;
      border-radius:22px;
      background:#080b18;
      isolation:isolate;
    }
    body.station-page .topbar { display:none; }
    body.station-page .main { padding:12px; }
    body.station-page .pages { margin-top:0; }
    body.station-page .station-wrap { min-height:calc(100vh - 24px); }
    .station-scene {
      position:absolute;
      inset:0;
      min-height:540px;
      overflow:hidden;
      background:#060815;
    }
    .station-canvas {
      position:absolute;
      z-index:0;
      inset:0;
      width:100%;
      height:100%;
      image-rendering:pixelated;
      image-rendering:crisp-edges;
    }
    .station-scene::before {
      content:"";
      position:absolute;
      z-index:1;
      inset:-18%;
      pointer-events:none;
      background:
        radial-gradient(circle at 16% 18%, #7d58ff30, transparent 22%),
        radial-gradient(circle at 82% 24%, #3e9dff28, transparent 24%),
        radial-gradient(circle at 52% 78%, #ffb35c18, transparent 24%);
      animation:cosmic-shift 11s ease-in-out infinite alternate;
      mix-blend-mode:screen;
    }
    .station-scene::after {
      content:"";
      position:absolute;
      inset:0;
      pointer-events:none;
      background:linear-gradient(180deg, #05081418 48%, #050814b8 100%);
      animation:station-light 14s ease-in-out infinite;
    }
    .space-window {
      position:absolute;
      z-index:1;
      overflow:hidden;
      pointer-events:none;
      opacity:.72;
      mix-blend-mode:screen;
    }
    .space-window::before,
    .space-window::after {
      content:"";
      position:absolute;
      inset:-35%;
      background-image:
        radial-gradient(circle at 10% 20%, #fff 0 1px, transparent 1.8px),
        radial-gradient(circle at 30% 70%, #8fc8ff 0 1px, transparent 1.8px),
        radial-gradient(circle at 55% 35%, #fff 0 1.2px, transparent 2px),
        radial-gradient(circle at 78% 78%, #d7b5ff 0 1px, transparent 1.8px),
        radial-gradient(circle at 92% 15%, #fff 0 .8px, transparent 1.5px);
      background-size:110px 90px, 150px 130px, 180px 145px, 210px 170px, 240px 190px;
      animation:star-drift 18s linear infinite;
    }
    .space-window::after {
      opacity:.55;
      transform:scale(.72);
      animation-duration:31s;
      animation-direction:reverse;
    }
    .window-left { left:3%; top:2%; width:34%; height:34%; clip-path:polygon(9% 2%, 90% 0, 76% 100%, 18% 96%); }
    .window-right { right:3%; top:2%; width:34%; height:34%; clip-path:polygon(10% 0, 91% 2%, 82% 96%, 24% 100%); }
    .window-side-left { left:0; top:30%; width:13%; height:43%; clip-path:polygon(0 0, 100% 13%, 72% 88%, 0 100%); opacity:.28; }
    .window-side-right { right:0; top:30%; width:13%; height:43%; clip-path:polygon(0 13%, 100% 0, 100% 100%, 28% 88%); opacity:.28; }
    .shooting-star {
      position:absolute;
      z-index:2;
      top:12%;
      left:8%;
      width:90px;
      height:1px;
      border-radius:999px;
      background:linear-gradient(90deg, transparent, #d9eeff);
      filter:drop-shadow(0 0 4px #8fc8ff);
      opacity:0;
      transform:rotate(-18deg);
      animation:shooting-star 8s ease-in-out infinite 2s;
      pointer-events:none;
    }
    .table-orbit {
      position:absolute;
      z-index:2;
      left:50%;
      top:51%;
      width:24%;
      aspect-ratio:1;
      transform:translate(-50%, -50%);
      border:2px solid #9fd8ff62;
      border-radius:50%;
      box-shadow:inset 0 0 28px #84cfff30, 0 0 42px #84cfff26;
      animation:table-breathe 5s ease-in-out infinite;
      pointer-events:none;
    }
    .table-orbit::before {
      content:"";
      position:absolute;
      inset:12%;
      border:1px dashed #f4c77f48;
      border-radius:50%;
      animation:ring-spin 30s linear infinite;
    }
    .station-hud {
      position:absolute;
      z-index:5;
      top:18px;
      left:18px;
      right:18px;
      display:flex;
      align-items:flex-start;
      justify-content:space-between;
      gap:14px;
      pointer-events:none;
    }
    .station-title,
    .station-stats,
    .station-feed {
      border:1px solid #ffffff20;
      background:#080b18d8;
      box-shadow:0 18px 50px #00000045;
      backdrop-filter:blur(12px);
    }
    .station-title {
      max-width:430px;
      padding:14px 16px;
      border-radius:16px;
      pointer-events:auto;
    }
    .station-title h2 { font-size:20px; margin-top:3px; }
    .station-title p { margin-top:6px; color:#aeb7c9; font-size:13px; }
    .station-title-row { display:flex; align-items:flex-start; justify-content:space-between; gap:14px; }
    .station-motion {
      flex:0 0 auto;
      padding:6px 9px;
      border-color:#ffffff28;
      background:#ffffff0b;
      color:#b9c7dc;
      font-size:11px;
    }
    .station-motion::before { content:""; display:inline-block; width:7px; height:7px; margin-right:6px; border-radius:50%; background:#78e39a; box-shadow:0 0 0 4px #78e39a18; }
    .station-motion.is-off::before { background:#778094; box-shadow:none; }
    .station-mode { color:#98bcff; font-size:10px; font-weight:900; letter-spacing:.17em; text-transform:uppercase; }
    .station-stats { display:flex; gap:6px; padding:7px; border-radius:14px; }
    .station-stat { min-width:68px; padding:7px 9px; border-radius:10px; background:#ffffff08; text-align:center; }
    .station-stat strong { display:block; font-size:16px; }
    .station-stat span { color:#8e98ab; font-size:10px; text-transform:uppercase; letter-spacing:.08em; }
    .station-agent {
      --pod-x:25%;
      --pod-y:31%;
      --table-x:42%;
      --table-y:52%;
      position:absolute;
      z-index:3;
      left:var(--pod-x);
      top:var(--pod-y);
      width:clamp(84px, 11vw, 148px);
      aspect-ratio:1;
      transform:translate(-50%, -50%);
      transition:left .9s cubic-bezier(.2,.8,.2,1), top .9s cubic-bezier(.2,.8,.2,1), opacity .35s ease, filter .35s ease;
      cursor:pointer;
      outline:none;
    }
    .station-agent.claude { --pod-x:75%; --pod-y:31%; --table-x:50%; --table-y:52%; }
    .station-agent.antigravity { --pod-x:24%; --pod-y:72%; --table-x:58%; --table-y:52%; }
    .station-agent.at-table { left:var(--table-x); top:var(--table-y); }
    .station-agent.is-walking { transition-duration:2.45s; transition-timing-function:linear; }
    .station-agent.is-offline { opacity:1; filter:none; }
    .station-agent.is-offline .astro-sprite { opacity:.56; filter:saturate(.48) brightness(.8); }
    .station-agent.is-offline .astro-sprite img { animation:sleep-float 3.8s ease-in-out infinite; }
    .station-agent.is-live { filter:drop-shadow(0 10px 18px #0008); }
    .station-agent:hover,
    .station-agent:focus-visible,
    .station-agent.is-selected { z-index:6; transform:translate(-50%, -50%) scale(1.08); }
    .station-agent.is-selected .agent-chip { border-color:#91bdff88; box-shadow:0 0 0 4px #91bdff18, 0 12px 34px #0008; }
    .astro-sprite { position:absolute; inset:0; overflow:hidden; image-rendering:pixelated; }
    .astro-sprite img {
      position:absolute;
      inset:0;
      width:100%;
      height:100%;
      object-fit:contain;
      image-rendering:pixelated;
      animation:astro-bob 2.4s ease-in-out infinite;
    }
    .astro-sprite .walk-frame { display:none; opacity:0; animation:none; }
    .station-agent.is-walking .astro-sprite .pose-main { opacity:0; }
    .station-agent.is-walking .astro-sprite .walk-frame { display:block; }
    .station-agent.is-walking .astro-sprite .walk-a { animation:walk-frame-a .42s steps(1, end) infinite; }
    .station-agent.is-walking .astro-sprite .walk-b { animation:walk-frame-b .42s steps(1, end) infinite; }
    .station-agent.is-walking.walk-left .astro-sprite .walk-frame { transform:scaleX(-1); }
    .astro-sprite.pose-offline img { animation:none; }
    .agent-chip {
      position:absolute;
      z-index:4;
      left:50%;
      bottom:-12px;
      min-width:112px;
      max-width:180px;
      padding:7px 10px;
      transform:translateX(-50%);
      border:1px solid #ffffff24;
      border-radius:12px;
      background:#080b18e8;
      text-align:center;
      box-shadow:0 10px 30px #0006;
    }
    .agent-chip strong { display:block; font-size:12px; }
    .agent-chip span { display:block; margin-top:2px; overflow:hidden; color:#9ca7ba; font-size:10px; white-space:nowrap; text-overflow:ellipsis; }
    .agent-signal {
      position:absolute;
      z-index:5;
      top:3px;
      right:3px;
      width:11px;
      height:11px;
      border:2px solid #080b18;
      border-radius:50%;
      background:#687186;
    }
    .crew-bubble {
      position:absolute;
      z-index:7;
      top:35%;
      left:50%;
      padding:8px 12px;
      transform:translate(-50%, 8px) scale(.94);
      border:1px solid #9bc8ff55;
      border-radius:12px;
      background:#080b18e8;
      color:#dbe9ff;
      font-size:11px;
      font-weight:800;
      letter-spacing:.05em;
      opacity:0;
      box-shadow:0 10px 30px #0007;
      transition:opacity .25s ease, transform .25s ease;
      pointer-events:none;
    }
    .crew-bubble.show { opacity:1; transform:translate(-50%, 0) scale(1); }
    .is-live .agent-signal { background:#74e39a; box-shadow:0 0 0 5px #74e39a18; animation:signal-pulse 1.6s ease-out infinite; }
    .station-feed {
      position:absolute;
      z-index:5;
      right:18px;
      bottom:18px;
      left:18px;
      display:flex;
      align-items:center;
      justify-content:space-between;
      gap:16px;
      min-height:62px;
      padding:12px 15px;
      border-radius:16px;
    }
    .station-feed-copy { min-width:0; }
    .station-feed-copy strong { display:block; font-size:13px; }
    .station-feed-copy span { display:block; margin-top:3px; overflow:hidden; color:#96a0b2; font-size:12px; white-space:nowrap; text-overflow:ellipsis; }
    .station-actions { display:flex; gap:8px; flex:0 0 auto; }
    @keyframes astro-bob { 0%,100% { margin-top:0; } 50% { margin-top:-5px; } }
    @keyframes walk-frame-a { 0%,49% { opacity:1; } 50%,100% { opacity:0; } }
    @keyframes walk-frame-b { 0%,49% { opacity:0; } 50%,100% { opacity:1; } }
    @keyframes sleep-float { 0%,100% { transform:translateY(0) rotate(-1deg); } 50% { transform:translateY(-7px) rotate(1deg); } }
    @keyframes signal-pulse { 0%,100% { box-shadow:0 0 0 3px #74e39a18; } 50% { box-shadow:0 0 0 9px #74e39a00; } }
    @keyframes star-drift { from { transform:translate3d(0, 0, 0); } to { transform:translate3d(14%, 9%, 0); } }
    @keyframes cosmic-shift { from { transform:translate3d(-2%, -1%, 0) scale(1); opacity:.45; } to { transform:translate3d(3%, 2%, 0) scale(1.06); opacity:.9; } }
    @keyframes station-light { 0%,100% { opacity:.9; } 50% { opacity:.66; } }
    @keyframes table-breathe { 0%,100% { opacity:.42; transform:translate(-50%, -50%) scale(.98); } 50% { opacity:.82; transform:translate(-50%, -50%) scale(1.03); } }
    @keyframes ring-spin { to { transform:rotate(360deg); } }
    @keyframes shooting-star {
      0%,68%,100% { opacity:0; transform:translate3d(0, 0, 0) rotate(-18deg); }
      71% { opacity:.85; }
      78% { opacity:0; transform:translate3d(520px, 145px, 0) rotate(-18deg); }
    }
    @keyframes state-flash { 0% { filter:brightness(1); } 35% { filter:brightness(1.35); } 100% { filter:brightness(1); } }
    .station-scene.state-change .station-hud,
    .station-scene.state-change .station-feed { animation:state-flash .7s ease-out; }
    .station-motion-off .station-scene::before,
    .station-motion-off .station-scene::after,
    .station-motion-off .station-scene *,
    .station-motion-off .station-scene *::before,
    .station-motion-off .station-scene *::after { animation-play-state:paused !important; }
    body {
      font-family:ui-monospace, "SFMono-Regular", Menlo, Monaco, Consolas, monospace;
      letter-spacing:-.02em;
      background:#080a11;
    }
    h1, h2, h3, strong { letter-spacing:-.045em; }
    button, input, textarea, select, code {
      border-radius:2px;
      font-family:inherit;
    }
    button {
      box-shadow:3px 3px 0 #05060a;
      transition:background .1s steps(2, end), border-color .1s steps(2, end), transform .1s steps(2, end);
    }
    button:hover { transform:translate(-1px, -1px); box-shadow:4px 4px 0 #05060a; }
    button:active { transform:translate(2px, 2px); box-shadow:1px 1px 0 #05060a; }
    .rail,
    .topbar,
    .card,
    .metric,
    .status-drawer,
    .next-step,
    .command,
    .agent-card,
    .task-row,
    .work-terminal-card,
    .terminal-pane,
    .empty,
    .version-card,
    .doc-card {
      border-radius:2px;
      box-shadow:4px 4px 0 #07080c;
    }
    .brand-mark,
    .badge,
    .workspace,
    .nav-item,
    .check-mark,
    .provider-mark { border-radius:2px; }
    .brand-mark {
      border-width:2px;
      box-shadow:3px 3px 0 #07080c;
      image-rendering:pixelated;
    }
    .nav-item { border-width:2px; }
    .nav-item.active { box-shadow:inset 3px 0 0 var(--accent); }
    .eyebrow,
    .station-mode,
    .station-stat span {
      font-family:inherit;
      letter-spacing:.12em;
      text-shadow:2px 2px 0 #05060a;
    }
    .station-wrap {
      border:3px solid #273552;
      border-radius:2px;
      box-shadow:6px 6px 0 #05060a, inset 0 0 0 2px #070a14;
      image-rendering:pixelated;
    }
    .station-scene {
      image-rendering:pixelated;
    }
    .station-scene::after {
      background:
        repeating-linear-gradient(0deg, #02040a18 0 1px, transparent 1px 4px),
        linear-gradient(180deg, #05081412 48%, #050814a8 100%);
      animation:station-light 14s steps(8, end) infinite;
    }
    .station-title,
    .station-stats,
    .station-feed,
    .agent-chip,
    .crew-bubble {
      border-width:2px;
      border-radius:2px;
      background:#080b18e8;
      backdrop-filter:none;
      box-shadow:5px 5px 0 #03050a;
      clip-path:polygon(0 0, calc(100% - 10px) 0, 100% 10px, 100% 100%, 10px 100%, 0 calc(100% - 10px));
    }
    .station-title { max-width:390px; }
    .station-title h2 { font-size:19px; text-transform:uppercase; }
    .station-motion {
      border-radius:2px;
      box-shadow:2px 2px 0 #03050a;
      text-transform:uppercase;
    }
    .station-stat { border-radius:0; border-left:1px solid #ffffff12; }
    .station-stat:first-child { border-left:0; }
    .station-agent,
    .astro-sprite,
    .astro-sprite img { image-rendering:pixelated; }
    .agent-chip strong { text-transform:uppercase; }
    .agent-signal { border-radius:1px; transform:rotate(45deg); }
    .is-live .agent-signal { animation:signal-pulse 1.6s steps(4, end) infinite; }
    .table-orbit,
    .table-orbit::before { animation-timing-function:steps(24, end); }
    .toast { border-radius:2px; box-shadow:4px 4px 0 #05060a; }
    @media (max-width: 680px) {
      .station-title { max-width:calc(100vw - 56px); }
      .station-title h2 { font-size:16px; }
    }
    @media (max-width: 820px) {
      .station-wrap { min-height:620px; }
      .station-hud { flex-direction:column; }
      .station-stats { align-self:flex-start; }
      .station-agent { width:104px; }
      .station-feed { align-items:flex-start; flex-direction:column; }
      .station-actions { width:100%; }
      .station-actions button { flex:1; }
    }
  </style>
</head>
<body data-coact-theme="orbit" data-active-page="station">
  <div class="theme-ambient" aria-hidden="true">
    <div class="ambient-scene ambient-orbit"><i class="ambient-creature ambient-star one"></i><i class="ambient-creature ambient-star two"></i><i class="ambient-creature ambient-comet"></i></div>
    <div class="ambient-scene ambient-ocean"><i class="ambient-creature ambient-fish one"></i><i class="ambient-creature ambient-fish two"></i><i class="ambient-creature ambient-hammer"></i><i class="ambient-creature ambient-whale"></i></div>
    <div class="ambient-scene ambient-ecodome"><i class="ambient-creature ambient-vine"></i><i class="ambient-creature ambient-leaf one"></i><i class="ambient-creature ambient-leaf two"></i><i class="ambient-creature ambient-butterfly"></i></div>
    <div class="ambient-scene ambient-wasteland"><i class="ambient-creature ambient-dust"></i><i class="ambient-creature ambient-tumbleweed"></i><i class="ambient-creature ambient-cactus"></i></div>
  </div>
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
        <button class="nav-item active" type="button" data-page-target="station"><span>Station</span><span>01</span></button>
        <button class="nav-item" type="button" data-page-target="overview"><span>Setup</span><span>02</span></button>
        <button class="nav-item" type="button" data-page-target="work"><span>Work</span><span>03</span></button>
      </nav>
      <div class="rail-footer">
        Local-only UI. <a href="https://github.com/tianyi-zhang-02/coact" target="_blank" rel="noreferrer">Docs on GitHub</a>.
      </div>
    </aside>

    <div class="main">
      <header class="topbar">
        <div class="headline">
          <div class="row wrap"><div class="eyebrow">Active project</div><div class="workspace" id="workspace" title="">Loading workspace…</div></div>
          <div class="project-bar">
            <select id="projectSelect" title="Active project"></select>
            <button class="small ghost" type="button" data-add-project>Add folder</button>
          </div>
        </div>
        <div class="status-panel">
          <div class="status-line">
            <span class="badge" id="version">coact —</span>
            <span class="badge" id="initBadge">checking</span>
            <button class="small ghost" type="button" data-toggle-ambient aria-pressed="true">Ambience on</button>
          </div>
          <div class="muted" id="updated">Syncing…</div>
        </div>
      </header>

      <main class="pages">
        <section class="page-view active" data-page="station">
          <section class="pixel-world-shell" aria-label="CoAct orbital crew pixel world">
            <div class="pixel-world-viewport"><canvas id="coactPixelBackground" width="768" height="512" aria-hidden="true"></canvas><canvas id="coactPixelWorld" width="768" height="512" aria-label="Animated pixel world showing live coding agents"></canvas></div>
            <header class="pixel-world-topbar">
              <div class="pixel-world-title"><small id="worldMode">Station syncing</small><strong id="worldTitle">CoAct Orbital Crew</strong><span id="worldSubtitle">Live agents move through one shared, protected workspace.</span></div>
              <div class="pixel-world-right">
                <div class="pixel-world-stats" aria-label="Workspace status">
                  <div class="pixel-world-stat"><b id="worldLive">0</b><span>live</span></div>
                  <div class="pixel-world-stat"><b id="worldTasks">0</b><span>tasks</span></div>
                  <div class="pixel-world-stat"><b id="worldLocks">0</b><span>locks</span></div>
                </div>
                <div class="pixel-world-controls"><button type="button" data-world-theme>Orbit</button><button type="button" data-world-quality title="Automatically protects character frame rate">Auto</button><button type="button" data-world-pause>Pause</button><button type="button" data-world-help>?</button></div>
              </div>
            </header>
            <div class="pixel-tooltip" role="tooltip"></div>
            <aside class="pixel-agent-panel" aria-live="polite">
              <header><h3 id="worldAgentName">Agent</h3><button type="button" data-world-close-agent>Close</button></header>
              <dl><dt>Status</dt><dd id="worldAgentStatus">—</dd><dt>Task</dt><dd id="worldAgentTask">—</dd><dt>Heartbeat</dt><dd id="worldAgentBeat">—</dd><dt>Animation</dt><dd id="worldAgentMode">—</dd></dl>
              <div class="pixel-agent-assist" id="worldAgentAssist"><strong>Teammate waiting</strong><p id="worldAgentAssistText">Another agent may benefit from help.</p><div class="pixel-agent-assist-actions"><button type="button" data-world-assist-offer>Offer help</button><button class="takeover" type="button" data-world-assist-handoff>Take over</button></div></div>
            </aside>
            <aside class="pixel-help">
              <header><h3>Flight manual</h3></header>
              <div class="pixel-help-grid"><div><kbd>T</kbd> Change environment</div><div><kbd>P</kbd> Pause animation</div><div><kbd>AUTO</kbd> Auto / Max / Lite graphics</div><div><kbd>?</kbd> Toggle this help</div><div><kbd>Esc</kbd> Close panels</div><div>Click crew or companions for live details.</div></div>
              <div style="margin-top:14px"><button type="button" data-world-close-help>Back to station</button></div>
            </aside>
            <footer class="pixel-world-footer">
              <div class="pixel-world-event"><strong id="worldEventTitle">Waiting for crew</strong><span id="worldEventText">Start agents with coact codex, coact claude, or coact antigravity.</span></div>
              <div class="pixel-world-actions"><button type="button" data-goto="overview">Setup</button><button class="primary" type="button" data-goto="work">Open board</button></div>
            </footer>
          </section>
        </section>

        <section class="page-view" data-page="overview">
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
                <p class="lead">CoAct is a local terminal hub and project-memory bridge for coding agents. Claude Code, Codex, and Antigravity keep their native CLI behavior, while CoAct gathers their project context into the same brief, task board, inbox, locks, and journal. On macOS CoAct can also launch native terminal sessions from the control center.</p>
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
                    <div class="step-item"><span class="step-num">3</span><div class="step-body"><strong>Start agents</strong><span>Launch Claude, Codex, or Antigravity. On macOS, CoAct opens native terminals; on other platforms, copy the generated command.</span></div></div>
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
                  <p>No. UI launch actions are allowlisted through built-in adapters such as Claude, Codex, and Antigravity. Generic shell execution is intentionally not exposed.</p>
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
            <section class="card span-12" id="routing-card">
              <div class="section-head">
                <div>
                  <div class="eyebrow">Coordinate</div>
                  <h2>Crew controls</h2>
                </div>
                <span class="badge">native terminals</span>
              </div>
              <div class="stack">
                <div class="guard-strip"><span>Task ownership</span><span>File conflict locks</span><span>Inbox + audit trail</span></div>
                <div id="workAgents" class="agent-list crew-strip"></div>
                <div class="quick-message">
                  <input id="msgFrom" type="hidden" value="human" />
                  <select id="msgTo" aria-label="Message recipient"><option value="codex">Codex</option><option value="claude">Claude</option><option value="antigravity">Antigravity</option></select>
                  <input id="msgText" placeholder="Send shared context to an agent" />
                  <button onclick="sendMessage()">Send note</button>
                </div>
              </div>
            </section>

            <details class="card span-12 board-details" id="tasks-card" open>
              <summary class="section-head board-summary">
                <div>
                  <div class="eyebrow">Board</div>
                  <h2>Tasks</h2>
                </div>
                <div class="row wrap"><span class="badge" id="taskCount">0 open</span><span class="badge">collapse</span></div>
              </summary>
              <div class="stack board-body" id="taskBoard">
                <div class="task-create"><input id="taskTitle" placeholder="Short Dashboard description" /><textarea id="taskPrompt" placeholder="Full prompt sent to the assigned agent. Defaults to the short description."></textarea><select id="taskOwner"><option value="">Unassigned</option><option value="codex">Codex</option><option value="claude">Claude</option><option value="antigravity">Antigravity</option></select><button onclick="addTask()">Add task</button></div>
                <div class="task-toolbar"><span class="muted" id="taskSummary">0 todo · 0 assigned · 0 active · 0 done</span><div class="task-filters"><button class="small ghost task-filter active" data-task-filter="open" onclick="setTaskFilter('open')">Open</button><button class="small ghost task-filter" data-task-filter="all" onclick="setTaskFilter('all')">All</button><button class="small ghost task-filter" data-task-filter="done" onclick="setTaskFilter('done')">Done</button></div></div>
                <div id="tasks" class="task-list"></div>
              </div>
            </details>

            <details class="card span-12 board-details" id="planning-card">
              <summary class="section-head board-summary">
                <div><div class="eyebrow">Plan together</div><h2>Lead planning</h2></div>
                <span class="badge">review by default</span>
              </summary>
              <div class="stack board-body">
                <div class="plan-create"><textarea id="planBrief" placeholder="Describe the goal, constraints, and expected result. The lead converts this into reviewed execution tasks."></textarea><select id="planLead"><option value="codex">Codex lead</option><option value="claude">Claude lead</option><option value="antigravity">Antigravity lead</option></select><select id="planApproval"><option value="review">Ask me before distributing</option><option value="auto">Auto-distribute (dangerous)</option></select><button onclick="startPlan()">Start planning</button></div>
                <div class="plan-participants"><strong>Planning pair</strong><label><input type="checkbox" data-plan-participant="codex" checked /> Codex</label><label><input type="checkbox" data-plan-participant="claude" checked /> Claude</label><label><input type="checkbox" data-plan-participant="antigravity" /> Antigravity</label></div>
                <div id="planStatus" class="plan-status muted">No planning run yet. Review mode lets the lead draft tasks, but a human must approve before distribution.</div>
              </div>
            </details>

            <details class="card span-12 board-details" id="brief-card">
              <summary class="section-head board-summary">
                <div><div class="eyebrow">Shared context</div><h2>Project brief</h2></div>
                <span class="badge">expand to edit</span>
              </summary>
              <div class="stack board-body">
                <textarea id="brief" placeholder="Goal, constraints, decisions, preferred agent split."></textarea>
                <div class="row wrap"><button onclick="saveBrief()">Save brief</button><span class="muted">Shared with every agent through <code>.coact/brief.md</code>.</span></div>
              </div>
            </details>

            <details class="card span-12 board-details" id="terminals-card">
              <summary class="section-head board-summary">
                <div><div class="eyebrow">Details</div><h2>Agent sessions</h2></div>
                <span class="badge" id="terminalCount">0 agents</span>
              </summary>
              <div class="stack board-body">
                <div id="workTerminals" class="work-terminal-list"></div>
                <p class="hint">Optional transcript and per-agent inbox details. Normal steering stays in each native terminal.</p>
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

  <script src="/world/world.js?v=27"></script>
  <script>
    const TOKEN = "__COACT_TOKEN__";
    let lastBrief = null;
    let toastTimer = null;
    let launchCommands = [];
    let lastAgents = [];
    let terminalMirrors = [];
    let fullTerminalMirrors = {};
    let lastDashboardSignature = '';
    let lastMirrorPoll = 0;
    let refreshInFlight = false;
    let lastStationTasks = [];
    let lastBoardTasks = [];
    let lastStationLocks = [];
    let lastStationLog = [];
    let lastStationSignature = '';
    let lastStationCrewSignature = '';
    let stationAtTable = false;
    let stationSceneMood = 'idle';
    let stationWorldTick = 0;
    let stationCanvasTime = 0;
    let stationCanvasLastFrame = 0;
    let stationCanvasStarted = false;
    let selectedStationAgent = '';
    const stationAgentPositions = {
      codex:{x:25,y:31},
      claude:{x:75,y:31},
      antigravity:{x:24,y:72}
    };
    const stationPodPositions = {
      codex:{x:25,y:31},
      claude:{x:75,y:31},
      antigravity:{x:24,y:72}
    };
    const stationTablePositions = {
      codex:{x:42,y:52},
      claude:{x:50,y:49},
      antigravity:{x:58,y:52}
    };
    const stationRoutes = {
      codex:[{x:25,y:31},{x:31,y:39},{x:38,y:43},{x:29,y:46}],
      claude:[{x:75,y:31},{x:69,y:39},{x:62,y:43},{x:71,y:46}],
      antigravity:[{x:24,y:72},{x:31,y:67},{x:38,y:61},{x:30,y:57}]
    };
    const stationStars = Array.from({length:96}, (_, index) => ({
      x:(index * 83 + 17) % 480,
      y:(index * 47 + 11) % 270,
      size:index % 11 === 0 ? 2 : 1,
      speed:.0015 + (index % 7) * .00025,
      phase:(index * 19) % 100
    }));
    const storedStationMotion = localStorage.getItem('coactStationMotion');
    let stationMotionEnabled = storedStationMotion === null
      ? !window.matchMedia('(prefers-reduced-motion: reduce)').matches
      : storedStationMotion !== 'off';
    const storedAmbient = localStorage.getItem('coactAmbientDecorations');
    let ambientEnabled = storedAmbient === null
      ? false
      : storedAmbient !== 'off';
    let activeMirrorAgent = localStorage.getItem('coactActiveMirrorAgent') || 'codex';
    let taskFilter = localStorage.getItem('coactTaskFilter') || 'open';
    let terminalFontSize = parseInt(localStorage.getItem('coactTerminalFontSize') || '13', 10);
    const DEFAULT_PAGE = "station";
    if (Number.isNaN(terminalFontSize)) terminalFontSize = 13;

    async function api(path, opts) {
      opts = opts || {};
      const headers = Object.assign({'Content-Type':'application/json','X-Coact-Token':TOKEN}, opts.headers || {});
      const res = await fetch(path, Object.assign({}, opts, {headers}));
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || res.statusText);
      return data;
    }
    window.addEventListener('coact:assist', async event => {
      const detail = event.detail || {};
      const waiting = String(detail.waiting || '');
      const helper = String(detail.helper || '');
      if (!waiting || !helper || waiting === helper) return;
      try {
        if (detail.action === 'offer') {
          await api('/api/messages', {method:'POST', body:JSON.stringify({
            from:'human',
            to:waiting,
            text:helper+' appears available to help with your current task. Share context with them, or ask me to approve a handoff.'
          })});
          showToast('Help offer sent to ' + waiting);
        } else if (detail.action === 'handoff') {
          if (!confirm('Move active tasks and release locks from ' + waiting + ' to ' + helper + '?')) return;
          const result = await api('/api/handoff', {method:'POST', body:JSON.stringify({
            from:waiting,
            to:helper,
            note:'Take over after reviewing the shared brief, board, inbox, and current locks.'
          })});
          showToast('Handoff complete · ' + ((result.tasks || []).length) + ' task(s) moved');
        } else {
          return;
        }
        window.CoActWorld?.assistHandled(detail.action);
        refresh();
      } catch (error) {
        showToast(error.message || String(error));
      }
    });
    function esc(s) { return String(s || '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
    function badge(text, type) { return '<span class="badge '+esc(type || '')+'">'+esc(text)+'</span>'; }
    function providerSpec(id) {
      const specs = {
        codex: {id:'codex', label:'OpenAI / Codex', mark:'O', brand:'openai', binary:'codex'},
        claude: {id:'claude', label:'Claude Code', mark:'C', brand:'claude', binary:'claude'},
        antigravity: {id:'antigravity', label:'Antigravity CLI', mark:'A', brand:'antigravity', binary:'agy'}
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
    function setStationMotion(enabled) {
      stationMotionEnabled = !!enabled;
      localStorage.setItem('coactStationMotion', stationMotionEnabled ? 'on' : 'off');
      document.body.classList.toggle('station-motion-off', !stationMotionEnabled);
      const button = document.getElementById('stationMotion');
      if (button) {
        button.textContent = stationMotionEnabled ? 'Ambient on' : 'Ambient off';
        button.classList.toggle('is-off', !stationMotionEnabled);
        button.setAttribute('aria-pressed', stationMotionEnabled ? 'true' : 'false');
      }
      if (stationMotionEnabled) setTimeout(advanceStationWorld, 120);
    }
    function stationPolygon(ctx, points, color, stroke) {
      ctx.beginPath();
      points.forEach((point, index) => index ? ctx.lineTo(point[0], point[1]) : ctx.moveTo(point[0], point[1]));
      ctx.closePath();
      ctx.fillStyle = color;
      ctx.fill();
      if (stroke) {
        ctx.strokeStyle = stroke;
        ctx.lineWidth = 2;
        ctx.stroke();
      }
    }
    function stationOctagon(ctx, cx, cy, rx, ry, color, stroke) {
      stationPolygon(ctx, [
        [cx-rx*.68,cy-ry],[cx+rx*.68,cy-ry],[cx+rx,cy-ry*.55],[cx+rx,cy+ry*.55],
        [cx+rx*.68,cy+ry],[cx-rx*.68,cy+ry],[cx-rx,cy+ry*.55],[cx-rx,cy-ry*.55]
      ], color, stroke);
    }
    function drawStationStars(ctx, time, alpha) {
      stationStars.forEach(star => {
        const x = (star.x + time * star.speed) % 480;
        const twinkle = ((Math.floor(time / 420) + star.phase) % 5 === 0) ? 1 : .48;
        ctx.globalAlpha = (alpha || 1) * twinkle;
        ctx.fillStyle = star.size === 2 ? '#b9ddff' : '#f3f7ff';
        ctx.fillRect(Math.floor(x), star.y, star.size, star.size);
      });
      ctx.globalAlpha = 1;
    }
    function drawStationWindow(ctx, points, time, phase) {
      ctx.save();
      stationPolygon(ctx, points, phase === 1 ? '#100d2d' : (phase === 2 ? '#071d29' : '#060a20'), '#435875');
      ctx.clip();
      drawStationStars(ctx, time * 1.35, .9);
      ctx.globalAlpha = .28;
      ctx.fillStyle = phase === 3 ? '#ff9d5c' : (phase === 1 ? '#7b52ff' : '#2f8dff');
      for (let index=0; index<18; index++) {
        const x = (index * 31 + Math.floor(time / 180)) % 520 - 20;
        const y = 10 + (index * 17) % 110;
        ctx.fillRect(x, y, 5 + index % 9, 2 + index % 4);
      }
      ctx.globalAlpha = 1;
      ctx.restore();
    }
    function drawStationCanvas(time) {
      const canvas = document.getElementById('stationCanvas');
      if (!canvas) return;
      const ctx = canvas.getContext('2d');
      ctx.imageSmoothingEnabled = false;
      const phase = Math.floor(time / 18000) % 4;
      const palettes = [
        ['#05081a','#101a34','#223252'],
        ['#090719','#1b1238','#3b275b'],
        ['#041319','#0c2830','#21434b'],
        ['#160b12','#34202d','#624039']
      ];
      const palette = palettes[phase];
      ctx.fillStyle = palette[0];
      ctx.fillRect(0,0,480,270);
      drawStationStars(ctx,time,.92);
      ctx.globalAlpha = .34;
      ctx.fillStyle = phase === 1 ? '#7542cc' : (phase === 3 ? '#e6844b' : '#2575b9');
      for (let index=0; index<24; index++) {
        const x = ((index * 29) + Math.floor(time / 230)) % 520 - 20;
        const y = 8 + (index * 13) % 250;
        ctx.fillRect(x,y,8 + index % 13,2 + index % 5);
      }
      ctx.globalAlpha = 1;

      const outer = [[7,55],[105,14],[375,14],[473,55],[480,168],[427,264],[53,264],[0,168]];
      const inner = [[28,66],[119,31],[361,31],[452,66],[461,159],[414,244],[66,244],[19,159]];
      stationPolygon(ctx,outer,'#c0c0ba','#ecded2');
      stationPolygon(ctx,inner,palette[1],'#2d405c');
      stationPolygon(ctx,[[47,76],[127,45],[353,45],[433,76],[440,154],[397,226],[83,226],[40,154]],palette[1],'#354b68');

      drawStationWindow(ctx,[[46,69],[113,34],[194,37],[167,91],[79,103]],time,phase);
      drawStationWindow(ctx,[[286,37],[367,34],[434,69],[401,103],[313,91]],time,phase);
      drawStationWindow(ctx,[[7,93],[54,103],[61,164],[18,184],[0,155]],time,phase);
      drawStationWindow(ctx,[[426,103],[473,93],[480,155],[462,184],[419,164]],time,phase);

      ctx.strokeStyle = '#304765';
      ctx.lineWidth = 1;
      for (let x=96; x<410; x+=36) {
        ctx.beginPath(); ctx.moveTo(x,104); ctx.lineTo(x-24,218); ctx.stroke();
      }
      for (let y=112; y<224; y+=24) {
        ctx.beginPath(); ctx.moveTo(62,y); ctx.lineTo(418,y); ctx.stroke();
      }

      const screenPulse = Math.floor(time / 500) % 2;
      stationOctagon(ctx,112,82,57,27,'#20283b','#7889a0');
      ctx.fillStyle = screenPulse ? '#4cc9ff' : '#245e8b'; ctx.fillRect(84,68,38,3); ctx.fillRect(89,74,27,2);
      stationOctagon(ctx,368,82,57,27,'#20283b','#7889a0');
      ctx.fillStyle = screenPulse ? '#ffb85c' : '#87582b'; ctx.fillRect(358,68,38,3); ctx.fillRect(364,74,27,2);
      stationOctagon(ctx,108,205,51,28,'#20283b','#7889a0');
      ctx.fillStyle = screenPulse ? '#65efd1' : '#277361'; ctx.fillRect(92,194,31,3); ctx.fillRect(96,200,22,2);

      const tableGlow = stationAtTable ? '#77d6ff' : (stationSceneMood === 'conflict' ? '#ff657a' : '#e7a95d');
      stationOctagon(ctx,240,150,64,42,'#292738','#7d6d64');
      stationOctagon(ctx,240,150,49,31,'#3a2e36',tableGlow);
      stationOctagon(ctx,240,150,28,17,stationAtTable ? '#21455b' : '#5a3d2b',tableGlow);
      const pulse = 4 + Math.floor((Math.sin(time / 620)+1)*3);
      ctx.strokeStyle = tableGlow; ctx.globalAlpha=.42; ctx.strokeRect(240-pulse,150-pulse,pulse*2,pulse*2); ctx.globalAlpha=1;

      ctx.fillStyle = '#69d5ff';
      for (let index=0; index<12; index++) {
        const x = 48 + index * 35;
        const on = (index + Math.floor(time / 360)) % 4 === 0;
        ctx.globalAlpha = on ? 1 : .24;
        ctx.fillRect(x,238,3,2);
      }
      ctx.globalAlpha = 1;

      const meteorPhase = (time % 9000) / 9000;
      if (meteorPhase > .7 && meteorPhase < .82) {
        const progress = (meteorPhase-.7)/.12;
        const x = 40 + progress*240;
        const y = 22 + progress*58;
        ctx.fillStyle='#f8fbff'; ctx.fillRect(Math.floor(x),Math.floor(y),7,2);
        ctx.fillStyle='#7ec8ff'; ctx.fillRect(Math.floor(x)-8,Math.floor(y)+2,10,1);
      }

      if (stationSceneMood === 'conflict') {
        ctx.globalAlpha = .12 + (Math.floor(time/400)%2)*.1;
        ctx.fillStyle='#ff304f'; ctx.fillRect(0,0,480,270); ctx.globalAlpha=1;
      } else if (stationSceneMood === 'celebrate') {
        ctx.fillStyle='#ffe074';
        for (let index=0; index<18; index++) ctx.fillRect((index*73+Math.floor(time/35))%480,(index*41+Math.floor(time/52))%270,2,2);
      }
    }
    function stationCanvasFrame(timestamp) {
      if (!stationCanvasLastFrame) stationCanvasLastFrame = timestamp;
      const delta = Math.min(50, timestamp - stationCanvasLastFrame);
      stationCanvasLastFrame = timestamp;
      if (stationMotionEnabled) stationCanvasTime += delta;
      drawStationCanvas(stationCanvasTime);
      requestAnimationFrame(stationCanvasFrame);
    }
    function initStationCanvas() {
      if (stationCanvasStarted) return;
      stationCanvasStarted = true;
      requestAnimationFrame(stationCanvasFrame);
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
      document.body.classList.toggle('station-page', page === 'station');
      document.body.dataset.activePage = page;
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
    function setAmbientEnabled(enabled) {
      ambientEnabled = !!enabled;
      document.body.classList.toggle('ambient-off', !ambientEnabled);
      localStorage.setItem('coactAmbientDecorations', ambientEnabled ? 'on' : 'off');
      const button = document.querySelector('[data-toggle-ambient]');
      if (button) {
        button.textContent = ambientEnabled ? 'Ambience on' : 'Ambience off';
        button.setAttribute('aria-pressed', String(ambientEnabled));
      }
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
        todo: tasks.filter(t => t.state === 'todo').length,
        claimed: tasks.filter(t => t.state === 'claimed').length
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
      if(refreshInFlight)return;
      refreshInFlight=true;
      try {
        const s = await api('/api/state');
        const tasks = s.tasks || [];
        const agents = s.agents || [];
        const locks = s.locks || [];
        const log = s.log || [];
        lastAgents = agents;
        document.getElementById('updated').textContent = 'Last sync ' + new Date().toLocaleTimeString();
        const dashboardSignature = JSON.stringify({
          workspace:s.workspace||'', initialized:!!s.initialized, mode:s.mode||'', version:s.version||'', brief:s.brief||'',
          tasks:tasks.map(t=>[t.id,t.title,t.state,t.owner]),
          agents:agents.map(a=>[a.id,a.live,a.status,a.current_task,a.enforcement]),
          locks:locks.map(l=>[l.path,l.owner,l.ttl_seconds,l.reason]),
          log:log.map(item=>[item.ts,item.agent,item.event,item.id,item.path,item.to]),
          projects:s.projects||[], versions:s.versions||[], manifest:s.manifest||null, plan:s.plan||null
        });
        if (dashboardSignature === lastDashboardSignature) {
          maybeLoadTerminalMirrors(false).catch(()=>{});
          return;
        }
        lastDashboardSignature = dashboardSignature;
        const stats = taskStats(tasks);
        const liveAgents = agents.filter(a => a.live).length;
        const wsEl = document.getElementById('workspace');
        const ws = s.workspace || '';
        wsEl.textContent = (ws.split('/').filter(Boolean).slice(-2).join('/') || ws) + (s.initialized ? '' : ' · not initialized');
        wsEl.title = ws;
        document.getElementById('version').textContent = 'coact ' + (s.version || 'dev');
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
        renderPlan(s.plan || null);
        renderAgents(agents);
        renderLocks(locks);
        renderLog(log);
        renderVersions(s.versions || [], s.manifest || null);
        renderGuide(s.versions || [], s.manifest || null);
        renderProjects(s.projects || [], s.workspace || '');
        renderNextStep(s, tasks, agents);
        renderChecklist(s, tasks, agents);
        renderStation(tasks, agents, locks, log);
        maybeLoadTerminalMirrors(false).catch(() => {});
      } catch (e) {
        document.getElementById('initState').className = 'badge bad';
        document.getElementById('initState').textContent = e.message;
      } finally {refreshInFlight=false;}
    }
    function renderStation(tasks, agents, locks, log) {
      window.CoActWorld?.update({tasks, agents, locks, log});
      return;
      const crew = document.getElementById('stationCrew');
      if (!crew) return;
      lastStationTasks = tasks;
      lastStationLocks = locks;
      lastStationLog = log;
      const ids = ['codex', 'claude', 'antigravity'];
      const live = agents.filter(a => a.live);
      const activeTasks = tasks.filter(t => t.state === 'doing');
      const recent = log.length ? log[log.length - 1] : null;
      const recentEvent = String(recent && recent.event || '').toLowerCase();
      const recentTS = Date.parse(recent && recent.ts || '');
      const recentAge = Number.isFinite(recentTS) ? Date.now() - recentTS : Infinity;
      const isFresh = recentAge < 3 * 60 * 1000;
      const conflict = locks.some(l => String(l.reason || '').toLowerCase().includes('conflict')) || (isFresh && /conflict|denied|blocked/.test(recentEvent));
      const planning = live.length > 1 && isFresh && /plan|review|proposal|handoff/.test(recentEvent);
      const messaging = live.length > 1 && recentAge < 45 * 1000 && recentEvent === 'msg.send';
      const celebrating = live.length > 0 && recentAge < 12 * 1000 && recentEvent === 'task.finish';
      const sameTask = live.length > 1 && live.every(a => a.current_task && a.current_task === live[0].current_task);
      const atTable = conflict || planning || messaging || sameTask;
      stationAtTable = atTable;
      stationSceneMood = conflict ? 'conflict' : (celebrating ? 'celebrate' : (atTable ? 'sync' : (live.length > 1 ? 'parallel' : (live.length === 1 ? 'active' : 'idle'))));
      const signature = agents.map(a => [a.id, a.live, a.status, a.current_task].join(':')).join('|') + '|' + locks.length + '|' + activeTasks.length + '|' + recentEvent;
      if (lastStationSignature && signature !== lastStationSignature) {
        const scene = document.getElementById('stationScene');
        scene.classList.remove('state-change');
        requestAnimationFrame(() => scene.classList.add('state-change'));
        setTimeout(() => scene.classList.remove('state-change'), 750);
      }
      lastStationSignature = signature;
      const mode = conflict ? 'Conflict review' : (messaging ? 'Crew sync' : (atTable ? 'Crew planning' : (live.length > 1 ? 'Parallel work' : (live.length === 1 ? 'Solo orbit' : 'Station idle'))));
      document.getElementById('stationMode').textContent = mode;
      document.getElementById('stationLive').textContent = live.length;
      document.getElementById('stationTasks').textContent = activeTasks.length;
      document.getElementById('stationLocks').textContent = locks.length;
      document.getElementById('stationSummary').textContent = atTable
        ? 'The live crew has moved to the shared table to align before execution.'
        : 'Each agent works in its native terminal while CoAct keeps project memory and ownership synchronized.';
      const taskTitles = {};
      tasks.forEach(t => { taskTitles[t.id] = t.title; });
      const assets = {codex:'astro-orbit', claude:'astro-nova', antigravity:'astro-comet'};
      const names = {codex:'Orbit · Codex', claude:'Nova · Claude', antigravity:'Comet · Antigravity'};
      ids.forEach(id => {
        const agent = agentFor(agents, id);
        if (!agent || !agent.live) stationAgentPositions[id] = Object.assign({}, stationPodPositions[id]);
      });
      const crewSignature = ids.map(id => {
        const agent = agentFor(agents, id) || {id:id, live:false, status:'offline', current_task:''};
        return [id, agent.live, agent.status, agent.current_task, atTable, celebrating].join(':');
      }).join('|');
      if (crewSignature !== lastStationCrewSignature) crew.innerHTML = ids.map(id => {
        const agent = agentFor(agents, id) || {id:id, live:false, status:'offline', current_task:''};
        const status = String(agent.status || (agent.live ? 'working' : 'offline')).toLowerCase();
        let pose = 'offline';
        if (agent.live && celebrating) pose = 'celebrate';
        else if (agent.live && atTable) pose = 'plan';
        else if (agent.live && /idle|wait|ready/.test(status)) pose = 'idle';
        else if (agent.live) pose = 'work';
        const task = taskTitles[agent.current_task] || agent.current_task || (agent.live ? status : 'offline');
        const position = stationAgentPositions[id];
        return '<div class="station-agent '+esc(id)+' '+(agent.live?'is-live':'is-offline')+(agent.live&&atTable?' at-table':'')+(selectedStationAgent===id?' is-selected':'')+'" style="left:'+position.x+'%;top:'+position.y+'%" role="button" tabindex="0" data-station-agent="'+esc(id)+'" aria-label="Inspect '+esc(names[id])+'"><div class="agent-signal"></div><div class="astro-sprite pose-'+pose+'"><img class="pose-main" src="/assets/'+assets[id]+'-'+pose+'.png?v=pixel-world-1" alt="" draggable="false"><img class="walk-frame walk-a" src="/assets/'+assets[id]+'-walk-a.png?v=pixel-world-1" alt="" draggable="false"><img class="walk-frame walk-b" src="/assets/'+assets[id]+'-walk-b.png?v=pixel-world-1" alt="" draggable="false"></div><div class="agent-chip"><strong>'+esc(names[id])+'</strong><span>'+esc(task)+'</span></div></div>';
      }).join('');
      lastStationCrewSignature = crewSignature;
      if (atTable) setTimeout(() => live.forEach(agent => moveStationAgent(agent.id, stationTablePositions[agent.id], true)), 80);
      const feedTitle = document.getElementById('stationFeedTitle');
      const feedText = document.getElementById('stationFeedText');
      if (selectedStationAgent) {
        renderStationSelection();
        return;
      } else if (!recent) {
        feedTitle.textContent = live.length ? 'Crew connected' : 'Waiting for crew';
        feedText.textContent = live.length ? 'CoAct is receiving agent heartbeats.' : 'Start an agent from Setup or a native terminal.';
        return;
      }
      const eventNames = {
        'task.add':'Task created',
        'task.claim':'Task claimed',
        'task.finish':'Task completed',
        'msg.send':'Message routed',
        'lock.acquire':'File protected',
        'lock.release':'File released',
        'agent.launch':'Agent launched',
        'plan.start':'Planning started',
        'plan.finish':'Plan finalized'
      };
      feedTitle.textContent = eventNames[recent.event] || String(recent.event || 'Workspace updated').replaceAll('.', ' ');
      const details = Object.keys(recent).filter(k => !['ts','event'].includes(k)).map(k => k+' '+recent[k]).join(' · ');
      feedText.textContent = details || 'Shared project state changed.';
    }
    function renderStationSelection() {
      const agent = agentFor(lastAgents, selectedStationAgent);
      if (!agent) return;
      const task = lastStationTasks.find(t => t.id === agent.current_task);
      document.getElementById('stationFeedTitle').textContent = providerSpec(agent.id).label + (agent.live ? ' · live' : ' · offline');
      document.getElementById('stationFeedText').textContent = [task ? task.title : (agent.current_task || 'no active task'), agent.status || 'idle', agent.beat || 'no recent heartbeat'].join(' · ');
    }
    function selectStationAgent(id) {
      selectedStationAgent = selectedStationAgent === id ? '' : id;
      document.querySelectorAll('[data-station-agent]').forEach(el => el.classList.toggle('is-selected', el.getAttribute('data-station-agent') === selectedStationAgent));
      if (selectedStationAgent) renderStationSelection();
      else renderStation(lastStationTasks, lastAgents, lastStationLocks, lastStationLog);
    }
    function moveStationAgent(id, target, walking) {
      const el = document.querySelector('[data-station-agent="'+id+'"]');
      if (!el || !target) return;
      const previous = stationAgentPositions[id] || target;
      stationAgentPositions[id] = {x:target.x, y:target.y};
      el.classList.toggle('walk-left', target.x < previous.x);
      if (walking) el.classList.add('is-walking');
      requestAnimationFrame(() => {
        el.style.left = target.x + '%';
        el.style.top = target.y + '%';
      });
      if (walking) setTimeout(() => el.classList.remove('is-walking'), 2500);
    }
    function showStationBubble(text, duration) {
      const bubble = document.getElementById('stationBubble');
      if (!bubble) return;
      bubble.textContent = text;
      bubble.classList.add('show');
      clearTimeout(showStationBubble.timer);
      showStationBubble.timer = setTimeout(() => bubble.classList.remove('show'), duration || 2600);
    }
    function advanceStationWorld() {
      if (!stationMotionEnabled || document.visibilityState === 'hidden') return;
      const live = lastAgents.filter(agent => agent.live);
      if (!live.length) return;
      stationWorldTick++;
      if (stationAtTable) {
        live.forEach(agent => moveStationAgent(agent.id, stationTablePositions[agent.id], false));
        if (stationWorldTick % 3 === 0) showStationBubble('shared context sync', 2200);
        return;
      }
      if (live.length > 1 && stationWorldTick % 5 === 0) {
        moveStationAgent(live[0].id, {x:45,y:39}, true);
        moveStationAgent(live[1].id, {x:55,y:39}, true);
        showStationBubble('quick context check-in', 3000);
        return;
      }
      live.forEach((agent, index) => {
        const route = stationRoutes[agent.id] || [stationPodPositions[agent.id]];
        const target = route[(stationWorldTick + index) % route.length];
        moveStationAgent(agent.id, target, true);
      });
    }
    function renderTasks(tasks) {
      lastBoardTasks = tasks;
      const stats = taskStats(tasks);
      const open = stats.total - stats.done;
      document.getElementById('taskCount').textContent = open + ' open · ' + stats.done + ' done';
      document.getElementById('taskSummary').textContent = stats.todo + ' todo · ' + stats.claimed + ' assigned · ' + stats.doing + ' active · ' + stats.done + ' done';
      document.querySelectorAll('[data-task-filter]').forEach(button => button.classList.toggle('active', button.dataset.taskFilter === taskFilter));
      const visible = tasks.filter(t => taskFilter === 'all' || (taskFilter === 'done' ? t.state === 'done' : t.state !== 'done'));
      const rank = {doing:0, claimed:1, todo:2, blocked:3, review:4, done:5};
      visible.sort((a,b) => (rank[a.state] ?? 9) - (rank[b.state] ?? 9) || a.id.localeCompare(b.id));
      document.getElementById('tasks').innerHTML = visible.map(t => {
        const stateType = t.state === 'done' ? 'ok' : (t.state === 'doing' ? 'warn' : (t.state === 'claimed' ? 'info' : ''));
        const line = '<span class="task-row-id">'+esc(t.id)+'</span><span class="task-title">'+esc(t.title)+'</span><span class="task-row-tags">'+badge(t.state, stateType)+badge(t.owner || 'unassigned')+'</span>';
        if (t.state === 'done') {
          return '<details class="task-row is-done"><summary>'+line+'</summary><div class="task-row-actions"><button class="small ghost" onclick="reopenTask(\''+esc(t.id)+'\')">Reopen</button></div></details>';
        }
        const owner = t.owner || '';
        let actions = '';
        if (t.state === 'todo') actions = '<button class="small ghost" onclick="assignTask(\''+esc(t.id)+'\')">Assign</button><button class="small" onclick="claimTask(\''+esc(t.id)+'\',\'\')">Assign & start</button>';
        if (t.state === 'claimed') actions = '<button class="small" onclick="claimTask(\''+esc(t.id)+'\',\''+esc(owner)+'\')">Start as '+esc(owner)+'</button><button class="small ghost" onclick="unassignTask(\''+esc(t.id)+'\')">Unassign</button>';
        if (t.state === 'doing') actions = '<button class="small" onclick="doneTask(\''+esc(t.id)+'\',\''+esc(owner)+'\')">Mark done</button>';
        return '<details class="task-row is-'+esc(t.state)+'"><summary>'+line+'</summary><div class="task-row-actions">'+actions+'</div></details>';
      }).join('') || '<div class="empty">'+(tasks.length ? 'No tasks match this filter.' : 'No tasks yet. Add one above, then assign it to an agent.')+'</div>';
    }
    function renderPlan(plan) {
      const el = document.getElementById('planStatus');
      if (!el) return;
      if (!plan) {
        el.className = 'plan-status muted';
        el.textContent = 'No planning run yet. Review mode lets the lead draft tasks, but a human must approve before distribution.';
        return;
      }
      const mode = plan.approval_mode === 'auto' ? badge('auto-distribute','warn') : badge('human review','ok');
      const status = badge(plan.status || 'pending', plan.status === 'approved' || plan.status === 'finalized' ? 'ok' : 'info');
      let action = '';
      if (plan.status === 'review' && plan.approval_mode === 'review') action = '<button class="small primary" onclick="approvePlan(\''+esc(plan.id)+'\')">Approve distribution</button>';
      else if (plan.status === 'approved') action = '<span class="hint">Approved. '+esc(plan.lead)+' will receive instructions to finalize and distribute.</span>';
      else if (plan.approval_mode === 'review') action = '<span class="hint">Lead drafts the final tasks, then runs <code>coact plan submit '+esc(plan.id)+'</code>.</span>';
      else action = '<span class="hint">Auto mode skips human approval. Agents still act only on their native CLI turns.</span>';
      el.className = 'plan-status';
      el.innerHTML = '<div class="row wrap"><strong>'+esc(plan.id)+'</strong>'+mode+status+badge('lead · '+esc(plan.lead))+'</div><p class="muted" style="margin-top:8px">'+esc(plan.brief || '')+'</p><div class="row wrap" style="margin-top:10px">'+action+'</div>';
    }
    function setTaskFilter(filter) {
      taskFilter = ['open','all','done'].includes(filter) ? filter : 'open';
      localStorage.setItem('coactTaskFilter', taskFilter);
      renderTasks(lastBoardTasks.slice());
    }
    function renderAgents(agents) {
      const live = agents.filter(a => a.live).length;
      document.getElementById('agentCount').textContent = live + ' live';
      document.getElementById('agents').innerHTML = agents.map(a => '<div class="agent-card"><div><div class="agent-name"><span class="badge '+(a.live?'ok':'bad')+'"><span class="dot"></span>'+(a.live?'live':'dead')+'</span><code>'+esc(a.id)+'</code></div><div class="agent-sub">'+esc(a.status || 'idle')+' · '+esc(a.enforcement || 'policy unknown')+'</div></div><div class="muted">'+esc(a.current_task || '-')+'<br>'+esc(a.beat || '-')+'</div></div>').join('') || '<div class="empty">No agents configured yet.</div>';
      const workAgents = document.getElementById('workAgents');
      if (workAgents) {
        workAgents.innerHTML = agents.map(a => '<div class="agent-card"><div><div class="agent-name"><span class="badge '+(a.live?'ok':'bad')+'">'+(a.live?'live':'offline')+'</span><code>'+esc(a.id)+'</code></div><div class="agent-sub">'+esc(a.status || 'idle')+'</div></div><div class="muted">'+esc(a.current_task || 'no task')+'</div></div>').join('') || '<div class="empty">No agents configured yet.</div>';
      }
      const terminalCount=document.getElementById('terminalCount');
      if(terminalCount)terminalCount.textContent=live+(live===1?' live agent':' live agents');
      syncAgentSelects(agents);
      if(terminalDetailsVisible())renderWorkTerminals(agents);
    }
    function syncAgentSelects(agents) {
      const configured = agents.map(agent => agent.id).filter(Boolean);
      const fill = (id, allowUnassigned) => {
        const select = document.getElementById(id);
        if (!select) return;
        const current = select.value;
        const options = configured.map(agent => '<option value="'+esc(agent)+'">'+esc(providerSpec(agent).label)+'</option>');
        if (allowUnassigned) options.unshift('<option value="">Unassigned</option>');
        select.innerHTML = options.join('');
        if ([...select.options].some(option => option.value === current)) select.value = current;
      };
      fill('msgTo', false);
      fill('taskOwner', true);
    }
    function renderWorkTerminals(agents) {
      const el = document.getElementById('workTerminals');
      if (!el) return;
      const focused = document.activeElement;
      const draft = focused && focused.id && focused.id.startsWith('agentInstruction-')
        ? {id: focused.id, value: focused.value, start: focused.selectionStart, end: focused.selectionEnd}
        : null;
      const ids = [...new Set(agents.map(agent => agent.id).concat(terminalMirrors.map(mirror => mirror.agent)))];
      const items = ids.map(id => {
        const spec = providerSpec(id);
        const agent = agentFor(agents, id) || {id:id, adapter:spec.binary, live:false};
        const mirror = mirrorFor(id);
        return {id, spec, agent, mirror};
      }).filter(item => item.agent.live || (item.mirror && item.mirror.exists));
      const count = document.getElementById('terminalCount');
      if (count) count.textContent = items.length + (items.length === 1 ? ' agent' : ' agents');
      if (!items.length) {
        el.innerHTML = '<div class="empty">No agent sessions found yet. Start agents from Setup or run <code>coact claude</code>, <code>coact codex</code>, or <code>coact antigravity</code> in native terminals.</div>';
        return;
      }
      if (!items.some(item => item.id === activeMirrorAgent)) {
        activeMirrorAgent = items[0].id;
        localStorage.setItem('coactActiveMirrorAgent', activeMirrorAgent);
      }
      const tabs = items.map(item => {
        const tabLabel = item.id === 'codex' ? 'Codex' : (item.id === 'claude' ? 'Claude' : item.spec.label);
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
      const transcript = hasFull ? esc(fullMirror.screen || fullMirror.tail || '') : '';
      const transcriptBlock = hasFull
        ? '<details open><summary>Current terminal screen</summary><pre class="work-terminal-output">'+transcript+'</pre></details>'
        : '<div class="hint" style="margin-top:12px">The current terminal screen is reconstructed from cursor movement and redraw events. Full-log rebuilding is loaded only when requested.</div>';
      const fullButton = mirror && mirror.exists
        ? (hasFull ? '<span class="badge ok">screen rebuilt</span>' : '<button class="small" data-full-transcript="'+esc(id)+'">Rebuild from full log</button>')
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
      const ids = [...new Set(agents.map(agent => agent.id).concat(launchCommands.map(command => command.agent)))];
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
      const mirrorText = mirror && (mirror.screen || mirror.tail)
        ? (mirror.screen || mirror.tail)
        : 'No terminal transcript yet. Start this agent through CoAct to mirror its raw session here.';
      const mirrorHTML = esc(mirrorText);
      const mirrorClass = mirror && (mirror.screen || mirror.tail) ? 'mirror-output' : 'mirror-output is-empty';
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
      const pane = '<div class="terminal-pane provider-'+esc(spec.brand)+' '+(agent.live?'is-live':'is-offline')+'"><div class="terminal-toolbar"><div class="terminal-title"><div class="provider-mark">'+esc(spec.mark)+'</div><div class="terminal-title-copy"><strong>'+esc(spec.label)+'</strong><span>'+esc(agent.adapter || spec.binary)+' · '+esc(status)+' · '+esc(cmd ? cmd.command : 'coact '+id)+'</span></div></div><div class="terminal-controls"><span class="badge '+(agent.live?'ok':'bad')+'">'+(agent.live?'live':'offline')+'</span><span class="badge '+(installed?'ok':'bad')+'">'+(installed?'installed':'missing')+'</span><div class="font-controls" aria-label="Terminal font size"><button type="button" data-font-delta="-1">A-</button><button type="button" data-font-reset>Reset</button><button type="button" data-font-delta="1">A+</button></div>'+startButton+copyButton+'</div></div><div class="terminal-meta-grid"><div class="terminal-meta"><span>Model</span><strong>'+esc(model)+'</strong></div><div class="terminal-meta"><span>Task</span><strong>'+esc(task)+'</strong></div><div class="terminal-meta"><span>Mirror</span><strong>'+esc(mirrorState)+'</strong></div><div class="terminal-meta"><span>Updated</span><strong>'+esc(mirrorTime)+'</strong></div></div><div class="mirror-meta"><span>Reconstructed terminal screen · output only</span><span>'+esc(terminalFontSize)+'px</span></div><pre class="'+mirrorClass+'" data-follow="true" style="font-size:'+esc(String(terminalFontSize))+'px">'+mirrorHTML+'</pre></div>';
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
    function selectedTaskOwner(preferred){ return preferred || document.getElementById('taskOwner').value || prompt('Choose an agent', 'codex'); }
    async function addTask(){ const description=document.getElementById('taskTitle').value; const prompt=document.getElementById('taskPrompt').value; const owner=document.getElementById('taskOwner').value; if(!description.trim()) return; await mutate(owner ? 'Task assigned to '+owner : 'Task added', () => api('/api/tasks',{method:'POST', body:JSON.stringify({description, prompt, owner})})); document.getElementById('taskTitle').value=''; document.getElementById('taskPrompt').value=''; }
    async function startPlan(){ const brief=document.getElementById('planBrief').value.trim(); if(!brief) return; const lead=document.getElementById('planLead').value; const approval_mode=document.getElementById('planApproval').value; if(approval_mode==='auto' && !confirm('Auto-distribute lets the lead create and assign tasks without asking you again. Continue?')) return; const participants=Array.from(document.querySelectorAll('[data-plan-participant]:checked')).map(input=>input.dataset.planParticipant); await mutate('Planning started with '+lead+' as lead', () => api('/api/plans',{method:'POST',body:JSON.stringify({brief,lead,approval_mode,participants})})); document.getElementById('planBrief').value=''; }
    async function approvePlan(id){ if(!confirm('Approve this plan and allow the lead to distribute its tasks?')) return; mutate('Plan approved', () => api('/api/plans/approve',{method:'POST',body:JSON.stringify({id})})); }
    async function assignTask(id){ const owner=selectedTaskOwner(''); if(!owner) return; mutate('Task assigned to '+owner, () => api('/api/tasks/'+id+'/assign',{method:'POST', body:JSON.stringify({owner})})); }
    async function claimTask(id,preferred){ const owner=selectedTaskOwner(preferred); if(!owner) return; mutate('Task started by '+owner, () => api('/api/tasks/'+id+'/claim',{method:'POST', body:JSON.stringify({owner})})); }
    async function doneTask(id,owner){ if(!owner) return; mutate('Task marked done', () => api('/api/tasks/'+id+'/done',{method:'POST', body:JSON.stringify({owner})})); }
    async function unassignTask(id){ mutate('Task returned to todo', () => api('/api/tasks/'+id+'/unassign',{method:'POST', body:'{}'})); }
    async function reopenTask(id){ mutate('Task reopened', () => api('/api/tasks/'+id+'/reopen',{method:'POST', body:'{}'})); }
    async function sendMessage(){ const text=document.getElementById('msgText').value; if(!text.trim()) return; const to=document.getElementById('msgTo').value; await mutate('Note sent to '+to, () => api('/api/messages',{method:'POST', body:JSON.stringify({from:document.getElementById('msgFrom').value,to,text})})); document.getElementById('msgText').value=''; }
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
    function terminalDetailsVisible(){
      const details=document.getElementById('terminals-card');
      return document.body.dataset.activePage==='work'&&!!details?.open;
    }
    async function maybeLoadTerminalMirrors(force){
      if(!terminalDetailsVisible())return;
      const now=Date.now();
      if(!force&&now-lastMirrorPoll<8000)return;
      lastMirrorPoll=now;
      await loadTerminalMirrors();
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
      maybeLoadTerminalMirrors(true).catch(() => {});
    }
    async function addProject(){
      const root = prompt('Project folder path?', document.getElementById('workspace').title || '');
      if (!root) return;
      await mutate('Project added', () => api('/api/projects',{method:'POST', body:JSON.stringify({root})}));
      loadLaunch().catch(e => showToast(e.message));
    }
    document.addEventListener('click', async ev => {
      const ambientToggle = ev.target.closest('[data-toggle-ambient]');
      if (ambientToggle) {
        setAmbientEnabled(!ambientEnabled);
        return;
      }
      const motion = ev.target.closest('[data-station-motion]');
      if (motion) {
        setStationMotion(!stationMotionEnabled);
        return;
      }
      const stationAgent = ev.target.closest('[data-station-agent]');
      if (stationAgent) {
        selectStationAgent(stationAgent.getAttribute('data-station-agent'));
        return;
      }
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
    document.addEventListener('toggle', ev => {
      if(ev.target.id==='terminals-card'&&ev.target.open)maybeLoadTerminalMirrors(true).catch(e=>showToast(e.message));
    }, true);
    document.addEventListener('keydown', ev => {
      const stationAgent = ev.target.closest && ev.target.closest('[data-station-agent]');
      if (stationAgent && (ev.key === 'Enter' || ev.key === ' ')) {
        ev.preventDefault();
        selectStationAgent(stationAgent.getAttribute('data-station-agent'));
      }
    });
    window.addEventListener('popstate', () => setPage(pageFromLocation(), false));
    setPage(pageFromLocation(), false);
    setAmbientEnabled(ambientEnabled);
    setStationMotion(stationMotionEnabled);
    loadLaunch().catch(e => showToast(e.message));
    refresh();
    setInterval(refresh, 2000);
  </script>
</body>
</html>`
