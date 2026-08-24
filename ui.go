package main

// indexHTML is the whole UI: three panes (methods, request, response), no build
// step, no dependencies. {{ADDR}} is substituted with the default target.
const indexHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>grpc-lab</title>
<style>
:root{--bg:#0f1115;--panel:#161a21;--line:#252b36;--fg:#d7dce5;--dim:#8b95a7;--acc:#5eead4;--err:#f87171;--ok:#4ade80}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:13px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace}
header{display:flex;gap:8px;align-items:center;padding:10px 14px;border-bottom:1px solid var(--line);background:var(--panel)}
header b{color:var(--acc);letter-spacing:.5px}
input,textarea,select,button{font:inherit;color:var(--fg);background:#0c0e12;border:1px solid var(--line);border-radius:6px;padding:6px 9px}
input:focus,textarea:focus{outline:1px solid var(--acc)}
button{cursor:pointer;background:#1d2530}
button:hover{border-color:var(--acc)}
button.primary{background:var(--acc);color:#062723;border-color:var(--acc);font-weight:600}
main{display:grid;grid-template-columns:280px 1fr 1fr;height:calc(100vh - 53px)}
section{overflow:auto;padding:10px 12px;border-right:1px solid var(--line)}
section:last-child{border-right:0}
h3{margin:14px 0 6px;font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:var(--dim)}
.svc{color:var(--dim);margin-top:10px;font-size:11px;word-break:break-all}
.m{padding:4px 8px;border-radius:5px;cursor:pointer;word-break:break-all}
.m:hover{background:#1b212b}
.m.sel{background:#1d3b38;color:var(--acc)}
textarea{width:100%;height:calc(100vh - 260px);resize:vertical;white-space:pre;overflow:auto}
pre{white-space:pre-wrap;word-break:break-word;background:#0c0e12;border:1px solid var(--line);border-radius:6px;padding:10px;margin:0;max-height:calc(100vh - 150px);overflow:auto}
.row{display:flex;gap:8px;align-items:center;margin-bottom:8px;flex-wrap:wrap}
.grow{flex:1;min-width:120px}
.pill{font-size:11px;padding:2px 8px;border-radius:999px;border:1px solid var(--line);color:var(--dim)}
.pill.ok{color:var(--ok);border-color:#1e4634}.pill.err{color:var(--err);border-color:#4a2020}
.saved{padding:3px 8px;border-radius:5px;cursor:pointer;color:var(--dim)}
.saved:hover{background:#1b212b;color:var(--fg)}
</style></head><body>
<header>
  <b>grpc-lab</b>
  <input id="addr" value="{{ADDR}}" title="target host:port" style="width:190px">
  <input id="token" placeholder="bearer token (optional)" class="grow">
  <button onclick="loadMethods()">Reload</button>
  <span id="hint" class="pill">reflection</span>
</header>
<main>
  <section>
    <h3>Methods</h3><div id="methods">loading…</div>
    <h3>Saved payloads</h3><div id="saved"></div>
  </section>
  <section>
    <div class="row">
      <span id="sel" class="pill">no method selected</span>
      <button onclick="loadTemplate()">Template</button>
      <button onclick="format()">Format</button>
      <button onclick="save()">Save…</button>
      <button class="primary" onclick="invoke()">Invoke</button>
    </div>
    <textarea id="body" spellcheck="false" placeholder="{}"></textarea>
  </section>
  <section>
    <div class="row"><h3 style="margin:0">Response</h3><span id="status" class="pill">idle</span></div>
    <pre id="out">—</pre>
  </section>
</main>
<script>
let method = "";
const $ = id => document.getElementById(id);
const addr = () => $("addr").value.trim();

async function loadMethods(){
  $("methods").textContent = "loading…";
  const r = await (await fetch("/api/methods?addr=" + encodeURIComponent(addr()))).json();
  if (r.error){ $("methods").innerHTML = '<span style="color:var(--err)">' + esc(r.error) + '</span>'; return; }
  $("methods").innerHTML = (r.services||[]).map(s =>
    '<div class="svc">' + esc(s.name) + '</div>' +
    (s.methods||[]).map(m => '<div class="m" data-m="' + esc(m) + '">' + esc(m.split(".").pop()) + '</div>').join("")
  ).join("");
  document.querySelectorAll(".m").forEach(el => el.onclick = () => pick(el));
}

function pick(el){
  document.querySelectorAll(".m").forEach(x => x.classList.remove("sel"));
  el.classList.add("sel");
  method = el.dataset.m;
  $("sel").textContent = method;
  loadTemplate();
}

async function loadTemplate(){
  if(!method) return;
  const r = await (await fetch("/api/template?addr=" + encodeURIComponent(addr()) + "&method=" + encodeURIComponent(method))).json();
  if(!$("body").value.trim() || confirmReplace()) $("body").value = r.template || "{}";
}
// Only clobber a non-empty editor when the developer says so -- losing a
// hand-built cart tree to a stray click is worse than an extra prompt.
function confirmReplace(){ return confirm("Replace the current request body with the template?"); }

function format(){
  try { $("body").value = JSON.stringify(JSON.parse($("body").value), null, 2); }
  catch(e){ setStatus("invalid JSON: " + e.message, false); }
}

async function invoke(){
  if(!method){ setStatus("pick a method first", false); return; }
  setStatus("calling…", null);
  $("out").textContent = "";
  const r = await (await fetch("/api/invoke", {
    method: "POST", headers: {"Content-Type":"application/json"},
    body: JSON.stringify({addr: addr(), method, payload: $("body").value, token: $("token").value.trim()})
  })).json();
  $("out").textContent = r.output || r.error || "(no output)";
  setStatus((r.ok ? "ok" : "error") + " · " + (r.ms||0) + "ms", r.ok);
}

async function save(){
  const name = prompt("save as (letters, digits, . _ -):");
  if(!name) return;
  const r = await (await fetch("/api/payloads", {
    method:"POST", headers:{"Content-Type":"application/json"},
    body: JSON.stringify({name, body: $("body").value})
  })).json();
  if(r.error){ setStatus(r.error, false); return; }
  setStatus("saved " + r.saved, true); loadSaved();
}

async function loadSaved(){
  const r = await (await fetch("/api/payloads")).json();
  $("saved").innerHTML = (r.payloads||[]).map(p =>
    '<div class="saved" data-b="' + esc(p.body) + '">' + esc(p.name) + '</div>').join("") || '<span style="color:var(--dim)">none</span>';
  document.querySelectorAll(".saved").forEach(el => el.onclick = () => {
    $("body").value = el.dataset.b;
    try { format(); } catch(e){}
  });
}

function setStatus(t, ok){
  const s = $("status"); s.textContent = t;
  s.className = "pill" + (ok === true ? " ok" : ok === false ? " err" : "");
}
function esc(s){ return String(s).replace(/[&<>"]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;"}[c])); }

loadMethods(); loadSaved();
</script></body></html>`
