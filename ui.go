package main

// indexHTML is the whole UI: three panes (methods, request, response), no build
// step, no dependencies. {{ADDR}} is substituted with the default target.
const indexHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>grpc-lab</title>
<style>
:root{--bg:#0f1115;--panel:#161a21;--line:#252b36;--fg:#d7dce5;--dim:#8b95a7;--acc:#5eead4;--err:#f87171;--ok:#4ade80;--warn:#fbbf24}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:13px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace}
header{display:flex;gap:8px;align-items:center;padding:10px 14px;border-bottom:1px solid var(--line);background:var(--panel)}
header b{color:var(--acc);letter-spacing:.5px}
input,textarea,select,button{font:inherit;color:var(--fg);background:#0c0e12;border:1px solid var(--line);border-radius:6px;padding:6px 9px}
input:focus,textarea:focus{outline:1px solid var(--acc)}
button{cursor:pointer;background:#1d2530}
button:hover{border-color:var(--acc)}
button.primary{background:var(--acc);color:#062723;border-color:var(--acc);font-weight:600}
button.small{padding:2px 7px;font-size:11px}
label.chk{color:var(--dim);font-size:11px;display:flex;gap:4px;align-items:center;white-space:nowrap}
main{display:grid;grid-template-columns:300px 1fr 1fr;height:calc(100vh - 53px)}
section{overflow:auto;padding:10px 12px;border-right:1px solid var(--line)}
section:last-child{border-right:0}
h3{margin:14px 0 6px;font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:var(--dim);display:flex;gap:8px;align-items:center}
.svc{color:var(--dim);margin-top:10px;font-size:11px;word-break:break-all}
.m{padding:4px 8px;border-radius:5px;cursor:pointer;word-break:break-all;display:flex;gap:6px;align-items:center}
.m:hover{background:#1b212b}
.m.sel{background:#1d3b38;color:var(--acc)}
.badge{font-size:9px;padding:0 5px;border-radius:4px;border:1px solid var(--line);color:var(--warn);flex:none}
textarea{width:100%;resize:vertical;white-space:pre;overflow:auto;tab-size:2}
#body{height:calc(100vh - 330px);min-height:160px}
textarea.aux{height:64px;font-size:12px}
pre{white-space:pre-wrap;word-break:break-word;background:#0c0e12;border:1px solid var(--line);border-radius:6px;padding:10px;margin:0;max-height:calc(100vh - 150px);overflow:auto}
.row{display:flex;gap:8px;align-items:center;margin-bottom:8px;flex-wrap:wrap}
.grow{flex:1;min-width:120px}
.pill{font-size:11px;padding:2px 8px;border-radius:999px;border:1px solid var(--line);color:var(--dim);word-break:break-all}
.pill.ok{color:var(--ok);border-color:#1e4634}.pill.err{color:var(--err);border-color:#4a2020}
.saved{padding:3px 8px;border-radius:5px;cursor:pointer;color:var(--dim);display:flex;justify-content:space-between;gap:6px}
.saved:hover{background:#1b212b;color:var(--fg)}
.saved .x{color:var(--dim);opacity:.5}.saved .x:hover{color:var(--err);opacity:1}
.hist{padding:3px 8px;border-radius:5px;cursor:pointer;color:var(--dim);font-size:11px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.hist:hover{background:#1b212b;color:var(--fg)}
.hist.ok::before{content:"● ";color:var(--ok)}.hist.err::before{content:"● ";color:var(--err)}
.tabs{display:flex;gap:2px}.tabs button{border-radius:6px 6px 0 0;border-bottom:0}.tabs button.on{background:#0c0e12;color:var(--acc)}
.anyrow{display:flex;gap:6px;align-items:center;margin:4px 0;font-size:12px}
.anyrow .p{color:var(--warn);flex:none}
details{margin-bottom:8px}summary{cursor:pointer;color:var(--dim);font-size:11px;text-transform:uppercase;letter-spacing:.08em}
.anyrow .p.bad{color:var(--err)}
details.t{margin:0}details.t>summary{cursor:pointer;list-style:none;text-transform:none;font-size:inherit;letter-spacing:0;color:inherit;white-space:pre}
details.t>summary::before{content:"\25BC";display:inline-block;width:14px;font-size:9px;color:var(--dim)}details.t:not([open])>summary::before{content:"\25B6"}
details.t:not([open])>summary .cl{display:inline}details.t>summary .cl{display:none}details.t>summary .cnt{color:var(--dim);font-size:11px}details.t[open]>summary .cnt{display:none}
.tb{margin-left:6px;padding-left:22px;border-left:1px solid var(--line)}.tb>div{white-space:pre}.tb>div.leaf,.te{padding-left:14px}.te{white-space:pre}
.leaf .v:hover{outline:1px dashed var(--acc);cursor:text}
#reqtree{display:none;border:1px solid var(--line);border-radius:6px;background:#0c0e12;padding:8px;height:calc(100vh - 330px);min-height:160px;overflow:auto}
.k{color:#7dd3fc}.s{color:#86efac}.n{color:#fb923c}.b{color:#f0abfc}
kbd{font-size:10px;color:var(--dim)}
.types{color:var(--dim);font-size:11px}
</style></head><body>
<header>
  <b>grpc-lab</b>
  <input id="addr" value="{{ADDR}}" title="target host:port" style="width:190px">
  <label class="chk"><input type="checkbox" id="tls">TLS (insecure)</label>
  <input id="token" placeholder="bearer token (optional)" class="grow">
  <button onclick="loadMethods(true)">Reload</button>
  <span id="hint" class="pill">reflection</span>
</header>
<main>
  <section>
    <input id="filter" placeholder="filter methods…" style="width:100%" oninput="renderMethods()">
    <div id="methods" style="margin-top:6px">loading…</div>
    <h3>Saved payloads</h3><div id="saved"></div>
    <h3>History <button class="small" onclick="clearHistory()">clear</button></h3><div id="history"></div>
  </section>
  <section>
    <div class="row">
      <span id="sel" class="pill">no method selected</span>
    </div>
    <div class="row">
      <button onclick="loadTemplate(true)" title="regenerate request skeleton from reflection">Template</button>
      <button onclick="format()">Format</button>
      <button id="reqview" onclick="toggleReq()" title="fold the request; double-click a value to edit it">Tree</button>
      <button onclick="scanAny()" title="find google.protobuf.Any fields and fill them with a concrete type">Any…</button>
      <button onclick="save()">Save…</button>
      <button class="primary" onclick="invoke()">Invoke <kbd>⌃⏎</kbd></button>
      <label class="chk"><input type="checkbox" id="emit">emit defaults</label>
      <label class="chk"><input type="checkbox" id="verbose">verbose</label>
    </div>
    <textarea id="body" spellcheck="false" placeholder="{}"></textarea>
    <div id="reqtree"></div>
    <div id="anybox"></div>
    <details><summary>Metadata headers</summary>
      <textarea id="headers" class="aux" spellcheck="false" placeholder="x-tenant-id: 1&#10;x-request-id: abc"></textarea></details>
    <details><summary>Variables ({{name}} in body)</summary>
      <textarea id="vars" class="aux" spellcheck="false" placeholder="customerId=30000&#10;sellerId=20"></textarea></details>
  </section>
  <section>
    <div class="row">
      <div class="tabs"><button class="on" data-t="out" onclick="tab('out')">Response</button><button data-t="desc" onclick="tab('desc')">Describe</button><button data-t="cmd" onclick="tab('cmd')">grpcurl</button></div>
      <span id="status" class="pill">idle</span>
      <button class="small" id="treebtn" onclick="treeMode=!treeMode;renderOut()">raw</button>
      <button class="small" onclick="foldAll(false)">collapse</button>
      <button class="small" onclick="foldAll(true)">expand</button>
      <button class="small" onclick="copy(lastOut)">copy</button>
    </div>
    <pre id="out">—</pre>
    <pre id="desc" style="display:none">—</pre>
    <pre id="cmd" style="display:none">—</pre>
  </section>
</main>
<datalist id="typelist"></datalist>
<script>
let method = "", services = [], meta = {}, types = [], schema = null, lastOut = "", treeMode = true, reqTree = false;
const $ = id => document.getElementById(id);
const addr = () => $("addr").value.trim();
const conn = () => "addr=" + encodeURIComponent(addr()) + "&tls=" + ($("tls").checked ? 1 : 0);
const api = async (url, opt) => (await fetch(url, opt)).json();

// ---- persistence: small conveniences survive a reload ----
const ls = { get: k => { try { return localStorage.getItem("grpclab:" + k) || ""; } catch(e){ return ""; } },
             set: (k, v) => { try { localStorage.setItem("grpclab:" + k, v); } catch(e){} } };
["addr","token","headers","vars"].forEach(k => { if (ls.get(k)) $(k).value = ls.get(k); $(k).oninput = () => ls.set(k, $(k).value); });
$("tls").checked = ls.get("tls") === "1"; $("tls").onchange = () => { ls.set("tls", $("tls").checked ? "1" : "0"); loadMethods(true); };
$("body").oninput = () => { if (method) ls.set("body:" + method, $("body").value); };
// Tab indents instead of leaving the editor.
$("body").addEventListener("keydown", e => { if (e.key === "Tab"){ e.preventDefault(); const t = e.target, a = t.selectionStart; t.value = t.value.slice(0, a) + "  " + t.value.slice(t.selectionEnd); t.selectionStart = t.selectionEnd = a + 2; } });

async function loadMethods(refresh){
  $("methods").textContent = "loading…"; types = [];
  const r = await api("/api/methods?" + conn());
  if (r.error){ $("methods").innerHTML = '<span style="color:var(--err)">' + esc(r.error) + '</span>'; return; }
  services = r.services || [];
  renderMethods();
  loadTypes(refresh);
}

function renderMethods(){
  const q = $("filter").value.trim().toLowerCase();
  $("methods").innerHTML = services.map(s => {
    const ms = s.methods.filter(m => !q || m.name.toLowerCase().includes(q));
    if (!ms.length) return "";
    return '<div class="svc">' + esc(s.name) + '</div>' + ms.map(m =>
      '<div class="m' + (m.name === method ? " sel" : "") + '" data-m="' + esc(m.name) + '">' + esc(m.name.split(".").pop()) +
      (m.clientStream && m.serverStream ? '<span class="badge">bidi</span>' : m.clientStream ? '<span class="badge">client stream</span>' : m.serverStream ? '<span class="badge">server stream</span>' : '') +
      '</div>').join("");
  }).join("") || '<span style="color:var(--dim)">no match</span>';
  document.querySelectorAll(".m").forEach(el => el.onclick = () => pick(el.dataset.m));
}

// Type index for the Any picker: every message the server knows about.
async function loadTypes(refresh){
  const r = await api("/api/types?" + conn() + (refresh ? "&refresh=1" : ""));
  types = r.types || [];
  $("typelist").innerHTML = types.map(t => '<option value="' + esc(t) + '">').join("");
  $("hint").textContent = "reflection · " + types.length + " types";
}

async function pick(name){
  method = name;
  meta = services.flatMap(s => s.methods).find(m => m.name === name) || {};
  renderMethods();
  $("sel").innerHTML = esc(name) + ' <span style="color:var(--dim)">' + esc(meta.input) + ' → ' + esc(meta.output) + '</span>';
  const saved = ls.get("body:" + name);
  $("body").value = ""; $("anybox").innerHTML = "";
  await loadTemplate(true);
  if (saved) $("body").value = saved;
  if (reqTree) renderReq();
}

async function loadTemplate(replace){
  if(!method) return;
  const r = await api("/api/template?" + conn() + "&method=" + encodeURIComponent(method));
  if (r.error){ setStatus(r.error, false); return; }
  $("desc").textContent = r.describe || "";
  schema = null; api("/api/schema?" + conn() + "&type=" + encodeURIComponent(r.input)).then(sr => { schema = sr.messages || null; });
  if (replace || !$("body").value.trim() || confirm("Replace the current request body with the template?")){
    let t = r.template || "{}";
    if (r.clientStream) t += "\n" + t; // stdin takes many messages: show two so the shape is obvious
    $("body").value = t; ls.set("body:" + method, t); $("anybox").innerHTML = "";
  }
}

function format(){
  const txt = $("body").value;
  try { $("body").value = parseMany(txt).map(o => JSON.stringify(o, null, 2)).join("\n"); ls.set("body:" + method, $("body").value); }
  catch(e){ setStatus("invalid JSON: " + e.message, false); }
  if (reqTree) renderReq();
}
// Client-streaming bodies are several JSON objects back to back; parse them all.
function parseMany(txt){
  const out = []; let s = txt.trim();
  while (s){
    let depth = 0, inStr = false, i = 0;
    for (; i < s.length; i++){
      const c = s[i];
      if (inStr){ if (c === "\\") i++; else if (c === '"') inStr = false; continue; }
      if (c === '"') inStr = true; else if (c === "{" || c === "[") depth++; else if (c === "}" || c === "]") { depth--; if (!depth){ i++; break; } }
    }
    out.push(JSON.parse(s.slice(0, i))); s = s.slice(i).trim();
  }
  return out;
}

// ---- google.protobuf.Any ----
// The schema (from reflection) says which positions are Any. Every such
// position in the body is listed; one without "@type" is flagged, because
// grpcurl's error for that case never names the field. Without a schema, fall
// back to scanning for "@type" keys.
function findAny(objs){
  const found = [];
  const walkSchema = (v, type, path) => {
    const fields = schema[type]; if (!fields || !v || typeof v !== "object") return;
    for (const k of Object.keys(v)){
      const f = fields[k]; if (!f) continue;
      const each = (x, p) => {
        if (f.kind === "any"){ if (x == null) return; // null = unset, grpcurl accepts it
          const ok = typeof x === "object" && "@type" in x;
          found.push({ path: p, type: ok ? String(x["@type"]).replace("type.googleapis.com/", "") : "", missing: !ok }); }
        else if (f.kind === "msg") walkSchema(x, f.type, p);
      };
      const p = path ? path + "." + k : k;
      if (f.map && v[k] && typeof v[k] === "object") Object.keys(v[k]).forEach(mk => each(v[k][mk], p + "." + mk));
      else if (f.repeated && Array.isArray(v[k])) v[k].forEach((x, i) => each(x, p + "[" + i + "]"));
      else each(v[k], p);
    }
  };
  const walkPlain = (v, path) => {
    if (Array.isArray(v)) v.forEach((x, i) => walkPlain(x, path + "[" + i + "]"));
    else if (v && typeof v === "object"){
      if ("@type" in v) found.push({ path, type: String(v["@type"]).replace("type.googleapis.com/", ""), missing: false });
      Object.keys(v).forEach(k => k !== "@type" && walkPlain(v[k], path ? path + "." + k : k));
    }
  };
  objs.forEach((o, i) => { const pre = objs.length > 1 ? "#" + i : "";
    schema && meta.input && schema[meta.input] ? walkSchema(o, meta.input, pre) : walkPlain(o, pre); });
  return found;
}
function scanAny(){
  let objs; try { objs = parseMany($("body").value); } catch(e){ setStatus("invalid JSON: " + e.message, false); return []; }
  const found = findAny(objs);
  if (!found.length){ $("anybox").innerHTML = '<div class="types">no google.protobuf.Any fields in body' + (schema ? "" : " (schema not loaded; scanned for @type)") + '</div>'; return found; }
  $("anybox").innerHTML = found.map((f, i) =>
    '<div class="anyrow"><span class="p' + (f.missing ? ' bad' : '') + '" title="' + (f.missing ? 'missing @type -- grpcurl will reject this' : 'google.protobuf.Any') + '">' + esc(f.path || "$") + '</span>' +
    '<input list="typelist" class="grow" id="any' + i + '" value="' + esc(f.type) + '" placeholder="pkg.Message (or delete the field)">' +
    '<button class="small" onclick="fillAny(' + i + ',' + JSON.stringify(f.path) + ')">fill</button></div>').join("");
  return found;
}
async function fillAny(i, path){
  const t = $("any" + i).value.trim(); if (!t) return;
  const r = await api("/api/describe?" + conn() + "&symbol=" + encodeURIComponent(t));
  if (r.error){ setStatus(r.error, false); return; }
  let tpl = {}; try { tpl = JSON.parse(r.template || "{}"); } catch(e){}
  const objs = parseMany($("body").value);
  let idx = 0, p = path;
  const m = /^#(\d+)\.?/.exec(p); if (m){ idx = +m[1]; p = p.slice(m[0].length); }
  setPath(objs[idx], p, Object.assign({ "@type": "type.googleapis.com/" + t }, tpl));
  $("body").value = objs.map(o => JSON.stringify(o, null, 2)).join("\n"); ls.set("body:" + method, $("body").value); if (reqTree) renderReq();
  setStatus("filled " + (path || "$") + " with " + t, true); scanAny();
}
function setPath(root, path, val){
  const keys = path ? path.replace(/\[(\d+)\]/g, ".$1").split(".").filter(Boolean) : [];
  if (!keys.length){ Object.keys(root).forEach(k => delete root[k]); Object.assign(root, val); return; }
  let o = root; keys.slice(0, -1).forEach(k => o = o[k]); o[keys[keys.length - 1]] = val;
}

// ---- invoke ----
function substitute(txt){
  const vars = {};
  $("vars").value.split("\n").forEach(l => { const i = l.indexOf("="); if (i > 0) vars[l.slice(0, i).trim()] = l.slice(i + 1).trim(); });
  return txt.replace(/\{\{\s*([\w.-]+)\s*\}\}/g, (m, k) => k in vars ? vars[k] : m);
}
async function invoke(){
  if(!method){ setStatus("pick a method first", false); return; }
  const payload = substitute($("body").value);
  let objs; try { objs = parseMany(payload); } catch(e){ setStatus("invalid JSON: " + e.message, false); return; }
  const missing = findAny(objs).filter(f => f.missing);
  if ($("anybox").innerHTML) scanAny();
  if (missing.length){ setStatus("Any without @type: " + missing.map(f => f.path).join(", ") + " -- fill it or delete the field", false); return; }
  setStatus("calling…", null); $("out").textContent = ""; tab("out");
  const r = await api("/api/invoke", { method: "POST", headers: {"Content-Type":"application/json"},
    body: JSON.stringify({ addr: addr(), method, payload, token: $("token").value.trim(), headers: $("headers").value,
      tls: $("tls").checked, emitDefaults: $("emit").checked, verbose: $("verbose").checked }) });
  lastOut = r.output || r.error || "(no output)"; renderOut();
  $("cmd").textContent = r.command || "";
  setStatus((r.ok ? "ok" : "error" + codeName(r.output)) + " · " + (r.ms||0) + "ms", r.ok);
  loadHistory();
}
const CODES = ["OK","CANCELLED","UNKNOWN","INVALID_ARGUMENT","DEADLINE_EXCEEDED","NOT_FOUND","ALREADY_EXISTS","PERMISSION_DENIED","RESOURCE_EXHAUSTED","FAILED_PRECONDITION","ABORTED","OUT_OF_RANGE","UNIMPLEMENTED","INTERNAL","UNAVAILABLE","DATA_LOSS","UNAUTHENTICATED"];
function codeName(out){ const m = /"code":\s*(\d+)/.exec(out || ""); return m && CODES[+m[1]] ? " " + CODES[+m[1]] : ""; }

// ---- saved payloads & history ----
async function save(){
  const name = prompt("save as (letters, digits, . _ -):", method ? method.split(".").pop() : "");
  if(!name) return;
  const r = await api("/api/payloads", { method:"POST", headers:{"Content-Type":"application/json"}, body: JSON.stringify({name, body: $("body").value}) });
  if(r.error){ setStatus(r.error, false); return; }
  setStatus("saved " + r.saved, true); loadSaved();
}
async function loadSaved(){
  const r = await api("/api/payloads");
  $("saved").innerHTML = (r.payloads||[]).map((p, i) =>
    '<div class="saved" data-i="' + i + '"><span>' + esc(p.name) + '</span><span class="x" title="delete">×</span></div>').join("") || '<span style="color:var(--dim)">none</span>';
  document.querySelectorAll(".saved").forEach(el => {
    const p = r.payloads[+el.dataset.i];
    el.onclick = () => { $("body").value = p.body; format(); };
    el.querySelector(".x").onclick = async e => { e.stopPropagation(); if(!confirm("delete " + p.name + "?")) return;
      await fetch("/api/payloads?name=" + encodeURIComponent(p.name), { method: "DELETE" }); loadSaved(); };
  });
}
async function loadHistory(){
  const r = await api("/api/history");
  $("history").innerHTML = (r.history||[]).map((h, i) =>
    '<div class="hist ' + (h.ok ? "ok" : "err") + '" data-i="' + i + '" title="' + esc(h.at) + '">' + esc(h.at.slice(11,19)) + ' ' + esc(h.method.split(".").slice(-2).join(".")) + ' ' + h.ms + 'ms</div>').join("") || '<span style="color:var(--dim)">none</span>';
  document.querySelectorAll(".hist").forEach(el => el.onclick = async () => {
    const h = r.history[+el.dataset.i];
    if (h.addr !== addr()){ $("addr").value = h.addr; ls.set("addr", h.addr); await loadMethods(); }
    if (!services.flatMap(s => s.methods).some(m => m.name === h.method)){ setStatus("method not on this target: " + h.method, false); return; }
    await pick(h.method); $("body").value = h.payload; format();
  });
}
async function clearHistory(){ await fetch("/api/history", { method: "DELETE" }); loadHistory(); }

// ---- misc ----
function tab(t){ ["out","desc","cmd"].forEach(x => { $(x).style.display = x === t ? "" : "none"; }); document.querySelectorAll(".tabs button").forEach(b => b.classList.toggle("on", b.dataset.t === t)); }
function copy(t){ navigator.clipboard.writeText(t).then(() => setStatus("copied", true)); }
function setStatus(t, ok){ const s = $("status"); s.textContent = t; s.className = "pill" + (ok === true ? " ok" : ok === false ? " err" : ""); }
function esc(s){ return String(s ?? "").replace(/[&<>"]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;"}[c])); }
// Collapsible JSON tree (JSON-Viewer style: arrows, guide lines, trailing
// commas, everything expanded). editable=true makes leaf values double-click
// editable; the change is written back into the editor text.
function renderOut(){
  $("treebtn").textContent = treeMode ? "raw" : "tree";
  let objs = null; if (treeMode) try { objs = parseMany(lastOut); } catch(e){}
  $("out").innerHTML = objs && objs.length ? objs.map(o => tree(o, "", true, "")).join('<hr style="border:0;border-top:1px dashed var(--line)">') : hl(lastOut);
}
function tree(v, key, last, path, editable){
  const comma = last ? "" : ",", label = key === "" ? "" : '<span class="k">"' + esc(key) + '"</span>: ';
  if (v === null || typeof v !== "object")
    return '<div class="leaf">' + label + '<span class="v" data-p="' + esc(path) + '"' + (editable ? ' ondblclick="editLeaf(this)"' : '') + '>' + hl(JSON.stringify(v)) + '</span>' + comma + '</div>';
  const arr = Array.isArray(v), keys = Object.keys(v), o = arr ? "[" : "{", c = arr ? "]" : "}";
  if (!keys.length) return '<div class="te">' + label + o + c + comma + '</div>';
  return '<details class="t" open><summary>' + label + o + '<span class="cnt"> \u2026 ' + keys.length + ' </span><span class="cl">' + c + comma + '</span></summary><div class="tb">' +
    keys.map((k, i) => tree(v[k], arr ? "" : k, i === keys.length - 1, path + (arr ? "[" + k + "]" : (path ? "." : "") + k), editable)).join("") +
    '</div><div class="te">' + c + comma + '</div></details>';
}
function foldAll(open){ document.querySelectorAll("#out details, #reqtree details").forEach(d => d.open = open); }

// Request tree: same renderer over the editor's JSON; edits round-trip.
function toggleReq(){ reqTree = !reqTree; $("reqview").textContent = reqTree ? "Edit" : "Tree"; $("body").style.display = reqTree ? "none" : ""; $("reqtree").style.display = reqTree ? "block" : "none"; if (reqTree) renderReq(); }
function renderReq(){
  let objs; try { objs = parseMany($("body").value); } catch(e){ $("reqtree").innerHTML = '<span style="color:var(--err)">' + esc("invalid JSON: " + e.message) + '</span>'; return; }
  $("reqtree").innerHTML = objs.map((o, i) => tree(o, "", true, objs.length > 1 ? "#" + i : "", true)).join('<hr style="border:0;border-top:1px dashed var(--line)">');
}
function editLeaf(el){
  const cur = el.textContent, nv = prompt("new value (JSON literal, e.g. \"text\", 42, true, null):", cur);
  if (nv === null || nv === cur) return;
  let val; try { val = JSON.parse(nv); } catch(e){ setStatus("not a JSON literal: " + nv, false); return; }
  const objs = parseMany($("body").value);
  let idx = 0, p = el.dataset.p; const m = /^#(\d+)\.?/.exec(p); if (m){ idx = +m[1]; p = p.slice(m[0].length); }
  setPath(objs[idx], p, val);
  $("body").value = objs.map(o => JSON.stringify(o, null, 2)).join("\n"); ls.set("body:" + method, $("body").value); renderReq();
}
// Cheap JSON colouring on the escaped text; grpcurl already pretty-prints.
function hl(s){ return String(s ?? "").replace(/[&<>]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;"}[c])).replace(/("(?:\\.|[^"\\])*")(\s*:)?|\b(true|false|null)\b|-?\b\d+(\.\d+)?([eE][+-]?\d+)?\b/g,
  (m, str, colon, kw) => str ? (colon ? '<span class="k">' + str + '</span>' + colon : '<span class="s">' + str + '</span>') : kw ? '<span class="b">' + kw + '</span>' : '<span class="n">' + m + '</span>'); }
document.addEventListener("keydown", e => { if ((e.ctrlKey || e.metaKey) && e.key === "Enter"){ e.preventDefault(); invoke(); } });

loadMethods(); loadSaved(); loadHistory();
</script></body></html>`
