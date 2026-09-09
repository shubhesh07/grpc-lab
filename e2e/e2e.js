// End-to-end drive of grpc-lab in a real browser against the reflecting test
// server in ./testsrv. Run ./run.sh (needs Go, node, playwright-core and a
// Playwright chromium). Covers: reflection, unary / streaming / error calls,
// metadata + auth, save-as into nested folders, dirty tab + ⌘S, reload
// persistence, tabs, Paste grpcurl, history, folders, pane collapse, sidebar
// toggle, splitter drag, copy, tree view, folder delete.
const path = require("path"), fs = require("fs"), os = require("os");
// playwright-core: from node_modules, or point PW_CORE at a copy (e.g. one in ~/.npm/_npx).
const { chromium } = require(process.env.PW_CORE || "playwright-core");
const exe = process.env.PW_CHROME; // optional: explicit chromium/headless-shell binary
const PAY = process.env.PAYLOADS || path.join(__dirname, "payloads");
const UI = "http://127.0.0.1:8097/";
let step = "", passed = 0, page, errors = [];
const ok = (cond, msg) => { if (!cond) throw new Error("FAIL [" + step + "] " + msg); passed++; };

(async () => {
  const browser = await chromium.launch({ executablePath: exe || undefined, headless: true });
  page = await browser.newPage({ viewport: { width: 1500, height: 900 } });
  page.on("pageerror", e => errors.push(String(e)));
  page.on("console", m => { if (m.type() === "error") errors.push("console: " + m.text()); });
  const dialogs = []; page.on("dialog", d => { const r = dialogs.shift(); r === undefined ? d.dismiss() : d.accept(r === true ? undefined : r); });
  const status = () => page.locator("#status").innerText();
  const waitStatus = re => page.waitForFunction(re => new RegExp(re).test(document.getElementById("status").textContent), re.source, { timeout: 15000 });
  const selectMethod = async name => { await page.selectOption("#msel", name); await page.waitForFunction(n => document.getElementById("msel").value === n && document.getElementById("body").value.trim().length > 0, name); };

  step = "load"; await page.goto(UI);
  await page.waitForFunction(() => document.querySelectorAll("#msel option").length > 3);
  await page.waitForFunction(() => document.getElementById("hint").textContent.includes("types"));
  ok(true, "reflection type count shown");

  step = "unary"; await selectMethod("grpc.testing.TestService.UnaryCall");
  ok((await page.locator("#sel").innerText()).includes("SimpleRequest"), "types line shows input type");
  await page.fill("#body", '{"responseSize": 4}');
  await page.click("button.primary"); await waitStatus(/ok/);
  const out1 = await page.locator("#out").innerText();
  ok(out1.includes("payload") && out1.includes("body"), "unary response rendered: " + out1.slice(0, 80));
  ok((await page.locator("#cmd").textContent()).includes("grpcurl"), "grpcurl tab filled");

  step = "health"; await selectMethod("grpc.health.v1.Health.Check");
  await page.keyboard.press("Control+Enter"); await waitStatus(/ok/);
  ok((await page.locator("#out").innerText()).includes("SERVING"), "health SERVING via ⌃⏎");

  step = "server-stream"; await selectMethod("grpc.testing.TestService.StreamingOutputCall");
  await page.fill("#body", '{"responseParameters":[{"size":1},{"size":2}]}');
  await page.click("button.primary"); await waitStatus(/ok/);
  ok((await page.locator("#out hr").count()) === 1, "two streamed messages separated");

  step = "error"; await selectMethod("grpc.testing.TestService.UnaryCall");
  await page.fill("#body", '{"responseStatus":{"code":5,"message":"nope"}}');
  await page.click("button.primary"); await waitStatus(/NOT_FOUND/);
  ok((await page.locator("#out").innerText()).includes("nope"), "gRPC error surfaced with code name");

  step = "metadata"; await page.click("#ptabs [data-p=pmeta]");
  await page.fill("#headers", "x-test: 1\nx-other: 2");
  ok((await page.locator("#ptabs [data-p=pmeta]").innerText()).includes("2"), "metadata badge counts headers");
  await page.click("#ptabs [data-p=pauth]"); await page.fill("#token", "abc");
  ok((await page.locator("#ptabs [data-p=pauth]").innerText()).includes("●"), "auth badge shows token");
  await page.click("#ptabs [data-p=pbody]");
  await page.fill("#body", '{"responseSize": 2}');
  await page.click("button.primary"); await waitStatus(/ok/);
  const cmd = await page.locator("#cmd").textContent();
  ok(cmd.includes("x-test: 1") && cmd.includes("authorization"), "headers + bearer in grpcurl command");

  step = "save-as"; await page.click("button:has-text(\"Save as…\")"); await page.waitForSelector("#dlg[open]"); await page.fill("#dlgfolder", "E2E/Cart"); await page.fill("#dlgname", "unary"); await page.click("#dlgok"); await waitStatus(/saved E2E\/Cart\/unary/);
  await page.waitForSelector("#saved details[data-p='E2E/Cart'] .saved");
  ok((await page.locator(".rtab.on").innerText()).startsWith("unary"), "tab named after file");
  const file = JSON.parse(fs.readFileSync(path.join(PAY, "E2E/Cart/unary.json"), "utf8"));
  ok(file["grpc-lab"].method === "grpc.testing.TestService.UnaryCall" && file["grpc-lab"].headers.includes("x-test") && file.body.responseSize === 2 && !JSON.stringify(file).includes("abc"), "file has envelope, headers, body; no token");

  step = "dirty+⌘S"; await page.fill("#body", '{"responseSize": 3}');
  ok((await page.locator(".rtab.on .dot").count()) === 1, "dirty dot after edit");
  await page.keyboard.press("Meta+s"); await waitStatus(/saved/);
  await page.waitForFunction(() => !document.querySelector(".rtab.on .dot"));
  ok(JSON.parse(fs.readFileSync(path.join(PAY, "E2E/Cart/unary.json"), "utf8")).body.responseSize === 3, "⌘S wrote to the same file");

  step = "reload"; await page.reload(); await page.waitForFunction(() => document.querySelectorAll("#msel option").length > 3);
  await page.waitForFunction(() => document.getElementById("msel").value === "grpc.testing.TestService.UnaryCall" && document.getElementById("body").value.includes("responseSize"));
  ok((await page.locator(".rtab.on").innerText()).startsWith("unary") && (await page.inputValue("#body")).includes('"responseSize": 3'), "tab, method and body survive reload: " + await page.locator(".rtab.on").innerText() + " / " + (await page.inputValue("#body")).slice(0,60) + " / ls=" + await page.evaluate(() => localStorage.getItem("grpclab:tabs")));
  ok(await page.locator("#saved details[data-p='E2E/Cart'] .saved.on").count() === 1, "sidebar highlights the open request");

  step = "tabs"; await page.click("#rtabs button"); await page.waitForFunction(() => document.querySelectorAll(".rtab").length === 2);
  ok((await page.locator(".rtab.on").innerText()).startsWith("new"), "new empty tab");
  await page.click("#saved details[data-p='E2E/Cart'] .saved");
  await page.waitForFunction(() => document.querySelector(".rtab.on").textContent.startsWith("unary"));
  ok((await page.locator(".rtab").count()) === 2, "opening a saved request focuses its existing tab (no duplicate)");

  step = "per-tab-target"; await page.fill("#addr", "127.0.0.1:1"); await page.press("#addr", "Enter"); await page.waitForTimeout(800);
  ok((await page.locator("#msel option").count()) <= 1, "bad target in this tab empties its method list");
  await page.click(".rtab >> nth=1"); await page.waitForFunction(() => document.getElementById("addr").value === "127.0.0.1:50077" && document.querySelectorAll("#msel option").length > 3);
  ok(true, "other tab keeps its own target and methods");
  await page.click(".rtab >> nth=0"); await page.waitForFunction(() => document.getElementById("addr").value === "127.0.0.1:1");
  await page.fill("#addr", "127.0.0.1:50077"); await page.press("#addr", "Enter"); await page.waitForFunction(() => document.querySelectorAll("#msel option").length > 3);
  ok(true, "target edit applies to the current tab only");

  step = "per-tab-response/a"; await page.click(".rtab >> nth=0"); await page.waitForFunction(() => document.getElementById("msel").value === "grpc.testing.TestService.UnaryCall" && document.getElementById("body").value.includes("responseSize"));
  step = "per-tab-response/b"; await page.click("button.primary"); await waitStatus(/ok/); const resp0 = await page.locator("#out").innerText();
  ok(resp0.includes("payload"), "tab 0 got its response");
  step = "per-tab-response/c"; await page.click(".rtab >> nth=1"); await page.waitForFunction(() => document.getElementById("status").textContent === "idle");
  ok((await page.locator("#out").innerText()).trim() === "—", "tab 1 shows no response of its own");
  await page.click(".rtab >> nth=0"); await page.waitForFunction(() => /ok/.test(document.getElementById("status").textContent));
  ok((await page.locator("#out").innerText()) === resp0, "tab 0 response restored on switch");

  step = "move"; await page.hover("#saved details[data-p='E2E/Cart'] .saved"); await page.click("#saved details[data-p='E2E/Cart'] .saved .mv");
  await page.waitForSelector("#dlg[open]"); ok((await page.inputValue("#dlgfolder")) === "E2E/Cart" && (await page.inputValue("#dlgname")) === "unary", "move dialog prefilled");
  await page.fill("#dlgfolder", "E2E/Moved"); await page.click("#dlgok"); await waitStatus(/moved to E2E\/Moved\/unary/);
  await page.waitForSelector("#saved details[data-p='E2E/Moved'] .saved");
  ok(fs.existsSync(path.join(PAY, "E2E/Moved/unary.json")) && !fs.existsSync(path.join(PAY, "E2E/Cart/unary.json")), "file moved on disk");
  ok((await page.locator(".rtab.on").innerText()).startsWith("unary"), "open tab still bound after move");
  await page.fill("#body", '{"responseSize": 5}'); await page.keyboard.press("Meta+s"); await waitStatus(/saved E2E\/Moved\/unary/);
  ok(JSON.parse(fs.readFileSync(path.join(PAY, "E2E/Moved/unary.json"), "utf8")).body.responseSize === 5, "⌘S after move writes to the new path");

  step = "paste-grpcurl"; dialogs.push(`grpcurl -plaintext -H "x-a: b" -d '{"service":""}' 127.0.0.1:50077 grpc.health.v1.Health/Check`);
  await page.click("button:has-text(\"Paste grpcurl\")"); await waitStatus(/imported/);
  ok((await page.locator("#msel").inputValue()) === "grpc.health.v1.Health.Check" && (await page.inputValue("#headers")).includes("x-a: b"), "grpcurl import sets method and headers");
  ok((await page.locator(".rtab").count()) === 3, "import opened a new tab");

  step = "history"; await page.click("#stabs [data-s=shist]");
  ok((await page.locator(".hist").count()) >= 4, "history has the calls");
  await page.click("#stabs [data-s=scoll]");

  step = "folders"; dialogs.push("Sub"); await page.hover("#saved details[data-p='E2E/Cart'] > summary"); await page.click("#saved details[data-p='E2E/Cart'] > summary .acts span >> nth=1");
  await page.waitForSelector("#saved details[data-p='E2E/Cart/Sub']");
  ok(fs.existsSync(path.join(PAY, "E2E/Cart/Sub")), "+ folder created nested dir");
  await page.click("#saved details[data-p='E2E/Cart/Sub'] > summary");
  await page.waitForFunction(() => (localStorage.getItem("grpclab:closed") || "").includes("E2E/Cart/Sub")); ok(true, "folder collapse remembered");

  step = "collapse-panes"; await page.click("#resfoldbtn"); ok(await page.locator("#respanel").isHidden(), "response pane collapses");
  await page.click("#resfoldbtn"); ok(await page.locator("#respanel").isVisible(), "response pane expands");
  await page.click("#reqfoldbtn"); ok(await page.locator("#reqpanel").isHidden(), "request pane collapses");
  await page.click("#resfoldbtn"); ok(await page.locator("#reqpanel").isVisible() && await page.locator("#respanel").isHidden(), "folding the other pane unfolds this one");
  await page.click("#resfoldbtn");
  await page.keyboard.press("Meta+b"); ok(await page.locator("#side").isHidden(), "⌘B hides sidebar");
  await page.reload(); await page.waitForSelector("#msel");
  ok(await page.locator("#side").isHidden(), "sidebar state remembered"); await page.keyboard.press("Meta+b"); ok(await page.locator("#side").isVisible(), "⌘B shows sidebar");

  step = "splitter"; const before = (await page.locator("#reqpanel").boundingBox()).height;
  const bar = await page.locator("#splitbar").boundingBox();
  await page.mouse.move(bar.x + bar.width - 120, bar.y + 4); await page.mouse.down(); await page.mouse.move(bar.x + bar.width - 120, bar.y - 150, { steps: 5 }); await page.mouse.up();
  const after = (await page.locator("#reqpanel").boundingBox()).height;
  ok(after < before - 100, "drag shrinks request pane " + before + " -> " + after);

  step = "sidebar-resize"; const sw = (await page.locator("#side").boundingBox()).width;
  const grip = await page.locator("#sidegrip").boundingBox();
  await page.mouse.move(grip.x + 3, grip.y + 200); await page.mouse.down(); await page.mouse.move(grip.x + 153, grip.y + 200, { steps: 5 }); await page.mouse.up();
  ok(Math.abs((await page.locator("#side").boundingBox()).width - sw - 150) < 10, "sidebar grip drag widens sidebar " + sw + " -> " + (await page.locator("#side").boundingBox()).width);
  await page.reload(); await page.waitForSelector("#msel");
  ok(Math.abs((await page.locator("#side").boundingBox()).width - sw - 150) < 10, "sidebar width remembered");
  await page.waitForFunction(() => document.querySelectorAll("#msel option").length > 3);

  step = "copy"; await page.click("button.primary"); await waitStatus(/ok/); await page.click("#splitbar button:has-text(\"copy\")"); await waitStatus(/copied|copy failed/); ok((await status()).startsWith("copied"), "copy reports success: " + await status());

  step = "tree-edit"; await page.click("#reqview"); ok(await page.locator("#reqtree").isVisible(), "tree view on");
  await page.click("#reqview"); ok(await page.locator("#body").isVisible(), "text view back");

  step = "delete-folder"; dialogs.push(true); await page.hover("#saved details[data-p='E2E'] > summary"); await page.click("#saved details[data-p='E2E'] > summary .acts .x");
  await page.waitForFunction(() => !document.querySelector("#saved details[data-p='E2E']"));
  ok(!fs.existsSync(path.join(PAY, "E2E")), "folder delete removed dir");

  await page.screenshot({ path: path.join(__dirname, "e2e.png") });
  await browser.close();
  if (errors.length) throw new Error("page errors: " + errors.join(" | "));
  console.log("E2E OK: " + passed + " checks passed");
})().catch(async e => { console.error(String(e.stack || e)); try { console.error("page errors:", errors); console.error("status:", await page.locator("#status").innerText(), "| step:", step, "| tab:", await page.locator(".rtab.on").innerText(), "| msel:", await page.locator("#msel").inputValue(), "| body:", (await page.inputValue("#body")).slice(0, 40), "| out:", (await page.locator("#out").innerText()).slice(0, 30)); await page.screenshot({ path: "fail.png" }); } catch(_){} process.exit(1); });
