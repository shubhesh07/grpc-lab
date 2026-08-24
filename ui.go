package main

// indexHTML is the whole UI: three panes (methods, request, response), no build
// step, no dependencies. {{ADDR}} is substituted with the default target.
const indexHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>grpc-lab</title>
<link rel="icon" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Crect width='32' height='32' rx='7' fill='%230f1115'/%3E%3Ctext x='16' y='22' font-size='17' font-family='Menlo,monospace' font-weight='700' text-anchor='middle' fill='%235eead4'%3Eg%3C/text%3E%3C/svg%3E">
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
.rtabs{display:flex;gap:4px;align-items:center;margin-bottom:8px;flex-wrap:wrap}.rtab{padding:3px 8px;border:1px solid var(--line);border-radius:6px;cursor:pointer;color:var(--dim);font-size:12px;max-width:220px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.rtab.on{color:var(--acc);border-color:var(--acc);background:#0c0e12}.rtab .x{margin-left:6px;opacity:.5}.rtab .x:hover{opacity:1;color:var(--err)}
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
.pen,.tog,.cnt,.del,details.t>summary::before{user-select:none;-webkit-user-select:none}
.del{color:var(--dim);opacity:0;cursor:pointer;margin-left:8px;font-size:11px}.leaf:hover>.del,summary:hover>.del,.te:hover>.del{opacity:.6}.del:hover{opacity:1;color:var(--err)}
.leaf:hover>.pen,summary:hover>.pen,.te:hover>.pen{opacity:.6}
.leaf.selrow,.te.selrow,details.t>summary.selrow{background:#1d3b38;border-radius:4px}
.pen{color:var(--dim);opacity:.35;cursor:pointer;margin-left:8px;font-size:11px}.pen:hover{opacity:1;color:var(--acc)}details.t>summary:hover .pen,.te:hover .pen{opacity:.8}
.ed{display:block;width:100%;min-height:80px;margin:4px 0;white-space:pre;font:inherit}.edrow{margin:2px 0 6px}.edrow button{margin-right:6px}
.leaf input.iv{padding:1px 4px;font:inherit;min-width:120px}
.hintline{color:var(--dim);font-size:11px;margin:4px 0}
.tog{color:var(--dim);opacity:0;cursor:pointer;margin-right:6px;font-size:11px}.leaf:hover>.tog,summary:hover>.tog,.te:hover>.tog{opacity:.6}.tog:hover,.tog.on{opacity:1;color:var(--warn)}.off,.k.off{color:var(--dim);text-decoration:line-through;opacity:.7}
#body{height:calc(100vh - 360px);min-height:160px;font-size:13px;line-height:1.45;tab-size:2}
#lint.bad{color:var(--err)}#lint.good{color:var(--ok)}
#reqtree{border:1px solid var(--line);border-radius:6px;background:#0c0e12;padding:8px;height:calc(100vh - 330px);min-height:160px;overflow:auto}
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
    <h3>Type sources <button class="small" onclick="addTypeSource()" title="pull descriptors from another server so its types decode inside Any">+ server</button> <button class="small" onclick="$('psfile').click()" title="upload a .protoset built from .proto files (buf build -o x.protoset)">+ file</button><input type="file" id="psfile" accept=".protoset,.binpb,.pb,.desc" style="display:none" onchange="uploadTypeSource(this)"></h3><div id="typesources"></div>
    <h3>History <button class="small" onclick="clearHistory()">clear</button></h3><div id="history"></div>
  </section>
  <section>
    <div class="rtabs" id="rtabs"></div>
    <div class="row">
      <span id="sel" class="pill">no method selected</span>
    </div>
    <div class="row">
      <button onclick="loadTemplate(true)" title="regenerate request skeleton from reflection">Template</button>
      <button id="reqview" onclick="toggleReq()" title="switch between text editor and foldable tree">Tree</button>
      <button id="fmtbtn" onclick="format()" title="pretty-print (comments are kept as-is)">Format</button>
      <span id="reqfold" style="display:none"><button class="small" onclick="foldReq(false)">collapse</button> <button class="small" onclick="foldReq(true)">expand</button></span>
      <button onclick="save()">Save…</button>
      <button onclick="pasteCurl()" title="paste a grpcurl command: target, headers, body and method are filled in">Paste grpcurl</button>
      <button class="primary" onclick="invoke()">Invoke <kbd>⌃⏎</kbd></button>
      <label class="chk"><input type="checkbox" id="emit">emit defaults</label>
      <label class="chk"><input type="checkbox" id="verbose">verbose</label>
    </div>
    <textarea id="body" spellcheck="false" placeholder="{}"></textarea>
    <div id="lint" class="hintline"></div>
    <div id="reqtree" style="display:none"></div>
    <div id="anybox"></div>
    <details><summary>Metadata headers</summary>
      <textarea id="headers" class="aux" spellcheck="false" placeholder="x-tenant-id: 1&#10;x-request-id: abc"></textarea></details>
    <details><summary>Variables ({{name}} in body)</summary>
      <textarea id="vars" class="aux" spellcheck="false" placeholder="customerId=123&#10;sellerId=1"></textarea></details>
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
// Request tabs: each is an independent workspace (method + body); the active
// one is what the editor shows. Persisted as one blob.
let tabs = [], cur = 0, selSegs = null, undoStack = [];
try { const t = JSON.parse(ls.get("tabs") || "null"); if (t && t.tabs && t.tabs.length){ tabs = t.tabs; cur = Math.min(t.cur || 0, tabs.length - 1); } } catch(e){}
if (!tabs.length) tabs = [{ method: "", body: "" }];
const $ = id => document.getElementById(id);
const ls = { get: k => { try { return localStorage.getItem("grpclab:" + k) || ""; } catch(e){ return ""; } },
             set: (k, v) => { try { localStorage.setItem("grpclab:" + k, v); } catch(e){} } };
const addr = () => $("addr").value.trim();
const conn = () => "addr=" + encodeURIComponent(addr()) + "&tls=" + ($("tls").checked ? 1 : 0);
const api = async (url, opt) => (await fetch(url, opt)).json();

// ---- persistence: small conveniences survive a reload ----
["addr","token","headers","vars"].forEach(k => { if (ls.get(k)) $(k).value = ls.get(k); $(k).oninput = () => ls.set(k, $(k).value); });
["emit","verbose"].forEach(k => { $(k).checked = ls.get(k) === "1"; $(k).onchange = () => ls.set(k, $(k).checked ? "1" : "0"); });
$("tls").checked = ls.get("tls") === "1"; $("tls").onchange = () => { ls.set("tls", $("tls").checked ? "1" : "0"); loadMethods(true); };
let anyTimer = 0;
$("body").oninput = () => { saveTab(); lint(); clearTimeout(anyTimer); anyTimer = setTimeout(() => scanAny(true), 400); };
// Tab indents instead of leaving the editor.
$("body").addEventListener("keydown", e => {
  const t = e.target, a = t.selectionStart;
  if (e.key === "Tab"){ e.preventDefault(); t.value = t.value.slice(0, a) + "  " + t.value.slice(t.selectionEnd); t.selectionStart = t.selectionEnd = a + 2; t.dispatchEvent(new Event("input")); }
  else if (e.key === "Enter" && !e.ctrlKey && !e.metaKey){
    e.preventDefault();
    const line = t.value.slice(0, a).split("\n").pop(), ind = /^\s*/.exec(line)[0], prev = t.value[a - 1], next = t.value[a];
    const open = prev === "{" || prev === "[", close = open && (next === "}" || next === "]");
    const ins = "\n" + ind + (open ? "  " : "") + (close ? "\n" + ind : "");
    t.value = t.value.slice(0, a) + ins + t.value.slice(t.selectionEnd);
    t.selectionStart = t.selectionEnd = a + 1 + ind.length + (open ? 2 : 0);
    t.dispatchEvent(new Event("input"));
  }
});
function lint(){
  const txt = $("body").value, el = $("lint");
  if (!txt.trim()){ el.textContent = ""; el.className = "hintline"; return; }
  try { const n = parseMany(txt).length; el.textContent = "valid JSON" + (n > 1 ? " \u00b7 " + n + " messages" : ""); el.className = "hintline good"; }
  catch(e){ const m = /position (\d+)/.exec(e.message); const ln = m ? stripComments(txt).slice(0, +m[1]).split("\n").length : null; el.textContent = "invalid JSON: " + e.message + (ln ? " (around line " + ln + ")" : ""); el.className = "hintline bad"; }
}

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
  const saved = tabs[cur].method === name ? tabs[cur].body : "";
  $("body").value = ""; $("anybox").innerHTML = ""; $("desc").textContent = "";
  await loadTemplate(true);
  if (method !== name) return;              // user already clicked elsewhere
  if (saved) $("body").value = saved;
  saveTab(); lint(); renderReq();
}

async function loadTemplate(replace){
  if(!method) return;
  const name = method; schema = null;
  const r = await api("/api/template?" + conn() + "&method=" + encodeURIComponent(name));
  if (method !== name) return;              // stale response for a method no longer selected
  if (r.error){ setStatus(r.error, false); return; }
  $("desc").textContent = r.describe || "";
  api("/api/schema?" + conn() + "&type=" + encodeURIComponent(r.input)).then(sr => { if (method === name){ schema = sr.messages || null; scanAny(true); } });
  if (replace || !$("body").value.trim() || confirm("Replace the current request body with the template?")){
    let t = r.template || "{}";
    if (r.clientStream) t += "\n" + t; // stdin takes many messages: show two so the shape is obvious
    $("body").value = t; saveTab(); $("anybox").innerHTML = ""; lint(); renderReq();
  }
}

function format(){
  const txt = $("body").value;
  if (hasComments(txt)){ try { parseMany(txt); } catch(e){ setStatus("invalid JSON: " + e.message, false); } if (reqTree) renderReq(); return; } // keep the developer's comments
  try { $("body").value = parseMany(txt).map(o => JSON.stringify(o, null, 2)).join("\n"); saveTab(); }
  catch(e){ setStatus("invalid JSON: " + e.message, false); }
  if (reqTree) renderReq();
}
// stripComments removes // and /* */ comments outside strings, so a developer
// can comment out part of a request instead of deleting it.
function stripComments(txt){
  let out = "", inStr = false;
  for (let i = 0; i < txt.length; i++){
    const c = txt[i], n = txt[i + 1];
    if (inStr){ out += c; if (c === "\\"){ out += n; i++; } else if (c === '"') inStr = false; continue; }
    if (c === '"'){ inStr = true; out += c; }
    else if (c === "/" && n === "/"){ while (i < txt.length && txt[i] !== "\n") i++; out += "\n"; }
    else if (c === "/" && n === "*"){ i += 2; while (i < txt.length && !(txt[i] === "*" && txt[i + 1] === "/")) i++; i++; }
    else out += c;
  }
  return out;
}
const hasComments = txt => stripComments(txt) !== txt;
// dropDisabled removes keys prefixed with "//" (disabled from the tree view).
function dropDisabled(v){
  if (Array.isArray(v)) return v.map(dropDisabled);
  if (v && typeof v === "object"){ const o = {}; for (const k of Object.keys(v)) if (!k.startsWith("//")) o[k] = dropDisabled(v[k]); return o; }
  return v;
}
// Client-streaming bodies are several JSON objects back to back; parse them all.
function parseMany(txt){
  const out = []; let s = stripComments(txt).trim();
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
  // segs: [msgIndex, key, key, ...] -- arrays, not dotted strings, so map keys
  // containing "." or "[" round-trip. showPath() renders them for display.
  const walkSchema = (v, type, segs) => {
    const fields = schema[type]; if (!fields || !v || typeof v !== "object") return;
    for (const k of Object.keys(v)){
      const f = fields[k]; if (!f || k.startsWith("//")) continue;
      const each = (x, p) => {
        if (f.kind === "any"){ if (x == null) return; // null = unset, grpcurl accepts it
          const ok = typeof x === "object" && "@type" in x;
          found.push({ segs: p, path: showPath(p), type: ok ? String(x["@type"]).replace("type.googleapis.com/", "") : "", missing: !ok }); }
        else if (f.kind === "msg") walkSchema(x, f.type, p);
      };
      const p = segs.concat(k);
      if (f.map && v[k] && typeof v[k] === "object") Object.keys(v[k]).forEach(mk => each(v[k][mk], p.concat(mk)));
      else if (f.repeated && Array.isArray(v[k])) v[k].forEach((x, i) => each(x, p.concat(i)));
      else each(v[k], p);
    }
  };
  const walkPlain = (v, segs) => {
    if (Array.isArray(v)) v.forEach((x, i) => walkPlain(x, segs.concat(i)));
    else if (v && typeof v === "object"){
      if ("@type" in v) found.push({ segs, path: showPath(segs), type: String(v["@type"]).replace("type.googleapis.com/", ""), missing: false });
      Object.keys(v).forEach(k => k !== "@type" && walkPlain(v[k], segs.concat(k)));
    }
  };
  objs.forEach((o, i) => schema && meta.input && schema[meta.input] ? walkSchema(o, meta.input, [i]) : walkPlain(o, [i]));
  return found;
}
// Runs automatically whenever the body changes; quiet=true keeps a half-typed
// body from spamming the status pill.
function scanAny(quiet){
  let objs; try { objs = parseMany($("body").value); } catch(e){ if (!quiet) setStatus("invalid JSON: " + e.message, false); return []; }
  const found = findAny(objs);
  if (!found.length){ $("anybox").innerHTML = ""; return found; }
  $("anybox").innerHTML = found.map((f, i) =>
    '<div class="anyrow"><span class="p' + (f.missing ? ' bad' : '') + '" title="' + (f.missing ? 'missing @type -- grpcurl will reject this' : 'google.protobuf.Any') + '">' + esc(f.path || "$") + '</span>' +
    '<input list="typelist" class="grow" id="any' + i + '" value="' + esc(f.type) + '" placeholder="pkg.Message (or delete the field)">' +
    '<button class="small" data-s="' + esc(JSON.stringify(f.segs)) + '" onclick="fillAny(' + i + ', JSON.parse(this.dataset.s))">fill</button></div>').join("");
  return found;
}
async function fillAny(i, segs){
  const t = $("any" + i).value.trim(); if (!t) return;
  const r = await api("/api/describe?" + conn() + "&symbol=" + encodeURIComponent(t));
  if (r.error){ setStatus(r.error, false); return; }
  let tpl = {}; try { tpl = JSON.parse(r.template || "{}"); } catch(e){}
  const objs = parseMany($("body").value);
  let cur = objs[segs[0]]; segs.slice(1).forEach(k => cur = cur && cur[k]);
  // Keep whatever the developer already typed: their values win over the template.
  const keep = cur && typeof cur === "object" ? Object.fromEntries(Object.entries(cur).filter(([k]) => k !== "@type" && k !== "value")) : {};
  setPath(objs, segs, Object.assign({ "@type": "type.googleapis.com/" + t }, deepMerge(tpl, keep)));
  setBody(objs); setStatus("filled " + showPath(segs) + " with " + t, true);
}
function deepMerge(base, over){
  if (Array.isArray(over) || over === null || typeof over !== "object" || typeof base !== "object" || base === null || Array.isArray(base)) return over;
  const out = Object.assign({}, base);
  for (const k of Object.keys(over)) out[k] = k in base ? deepMerge(base[k], over[k]) : over[k];
  return out;
}
// segs[0] is the message index (client streams carry several); the rest walk into it.
function setPath(objs, segs, val){
  if (segs.length === 1){ const root = objs[segs[0]]; Object.keys(root).forEach(k => delete root[k]); Object.assign(root, val); return; }
  let o = objs[segs[0]]; segs.slice(1, -1).forEach(k => o = o[k]); o[segs[segs.length - 1]] = val;
}
function showPath(segs){
  const s = segs.slice(1).map(k => typeof k === "number" ? "[" + k + "]" : "." + k).join("").replace(/^\./, "");
  return (segs[0] ? "#" + segs[0] + " " : "") + (s || "$");
}

// ---- invoke ----
function substitute(txt){
  const vars = {};
  $("vars").value.split("\n").forEach(l => { const i = l.indexOf("="); if (i > 0) vars[l.slice(0, i).trim()] = l.slice(i + 1).trim(); });
  return txt.replace(/\{\{\s*([\w.-]+)\s*\}\}/g, (m, k) => k in vars ? vars[k] : m);
}
async function invoke(){
  if(!method){ setStatus("pick a method first", false); return; }
  let objs; try { objs = parseMany(substitute($("body").value)); } catch(e){ setStatus("invalid JSON: " + e.message, false); return; }
  objs = objs.map(dropDisabled);
  const payload = objs.map(o => JSON.stringify(o, null, 2)).join("\n");
  const missing = findAny(objs).filter(f => f.missing);
  scanAny(true);
  if (missing.length){ setStatus("Any without @type: " + missing.map(f => f.path).join(", ") + " -- fill it or delete the field", false); return; }
  setStatus("calling…", null); $("out").textContent = ""; tab("out");
  const r = await api("/api/invoke", { method: "POST", headers: {"Content-Type":"application/json"},
    body: JSON.stringify({ addr: addr(), method, payload, token: $("token").value.trim(), headers: $("headers").value,
      tls: $("tls").checked, emitDefaults: $("emit").checked, verbose: $("verbose").checked }) });
  lastOut = r.output || r.error || "(no output)"; renderOut();
  const un = /"@error":\s*"([\w.]+) is not recognized/.exec(lastOut);
  if (un) setStatus("response has an Any of " + un[1] + " the target cannot reflect \u2014 add its server under Type sources", false);
  $("cmd").textContent = r.command || "";
  if (!un) setStatus((r.ok ? "ok" : "error" + codeName(r.output)) + " · " + (r.ms||0) + "ms", r.ok);
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
// Type sources: descriptor sets from other servers, so an Any carrying e.g.
// customer-service's CustomerInfo decodes even when the target can't reflect it.
async function loadTypeSources(){
  const r = await api("/api/typesources");
  $("typesources").innerHTML = (r.sources||[]).map(n => '<div class="saved" data-n="' + esc(n) + '"><span title="descriptors from ' + esc(n.replace(/_/g, ":")) + '">' + esc(n.replace(/_/g, ":")) + '</span><span class="x" title="remove">×</span></div>').join("") ||
    '<span style="color:var(--dim)" title="if a response shows @error ... is not recognized, add that type\'s server here">none</span>';
  document.querySelectorAll("#typesources .x").forEach(x => x.onclick = async () => { await fetch("/api/typesources?name=" + encodeURIComponent(x.parentNode.dataset.n), { method: "DELETE" }); loadTypeSources(); loadTypes(true); });
}
async function addTypeSource(){
  const a = prompt("host:port of the server whose types should decode (uses the TLS checkbox):"); if (!a) return;
  setStatus("fetching descriptors from " + a + "\u2026", null);
  const r = await api("/api/typesources", { method: "POST", headers: {"Content-Type":"application/json"}, body: JSON.stringify({ addr: a.trim(), tls: $("tls").checked }) });
  if (r.error){ setStatus(r.error, false); return; }
  setStatus("added type source " + a, true); loadTypeSources(); loadTypes(true);
}
async function uploadTypeSource(inp){
  const f = inp.files[0]; if (!f) return; inp.value = "";
  const name = f.name.replace(/\.(protoset|binpb|pb|desc)$/, "");
  const r = await api("/api/typesources?name=" + encodeURIComponent(name), { method: "PUT", body: f });
  if (r.error){ setStatus(r.error, false); return; }
  setStatus("added type source " + name, true); loadTypeSources(); loadTypes(true);
}
async function clearHistory(){ await fetch("/api/history", { method: "DELETE" }); loadHistory(); }

// ---- request tabs ----
function saveTab(){ tabs[cur] = { method, body: $("body").value }; ls.set("tabs", JSON.stringify({ tabs, cur })); renderTabs(); }
function renderTabs(){
  $("rtabs").innerHTML = tabs.map((t, i) => '<span class="rtab' + (i === cur ? " on" : "") + '" data-i="' + i + '">' + esc(t.method ? t.method.split(".").pop() : "new") +
    (tabs.length > 1 ? '<span class="x" title="close">×</span>' : '') + '</span>').join("") + '<button class="small" onclick="newTab()" title="new request tab">+</button>';
  document.querySelectorAll(".rtab").forEach(el => {
    el.onclick = () => switchTab(+el.dataset.i);
    const x = el.querySelector(".x"); if (x) x.onclick = e => { e.stopPropagation(); closeTab(+el.dataset.i); };
  });
}
async function switchTab(i){
  if (i === cur) return;
  tabs[cur] = { method, body: $("body").value }; cur = i;
  const t = tabs[cur]; method = ""; $("anybox").innerHTML = "";
  if (t.method && services.flatMap(s => s.methods).some(m => m.name === t.method)) await pick(t.method); else { $("body").value = t.body; $("sel").textContent = "no method selected"; renderMethods(); renderReq(); }
  saveTab();
}
function newTab(){ tabs[cur] = { method, body: $("body").value }; tabs.push({ method: "", body: "" }); switchTab(tabs.length - 1); }
function closeTab(i){ tabs.splice(i, 1); if (cur >= i) cur = Math.max(0, cur - 1); const t = tabs[cur]; cur = -1; switchTab(tabs.indexOf(t)); }

// ---- paste a grpcurl command ----
// Tokenises like a shell (quotes, backslash-newline), then picks out what the
// UI understands. Anything else (-cacert, -proto, ...) is ignored, not fatal.
function shellSplit(cmd){
  const out = []; let tok = "", q = null, has = false;
  cmd = cmd.replace(/\\\r?\n/g, " ");
  for (let i = 0; i < cmd.length; i++){
    const c = cmd[i];
    if (q){ if (c === q) q = null; else if (q === '"' && c === "\\" && i + 1 < cmd.length) tok += cmd[++i]; else tok += c; continue; }
    if (c === "'" || c === '"'){ q = c; has = true; }
    else if (c === "\\" && i + 1 < cmd.length) { tok += cmd[++i]; has = true; }
    else if (/\s/.test(c)){ if (has || tok){ out.push(tok); tok = ""; has = false; } }
    else { tok += c; has = true; }
  }
  if (has || tok) out.push(tok);
  return out;
}
function parseGrpcurl(cmd){
  const a = shellSplit(cmd.trim()); const r = { headers: [], tls: false, body: "", token: "" };
  let i = a[0] === "grpcurl" ? 1 : 0; const pos = [];
  const withVal = new Set(["-d","-H","-rpc-header","-max-time","-connect-timeout","-authority","-cacert","-cert","-key","-import-path","-proto","-protoset","-servername","-format","-max-msg-sz","-unix","-user-agent","-reflect-header"]);
  for (; i < a.length; i++){
    let t = a[i];
    if (t.startsWith("-")){
      t = t.replace(/^--/, "-"); let v = null; const eq = t.indexOf("=");
      if (eq > 0){ v = t.slice(eq + 1); t = t.slice(0, eq); } else if (withVal.has(t)) v = a[++i];
      if (t === "-d") r.body = v; else if (t === "-H" || t === "-rpc-header"){ const m = /^authorization:\s*bearer\s+(.+)$/i.exec(v || ""); if (m) r.token = m[1]; else r.headers.push(v); }
      else if (t === "-insecure" || t === "-cacert" || t === "-cert") r.tls = true;
      else if (t === "-plaintext") r.tls = false;
      else if (t === "-emit-defaults") r.emit = true; else if (t === "-v") r.verbose = true;
    } else pos.push(t);
  }
  if (pos.length < 2) throw new Error("expected: grpcurl [flags] host:port service.Method");
  r.addr = pos[pos.length - 2]; r.method = pos[pos.length - 1].replace("/", ".");
  if (r.body === "@") throw new Error("-d @ reads stdin; paste the body inline instead");
  return r;
}
async function pasteCurl(){
  const cmd = prompt("paste a grpcurl command:"); if (!cmd) return;
  let r; try { r = parseGrpcurl(cmd); } catch(e){ setStatus(e.message, false); return; }
  if (r.addr !== addr() || r.tls !== $("tls").checked){ $("addr").value = r.addr; ls.set("addr", r.addr); $("tls").checked = r.tls; ls.set("tls", r.tls ? "1" : "0"); await loadMethods(); }
  if (r.token) $("token").value = r.token; ls.set("token", $("token").value);
  $("headers").value = r.headers.join("\n"); ls.set("headers", $("headers").value);
  if (r.emit != null) $("emit").checked = !!r.emit; if (r.verbose != null) $("verbose").checked = !!r.verbose;
  if (!services.flatMap(s => s.methods).some(m => m.name === r.method)){ setStatus("method not on " + r.addr + ": " + r.method, false); return; }
  if (method) newTab();
  await pick(r.method);
  $("body").value = r.body || "{}"; format(); saveTab();
  setStatus("imported " + r.method + " — Invoke to run", true);
}

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
  $("out").innerHTML = objs && objs.length ? objs.map((o, i) => tree(o, "", true, [i])).join('<hr style="border:0;border-top:1px dashed var(--line)">') : hl(lastOut);
}
function tree(v, key, last, segs, editable){
  const off = key !== "" && typeof key === "string" && key.startsWith("//");
  const comma = last ? "" : ",";
  const label = key === "" ? "" : (editable && typeof segs[segs.length - 1] === "string" ? '<span class="tog' + (off ? " on" : "") + '" title="' + (off ? "enable" : "disable (kept in body, not sent)") + '" data-s="' + esc(JSON.stringify(segs)) + '" onclick="toggleKey(JSON.parse(this.dataset.s))">//</span>' : '') +
    '<span class="k' + (off ? " off" : "") + '">"' + esc(key) + '"</span>: ';
  if (off) return '<div class="leaf off">' + label + esc(JSON.stringify(v)) + comma + '</div>';
  const sj = esc(JSON.stringify(segs));
  const rowSel = editable ? ' data-s="' + sj + '" onclick="if(!event.target.closest(\'.pen,.tog,.del,.v,input\'))selectRow(this,JSON.parse(this.dataset.s))"' : '';
  const del = editable && segs.length > 1 ? '<span class="del" title="delete (or select the row and press Delete)" data-s="' + sj + '" onclick="event.preventDefault();deleteNode(JSON.parse(this.dataset.s))">\u2715</span>' : '';
  if (v === null || typeof v !== "object")
    return '<div class="leaf"' + rowSel + '>' + label + '<span class="v" data-s="' + sj + '"' + (editable ? ' ondblclick="editLeaf(this)" title="double-click to edit"' : '') + '>' + hl(JSON.stringify(v)) + '</span>' + comma + del + '</div>';
  const arr = Array.isArray(v), keys = Object.keys(v), o = arr ? "[" : "{", c = arr ? "]" : "}";
  const pen = editable ? '<span class="pen" title="edit this ' + (arr ? "array" : "object") + ' as JSON" data-s="' + sj + '" onclick="event.preventDefault();editNode(JSON.parse(this.dataset.s))">\u270E</span>' : '';
  const wrap = editable ? ' data-n="' + sj + '"' : '';
  if (!keys.length) return '<div class="te"' + wrap + rowSel + '>' + label + o + c + comma + pen + del + '</div>';
  return '<div' + wrap + '><details class="t" open><summary' + rowSel + '>' + label + o + '<span class="cnt"> \u2026 ' + keys.length + ' </span><span class="cl">' + c + comma + '</span>' + pen + del + '</summary><div class="tb">' +
    keys.map((k, i) => tree(v[k], arr ? "" : k, i === keys.length - 1, segs.concat(arr ? i : k), editable)).join("") +
    '</div><div class="te">' + c + comma + '</div></details></div>';
}
function foldAll(open){ document.querySelectorAll("#out details").forEach(d => d.open = open); }

// Request tree: same renderer over the editor's JSON; edits round-trip.
// #body is the source of truth; the tree renders from it and writes back.
function toggleReq(){
  reqTree = !reqTree; $("reqview").textContent = reqTree ? "Text" : "Tree";
  $("body").style.display = reqTree ? "none" : ""; $("lint").style.display = reqTree ? "none" : ""; $("fmtbtn").style.display = reqTree ? "none" : "";
  $("reqtree").style.display = reqTree ? "block" : "none"; $("reqfold").style.display = reqTree ? "" : "none";
  if (reqTree) renderReq(); else lint();
}
function foldReq(open){ document.querySelectorAll("#reqtree details").forEach(d => d.open = open); }
function renderReq(){
  if (!reqTree) return;
  let objs; try { objs = parseMany($("body").value || "{}"); } catch(e){ $("reqtree").innerHTML = '<span style="color:var(--err)">' + esc("invalid JSON: " + e.message) + '</span> <button class="small" onclick="editNode([0])">fix</button>'; return; }
  if (!objs.length) objs = [{}];
  $("reqtree").innerHTML = objs.map((o, i) => tree(o, "", true, [i], true)).join('<hr style="border:0;border-top:1px dashed var(--line)">') +
    (meta.clientStream ? '<div class="hintline">client stream: each message is sent in order <button class="small" onclick="addMessage()">+ message</button></div>' : '') +
    '<div class="hintline">click a row + Delete removes it (\u2318Z undo) \u00b7 double-click a value to edit \u00b7 \u270E edits an object as JSON \u00b7 // disables a key \u00b7 Text for the plain editor</div>';
}
function setBody(objs){ pushUndo(); $("body").value = objs.map(o => JSON.stringify(o, null, 2)).join("\n"); saveTab(); lint(); renderReq(); scanAny(true); }
function pushUndo(){ undoStack.push($("body").value); if (undoStack.length > 30) undoStack.shift(); }
function undo(){ if (!undoStack.length){ setStatus("nothing to undo", false); return; } $("body").value = undoStack.pop(); saveTab(); renderReq(); scanAny(true); setStatus("undone", true); }
// deleteNode removes a key or array element; the root resets to {}.
function deleteNode(segs){
  const objs = parseMany($("body").value || "{}");
  if (segs.length === 1) objs[segs[0]] = {};
  else { let o = objs[segs[0]]; segs.slice(1, -1).forEach(k => o = o[k]); const k = segs[segs.length - 1]; Array.isArray(o) ? o.splice(k, 1) : delete o[k]; }
  selSegs = null; setBody(objs); setStatus("deleted " + showPath(segs) + " \u00b7 \u2318Z to undo", true);
}
function selectRow(el, segs){ selSegs = segs; document.querySelectorAll("#reqtree .selrow").forEach(x => x.classList.remove("selrow")); el.classList.add("selrow"); }
function addMessage(){ const objs = parseMany($("body").value || "{}"); objs.push(JSON.parse(JSON.stringify(objs[objs.length - 1] || {}))); setBody(objs); }
// editNode swaps a subtree (or the root) for a textarea holding its JSON.
function editNode(segs){
  const objs = parseMany($("body").value || "{}");
  let v = objs[segs[0]]; segs.slice(1).forEach(k => v = v[k]);
  const holder = document.querySelector('#reqtree [data-n="' + CSS.escape(JSON.stringify(segs)) + '"]') || $("reqtree");
  const ta = document.createElement("textarea"); ta.className = "ed"; ta.value = JSON.stringify(v, null, 2); ta.rows = Math.min(30, ta.value.split("\n").length + 1); ta.spellcheck = false;
  const row = document.createElement("div"); row.className = "edrow";
  const ok = document.createElement("button"); ok.className = "small primary"; ok.textContent = "apply"; const no = document.createElement("button"); no.className = "small"; no.textContent = "cancel";
  const cp = document.createElement("button"); cp.className = "small"; cp.textContent = "copy JSON"; cp.onclick = () => copy(ta.value);
  row.append(ok, no, cp);
  if (segs.length > 1){ const del = document.createElement("button"); del.className = "small"; del.textContent = "delete"; del.title = "remove this key / element"; row.append(del);
    del.onclick = () => { let o = objs[segs[0]]; segs.slice(1, -1).forEach(k => o = o[k]); const k = segs[segs.length - 1]; Array.isArray(o) ? o.splice(k, 1) : delete o[k]; setBody(objs); setStatus("deleted " + showPath(segs), true); }; }
  holder.replaceChildren(ta, row); ta.focus(); ta.select();
  no.onclick = renderReq;
  ok.onclick = () => {
    const cleaned = ta.value.replace(/\u270E/g, "").replace(/^(\s*)\/\/\s*(?=")/gm, "$1"); // text copied from the tree
    let val; try { const many = parseMany(cleaned); if (many.length !== 1) throw new Error("exactly one JSON value expected"); val = many[0]; } catch(e){ setStatus("invalid JSON: " + e.message, false); return; }
    if (segs.length === 1) objs[segs[0]] = val; else setPath(objs, segs, val);
    setBody(objs); setStatus("updated " + showPath(segs), true);
  };
  ta.onkeydown = e => { if ((e.ctrlKey || e.metaKey) && e.key === "Enter"){ e.preventDefault(); e.stopPropagation(); ok.onclick(); } else if (e.key === "Escape") renderReq(); else if (e.key === "Tab"){ e.preventDefault(); const a = ta.selectionStart; ta.value = ta.value.slice(0, a) + "  " + ta.value.slice(ta.selectionEnd); ta.selectionStart = ta.selectionEnd = a + 2; } };
}
function toggleKey(segs){
  const objs = parseMany($("body").value);
  let o = objs[segs[0]]; segs.slice(1, -1).forEach(k => o = o[k]);
  const k = segs[segs.length - 1], nk = k.startsWith("//") ? k.slice(2) : "//" + k;
  const rebuilt = {}; for (const key of Object.keys(o)) rebuilt[key === k ? nk : key] = o[key]; // keep key order
  Object.keys(o).forEach(key => delete o[key]); Object.assign(o, rebuilt);
  setBody(objs);
}
function editLeaf(el){
  const cur = el.textContent, segs = JSON.parse(el.dataset.s);
  const inp = document.createElement("input"); inp.className = "iv"; inp.value = cur; inp.size = Math.max(8, cur.length + 2); inp.title = "JSON literal: \"text\", 42, true, null · Enter applies, Esc cancels";
  el.replaceWith(inp); inp.focus(); inp.select();
  const commit = () => {
    const nv = inp.value.trim(); if (nv === cur){ renderReq(); return; }
    let val; try { val = JSON.parse(nv); } catch(e){ val = nv; } // unquoted text becomes a string
    const objs = parseMany($("body").value); setPath(objs, segs, val); setBody(objs);
  };
  inp.onkeydown = e => { if (e.key === "Enter"){ e.preventDefault(); e.stopPropagation(); commit(); } else if (e.key === "Escape") renderReq(); };
  inp.onblur = commit;
}
// Cheap JSON colouring on the escaped text; grpcurl already pretty-prints.
function hl(s){ return String(s ?? "").replace(/[&<>]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;"}[c])).replace(/("(?:\\.|[^"\\])*")(\s*:)?|\b(true|false|null)\b|-?\b\d+(\.\d+)?([eE][+-]?\d+)?\b/g,
  (m, str, colon, kw) => str ? (colon ? '<span class="k">' + str + '</span>' + colon : '<span class="s">' + str + '</span>') : kw ? '<span class="b">' + kw + '</span>' : '<span class="n">' + m + '</span>'); }
document.addEventListener("keydown", e => {
  if ((e.ctrlKey || e.metaKey) && e.key === "Enter"){ e.preventDefault(); invoke(); return; }
  if (e.target && e.target.closest && e.target.closest("input,textarea")) return;
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "z"){ e.preventDefault(); undo(); }
  else if ((e.key === "Delete" || e.key === "Backspace") && selSegs && selSegs.length > 1){ e.preventDefault(); deleteNode(selSegs); }
});

renderTabs();
loadMethods().then(() => { const t = tabs[cur]; if (t.method && services.flatMap(s => s.methods).some(m => m.name === t.method)) pick(t.method); else { $("body").value = t.body; renderReq(); } });
loadSaved(); loadHistory(); loadTypeSources();
</script></body></html>`
