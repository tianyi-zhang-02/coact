(() => {
  'use strict';

  const WIDTH = 768;
  const HEIGHT = 512;
  const MOVING_FRAME_INTERVAL = 16;
  const IDLE_FRAME_INTERVAL = 50;
  const CHROMIUM_MOVING_FRAME_INTERVAL = 33;
  const CHROMIUM_PRESSURE_FRAME_INTERVAL = 40;
  const CHROMIUM_IDLE_FRAME_INTERVAL = 75;
  const LITE_MOVING_FRAME_INTERVAL = 40;
  const LITE_IDLE_FRAME_INTERVAL = 100;
  const REDUCED_MOTION_FRAME_INTERVAL = 100;
  const BACKGROUND_FRAME_INTERVAL = 50;
  const CHROMIUM_BACKGROUND_FRAME_INTERVAL = 180;
  const LITE_BACKGROUND_FRAME_INTERVAL = 400;
  const SLOW_BACKGROUND_FRAME_INTERVAL = 220;
  const REDUCED_MOTION_BACKGROUND_INTERVAL = 180;
  const IS_CHROMIUM = /(?:Chrome|Chromium|CriOS|Edg)\//.test(navigator.userAgent || '') && !/(?:OPR|Opera)\//.test(navigator.userAgent || '');
  const AIRLOCK = {x:384,y:154};
  const TABLE_SEATS = [{x:290,y:250},{x:384,y:184},{x:478,y:250},{x:384,y:330}];
  const AGENT_SLOTS = [
    {x:215,y:205,room:'Systems lab'}, {x:550,y:205,room:'Navigation lab'},
    {x:190,y:356,room:'Engineering bay'}, {x:565,y:365,room:'Expansion lounge'},
    {x:330,y:165,room:'Bridge port'}, {x:438,y:165,room:'Bridge starboard'}
  ];
  const WAYPOINTS = {
    airlock:{x:384,y:154}, north:{x:384,y:184}, northWest:{x:330,y:205}, northEast:{x:438,y:205},
    west:{x:290,y:250}, east:{x:478,y:250}, southWest:{x:320,y:315}, southEast:{x:448,y:315}, south:{x:384,y:330}, bottom:{x:384,y:394},
    codexEntry:{x:275,y:205}, codexDesk:{x:215,y:205}, codexWindow:{x:174,y:174},
    claudeEntry:{x:493,y:205}, claudeDesk:{x:550,y:205}, claudeWindow:{x:594,y:174},
    antigravityEntry:{x:288,y:350}, antigravityDesk:{x:190,y:356}, antigravityWindow:{x:150,y:320},
    loungeEntry:{x:480,y:350}, lounge:{x:565,y:365}, loungeWindow:{x:620,y:320},
    bridgePort:{x:330,y:165}, bridgeStarboard:{x:438,y:165}
  };
  const ROUTES = [
    ['airlock','north'],['north','northWest'],['north','northEast'],['north','bridgePort'],['north','bridgeStarboard'],
    ['northWest','west'],['northEast','east'],['west','southWest'],['east','southEast'],['southWest','south'],['southEast','south'],['south','bottom'],
    ['northWest','codexEntry'],['codexEntry','codexDesk'],['codexDesk','codexWindow'],
    ['northEast','claudeEntry'],['claudeEntry','claudeDesk'],['claudeDesk','claudeWindow'],
    ['southWest','antigravityEntry'],['antigravityEntry','antigravityDesk'],['antigravityDesk','antigravityWindow'],
    ['southEast','loungeEntry'],['loungeEntry','lounge'],['lounge','loungeWindow']
  ];
  const BASE_CREW = {
    codex:{name:'Orbit · Codex',short:'CODEX',row:0,slot:0,color:'#42c9ff'},
    claude:{name:'Nova · Claude',short:'CLAUDE',row:1,slot:1,color:'#ff9b3f'},
    antigravity:{name:'Comet · Antigravity',short:'AGY',row:2,slot:2,color:'#a886ff'}
  };
  const EXTRA_COLORS = ['#61e6b0','#ff70ae','#7692ff','#ffd35d','#49d7e8','#cc7dff'];
  const MODES = ['ORBIT','OCEAN','ECODOME','WASTELAND'];
  const SCENES = [
    {asset:'station-orbit.png',crew:'crew-atlas-v2.png',accent:'#4cd9ff',soft:'#a8f3ff',label:'DEEP ORBIT'},
    {asset:'station-ocean.png',crew:'crew-atlas-ocean.png',accent:'#27dbea',soft:'#b9fff4',label:'SHALLOW SEA LAB'},
    {asset:'station-ecodome.png',crew:'crew-atlas-ecodome.png',accent:'#71dc72',soft:'#e1ffc4',label:'WATERFALL ECODOME'},
    {asset:'station-wasteland.png',crew:'crew-atlas-wasteland.png',accent:'#f29a45',soft:'#ffe0a3',label:'WASTELAND OUTPOST'}
  ];
  const SERVICE_THEMES = [
    {},
    {sentinel:['Clawdia','Crab lock guardian'],relay:['Finley','Fish message courier'],mnemo:['Juno','Jellyfish memory keeper'],ledger:['Shelly','Turtle audit archivist']},
    {sentinel:['Pip','Beetle lock guardian'],relay:['Zip','Hummingbird courier'],mnemo:['Sprig','Seedling memory keeper'],ledger:['Professor Hoot','Owl audit archivist']},
    {sentinel:['Tin Deputy','Frontier lock guardian'],relay:['Dusty','Roadrunner courier'],mnemo:['Glow','Lantern memory keeper'],ledger:['Tumble','Armadillo audit archivist']}
  ];
  const FRAME = {
    down:0,up:1,left:2,right:3,walkDownA:4,walkDownB:5,walkUpA:6,walkUpB:7,
    runDown:8,typeA:9,typeB:10,sit:11,wait:12,sleep:13,plan:14,celebrate:15
  };
  const CELL = {w:96,h:176};
  const clamp = (value,min,max) => Math.max(min,Math.min(max,value));
  const humanEvent = event => ({
    'task.add':'Task created','task.claim':'Task claimed','task.finish':'Task completed','task.done':'Task completed',
    'msg.send':'Message routed','lock.acquire':'File protected','lock.release':'File released','agent.launch':'Crew launched',
    'session.start':'Crew arrived','session.stop':'Crew departed','plan.start':'Planning started','plan.finish':'Plan finalized'
  }[event] || String(event || 'Station updated').replaceAll('.',' '));

  function loadImage(src) {
    return new Promise((resolve,reject) => {
      const image = new Image();
      image.onload = () => resolve(image);
      image.onerror = () => reject(new Error('unable to load '+src));
      image.src = src;
    });
  }

  class OrbitalWorld {
    constructor(root) {
      this.root = root;
      this.canvas = root.querySelector('#coactPixelWorld');
      this.ctx = this.canvas.getContext('2d',{alpha:true,desynchronized:true});
      this.ctx.imageSmoothingEnabled = false;
      this.backgroundCanvas = root.querySelector('#coactPixelBackground')||document.createElement('canvas');
      this.backgroundCanvas.width = WIDTH;
      this.backgroundCanvas.height = HEIGHT;
      this.backgroundCtx = this.backgroundCanvas.getContext('2d',{alpha:false,desynchronized:true});
      this.backgroundCtx.imageSmoothingEnabled = false;
      this.backgroundReady = false;
      this.lastBackgroundDrawTime = -Infinity;
      this.state = {agents:[],tasks:[],locks:[],log:[]};
      this.crew = new Map();
      this.specs = new Map(Object.entries(BASE_CREW));
      this.routeGraph = this.buildRouteGraph();
      this.modeIndex = Number(localStorage.getItem('coactWorldTheme') || 0);
      if (!Number.isFinite(this.modeIndex)) this.modeIndex = 0;
      this.modeIndex = ((this.modeIndex % MODES.length) + MODES.length) % MODES.length;
      this.chromium = IS_CHROMIUM;
      this.root.classList.toggle('is-chromium',this.chromium);
      this.reducedMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;
      this.qualityPreference = ['auto','smooth','lite'].includes(localStorage.getItem('coactWorldQuality'))?localStorage.getItem('coactWorldQuality'):'auto';
      this.paused = false;
      this.selected = '';
      this.hover = null;
      this.worldTime = 0;
      this.previousModeIndex = this.modeIndex;
      this.sceneChangedAt = -2000;
      this.lastTime = performance.now();
      this.lastAmbient = 0;
      this.frameCount = 0;
      this.lastWatchdogFrame = 0;
      this.lastDrawTime = 0;
      this.averageDrawCost = 0;
      this.averageFrameGap = 16.7;
      this.slowDrawCount = 0;
      this.performancePressure = 0;
      this.lowPowerBackground = this.qualityPreference==='lite'||(this.chromium&&this.qualityPreference==='auto');
      this.assetsReady = false;
      this.assetsError = '';
      this.hasSyncedState = false;
      this.lastFinishKey = '';
      this.assistSuggestion = null;
      this.dismissedAssists = new Set();
      const storedAssistDelay = Number(localStorage.getItem('coactAssistAfterMs') || 90000);
      this.assistAfterMs = Number.isFinite(storedAssistDelay) ? clamp(storedAssistDelay,15000,600000) : 90000;
      this.services = this.createServices();
      this.stars = Array.from({length:90},(_,index)=>({
        x:(index*83+17)%WIDTH,y:(index*47+11)%170,size:index%17===0?2:1,depth:.15+(index%6)*.08,phase:index%9
      }));
      this.bind();
      this.syncControls();
      this.loadAssets();
      this.scheduleFrame();
      this.watchdog = window.setInterval(()=>{
        if (document.hidden || this.paused) return;
        if (this.frameCount===this.lastWatchdogFrame) {
          const motionDelta=this.reducedMotion?320:500;
          this.worldTime+=motionDelta;this.step(motionDelta);this.safeDraw();
        }
        this.lastWatchdogFrame=this.frameCount;
      },500);
    }

    scheduleFrame(delay=0) {
      if(delay>0)window.setTimeout(()=>requestAnimationFrame(time=>this.frame(time)),delay);
      else requestAnimationFrame(time=>this.frame(time));
    }

    async loadAssets() {
      try {
        const assets = await Promise.all([
          ...SCENES.map(scene=>loadImage('/world/assets/'+scene.asset+'?v=7')),
          ...SCENES.map(scene=>loadImage('/world/assets/'+scene.crew+'?v=13'))
        ]);
        this.stations = assets.slice(0,SCENES.length);
        this.atlases = assets.slice(SCENES.length);
        this.assetsReady = true;
        this.backgroundReady = false;
        this.safeDraw();
      } catch (error) {
        this.assetsError = error instanceof Error ? error.message : String(error);
        this.safeDraw();
      }
    }

    bind() {
      this.canvas.addEventListener('pointermove',event=>this.onPointer(event));
      this.canvas.addEventListener('pointerleave',()=>{this.hover=null;this.hideTooltip();});
      this.canvas.addEventListener('click',event=>this.onClick(event));
      this.root.querySelector('[data-world-pause]')?.addEventListener('click',()=>{
        this.paused=!this.paused;this.syncControls();this.safeDraw();
      });
      this.root.querySelector('[data-world-quality]')?.addEventListener('click',()=>{
        const modes=['auto','smooth','lite'];this.qualityPreference=modes[(modes.indexOf(this.qualityPreference)+1)%modes.length];localStorage.setItem('coactWorldQuality',this.qualityPreference);this.performancePressure=0;this.setLowPowerBackground(this.qualityPreference==='lite'||(this.chromium&&this.qualityPreference==='auto'));this.syncControls();this.backgroundReady=false;this.safeDraw();
      });
      this.root.querySelector('[data-world-theme]')?.addEventListener('click',()=>{
        this.previousModeIndex=this.modeIndex;this.modeIndex=(this.modeIndex+1)%MODES.length;this.sceneChangedAt=this.worldTime;this.backgroundReady=false;localStorage.setItem('coactWorldTheme',String(this.modeIndex));this.syncControls();if(this.selected)this.renderSelected();this.safeDraw();
      });
      this.root.querySelector('[data-world-help]')?.addEventListener('click',()=>this.root.querySelector('.pixel-help')?.classList.toggle('show'));
      this.root.querySelector('[data-world-close-help]')?.addEventListener('click',()=>this.root.querySelector('.pixel-help')?.classList.remove('show'));
      this.root.querySelector('[data-world-close-agent]')?.addEventListener('click',()=>this.select(''));
      this.root.querySelector('[data-world-assist-offer]')?.addEventListener('click',()=>this.requestAssist('offer'));
      this.root.querySelector('[data-world-assist-handoff]')?.addEventListener('click',()=>this.requestAssist('handoff'));
      window.addEventListener('keydown',event=>{
        if (event.target && /input|textarea|select/i.test(event.target.tagName)) return;
        if (event.key==='p') this.root.querySelector('[data-world-pause]')?.click();
        if (event.key==='t') this.root.querySelector('[data-world-theme]')?.click();
        if (event.key==='?') this.root.querySelector('[data-world-help]')?.click();
        if (event.key==='Escape') {this.select('');this.root.querySelector('.pixel-help')?.classList.remove('show');}
      });
    }

    syncControls() {
      this.root.classList.toggle('is-paused',this.paused);
      const pause=this.root.querySelector('[data-world-pause]');
      if (pause) {pause.textContent=this.paused?'Resume':'Pause';pause.setAttribute('aria-pressed',String(!this.paused));}
      const quality=this.root.querySelector('[data-world-quality]');
      if(quality){const adaptive=this.qualityPreference==='auto'&&this.lowPowerBackground,chromeAuto=this.chromium&&this.qualityPreference==='auto';quality.textContent=this.qualityPreference==='smooth'?'MAX':this.qualityPreference==='lite'?'LITE':chromeAuto?'AUTO·30':adaptive?'AUTO·LITE':'AUTO';quality.title=chromeAuto?'Chrome optimized: stable 30 FPS characters with a reduced background refresh rate':this.qualityPreference==='auto'?'Automatically protects character frame rate':this.qualityPreference==='smooth'?'Maximum background detail':'Reduced background detail';}
      const mode=this.root.querySelector('[data-world-theme]');
      if (mode) mode.textContent=MODES[this.modeIndex];
      const titles=['CoAct Orbital Crew','CoAct Shallow Sea Lab','CoAct Waterfall Habitat','CoAct Wasteland Outpost'];
      const subtitles=['Agents coordinate across a shared deep-space workspace.','Fish, divers, and agents move through one protected sea lab.','Agents collaborate inside a living waterfall research habitat.','Crew, companions, and audit services share one frontier outpost.'];
      this.text('worldTitle',titles[this.modeIndex]);this.text('worldSubtitle',subtitles[this.modeIndex]);
      document.body.dataset.coactTheme=MODES[this.modeIndex].toLowerCase();
    }

    createServices() {
      return [
        {id:'sentinel',name:'Sentinel',role:'Lock guardian',color:'#ffd05d',px:320,py:315,path:[],wait:0,status:'No active locks'},
        {id:'relay',name:'Relay',role:'Message courier',color:'#65d9ff',px:448,py:315,path:[],wait:0,status:'Inbox standing by'},
        {id:'mnemo',name:'Mnemo',role:'Memory synchronizer',color:'#aa8dff',px:290,py:250,path:[],wait:0,status:'Shared context ready'},
        {id:'ledger',name:'Ledger',role:'Audit recorder',color:'#65e39e',px:478,py:250,path:[],wait:0,status:'Journal ready'}
      ];
    }

    buildRouteGraph() {
      const graph=new Map(Object.keys(WAYPOINTS).map(id=>[id,[]]));
      ROUTES.forEach(([left,right])=>{graph.get(left).push(right);graph.get(right).push(left);});
      return graph;
    }

    specFor(id) {
      if (this.specs.has(id)) return this.specs.get(id);
      const index=this.specs.size;
      const spec={name:'Crew · '+id,short:String(id).toUpperCase().slice(0,9),row:index%3,color:EXTRA_COLORS[index%EXTRA_COLORS.length]};
      this.specs.set(id,spec);
      return spec;
    }

    update(next) {
      this.state={agents:next.agents||[],tasks:next.tasks||[],locks:next.locks||[],log:next.log||[]};
      this.state.agents.forEach(agent=>this.specFor(agent.id));
      const live=this.state.agents.filter(agent=>agent.live).length;
      const active=this.state.tasks.filter(task=>task.state==='doing').length;
      this.text('worldLive',live);this.text('worldTasks',active);this.text('worldLocks',this.state.locks.length);
      this.text('worldMode',live>1?'MULTI-AGENT ORBIT':live===1?'SOLO ORBIT':'STATION IDLE');
      const latest=this.state.log[this.state.log.length-1];
      this.text('worldEventTitle',latest?humanEvent(latest.event):(live?'Crew connected':'Waiting for crew'));
      this.text('worldEventText',latest?this.describeEvent(latest):(live?'CoAct is receiving live heartbeats.':'Start agents with coact codex, coact claude, or coact antigravity.'));
      [...this.specs.keys()].forEach(id=>this.reduceCrew(id,this.state.agents.find(agent=>agent.id===id)));
      this.captureCompletion();
      this.updateAssistSuggestion();
      this.updateServices();
      if (this.selected) this.renderSelected();
      if (this.paused) this.safeDraw();
    }

    text(id,value) {const element=this.root.querySelector('#'+id);if(element)element.textContent=String(value);}
    describeEvent(event) {return Object.keys(event).filter(name=>!['ts','event'].includes(name)).map(name=>name+' '+event[name]).join(' · ')||'Shared project state changed.';}

    reduceCrew(id,agent) {
      let actor=this.crew.get(id);
      const live=!!agent?.live;
      if (!actor) {
        actor={id,px:AIRLOCK.x,py:AIRLOCK.y,path:[],live:false,mode:'hidden',direction:'down',step:0,walkDistance:0,agent:null,ambientTarget:null,ambientUntil:0,celebrateStarted:0,celebrateUntil:0,waitingSince:0};
        this.crew.set(id,actor);
      }
      const wasLive=actor.live;actor.agent=agent||null;actor.live=live;
      if (live&&!wasLive) {
        actor.px=AIRLOCK.x;actor.py=AIRLOCK.y;actor.mode='arriving';this.route(actor,this.targetFor(actor));
      } else if (!live&&wasLive) {
        actor.mode='departing';this.route(actor,AIRLOCK);
      } else if (live) {
        const target=this.targetFor(actor),goal=actor.path.at(-1)||{x:actor.px,y:actor.py};
        if (goal.x!==target.x||goal.y!==target.y) this.route(actor,target);
        if (!actor.path.length) actor.mode=this.resolvedMode(actor);
      }
    }

    modeFor(agent) {
      const status=String(agent?.status||'working').toLowerCase();
      if (/wait|permission|input|blocked/.test(status)) return 'waiting';
      if (/idle|ready|sleep/.test(status)) return 'idle';
      if (/plan|review|sync/.test(status)) return 'planning';
      if (/done|complete|finish/.test(status)) return 'celebrating';
      return 'working';
    }

    resolvedMode(actor) {
      if(actor.celebrateUntil>this.worldTime)return'celebrating';
      return this.modeFor(actor.agent);
    }

    taskFor(actor) {
      return this.state.tasks.find(task=>task.id===actor.agent?.current_task)||null;
    }

    activityFor(actor) {
      const task=this.taskFor(actor),text=[task?.title,task?.description,task?.notes,actor.agent?.status].filter(Boolean).join(' ').toLowerCase();
      if(/test|verify|validation|qa|benchmark|测试|验证|验收/.test(text))return'testing';
      if(/security|audit|threat|policy|permission|安全|审计|权限/.test(text))return'security';
      if(/review|inspect|quality|critique|评审|检查|质量/.test(text))return'reviewing';
      if(/debug|trace|repro|diagnos|排查|调试|定位/.test(text))return'debugging';
      if(/document|docs|readme|copy|guide|write|文档|说明|写作/.test(text))return'writing';
      if(/analysis|metric|report|data|profil|分析|指标|报告|数据/.test(text))return'analyzing';
      if(/design|plan|architect|proposal|ux|ui|设计|规划|架构/.test(text))return actor.id==='codex'?'prototyping':actor.id==='antigravity'?'researching':'designing';
      if(/research|read|explore|investigate|learn|调研|研究|阅读/.test(text))return'researching';
      if(/message|handoff|sync|coordinate|release|publish|沟通|同步|发布/.test(text))return'coordinating';
      if(/build|code|implement|fix|refactor|develop|开发|实现|修复/.test(text))return'coding';
      if(actor.id==='claude')return'designing';
      if(actor.id==='antigravity')return'researching';
      return'coding';
    }

    actorPhase(actor) {
      return [...String(actor.id)].reduce((total,character)=>total+character.charCodeAt(0),0)%7;
    }

    captureCompletion() {
      const finished=[...this.state.log].reverse().find(event=>event.event==='task.finish'||event.event==='task.done');
      const key=finished?[finished.ts,finished.agent,finished.event,finished.id].join('|'):'';
      if(!this.hasSyncedState){this.hasSyncedState=true;this.lastFinishKey=key;return;}
      if(!finished||key===this.lastFinishKey)return;
      this.lastFinishKey=key;
      const age=Date.now()-Date.parse(finished.ts||0);
      if(!Number.isFinite(age)||age>20000)return;
      const actor=this.crew.get(finished.agent);
      if(!actor||!actor.live)return;
      actor.celebrateStarted=this.worldTime;
      actor.celebrateUntil=this.worldTime+4800;
      if(!actor.path.length)actor.mode='celebrating';
    }

    latestActorEvent(id) {
      return [...this.state.log].reverse().find(event=>event.agent===id)||null;
    }

    activeTaskFor(id) {
      return this.state.tasks.find(task=>task.owner===id&&task.state==='doing')||null;
    }

    waitingContext(actor,now) {
      const status=String(actor.agent?.status||'').toLowerCase();
      if(/wait|permission|input|blocked/.test(status)){
        if(!actor.waitingSince)actor.waitingSince=now;
        return{since:actor.waitingSince,reason:status||'waiting'};
      }
      actor.waitingSince=0;
      const latest=this.latestActorEvent(actor.id);
      if(!latest||!['msg.send','lock.denied'].includes(latest.event))return null;
      const since=Date.parse(latest.ts||0);
      if(!Number.isFinite(since)||now-since>30*60*1000)return null;
      const peer=latest.event==='msg.send'?latest.to:latest.held_by;
      const peerProgress=peer&&this.state.log.some(event=>event.agent===peer&&Date.parse(event.ts||0)>since);
      if(peerProgress)return null;
      const reason=latest.event==='msg.send'?'waiting after messaging '+(latest.to||'a teammate'):'blocked by '+(latest.held_by||'a protected path');
      return{since,reason};
    }

    updateAssistSuggestion() {
      const now=Date.now(),live=[...this.crew.values()].filter(actor=>actor.live&&actor.mode!=='hidden');
      this.assistSuggestion=null;
      for(const waiting of live){
        const context=this.waitingContext(waiting,now);
        if(!context||now-context.since<this.assistAfterMs)continue;
        const peers=live.filter(actor=>actor.id!==waiting.id);
        const helper=peers.find(actor=>!this.activeTaskFor(actor.id))||peers[0];
        if(!helper)continue;
        const key=[waiting.id,helper.id,context.since].join(':');
        if(this.dismissedAssists.has(key))continue;
        this.assistSuggestion={key,waiting:waiting.id,helper:helper.id,reason:context.reason,since:context.since};
        break;
      }
    }

    requestAssist(action) {
      const suggestion=this.assistSuggestion;
      if(!suggestion)return;
      window.dispatchEvent(new CustomEvent('coact:assist',{detail:{action,waiting:suggestion.waiting,helper:suggestion.helper}}));
    }

    assistHandled() {
      if(this.assistSuggestion)this.dismissedAssists.add(this.assistSuggestion.key);
      this.assistSuggestion=null;
      if(this.selected)this.renderSelected();
      this.safeDraw();
    }

    targetFor(actor) {
      if(actor.ambientTarget&&this.worldTime<actor.ambientUntil)return actor.ambientTarget;
      actor.ambientTarget=null;
      const latest=this.state.log.at(-1)||{};
      const recent=Date.now()-Date.parse(latest.ts||0)<90000;
      const collaborate=this.state.agents.filter(agent=>agent.live).length>1&&recent&&/plan|review|proposal|handoff|msg/.test(String(latest.event||''));
      const slot=this.specFor(actor.id).slot??([...this.specs.keys()].indexOf(actor.id)%AGENT_SLOTS.length);
      if (collaborate) return TABLE_SEATS[slot%TABLE_SEATS.length];
      return AGENT_SLOTS[slot%AGENT_SLOTS.length];
    }

    route(entity,target) {
      entity.path=this.findPath({x:entity.px,y:entity.py},target);
      entity.step=0;entity.segmentKey='';
      if(entity.path.length&&entity.mode!=='departing')entity.mode='walking';
    }

    findPath(start,goal) {
      const nearest=point=>Object.keys(WAYPOINTS).reduce((best,id)=>{
        const node=WAYPOINTS[id],distance=Math.hypot(point.x-node.x,point.y-node.y);
        return !best||distance<best.distance?{id,distance}:best;
      },null).id;
      const startID=nearest(start),goalID=nearest(goal),open=[{id:startID,g:0,f:0}],came=new Map(),scores=new Map([[startID,0]]);
      while(open.length) {
        open.sort((a,b)=>a.f-b.f);const current=open.shift();
        if(current.id===goalID){const ids=[];let cursor=current.id;while(cursor!==startID){ids.unshift(cursor);cursor=came.get(cursor);}const path=ids.map(id=>({...WAYPOINTS[id]}));const last=path.at(-1)||start;if(Math.hypot(last.x-goal.x,last.y-goal.y)>2)path.push({...goal});return path;}
        for(const id of this.routeGraph.get(current.id)||[]){const from=WAYPOINTS[current.id],to=WAYPOINTS[id],next=current.g+Math.hypot(to.x-from.x,to.y-from.y);if(next>=(scores.get(id)??Infinity))continue;came.set(id,current.id);scores.set(id,next);open.push({id,g:next,f:next+Math.hypot(goal.x-to.x,goal.y-to.y)});}
      }
      return [];
    }

    updateServices() {
      const latest=this.state.log.at(-1)||{},live=this.state.agents.filter(agent=>agent.live);
      const sentinel=this.services[0];sentinel.status=this.state.locks.length?this.state.locks.length+' protected path'+(this.state.locks.length===1?'':'s'):'No active locks';
      const owner=this.state.locks[0]?.owner,ownerIndex=live.findIndex(agent=>agent.id===owner);if(ownerIndex>=0)this.routeService(sentinel,AGENT_SLOTS[ownerIndex%AGENT_SLOTS.length]);
      const relay=this.services[1];relay.status=latest.event==='msg.send'?'Delivering '+(latest.to||'crew'):'Inbox standing by';
      if(latest.event==='msg.send'){const index=live.findIndex(agent=>agent.id===latest.to);if(index>=0)this.routeService(relay,AGENT_SLOTS[index%AGENT_SLOTS.length]);}
      this.services[2].status=live.length?live.length+' live context stream'+(live.length===1?'':'s'):'Shared context ready';
      this.services[3].status=this.state.log.length+' audited event'+(this.state.log.length===1?'':'s');
    }

    routeService(bot,target) {
      const goal=bot.path.at(-1)||{x:bot.px,y:bot.py};if(goal.x===target.x&&goal.y===target.y)return;
      bot.path=this.findPath({x:bot.px,y:bot.py},target);
    }

    frame(time) {
      const rawDelta=time-this.lastTime,delta=Math.min(50,rawDelta);this.lastTime=time;
      const page=this.root.closest('.page-view');
      if(document.hidden||(page&&!page.classList.contains('active'))){this.frameCount++;this.scheduleFrame(180);return;}
      if(this.paused){this.frameCount++;this.scheduleFrame(180);return;}
      if(rawDelta>0&&rawDelta<120){this.averageFrameGap=this.averageFrameGap*.94+rawDelta*.06;this.updatePerformancePressure();}
      if(!this.paused){const motionDelta=this.reducedMotion?delta*.7:delta;this.worldTime+=motionDelta;this.step(motionDelta);}
      this.frameCount++;
      const entitiesMoving=[...this.crew.values()].some(actor=>actor.live&&actor.path.length)||this.services.some(bot=>bot.path.length);
      const frameInterval=this.foregroundInterval(entitiesMoving);
      if(time-this.lastDrawTime>=frameInterval){this.lastDrawTime=time;this.safeDraw();}
      this.scheduleFrame();
    }

    foregroundInterval(entitiesMoving) {
      if(this.reducedMotion)return REDUCED_MOTION_FRAME_INTERVAL;
      if(this.qualityPreference==='lite')return entitiesMoving?LITE_MOVING_FRAME_INTERVAL:LITE_IDLE_FRAME_INTERVAL;
      if(this.chromium&&this.qualityPreference==='auto')return entitiesMoving?(this.performancePressure>=18?CHROMIUM_PRESSURE_FRAME_INTERVAL:CHROMIUM_MOVING_FRAME_INTERVAL):CHROMIUM_IDLE_FRAME_INTERVAL;
      return entitiesMoving?MOVING_FRAME_INTERVAL:IDLE_FRAME_INTERVAL;
    }

    backgroundInterval() {
      if(this.reducedMotion)return REDUCED_MOTION_BACKGROUND_INTERVAL;
      if(this.qualityPreference==='lite')return LITE_BACKGROUND_FRAME_INTERVAL;
      if(this.chromium&&this.qualityPreference==='auto')return CHROMIUM_BACKGROUND_FRAME_INTERVAL;
      return this.lowPowerBackground?SLOW_BACKGROUND_FRAME_INTERVAL:BACKGROUND_FRAME_INTERVAL;
    }

    step(delta) {
      this.crew.forEach(actor=>{
        if(!actor.path.length&&actor.mode==='celebrating'&&actor.celebrateUntil<=this.worldTime)actor.mode=this.modeFor(actor.agent);
        this.stepEntity(actor,delta,.072,true);
      });
      this.services.forEach(bot=>this.stepService(bot,delta));
      if(this.worldTime-this.lastAmbient>6200){this.lastAmbient=this.worldTime;this.ambientRoutes();}
    }

    stepEntity(entity,delta,speedFactor,isCrew) {
      if(!entity.path.length)return;
      let remaining=speedFactor*delta;
      while(entity.path.length&&remaining>0){
        const next=entity.path[0],dx=next.x-entity.px,dy=next.y-entity.py,distance=Math.hypot(dx,dy);
        const segmentKey=next.x+':'+next.y;
        if(entity.segmentKey!==segmentKey){entity.segmentKey=segmentKey;if(Math.abs(dx)>Math.abs(dy))entity.direction=dx<0?'left':'right';else if(distance>0)entity.direction=dy<0?'up':'down';}
        if(distance<=remaining){entity.px=next.x;entity.py=next.y;entity.path.shift();entity.segmentKey='';entity.step++;entity.walkDistance=(entity.walkDistance||0)+distance;remaining-=distance;}
        else{entity.px+=dx/distance*remaining;entity.py+=dy/distance*remaining;entity.walkDistance=(entity.walkDistance||0)+remaining;remaining=0;}
      }
      if(!entity.path.length&&isCrew){if(entity.mode==='departing'){entity.mode='hidden';entity.live=false;}else entity.mode=this.resolvedMode(entity);}
    }

    stepService(bot,delta) {
      if(bot.wait>0){bot.wait-=delta;return;}
      if(!bot.path.length){const routes=[[WAYPOINTS.southWest,WAYPOINTS.northWest,WAYPOINTS.southEast],[WAYPOINTS.southEast,WAYPOINTS.west,WAYPOINTS.east],[WAYPOINTS.west,WAYPOINTS.east,WAYPOINTS.north],[WAYPOINTS.east,WAYPOINTS.lounge,WAYPOINTS.south]];const route=routes[this.services.indexOf(bot)],target=route[(Math.floor(this.worldTime/6000)+this.services.indexOf(bot))%route.length];this.routeService(bot,target);if(!bot.path.length){bot.wait=1200;return;}}
      this.stepEntity(bot,delta,.034,false);if(!bot.path.length)bot.wait=900;
    }

    ambientRoutes() {
      const available=[...this.crew.values()].filter(actor=>actor.live&&!actor.path.length&&!['planning','celebrating','departing'].includes(actor.mode));
      const destinations=[WAYPOINTS.west,WAYPOINTS.east,WAYPOINTS.south,WAYPOINTS.southWest,WAYPOINTS.southEast,WAYPOINTS.north];
      available.forEach((actor,index)=>{
        const cycle=Math.floor(this.worldTime/6200)+index;
        const target=cycle%2===0?destinations[cycle%destinations.length]:AGENT_SLOTS[[...this.specs.keys()].indexOf(actor.id)%AGENT_SLOTS.length];
        actor.ambientTarget=target;actor.ambientUntil=this.worldTime+(this.reducedMotion?7200:4200);this.route(actor,target);
      });
    }

    safeDraw() {
      const started=performance.now();
      try {this.draw();}
      catch(error){const message=error instanceof Error?error.message:String(error);this.drawFailure(message);console.error('CoAct orbital renderer failed',error);}
      finally {
        const cost=performance.now()-started;
        this.averageDrawCost=this.averageDrawCost*.88+cost*.12;
        if(this.averageDrawCost>11)this.slowDrawCount=Math.min(30,this.slowDrawCount+1);
        else if(this.averageDrawCost<7)this.slowDrawCount=Math.max(0,this.slowDrawCount-1);
        this.updatePerformancePressure();
      }
    }

    updatePerformancePressure() {
      if(this.qualityPreference==='smooth'){this.performancePressure=0;this.setLowPowerBackground(false);return;}
      if(this.qualityPreference==='lite'){this.performancePressure=30;this.setLowPowerBackground(true);return;}
      const slow=this.averageFrameGap>24||this.averageDrawCost>11||this.slowDrawCount>=8;
      this.performancePressure=clamp(this.performancePressure+(slow?1:-.05),0,30);
      if(this.chromium){this.setLowPowerBackground(true);return;}
      if(!this.lowPowerBackground&&this.performancePressure>=12)this.setLowPowerBackground(true);
      else if(this.lowPowerBackground&&this.performancePressure<=3)this.setLowPowerBackground(false);
    }

    setLowPowerBackground(enabled) {
      enabled=!!enabled;if(this.lowPowerBackground===enabled)return;this.lowPowerBackground=enabled;this.backgroundReady=false;this.syncControls();
    }

    draw() {
      const ctx=this.ctx;ctx.clearRect(0,0,WIDTH,HEIGHT);
      if(!this.assetsReady){this.drawLoading();return;}
      ctx.imageSmoothingEnabled=false;
      const transition=clamp((this.worldTime-this.sceneChangedAt)/650,0,1),backgroundInterval=this.backgroundInterval();
      if(!this.backgroundReady||this.worldTime-this.lastBackgroundDrawTime>=backgroundInterval)this.renderBackground(transition);
      const entities=[...this.services.map(bot=>({kind:'service',value:bot,y:bot.py})),...[...this.crew.values()].filter(actor=>actor.mode!=='hidden').map(actor=>({kind:'crew',value:actor,y:actor.py}))].sort((a,b)=>a.y-b.y);
      entities.forEach(entity=>entity.kind==='crew'?this.drawCrew(entity.value):this.drawService(entity.value));
      this.drawHoverLabel();
      this.drawModeAtmosphere();
    }

    renderBackground(transition) {
      const outputCtx=this.ctx;this.ctx=this.backgroundCtx;
      try {
        const ctx=this.ctx;ctx.clearRect(0,0,WIDTH,HEIGHT);ctx.imageSmoothingEnabled=false;
      ctx.drawImage(this.stations[this.modeIndex],0,0,WIDTH,HEIGHT);
      if(transition<1&&this.previousModeIndex!==this.modeIndex){ctx.save();ctx.globalAlpha=1-transition;ctx.drawImage(this.stations[this.previousModeIndex],0,0,WIDTH,HEIGHT);ctx.restore();}
      if(!this.lowPowerBackground)this.drawWindowMotion();
      this.drawInteriorAnimation();
      if(!this.lowPowerBackground)this.drawBackgroundNPCs();
        this.backgroundReady=true;this.lastBackgroundDrawTime=this.worldTime;
      } finally {this.ctx=outputCtx;}
    }

    drawLoading() {
      const ctx=this.ctx;ctx.fillStyle='#030713';ctx.fillRect(0,0,WIDTH,HEIGHT);ctx.fillStyle='#78bfff';ctx.font='bold 12px ui-monospace,monospace';ctx.fillText(this.assetsError?'ASSET LOAD ERROR':'PREPARING ORBITAL STATION',32,HEIGHT/2);ctx.fillStyle='#b4c3d8';ctx.font='9px ui-monospace,monospace';ctx.fillText(this.assetsError||'Loading high-fidelity pixel assets…',32,HEIGHT/2+22);
    }

    drawFailure(message) {const ctx=this.ctx;ctx.fillStyle='#030713';ctx.fillRect(0,0,WIDTH,HEIGHT);ctx.fillStyle='#ff6f7f';ctx.font='bold 11px ui-monospace,monospace';ctx.fillText('ORBITAL RENDER ERROR',32,HEIGHT/2);ctx.fillStyle='#dce7f6';ctx.font='8px ui-monospace,monospace';ctx.fillText(message.slice(0,100),32,HEIGHT/2+20);this.text('worldEventTitle','Renderer needs attention');this.text('worldEventText',message);}

    drawWindowMotion() {
      const ctx=this.ctx,windows=[[[168,38],[333,38],[333,137],[174,146]],[[460,38],[612,43],[624,147],[458,137]],[[35,132],[105,90],[106,234],[42,258]],[[663,91],[730,132],[726,258],[660,233]]];
      ctx.save();ctx.globalCompositeOperation='screen';
      windows.forEach((points,windowIndex)=>{ctx.save();ctx.beginPath();points.forEach((point,index)=>index?ctx.lineTo(...point):ctx.moveTo(...point));ctx.closePath();ctx.clip();
        if(this.modeIndex===0)this.drawOrbitWindow(points,windowIndex);
        else if(this.modeIndex===1)this.drawOceanWindow(points,windowIndex);
        else if(this.modeIndex===2)this.drawEcodomeWindow(points,windowIndex);
        else this.drawWastelandWindow(points,windowIndex);
        ctx.restore();});
      ctx.restore();ctx.globalAlpha=1;
    }

    drawStars(color,windowIndex,speed=.012,alpha=.25) {
      const ctx=this.ctx;
      this.stars.forEach(star=>{const x=(star.x+this.worldTime*star.depth*speed+windowIndex*73)%WIDTH,y=star.y+(windowIndex%2)*18;ctx.globalAlpha=alpha+((Math.floor(this.worldTime/350)+star.phase)%5===0?.55:.1);ctx.fillStyle=color;ctx.fillRect(Math.floor(x),Math.floor(y),star.size,star.size);});
    }

    drawOrbitWindow(points,windowIndex) {
      const ctx=this.ctx;this.drawStars(windowIndex%2?'#b8c8ff':'#80dcff',windowIndex,.012,.24);
      const cycle=(this.worldTime/14+windowIndex*180)%900;ctx.globalAlpha=.58;ctx.fillStyle='#dff8ff';ctx.fillRect(Math.floor(points[0][0]-120+cycle),points[0][1]+24,46,1);ctx.globalAlpha=.22;ctx.fillRect(Math.floor(points[0][0]-128+cycle),points[0][1]+25,62,1);
    }

    drawOceanWindow(points,windowIndex) {
      const ctx=this.ctx,left=points[0][0]-50,top=points[0][1]+6,time=this.worldTime;
      for(let index=0;index<12;index++){const x=left+(index*41+windowIndex*29)%255,y=top+105-((time*.018+index*17+windowIndex*11)%118),size=index%5===0?3:2;ctx.globalAlpha=.22+(index%3)*.08;ctx.fillStyle='#d8ffff';ctx.fillRect(Math.floor(x),Math.floor(y),size,size);if(size===3)ctx.fillStyle='#66ddeb',ctx.fillRect(Math.floor(x)+1,Math.floor(y)+1,1,1);}
      for(let fish=0;fish<4;fish++){const direction=(fish+windowIndex)%2?1:-1,travel=(time*.014*(1+fish*.12)+fish*67+windowIndex*43)%260,x=direction>0?left+travel:left+240-travel,y=top+24+fish*19+Math.sin(time/420+fish)*5,color=['#ffcf5d','#ff7f68','#6ff2c4','#9ed5ff'][fish];ctx.globalAlpha=.62;ctx.fillStyle=color;ctx.fillRect(Math.floor(x),Math.floor(y),6,3);ctx.fillRect(Math.floor(x)+(direction>0?-2:6),Math.floor(y)+1,2,1);ctx.fillStyle='#17364a';ctx.fillRect(Math.floor(x)+(direction>0?4:1),Math.floor(y),1,1);}
      ctx.globalAlpha=.08;ctx.fillStyle='#b8fff7';const caustic=Math.floor(time/38)%36;for(let line=0;line<4;line++)ctx.fillRect(left-20+caustic+line*54,top+15+line*22,34,2);
    }

    drawEcodomeWindow(points,windowIndex) {
      const ctx=this.ctx,left=points[0][0]-45,top=points[0][1]+4,time=this.worldTime;
      ctx.globalAlpha=.1+(Math.sin(time/510+windowIndex)+1)*.04;ctx.fillStyle='#f2fff2';for(let mist=0;mist<6;mist++){const x=left+(mist*43+time*.006)%250,y=top+70+Math.sin(time/360+mist)*12;ctx.fillRect(Math.floor(x),Math.floor(y),34,7);}
      for(let fall=0;fall<4;fall++){ctx.globalAlpha=.16+fall*.025;ctx.fillStyle=fall%2?'#aeefff':'#e8ffff';const x=left+35+fall*57;ctx.fillRect(Math.floor(x),top,2,104);ctx.fillRect(Math.floor(x)+3,top+((time/24+fall*29)%80),1,23);}
      for(let leaf=0;leaf<7;leaf++){const x=left+(leaf*37+time*.009)%250,y=top+(leaf*19+time*.013)%105,wave=Math.sin(time/260+leaf)*5;ctx.globalAlpha=.5;ctx.fillStyle=leaf%2?'#9be46e':'#f2dc72';ctx.fillRect(Math.floor(x+wave),Math.floor(y),3,2);}
      ctx.globalAlpha=.55;ctx.fillStyle='#f7ffff';const bird=(time*.01+windowIndex*83)%250;ctx.fillRect(Math.floor(left+bird),top+20,3,1);ctx.fillRect(Math.floor(left+bird)+4,top+20,3,1);ctx.fillRect(Math.floor(left+bird)+3,top+21,1,1);
    }

    drawWastelandWindow(points,windowIndex) {
      const ctx=this.ctx,left=points[0][0]-55,top=points[0][1]+4,time=this.worldTime,drift=(time/11+windowIndex*53)%260;
      for(let streak=0;streak<9;streak++){ctx.globalAlpha=.08+streak*.012;ctx.fillStyle=streak%2?'#ffd17b':'#d87836';ctx.fillRect(Math.floor(left-100+drift+streak*31),top+12+streak*11,48+streak*3,2);}
      for(let mote=0;mote<15;mote++){const x=left+(mote*47+time*.025*(1+mote%3))%255,y=top+(mote*23+Math.sin(time/250+mote)*8)%105;ctx.globalAlpha=.24;ctx.fillStyle=mote%3?'#e9a45c':'#ffe0a0';ctx.fillRect(Math.floor(x),Math.floor(y),mote%5===0?2:1,1);}
      const tumble=(time*.018+windowIndex*61)%245,bounce=Math.abs(Math.sin(time/330))*7;ctx.globalAlpha=.55;ctx.strokeStyle='#b96b35';ctx.lineWidth=2;ctx.beginPath();ctx.ellipse(Math.floor(left+tumble),Math.floor(top+86-bounce),7,7,0,0,Math.PI*2);ctx.stroke();
    }

    drawInteriorAnimation() {
      const ctx=this.ctx,scene=SCENES[this.modeIndex],pulse=.45+(Math.sin(this.worldTime/420)+1)*.22;
      ctx.save();ctx.globalCompositeOperation='screen';ctx.globalAlpha=pulse;
      const monitors=[[183,151,37,3,scene.accent],[221,161,24,2,scene.soft],[520,149,36,3,scene.accent],[557,160,23,2,scene.soft],[159,332,28,3,scene.accent],[540,355,25,2,scene.soft]];
      monitors.forEach(([x,y,w,h,color],index)=>{ctx.fillStyle=color;ctx.fillRect(x,y,w,h);if((Math.floor(this.worldTime/500)+index)%3===0)ctx.fillRect(x+3,y+5,Math.max(5,w-9),1);});
      ctx.strokeStyle=scene.accent;ctx.lineWidth=1;const radius=34+(Math.sin(this.worldTime/520)+1)*3;ctx.beginPath();ctx.ellipse(384,246,radius,Math.round(radius*.45),0,0,Math.PI*2);ctx.stroke();ctx.globalAlpha=.28;ctx.beginPath();ctx.ellipse(384,246,radius+13,(radius+13)*.45,0,0,Math.PI*2);ctx.stroke();
      const angle=this.worldTime/(this.modeIndex===3?470:850),orbX=384+Math.cos(angle)*26,orbY=239+Math.sin(angle)*10;ctx.globalAlpha=.9;ctx.fillStyle=scene.soft;ctx.fillRect(Math.round(orbX)-1,Math.round(orbY)-1,3,3);
      ctx.globalAlpha=.25+(Math.sin(this.worldTime/700)+1)*.18;ctx.fillStyle=scene.accent;ctx.fillRect(374,221,20,31);
      if(this.modeIndex===1){ctx.globalAlpha=.18;for(let index=0;index<4;index++){ctx.beginPath();ctx.ellipse(384,246,52+index*9,20+index*4,0,0,Math.PI*2);ctx.stroke();}}
      if(this.modeIndex===2){ctx.globalAlpha=.26;for(let x=338;x<=430;x+=6){const y=244+Math.sin((x+this.worldTime/18)/13)*9;ctx.fillRect(x,Math.floor(y),4,12);}}
      if(this.modeIndex===3){ctx.globalAlpha=.3+(Math.sin(this.worldTime/170)+1)*.15;ctx.fillStyle='#ff794d';ctx.fillRect(302,299,164,2);ctx.fillRect(324,318,120,2);}
      ctx.globalAlpha=.55;ctx.fillStyle=Math.floor(this.worldTime/650)%2?scene.accent:scene.soft;ctx.fillRect(375,68,18,2);ctx.restore();ctx.globalAlpha=1;
    }

    drawBackgroundNPCs() {
      if(this.modeIndex!==3)return;
      const ctx=this.ctx,atlas=this.atlases[3],pulse=Math.floor(this.worldTime/900)%2;
      const drawNPC=(row,frame,x,y,width=29,height=53)=>ctx.drawImage(atlas,frame*CELL.w,row*CELL.h,CELL.w,CELL.h,Math.round(x-width/2),Math.round(y-height+8),width,height);
      const live=[...this.crew.values()].filter(actor=>actor.live&&actor.mode!=='hidden');
      const busy=zone=>live.some(actor=>actor.px>=zone.left&&actor.px<=zone.right&&actor.py>=zone.top&&actor.py<=zone.bottom);
      const zones=[
        {left:470,right:640,top:120,bottom:250,people:[[520,166],[568,166],[626,158]],table:[544,153]},
        {left:470,right:650,top:285,bottom:420,people:[[528,374],[574,374],[622,360]],table:[552,356]},
        {left:115,right:300,top:285,bottom:420,people:[[152,370],[198,370],[248,354]],table:[176,352]}
      ];
      const zone=zones.find(candidate=>!busy(candidate));
      if(!zone)return;
      ctx.save();ctx.globalAlpha=.78;
      drawNPC(1,pulse?FRAME.typeA:FRAME.typeB,...zone.people[0],27,49);
      drawNPC(2,FRAME.sit,...zone.people[1],27,49);
      drawNPC(0,FRAME.down,...zone.people[2],28,51);
      const [tableX,tableY]=zone.table;ctx.fillStyle='#f0c05e';ctx.fillRect(tableX,tableY,2,2);ctx.fillRect(tableX+7,tableY+2,2,2);ctx.fillStyle='#d85442';ctx.fillRect(tableX+12,tableY-1,3,2);
      const [lampX,lampY]=zone.people[2];ctx.globalAlpha=.35+(Math.sin(this.worldTime/240)+1)*.15;ctx.fillStyle='#ffb64f';ctx.fillRect(lampX+7,lampY-29,4,7);ctx.fillStyle='#ffe39a';ctx.fillRect(lampX+8,lampY-30,2,5);
      ctx.restore();
    }

    spriteFrame(actor) {
      if(actor.path.length)return actor.direction==='up'?FRAME.up:['left','right'].includes(actor.direction)?FRAME.right:FRAME.down;
      if(actor.mode==='working'){const activity=this.activityFor(actor),phase=Math.floor((this.worldTime+this.actorPhase(actor)*170)/520);if(activity==='designing')return phase%3===0?FRAME.typeA:FRAME.plan;if(activity==='prototyping')return phase%3===0?FRAME.plan:FRAME.typeB;if(activity==='reviewing'||activity==='writing')return phase%3===0?FRAME.wait:FRAME.plan;if(activity==='testing'||activity==='debugging')return phase%3===0?FRAME.wait:FRAME.typeB;if(activity==='security'||activity==='analyzing')return phase%4===0?FRAME.wait:FRAME.plan;if(activity==='researching')return phase%4===0?FRAME.sit:FRAME.plan;if(activity==='coordinating')return phase%3===0?FRAME.wait:FRAME.typeA;return phase%2?FRAME.typeA:FRAME.typeB;}
      if(actor.mode==='waiting')return FRAME.wait;
      if(actor.mode==='idle')return FRAME.sleep;
      if(actor.mode==='planning')return FRAME.plan;
      if(actor.mode==='celebrating')return FRAME.celebrate;
      return FRAME[actor.direction]??FRAME.down;
    }

    drawCrew(actor) {
      const ctx=this.ctx,atlas=this.atlases[this.modeIndex],spec=this.specFor(actor.id),frame=this.spriteFrame(actor),sourceX=frame*CELL.w,sourceY=spec.row*CELL.h;
      const width=40,height=73,x=Math.round(actor.px-width/2),y=Math.round(actor.py-height+9),walking=actor.path.length>0,horizontal=walking&&['left','right'].includes(actor.direction),flip=horizontal&&actor.direction==='left';
      ctx.save();ctx.globalAlpha=actor.mode==='departing'?.82:1;
      ctx.fillStyle='#02050b66';ctx.beginPath();ctx.ellipse(actor.px,actor.py+6,12,4,0,0,Math.PI*2);ctx.fill();
      if(walking)this.drawWalkingCrew(actor,atlas,sourceX,sourceY,x,y,width,height,flip);
      else ctx.drawImage(atlas,sourceX,sourceY,CELL.w,CELL.h,x,y,width,height);
      if(this.selected===actor.id){ctx.strokeStyle='#f7fbff';ctx.lineWidth=2;ctx.strokeRect(x-3,y-3,width+6,height+6);}
      if(actor.path.length){ctx.globalAlpha=.85;ctx.fillStyle=spec.color;ctx.fillRect(Math.round(actor.px)-7,Math.round(actor.py)-66,14,2);ctx.globalAlpha=1;}
      if(actor.mode==='working'&&!actor.path.length){const dot=Math.floor(this.worldTime/260)%3;ctx.fillStyle=spec.color;for(let index=0;index<3;index++)ctx.fillRect(Math.round(actor.px)-6+index*5,Math.round(actor.py)-67-(index===dot?2:0),2,2);}
      if(actor.celebrateUntil>this.worldTime)this.drawCelebration(actor);
      if(this.assistSuggestion?.helper===actor.id)this.drawAssistBubble(actor);
      ctx.restore();
    }

    drawWalkingCrew(actor,atlas,sourceX,sourceY,x,y,width,height,flip) {
      const ctx=this.ctx,split=130,destSplit=Math.round(height*split/CELL.h),legHeight=height-destSplit,phase=Math.floor((actor.walkDistance||0)/5)%4;
      const offsets=[{x:0,y:0},{x:-1,y:-1},{x:0,y:0},{x:1,y:0}],left=offsets[phase],right=offsets[(phase+2)%4];
      ctx.save();ctx.translate(flip?x+width:x,y);if(flip)ctx.scale(-1,1);
      ctx.drawImage(atlas,sourceX,sourceY,CELL.w,split,0,0,width,destSplit);
      ctx.drawImage(atlas,sourceX,sourceY+split,CELL.w/2,CELL.h-split,left.x,destSplit+left.y,width/2,legHeight);
      ctx.drawImage(atlas,sourceX+CELL.w/2,sourceY+split,CELL.w/2,CELL.h-split,width/2+right.x,destSplit+right.y,width/2,legHeight);
      ctx.restore();
    }

    drawCelebration(actor) {
      const ctx=this.ctx,age=this.worldTime-actor.celebrateStarted,scene=SCENES[this.modeIndex],originX=Math.round(actor.px),originY=Math.round(actor.py-34);
      ctx.save();
      const ring=14+(age%900)/75;
      ctx.globalAlpha=.8*(1-(age%900)/900);
      ctx.strokeStyle=scene.accent;ctx.lineWidth=2;ctx.beginPath();ctx.ellipse(originX,actor.py+5,ring,ring*.28,0,0,Math.PI*2);ctx.stroke();
      const colors=[scene.accent,scene.soft,'#ffd76e','#ffffff'];
      for(let index=0;index<18;index++){
        const local=(age+index*137)%1500,progress=local/1500,angle=index/18*Math.PI*2+this.actorPhase(actor),radius=10+progress*38;
        const x=Math.round(originX+Math.cos(angle)*radius),y=Math.round(originY+Math.sin(angle)*radius-progress*20),size=index%4===0?3:2;
        ctx.globalAlpha=1-progress;ctx.fillStyle=colors[index%colors.length];
        if(this.modeIndex===0){ctx.fillRect(x-size,y,1+size*2,1);ctx.fillRect(x,y-size,1,1+size*2);}
        else if(this.modeIndex===1){ctx.strokeStyle=ctx.fillStyle;ctx.strokeRect(x-size,y-size,size*2,size*2);}
        else if(this.modeIndex===2){ctx.fillRect(x-size,y,size*2+1,2);ctx.fillRect(x,y-2,1,5);}
        else{ctx.fillRect(x,y,size,size);ctx.fillRect(x-size,y+2,2,1);}
      }
      ctx.globalAlpha=1;ctx.fillStyle=scene.soft;ctx.font='bold 8px ui-monospace,monospace';ctx.textAlign='center';ctx.fillText('TASK COMPLETE',originX,Math.max(76,originY-38));ctx.textAlign='start';ctx.restore();
    }

    drawAssistBubble(actor) {
      const suggestion=this.assistSuggestion;if(!suggestion)return;
      const ctx=this.ctx,waiting=this.specFor(suggestion.waiting),width=104,height=31,left=clamp(Math.round(actor.px-width/2),6,WIDTH-width-6),top=Math.max(74,Math.round(actor.py-108));
      ctx.save();ctx.fillStyle='#07101ef2';ctx.fillRect(left,top,width,height);ctx.strokeStyle='#f0ca70';ctx.lineWidth=2;ctx.strokeRect(left,top,width,height);ctx.fillStyle='#07101ef2';ctx.fillRect(Math.round(actor.px)-3,top+height,7,6);ctx.fillStyle='#f7dda0';ctx.font='bold 8px ui-monospace,monospace';ctx.textAlign='center';ctx.fillText('NEED A HAND?',left+width/2,top+12);ctx.fillStyle='#aebbd0';ctx.font='7px ui-monospace,monospace';ctx.fillText(waiting.short+' IS WAITING · CLICK',left+width/2,top+23);ctx.textAlign='start';ctx.restore();
    }

    drawService(bot) {
      const ctx=this.ctx,x=Math.round(bot.px),y=Math.round(bot.py),bob=Math.floor(this.worldTime/230+this.services.indexOf(bot))%2;
      ctx.fillStyle='#02050b88';ctx.beginPath();ctx.ellipse(x,y+12,12,4,0,0,Math.PI*2);ctx.fill();
      if(bot.path.length){ctx.globalAlpha=.18;ctx.fillStyle=bot.color;ctx.fillRect(x-13,y+13,26,2);ctx.globalAlpha=1;}
      if(this.modeIndex===1)this.drawOceanCompanion(ctx,bot,x,y,bob);
      else if(this.modeIndex===2)this.drawEcoCompanion(ctx,bot,x,y,bob);
      else if(this.modeIndex===3)this.drawWastelandCompanion(ctx,bot,x,y,bob);
      else if(bot.id==='sentinel')this.drawSentinel(ctx,x,y,bob,bot.color);
      else if(bot.id==='relay')this.drawRelay(ctx,x,y,bob,bot.color);
      else if(bot.id==='mnemo')this.drawMnemo(ctx,x,y,bob,bot.color);
      else this.drawLedger(ctx,x,y,bob,bot.color);
      if(this.selected==='service:'+bot.id){ctx.strokeStyle='#fff';ctx.strokeRect(x-17,y-20,34,39);}
    }

    drawOceanCompanion(ctx,bot,x,y,bob) {
      const top=y-13+bob;
      if(bot.id==='sentinel'){ctx.fillStyle='#e96d49';ctx.fillRect(x-9,top+8,18,12);ctx.fillStyle='#ff9c68';ctx.fillRect(x-6,top+5,12,6);ctx.fillStyle='#142536';ctx.fillRect(x-4,top+6,2,2);ctx.fillRect(x+2,top+6,2,2);ctx.fillStyle='#ffb276';ctx.fillRect(x-16,top+9,7,6);ctx.fillRect(x+9,top+9,7,6);ctx.fillStyle='#e96d49';ctx.fillRect(x-14,top+6,4,4);ctx.fillRect(x+10,top+6,4,4);for(let leg=0;leg<3;leg++){ctx.fillRect(x-10+leg*5,top+20,4,3);ctx.fillRect(x+6-leg*5,top+20,4,3);}}
      else if(bot.id==='relay'){ctx.fillStyle='#44d5ee';ctx.fillRect(x-10,top+7,19,11);ctx.fillStyle='#9efff1';ctx.fillRect(x-7,top+5,12,5);ctx.fillStyle='#1d4360';ctx.fillRect(x+5,top+8,2,2);ctx.fillStyle='#38a9d2';ctx.fillRect(x-15,top+9,5,3);ctx.fillRect(x-17,top+7,4,7);ctx.fillStyle='#ffd96a';ctx.fillRect(x-4,top+17,9,5);ctx.fillStyle='#805938';ctx.fillRect(x-2,top+18,5,2);}
      else if(bot.id==='mnemo'){ctx.globalAlpha=.82;ctx.fillStyle='#b894ff';ctx.beginPath();ctx.ellipse(x,top+10,11,9,0,Math.PI,0);ctx.fill();ctx.fillStyle='#eadfff';ctx.fillRect(x-6,top+8,12,6);ctx.fillStyle='#4b367c';ctx.fillRect(x-3,top+10,2,2);ctx.fillRect(x+2,top+10,2,2);ctx.fillStyle='#a879ee';for(let tentacle=0;tentacle<4;tentacle++)ctx.fillRect(x-8+tentacle*5,top+14,2,10+(tentacle%2)*3);ctx.globalAlpha=1;}
      else{ctx.fillStyle='#4f9b69';ctx.beginPath();ctx.ellipse(x,top+13,11,9,0,0,Math.PI*2);ctx.fill();ctx.fillStyle='#9bd07d';ctx.fillRect(x-7,top+8,14,9);ctx.fillStyle='#315a45';ctx.fillRect(x-4,top+10,8,5);ctx.fillStyle='#8fcba6';ctx.fillRect(x+9,top+10,5,5);ctx.fillStyle='#15372d';ctx.fillRect(x+12,top+11,1,1);ctx.fillStyle='#5e8466';ctx.fillRect(x-9,top+20,5,3);ctx.fillRect(x+4,top+20,5,3);}
    }

    drawEcoCompanion(ctx,bot,x,y,bob) {
      const top=y-13+bob;
      if(bot.id==='sentinel'){ctx.fillStyle='#563b2d';ctx.fillRect(x-8,top+7,16,16);ctx.fillStyle='#df5b4f';ctx.fillRect(x-7,top+5,14,13);ctx.fillStyle='#352a24';ctx.fillRect(x-1,top+5,2,13);ctx.fillRect(x-5,top+8,2,2);ctx.fillRect(x+3,top+8,2,2);ctx.fillStyle='#86c85d';ctx.fillRect(x-10,top+20,6,3);ctx.fillRect(x+4,top+20,6,3);}
      else if(bot.id==='relay'){ctx.fillStyle='#55cfc0';ctx.fillRect(x-7,top+9,14,8);ctx.fillStyle='#f0e0a0';ctx.fillRect(x+5,top+10,6,2);ctx.fillStyle='#315b4f';ctx.fillRect(x-10,top+12,4,3);ctx.fillStyle='#dfffe8';ctx.globalAlpha=.48;ctx.fillRect(x-12,top+4,8,8);ctx.fillRect(x+4,top+4,8,8);ctx.globalAlpha=1;ctx.fillStyle='#75dd75';ctx.fillRect(x-3,top+17,6,5);}
      else if(bot.id==='mnemo'){ctx.fillStyle='#9b6843';ctx.fillRect(x-8,top+13,16,11);ctx.fillStyle='#d5a36b';ctx.fillRect(x-10,top+11,20,4);ctx.fillStyle='#6cca67';ctx.fillRect(x-2,top+3,4,10);ctx.fillRect(x+1,top+4,7,4);ctx.fillRect(x-8,top+5,7,4);ctx.fillStyle='#eff8d3';ctx.fillRect(x-4,top+16,8,4);ctx.fillStyle='#467447';ctx.fillRect(x-2,top+17,4,2);}
      else{ctx.fillStyle='#8f6540';ctx.fillRect(x-10,top+6,20,17);ctx.fillStyle='#d7b681';ctx.fillRect(x-7,top+8,14,10);ctx.fillStyle='#fff0c9';ctx.fillRect(x-5,top+10,4,4);ctx.fillRect(x+1,top+10,4,4);ctx.fillStyle='#30291f';ctx.fillRect(x-3,top+11,2,2);ctx.fillRect(x+1,top+11,2,2);ctx.fillStyle='#7a4f32';ctx.fillRect(x-12,top+8,4,8);ctx.fillRect(x+8,top+8,4,8);ctx.fillStyle='#6db36a';ctx.fillRect(x-5,top+19,10,4);}
    }

    drawWastelandCompanion(ctx,bot,x,y,bob) {
      const top=y-13+bob;
      if(bot.id==='sentinel'){ctx.fillStyle='#7a6652';ctx.fillRect(x-9,top+7,18,17);ctx.fillStyle='#d6b676';ctx.fillRect(x-7,top+9,14,7);ctx.fillStyle='#342d26';ctx.fillRect(x-3,top+11,2,2);ctx.fillRect(x+1,top+11,2,2);ctx.fillStyle='#5a402d';ctx.fillRect(x-13,top+4,26,4);ctx.fillRect(x-8,top,16,5);ctx.fillStyle='#e8b14d';ctx.fillRect(x-2,top+17,4,4);ctx.fillStyle='#4d4338';ctx.fillRect(x-8,top+24,6,3);ctx.fillRect(x+2,top+24,6,3);}
      else if(bot.id==='relay'){ctx.fillStyle='#c88a49';ctx.fillRect(x-8,top+10,15,10);ctx.fillStyle='#f1c675';ctx.fillRect(x+6,top+8,8,5);ctx.fillStyle='#2f2a25';ctx.fillRect(x+11,top+9,1,1);ctx.fillStyle='#8e5e36';ctx.fillRect(x-13,top+8,6,5);ctx.fillStyle='#3e5360';ctx.fillRect(x-3,top+19,3,8);ctx.fillRect(x+5,top+19,3,8);ctx.fillStyle='#e2a84d';ctx.fillRect(x-4,top+6,2,5);}
      else if(bot.id==='mnemo'){ctx.fillStyle='#4a382b';ctx.fillRect(x-8,top+4,16,21);ctx.fillStyle='#96704a';ctx.fillRect(x-10,top+7,20,4);ctx.globalAlpha=.45+(Math.sin(this.worldTime/180)+1)*.2;ctx.fillStyle='#ffb13d';ctx.fillRect(x-5,top+10,10,11);ctx.fillStyle='#fff0a1';ctx.fillRect(x-2,top+12,4,6);ctx.globalAlpha=1;ctx.fillStyle='#2d251e';ctx.fillRect(x-6,top+22,12,3);ctx.fillRect(x-2,top,4,4);}
      else{ctx.fillStyle='#8c715b';ctx.beginPath();ctx.ellipse(x,top+15,12,9,0,0,Math.PI*2);ctx.fill();ctx.fillStyle='#c9a77b';for(let plate=0;plate<4;plate++)ctx.fillRect(x-8+plate*5,top+10+Math.abs(2-plate),4,8);ctx.fillStyle='#5d493a';ctx.fillRect(x+10,top+13,5,4);ctx.fillStyle='#2d261f';ctx.fillRect(x+13,top+14,1,1);ctx.fillStyle='#6d5847';ctx.fillRect(x-8,top+22,5,3);ctx.fillRect(x+3,top+22,5,3);}
    }

    drawSentinel(ctx,x,y,bob,color) {
      const top=y-13+bob;
      ctx.fillStyle='#171d27';ctx.fillRect(x-10,top+5,20,18);ctx.fillStyle='#576578';ctx.fillRect(x-8,top+3,16,9);ctx.fillStyle='#d8e2ee';ctx.fillRect(x-6,top+5,12,6);ctx.fillStyle=color;ctx.fillRect(x-4,top+7,8,2);ctx.fillStyle='#10151d';ctx.fillRect(x-2,top+7,4,2);
      ctx.fillStyle='#303b4a';ctx.fillRect(x-8,top+12,16,9);ctx.fillStyle=color;ctx.fillRect(x-5,top+14,10,4);ctx.fillStyle='#151b24';ctx.fillRect(x-3,top+15,6,2);
      ctx.fillStyle='#2b3441';ctx.fillRect(x-8,top+23,6,4);ctx.fillRect(x+2,top+23,6,4);ctx.fillStyle='#91a0b3';ctx.fillRect(x-6,top+25,4,2);ctx.fillRect(x+2,top+25,4,2);
      ctx.fillStyle='#394657';ctx.fillRect(x-14,top+10,4,12);ctx.fillRect(x+10,top+11,4,8);ctx.fillStyle=color;ctx.fillRect(x-13,top+12,2,8);ctx.fillRect(x-1,top,2,3);
    }

    drawRelay(ctx,x,y,bob,color) {
      const top=y-13+bob;
      ctx.fillStyle='#5b6d82';ctx.fillRect(x-14,top+7,4,10);ctx.fillRect(x+10,top+7,4,10);ctx.fillStyle=color;ctx.fillRect(x-15,top+9,3,6);ctx.fillRect(x+12,top+9,3,6);
      ctx.fillStyle='#dce8f4';ctx.fillRect(x-10,top+4,20,18);ctx.fillStyle='#73869c';ctx.fillRect(x-8,top+2,16,5);ctx.fillStyle=color;ctx.fillRect(x-6,top+6,12,5);ctx.fillStyle='#102233';ctx.fillRect(x-3,top+8,6,2);
      ctx.fillStyle='#f8f2df';ctx.fillRect(x-8,top+13,16,8);ctx.fillStyle=color;ctx.fillRect(x-6,top+15,12,1);ctx.fillRect(x-4,top+17,8,1);ctx.fillStyle='#526377';ctx.fillRect(x-8,top+22,6,4);ctx.fillRect(x+2,top+22,6,4);
      ctx.fillStyle=color;ctx.fillRect(x-1,top-1,2,3);ctx.fillRect(x,top-3,5,2);
    }

    drawMnemo(ctx,x,y,bob,color) {
      const top=y-13+bob;
      ctx.globalAlpha=.28;ctx.strokeStyle=color;ctx.lineWidth=1;ctx.strokeRect(x-13,top+7,26,12);ctx.globalAlpha=1;
      ctx.fillStyle='#292441';ctx.fillRect(x-10,top+4,20,19);ctx.fillStyle='#7667a4';ctx.fillRect(x-8,top+2,16,6);ctx.fillStyle='#ece9ff';ctx.fillRect(x-7,top+7,14,10);ctx.fillStyle=color;ctx.fillRect(x-5,top+9,10,6);ctx.fillStyle='#211846';ctx.fillRect(x-2,top+10,4,4);ctx.fillStyle='#fff';ctx.fillRect(x-1,top+10,2,2);
      ctx.fillStyle='#51476e';ctx.fillRect(x-8,top+18,16,5);ctx.fillStyle=color;ctx.fillRect(x-5,top+19,10,2);ctx.fillStyle='#332d48';ctx.fillRect(x-8,top+23,6,4);ctx.fillRect(x+2,top+23,6,4);ctx.fillStyle=color;ctx.fillRect(x-1,top-1,2,4);ctx.fillRect(x-3,top-2,6,2);
    }

    drawLedger(ctx,x,y,bob,color) {
      const top=y-13+bob;
      ctx.fillStyle='#26372f';ctx.fillRect(x-11,top+5,22,18);ctx.fillStyle='#63796b';ctx.fillRect(x-9,top+2,18,7);ctx.fillStyle='#dfe9df';ctx.fillRect(x-7,top+4,14,5);ctx.fillStyle='#121c17';ctx.fillRect(x-5,top+5,10,2);
      ctx.fillStyle=color;ctx.fillRect(x-8,top+11,16,7);ctx.fillStyle='#173427';ctx.fillRect(x-5,top+13,10,1);ctx.fillRect(x-4,top+16,8,1);ctx.fillStyle='#8ca394';ctx.fillRect(x-7,top+19,14,3);
      ctx.fillStyle='#34483d';ctx.fillRect(x-9,top+23,6,4);ctx.fillRect(x+3,top+23,6,4);ctx.fillStyle=color;ctx.fillRect(x-6,top+24,3,2);ctx.fillRect(x+3,top+24,3,2);ctx.fillRect(x-1,top-1,2,3);
    }

    drawModeAtmosphere() {
      const age=this.worldTime-this.sceneChangedAt;if(age<0||age>1400)return;const ctx=this.ctx,scene=SCENES[this.modeIndex],fade=age<180?age/180:1-(age-180)/1220;ctx.save();ctx.globalAlpha=clamp(fade,0,1)*.9;ctx.fillStyle='#030713dd';ctx.fillRect(316,446,136,28);ctx.strokeStyle=scene.accent;ctx.strokeRect(316,446,136,28);ctx.fillStyle=scene.soft;ctx.font='bold 9px ui-monospace,monospace';ctx.textAlign='center';ctx.fillText(scene.label,384,463);ctx.textAlign='start';ctx.restore();
    }

    servicePresentation(bot) {
      const themed=SERVICE_THEMES[this.modeIndex]?.[bot.id];return themed?{name:themed[0],role:themed[1]}:{name:bot.name,role:bot.role};
    }

    canvasPoint(event) {const rect=this.canvas.getBoundingClientRect();return{x:(event.clientX-rect.left)*WIDTH/rect.width,y:(event.clientY-rect.top)*HEIGHT/rect.height,cx:event.clientX-rect.left,cy:event.clientY-rect.top};}
    hit(point) {
      if(this.assistSuggestion){const helper=this.crew.get(this.assistSuggestion.helper);if(helper){const left=clamp(helper.px-52,6,WIDTH-110),top=Math.max(74,helper.py-108);if(point.x>=left&&point.x<=left+104&&point.y>=top&&point.y<=top+37)return{type:'crew',actor:helper,label:'Offer help or take over'};}}
      for(const actor of this.crew.values())if(actor.mode!=='hidden'&&Math.abs(point.x-actor.px)<22&&point.y>actor.py-68&&point.y<actor.py+12)return{type:'crew',actor,label:this.specFor(actor.id).name};
      for(const bot of this.services)if(Math.abs(point.x-bot.px)<18&&Math.abs(point.y-bot.py)<24){const presentation=this.servicePresentation(bot);return{type:'service',bot,label:presentation.name+' · '+presentation.role};}
      return null;
    }
    onPointer(event) {const point=this.canvasPoint(event),hit=this.hit(point);this.hover=hit;this.canvas.style.cursor=hit?'pointer':'crosshair';if(!hit){this.hideTooltip();if(this.paused)this.safeDraw();return;}const tooltip=this.root.querySelector('.pixel-tooltip');if(!tooltip)return;tooltip.textContent=hit.label;tooltip.style.left=clamp(point.cx+16,8,this.root.clientWidth-260)+'px';tooltip.style.top=clamp(point.cy+16,8,this.root.clientHeight-50)+'px';tooltip.classList.add('show');if(this.paused)this.safeDraw();}
    hideTooltip(){this.root.querySelector('.pixel-tooltip')?.classList.remove('show');}
    onClick(event){const hit=this.hit(this.canvasPoint(event));if(hit?.type==='crew')this.select(hit.actor.id);else if(hit?.type==='service')this.select('service:'+hit.bot.id);else this.select('');}
    select(id){this.selected=id;this.root.querySelector('.pixel-agent-panel')?.classList.toggle('show',!!id);if(id)this.renderSelected();if(this.paused)this.safeDraw();}
    renderSelected(){
      const assist=this.root.querySelector('#worldAgentAssist');
      assist?.classList.remove('show');
      if(this.selected.startsWith('service:')){
        const bot=this.services.find(item=>item.id===this.selected.slice(8));if(!bot)return;
        const presentation=this.servicePresentation(bot);this.text('worldAgentName',presentation.name);this.text('worldAgentStatus',bot.status);this.text('worldAgentTask',presentation.role);this.text('worldAgentBeat','local service · always on');this.text('worldAgentMode',bot.path.length?'patrolling':'monitoring');return;
      }
      const actor=this.crew.get(this.selected);if(!actor)return;
      const agent=actor.agent||{},task=this.taskFor(actor);this.text('worldAgentName',this.specFor(actor.id).name);this.text('worldAgentStatus',actor.live?(agent.status||actor.mode):'offline');this.text('worldAgentTask',task?.title||agent.current_task||'no active task');this.text('worldAgentBeat',agent.beat||'no recent heartbeat');this.text('worldAgentMode',actor.live?this.activityFor(actor):actor.mode);
      if(assist&&this.assistSuggestion?.helper===actor.id){
        const suggestion=this.assistSuggestion,waiting=this.specFor(suggestion.waiting),seconds=Math.max(1,Math.floor((Date.now()-suggestion.since)/1000));
        this.text('worldAgentAssistText',waiting.name+' has waited '+seconds+'s ('+suggestion.reason+'). Offer help first, or explicitly transfer active tasks and locks.');
        assist.classList.add('show');
      }
    }
    drawHoverLabel(){const hit=this.hover;if(!hit)return;const ctx=this.ctx;let x,y,title,subtitle;if(hit.type==='crew'){x=hit.actor.px;y=hit.actor.py-69;title=this.specFor(hit.actor.id).name;subtitle=hit.actor.path.length?'walking':this.activityFor(hit.actor);}else{const presentation=this.servicePresentation(hit.bot);x=hit.bot.px;y=hit.bot.py-28;title=presentation.name;subtitle=presentation.role;}ctx.font='bold 8px ui-monospace,monospace';const width=Math.max(ctx.measureText(title).width,ctx.measureText(subtitle).width)+14,left=clamp(x-width/2,4,WIDTH-width-4),top=clamp(y-25,4,HEIGHT-30);ctx.fillStyle='#050914ee';ctx.fillRect(left,top,width,24);ctx.strokeStyle='#91a9cf';ctx.strokeRect(left,top,width,24);ctx.fillStyle='#f5f8ff';ctx.textAlign='center';ctx.fillText(title,left+width/2,top+10);ctx.fillStyle='#97a8c1';ctx.font='7px ui-monospace,monospace';ctx.fillText(subtitle,left+width/2,top+19);ctx.textAlign='start';}
  }

  window.CoActWorld={
    instance:null,
    init(){const root=document.querySelector('.pixel-world-shell');if(root&&!this.instance)this.instance=new OrbitalWorld(root);return this.instance;},
    update(state){this.init()?.update(state||{});},
    assistHandled(action){this.instance?.assistHandled(action);}
  };
  document.readyState==='loading'?document.addEventListener('DOMContentLoaded',()=>window.CoActWorld.init()):window.CoActWorld.init();
})();
