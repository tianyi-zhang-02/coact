package ui

const indexHTML = `<!doctype html>
<html lang="__COACT_LANG__">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>CoAct Control Center</title>
  <style>
    :root { color-scheme: light dark; --bg:#0b1020; --panel:#121a2f; --text:#e8eefc; --muted:#9aa7bd; --line:#25304a; --accent:#7c9cff; --ok:#42d392; --warn:#ffcb6b; --bad:#ff6b6b; }
    body { margin:0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background:var(--bg); color:var(--text); }
    header { padding:24px 28px; border-bottom:1px solid var(--line); display:flex; justify-content:space-between; gap:20px; align-items:center; }
    h1 { margin:0; font-size:22px; }
    h2 { margin:0 0 12px; font-size:16px; }
    main { padding:24px; display:grid; grid-template-columns: 1.1fr 1fr; gap:18px; }
    section { background:var(--panel); border:1px solid var(--line); border-radius:14px; padding:16px; box-shadow:0 16px 32px #0004; }
    button, input, textarea, select { font:inherit; border-radius:10px; border:1px solid var(--line); background:#0d1428; color:var(--text); padding:9px 10px; }
    button { cursor:pointer; background:#1a2b58; border-color:#32477a; }
    button:hover { background:#24386e; }
    textarea { width:100%; min-height:110px; box-sizing:border-box; resize:vertical; }
    input, select { width:100%; box-sizing:border-box; }
    table { width:100%; border-collapse:collapse; font-size:14px; }
    th, td { text-align:left; padding:8px 6px; border-bottom:1px solid var(--line); vertical-align:top; }
    th { color:var(--muted); font-weight:600; }
    code { background:#0a0f20; border:1px solid var(--line); border-radius:8px; padding:2px 6px; }
    .muted { color:var(--muted); }
    .grid2 { display:grid; grid-template-columns:1fr 1fr; gap:10px; }
    .row { display:flex; gap:8px; align-items:center; }
    .stack { display:flex; flex-direction:column; gap:10px; }
    .pill { display:inline-flex; border-radius:999px; padding:2px 8px; font-size:12px; border:1px solid var(--line); color:var(--muted); }
    .live { color:var(--ok); }
    .dead { color:var(--bad); }
    .full { grid-column:1 / -1; }
    .right { text-align:right; }
    .log { max-height:260px; overflow:auto; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size:12px; white-space:pre-wrap; }
    .error { color:var(--bad); }
    @media (max-width: 960px) { main { grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <header>
    <div>
      <h1>CoAct Control Center</h1>
      <div class="muted" id="workspace">Loading…</div>
    </div>
    <div class="right">
      <div id="version"></div>
      <div class="muted" id="updated"></div>
    </div>
  </header>
  <main>
    <section>
      <h2>Setup / Doctor</h2>
      <div class="stack">
        <div id="initState" class="muted">Checking project…</div>
        <button onclick="initProject()">Initialize this repo</button>
        <div class="muted">Run <code>coact doctor</code> in a terminal for full local checks.</div>
      </div>
    </section>

    <section>
      <h2>Launch Commands</h2>
      <div id="launch" class="stack muted">Loading commands…</div>
    </section>

    <section class="full">
      <h2>Project Brief</h2>
      <div class="stack">
        <textarea id="brief" placeholder="Describe the goal, constraints, and preferred agent split. This is saved to .coact/brief.md."></textarea>
        <div class="row"><button onclick="saveBrief()">Save brief</button><span class="muted">Human-controlled shared context for agents.</span></div>
      </div>
    </section>

    <section>
      <h2>Tasks</h2>
      <div class="stack">
        <div class="row"><input id="taskTitle" placeholder="New task title" /><button onclick="addTask()">Add</button></div>
        <table><thead><tr><th>ID</th><th>State</th><th>Owner</th><th>Task</th><th></th></tr></thead><tbody id="tasks"></tbody></table>
      </div>
    </section>

    <section>
      <h2>Agents</h2>
      <table><thead><tr><th>Agent</th><th>Status</th><th>Task</th><th>Beat</th></tr></thead><tbody id="agents"></tbody></table>
    </section>

    <section>
      <h2>Messages</h2>
      <div class="stack">
        <div class="grid2"><input id="msgTo" placeholder="to: claude/codex" /><input id="msgFrom" value="human" /></div>
        <textarea id="msgText" placeholder="Message to agent"></textarea>
        <button onclick="sendMessage()">Send message</button>
      </div>
    </section>

    <section>
      <h2>Locks</h2>
      <table><thead><tr><th>Path</th><th>Owner</th><th>TTL</th></tr></thead><tbody id="locks"></tbody></table>
    </section>

    <section>
      <h2>Versions <span class="pill">experimental</span></h2>
      <div id="versions" class="stack muted">Managed versions appear here after <code>coact update</code> (checksum-verified over HTTPS, not yet signed).</div>
    </section>

    <section>
      <h2>Activity Log</h2>
      <div id="log" class="log muted">No events yet.</div>
    </section>
  </main>

  <script>
    const TOKEN = "__COACT_TOKEN__";
    let lastBrief = null;
    async function api(path, opts) {
      opts = opts || {};
      const headers = Object.assign({'Content-Type':'application/json','X-Coact-Token':TOKEN}, opts.headers || {});
      const res = await fetch(path, Object.assign({}, opts, {headers}));
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || res.statusText);
      return data;
    }
    function esc(s) { return String(s || '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
    async function refresh() {
      try {
        const s = await api('/api/state');
        document.getElementById('workspace').textContent = s.workspace + (s.initialized ? '' : ' (not initialized)');
        document.getElementById('version').innerHTML = '<span class="pill">coact ' + esc(s.version) + '</span>';
        document.getElementById('updated').textContent = new Date().toLocaleTimeString();
        document.getElementById('initState').innerHTML = s.initialized ? '<span class="live">Initialized</span>' : '<span class="error">Not initialized</span>';
        if (lastBrief === null || document.activeElement.id !== 'brief') {
          document.getElementById('brief').value = s.brief || '';
          lastBrief = s.brief || '';
        }
        renderTasks(s.tasks || []);
        renderAgents(s.agents || []);
        renderLocks(s.locks || []);
        renderLog(s.log || []);
        renderVersions(s.versions || [], s.manifest || null);
      } catch (e) {
        document.getElementById('initState').innerHTML = '<span class="error">' + esc(e.message) + '</span>';
      }
    }
    function renderTasks(tasks) {
      document.getElementById('tasks').innerHTML = tasks.map(t => '<tr><td><code>'+esc(t.id)+'</code></td><td>'+esc(t.state)+'</td><td>'+esc(t.owner||'-')+'</td><td>'+esc(t.title)+'</td><td><div class="row"><button onclick="claimTask(\''+esc(t.id)+'\')">Claim</button><button onclick="doneTask(\''+esc(t.id)+'\')">Done</button></div></td></tr>').join('') || '<tr><td colspan="5" class="muted">No tasks.</td></tr>';
    }
    function renderAgents(agents) {
      document.getElementById('agents').innerHTML = agents.map(a => '<tr><td><code>'+esc(a.id)+'</code><br><span class="muted">'+esc(a.enforcement)+'</span></td><td class="'+(a.live?'live':'dead')+'">'+(a.live?'live':'dead')+'<br><span class="muted">'+esc(a.status||'')+'</span></td><td>'+esc(a.current_task||'-')+'</td><td>'+esc(a.beat||'-')+'</td></tr>').join('');
    }
    function renderLocks(locks) {
      document.getElementById('locks').innerHTML = locks.map(l => '<tr><td>'+esc(l.path)+'</td><td>'+esc(l.owner)+'</td><td>'+esc(l.ttl_seconds)+'s</td></tr>').join('') || '<tr><td colspan="3" class="muted">No active locks.</td></tr>';
    }
    function renderLog(log) {
      document.getElementById('log').textContent = log.map(r => [r.ts, r.agent, r.event, Object.keys(r).filter(k => !['ts','agent','event'].includes(k)).sort().map(k => k+'='+r[k]).join(' ')].join('  ')).join('\n') || 'No events yet.';
    }
    function renderVersions(versions, manifest) {
      const local = versions.length ? versions.map(v => '<div><code>'+esc(v.version)+'</code> '+(v.active?'<span class="pill">active</span> ':'')+'<span class="pill">'+esc(v.path)+'</span></div>').join('') : '<div class="muted">No managed versions yet. Run <code>coact update</code>.</div>';
      const supports = manifest && manifest.supports ? manifest.supports : {};
      const meta = manifest ? '<div><strong>'+esc(manifest.version)+'</strong> <span class="pill">'+esc(manifest.channel)+'</span> <span class="pill">'+esc(manifest.stability)+'</span>'+(manifest.recommended?' <span class="pill">recommended</span>':'')+'</div><div class="muted">'+esc(manifest.summary||'')+'</div><div class="muted">Supports: '+esc((supports.agents||[]).join(', '))+' · realtime: '+esc(supports.realtime||'')+' · autopilot: '+esc(supports.autopilot||'')+'</div><div>'+((manifest.notes||[]).map(n => '<div>• '+esc(n)+'</div>').join(''))+'</div>' : '';
      document.getElementById('versions').innerHTML = meta + '<hr style="border:0;border-top:1px solid var(--line);width:100%">' + local + '<div class="muted">Commands: <code>coact update --channel stable</code> · <code>coact versions</code> · <code>coact switch &lt;version&gt;</code></div>';
    }
    async function initProject(){ await api('/api/init',{method:'POST', body:'{}'}); refresh(); }
    async function saveBrief(){ await api('/api/brief',{method:'POST', body:JSON.stringify({text:document.getElementById('brief').value})}); refresh(); }
    async function addTask(){ const title=document.getElementById('taskTitle').value; if(!title.trim()) return; await api('/api/tasks',{method:'POST', body:JSON.stringify({title})}); document.getElementById('taskTitle').value=''; refresh(); }
    async function claimTask(id){ const owner=prompt('Owner agent?', 'claude'); if(!owner) return; await api('/api/tasks/'+id+'/claim',{method:'POST', body:JSON.stringify({owner})}); refresh(); }
    async function doneTask(id){ const owner=prompt('Owner agent?', 'human'); if(!owner) return; await api('/api/tasks/'+id+'/done',{method:'POST', body:JSON.stringify({owner})}); refresh(); }
    async function sendMessage(){ await api('/api/messages',{method:'POST', body:JSON.stringify({from:document.getElementById('msgFrom').value,to:document.getElementById('msgTo').value,text:document.getElementById('msgText').value})}); document.getElementById('msgText').value=''; refresh(); }
    async function loadLaunch(){ const d = await api('/api/launch-commands'); document.getElementById('launch').innerHTML = d.commands.map(c => '<div><strong>'+esc(c.agent)+'</strong><br><code>'+esc(c.command)+'</code></div>').join(''); }
    loadLaunch(); refresh(); setInterval(refresh, 1500);
  </script>
</body>
</html>`
