// grpc-lab is a small browser UI for exercising gRPC services during
// development.
//
// It deliberately does not speak gRPC itself: every call shells out to
// grpcurl. That is the whole point. grpcurl resolves google.protobuf.Any
// payloads by asking the server's reflection service for the concrete type on
// demand, which is exactly what grpcui cannot do -- grpcui builds its type
// registry from the method's descriptor closure, and an Any is opaque to that,
// so a request carrying one fails in the browser before it reaches the server.
// Delegating to grpcurl inherits correct behaviour for free and leaves this
// tool with no protobuf dependencies at all (stdlib only).
//
//	go run . -addr localhost:8082
//
// Requires grpcurl on PATH and server reflection enabled on the target.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	defaultAddr = flag.String("addr", "localhost:8086", "default gRPC target")
	uiPort      = flag.Int("port", 8090, "port for this UI")
	bindHost    = flag.String("bind", "127.0.0.1", "interface for this UI; 0.0.0.0 only inside a container")
	payloadDir  = flag.String("payloads", "payloads", "directory of saved request bodies")
	callTimeout = flag.Duration("timeout", 30*time.Second, "per-call timeout")
	protosetDir = flag.String("protosets", "protosets", "directory of extra .protoset files (type sources) added to every call")
)

// safeName keeps saved payloads inside payloadDir. Names arrive from the
// browser, so anything with a path separator or dot-dot is rejected outright
// rather than sanitised -- an obvious rule beats a clever one.
var safeName = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// rpcLine matches one line of `grpcurl describe <service>`:
//
//	rpc Name ( stream .pkg.In ) returns ( stream .pkg.Out );
var rpcLine = regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(stream\s+)?\.?([\w.]+)\s*\)\s*returns\s*\(\s*(stream\s+)?\.?([\w.]+)\s*\)`)

// typeRef matches a fully-qualified message/enum reference inside describe
// output, e.g. `.shared.common.v1.Seller seller = 1;` or `map<string, .a.B>`.
var typeRef = regexp.MustCompile(`\.([a-zA-Z_][\w]*(?:\.[a-zA-Z_][\w]*)+)`)

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		// Go stops parsing flags at the first bare argument, so a stray "." or
		// typo would silently discard every flag after it.
		log.Fatalf("unexpected argument %q -- usage: grpc-lab [-addr host:port] [-port N] [-payloads dir] [-timeout 30s] [-bind ip]", flag.Arg(0))
	}
	if _, err := exec.LookPath("grpcurl"); err != nil {
		log.Fatal("grpcurl not found on PATH. Install: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest")
	}
	if err := os.MkdirAll(*payloadDir, 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(*protosetDir, 0o755); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/methods", handleMethods)
	http.HandleFunc("/api/template", handleTemplate)
	http.HandleFunc("/api/describe", handleDescribe)
	http.HandleFunc("/api/types", handleTypes)
	http.HandleFunc("/api/schema", handleSchema)
	http.HandleFunc("/api/invoke", handleInvoke)
	http.HandleFunc("/api/payloads", handlePayloads)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/typesources", handleTypeSources)

	addr := fmt.Sprintf("%s:%d", *bindHost, *uiPort)
	log.Printf("grpc-lab  http://%s   target=%s   payloads=%s", addr, *defaultAddr, *payloadDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// grpcurlRun executes grpcurl and returns combined output. Errors are returned
// alongside the output, never instead of it: grpcurl writes the useful part of
// a failure (status code, message, details) to stderr, so discarding it on a
// non-zero exit would throw away exactly what the developer needs to see.
func grpcurlRun(ctx context.Context, stdin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "grpcurl", append(protosetArgs(), args...)...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// protosetArgs adds every saved type source to a grpcurl invocation. A server
// can only reflect the types it links in; an Any packed with a type from
// another service (say CustomerInfo from customer-service) comes back as
// @error/@value unless grpcurl also has that service's descriptors.
// -use-reflection keeps the target's own reflection active alongside them.
func protosetArgs() []string {
	files, _ := filepath.Glob(filepath.Join(*protosetDir, "*.protoset"))
	if len(files) == 0 {
		return nil
	}
	args := []string{"-use-reflection"}
	for _, f := range files {
		args = append(args, "-protoset", f)
	}
	return args
}

// handleTypeSources lists (GET), fetches (POST {addr,tls}) or removes (DELETE
// ?name=) descriptor sets pulled from other servers' reflection.
func handleTypeSources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		files, _ := filepath.Glob(filepath.Join(*protosetDir, "*.protoset"))
		names := []string{}
		for _, f := range files {
			names = append(names, strings.TrimSuffix(filepath.Base(f), ".protoset"))
		}
		writeJSON(w, map[string]any{"sources": names})
	case http.MethodPut:
		// Upload a descriptor set built from .proto files (buf build -o x.protoset
		// or protoc --descriptor_set_out --include_imports) for types no server
		// reflects. ?name= is the file's base name.
		name := strings.TrimSuffix(r.URL.Query().Get("name"), ".protoset")
		if !safeName.MatchString(name) {
			writeJSON(w, map[string]any{"error": "name must match [A-Za-z0-9._-]{1,64}"})
			return
		}
		b, err := io.ReadAll(io.LimitReader(r.Body, 32<<20))
		if err != nil || len(b) == 0 {
			writeJSON(w, map[string]any{"error": "empty upload"})
			return
		}
		out := filepath.Join(*protosetDir, name+".protoset")
		if err := os.WriteFile(out, b, 0o644); err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		// Reject anything grpcurl cannot read, so a bad file never breaks every call.
		if probe, perr := exec.Command("grpcurl", "-protoset", out, "list").CombinedOutput(); perr != nil && !strings.Contains(string(probe), "No services") {
			os.Remove(out)
			writeJSON(w, map[string]any{"error": "not a descriptor set grpcurl can read: " + strings.TrimSpace(string(probe))})
			return
		}
		typeCache.Range(func(k, _ any) bool { typeCache.Delete(k); return true })
		writeJSON(w, map[string]any{"added": name})
	case http.MethodPost:
		var req struct {
			Addr string `json:"addr"`
			TLS  bool   `json:"tls"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Addr == "" {
			writeJSON(w, map[string]any{"error": "addr is required"})
			return
		}
		name := strings.NewReplacer(":", "_", "/", "_").Replace(req.Addr)
		if !safeName.MatchString(name) {
			writeJSON(w, map[string]any{"error": "bad address"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		out := filepath.Join(*protosetDir, name+".protoset")
		// Plain exec, not grpcurlRun: the fetch must not depend on the other sources.
		cmd := exec.CommandContext(ctx, "grpcurl", append(connArgs(req.TLS), "-protoset-out", out, req.Addr, "describe")...)
		if b, err := cmd.CombinedOutput(); err != nil {
			writeJSON(w, map[string]any{"error": strings.TrimSpace(string(b))})
			return
		}
		typeCache.Range(func(k, _ any) bool { typeCache.Delete(k); return true })
		writeJSON(w, map[string]any{"added": name})
	case http.MethodDelete:
		name := r.URL.Query().Get("name")
		if !safeName.MatchString(name) {
			writeJSON(w, map[string]any{"error": "bad name"})
			return
		}
		if err := os.Remove(filepath.Join(*protosetDir, name+".protoset")); err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		typeCache.Range(func(k, _ any) bool { typeCache.Delete(k); return true })
		writeJSON(w, map[string]any{"deleted": name})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func addrOf(r *http.Request) string {
	if a := r.URL.Query().Get("addr"); a != "" {
		return a
	}
	return *defaultAddr
}

// connArgs picks the transport flags. plaintext is the dev default; tls=1
// uses TLS without verifying the certificate, which is what a local port
// forward to a cluster ingress needs.
func connArgs(tls bool) []string {
	if tls {
		return []string{"-insecure"}
	}
	return []string{"-plaintext"}
}

func tlsOf(r *http.Request) bool { return r.URL.Query().Get("tls") == "1" }

type method struct {
	Name         string `json:"name"` // fully qualified: pkg.Service.Method
	Input        string `json:"input"`
	Output       string `json:"output"`
	ClientStream bool   `json:"clientStream"`
	ServerStream bool   `json:"serverStream"`
}

// parseMethods turns `grpcurl describe <service>` output into methods.
func parseMethods(service, describe string) []method {
	var ms []method
	for _, m := range rpcLine.FindAllStringSubmatch(describe, -1) {
		ms = append(ms, method{
			Name:         service + "." + m[1],
			Input:        m[3],
			Output:       m[5],
			ClientStream: m[2] != "",
			ServerStream: m[4] != "",
		})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].Name < ms[j].Name })
	return ms
}

// isNoise hides the reflection service itself; health stays, Check is worth a click.
func isNoise(service string) bool {
	return strings.HasPrefix(service, "grpc.reflection.")
}

// handleMethods lists every method on the target with its request/response
// types and streaming shape, so the UI never needs a .proto file.
func handleMethods(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	conn := connArgs(tlsOf(r))

	out, err := grpcurlRun(ctx, "", append(conn, addrOf(r), "list")...)
	if err != nil {
		writeJSON(w, map[string]any{"error": strings.TrimSpace(out)})
		return
	}

	type svc struct {
		Name    string   `json:"name"`
		Methods []method `json:"methods"`
	}
	var services []svc
	for name, d := range describeAll(ctx, conn, addrOf(r), strings.Fields(out)) {
		if !isNoise(name) {
			services = append(services, svc{Name: name, Methods: parseMethods(name, d)})
		}
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	writeJSON(w, map[string]any{"services": services})
}

// describeSymbol returns the proto text for a symbol and, for messages, the
// JSON template grpcurl derives from it.
func describeSymbol(ctx context.Context, addr string, tls bool, symbol string) (text, template string, err error) {
	out, err := grpcurlRun(ctx, "", append(connArgs(tls), "-msg-template", addr, "describe", symbol)...)
	if err != nil {
		return strings.TrimSpace(out), "", err
	}
	text = strings.TrimSpace(out)
	if i := strings.Index(out, "Message template:"); i >= 0 {
		text = strings.TrimSpace(out[:i])
		template = strings.TrimSpace(out[i+len("Message template:"):])
	}
	return text, template, nil
}

// handleTemplate resolves the method's real input type through reflection
// (no <Method>Request guessing) and returns its skeleton plus the proto text
// of both request and response so the developer can read field docs in place.
func handleTemplate(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("method")
	if name == "" {
		writeJSON(w, map[string]any{"error": "method is required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	addr, tls := addrOf(r), tlsOf(r)

	sig, _, err := describeSymbol(ctx, addr, tls, name)
	if err != nil {
		writeJSON(w, map[string]any{"error": sig})
		return
	}
	m := rpcLine.FindStringSubmatch(sig)
	if m == nil {
		writeJSON(w, map[string]any{"error": "cannot parse method signature:\n" + sig})
		return
	}
	inText, template, _ := describeSymbol(ctx, addr, tls, m[3])
	outText, _, _ := describeSymbol(ctx, addr, tls, m[5])
	if template == "" {
		template = "{}"
	}
	writeJSON(w, map[string]any{
		"template":     template,
		"input":        m[3],
		"output":       m[5],
		"clientStream": m[2] != "",
		"serverStream": m[4] != "",
		"describe":     sig + "\n\n" + inText + "\n\n" + outText,
	})
}

// handleDescribe describes any symbol -- the UI uses it to expand a
// google.protobuf.Any into the concrete type's template.
func handleDescribe(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Query().Get("symbol"), "type.googleapis.com/")
	if symbol == "" {
		writeJSON(w, map[string]any{"error": "symbol is required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	text, template, err := describeSymbol(ctx, addrOf(r), tlsOf(r), symbol)
	if err != nil {
		writeJSON(w, map[string]any{"error": text})
		return
	}
	writeJSON(w, map[string]any{"describe": text, "template": template})
}

// typeCache remembers the message-type index per target so the Any picker is
// instant after the first load. Reload clears it.
var typeCache sync.Map // addr -> []string

// handleTypes walks every message reachable from every service through
// reflection and returns the sorted set of fully-qualified names. This is the
// list a developer picks from when filling an Any -- the server itself is the
// source of truth for which concrete types exist.
func handleTypes(w http.ResponseWriter, r *http.Request) {
	addr, tls := addrOf(r), tlsOf(r)
	key := fmt.Sprintf("%s|%v", addr, tls)
	if r.URL.Query().Get("refresh") == "1" {
		typeCache.Delete(key)
	}
	if v, ok := typeCache.Load(key); ok {
		writeJSON(w, map[string]any{"types": v, "cached": true})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	conn := connArgs(tls)

	out, err := grpcurlRun(ctx, "", append(conn, addr, "list")...)
	if err != nil {
		writeJSON(w, map[string]any{"error": strings.TrimSpace(out)})
		return
	}
	seen := map[string]bool{}
	var queue []string
	for _, s := range strings.Fields(out) {
		if !isNoise(s) {
			queue = append(queue, s)
			seen[s] = true
		}
	}
	var types []string
	for len(queue) > 0 {
		descs := describeAll(ctx, conn, addr, queue)
		queue = nil
		for _, sym := range sortedKeys(descs) {
			d := descs[sym]
			if strings.Contains(d, " is a message:") {
				types = append(types, sym)
			}
			for _, ref := range extractTypeRefs(d) {
				if !seen[ref] {
					seen[ref] = true
					queue = append(queue, ref)
				}
			}
		}
	}
	sort.Strings(types)
	typeCache.Store(key, types)
	writeJSON(w, map[string]any{"types": types})
}

// describeAll describes symbols concurrently (bounded); unresolvable ones
// (e.g. option types like google.api.http) are simply absent from the result.
func describeAll(ctx context.Context, conn []string, addr string, syms []string) map[string]string {
	out := map[string]string{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for _, sym := range syms {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			d, err := grpcurlRun(ctx, "", append(conn, addr, "describe", sym)...)
			if err != nil {
				return
			}
			mu.Lock()
			out[sym] = d
			mu.Unlock()
		}(sym)
	}
	wg.Wait()
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// extractTypeRefs pulls `.pkg.Type` references out of describe output.
func extractTypeRefs(describe string) []string {
	var refs []string
	for _, m := range typeRef.FindAllStringSubmatch(describe, -1) {
		refs = append(refs, m[1])
	}
	return refs
}

// parseHeaders turns "k: v" lines into grpcurl -H arguments. Blank lines and
// `#` comments are skipped so a developer can keep a scratch list.
func parseHeaders(text string) []string {
	var args []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, ":") {
			continue
		}
		args = append(args, "-H", line)
	}
	return args
}

// shellQuote renders args as a pasteable shell command.
var plainArg = regexp.MustCompile(`^[\w./:@=-]+$`)

func shellQuote(args []string) string {
	q := make([]string, len(args))
	for i, a := range args {
		if plainArg.MatchString(a) {
			q[i] = a
		} else {
			q[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		}
	}
	return strings.Join(q, " ")
}

type historyEntry struct {
	At      string `json:"at"`
	Addr    string `json:"addr"`
	Method  string `json:"method"`
	Payload string `json:"payload"`
	OK      bool   `json:"ok"`
	Ms      int64  `json:"ms"`
}

var historyMu sync.Mutex

func historyPath() string { return filepath.Join(*payloadDir, ".history.jsonl") }

func appendHistory(e historyEntry) {
	historyMu.Lock()
	defer historyMu.Unlock()
	f, err := os.OpenFile(historyPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(e)
}

// handleHistory returns the most recent calls, newest first, so a request
// from ten minutes ago is one click away instead of a rebuild.
func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		historyMu.Lock()
		_ = os.Remove(historyPath())
		historyMu.Unlock()
		writeJSON(w, map[string]any{"cleared": true})
		return
	}
	historyMu.Lock()
	b, _ := os.ReadFile(historyPath())
	historyMu.Unlock()
	var entries []historyEntry
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var e historyEntry
		if json.Unmarshal([]byte(line), &e) == nil {
			entries = append(entries, e)
		}
	}
	const keep = 100
	if len(entries) > keep {
		entries = entries[len(entries)-keep:]
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	writeJSON(w, map[string]any{"history": entries})
}

func handleInvoke(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Addr         string `json:"addr"`
		Method       string `json:"method"`
		Payload      string `json:"payload"`
		Token        string `json:"token"`
		Headers      string `json:"headers"`
		TLS          bool   `json:"tls"`
		EmitDefaults bool   `json:"emitDefaults"`
		Verbose      bool   `json:"verbose"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	if req.Method == "" {
		writeJSON(w, map[string]any{"error": "method is required"})
		return
	}
	if req.Addr == "" {
		req.Addr = *defaultAddr
	}
	if strings.TrimSpace(req.Payload) == "" {
		req.Payload = "{}"
	}

	ctx, cancel := context.WithTimeout(r.Context(), *callTimeout)
	defer cancel()

	args := append(connArgs(req.TLS), "-format-error", "-max-time", fmt.Sprintf("%.0f", callTimeout.Seconds()))
	if req.EmitDefaults {
		args = append(args, "-emit-defaults")
	}
	if req.Verbose {
		args = append(args, "-v")
	}
	if req.Token != "" {
		args = append(args, "-H", "authorization: Bearer "+req.Token)
	}
	args = append(args, parseHeaders(req.Headers)...)
	args = append(args, "-d", "@", req.Addr, req.Method)

	started := time.Now()
	out, err := grpcurlRun(ctx, req.Payload, args...)
	ms := time.Since(started).Milliseconds()
	if ctx.Err() == context.DeadlineExceeded {
		// grpcurl was killed, so it printed nothing; say why instead of showing a blank pane.
		out = fmt.Sprintf("timed out after %s (-timeout flag)\n%s", *callTimeout, out)
	}
	appendHistory(historyEntry{At: started.Format(time.RFC3339), Addr: req.Addr, Method: req.Method, Payload: req.Payload, OK: err == nil, Ms: ms})

	// Pasteable equivalent, with the body inline instead of on stdin.
	cmdArgs := append(append([]string{"grpcurl"}, protosetArgs()...), args[:len(args)-4]...)
	cmdArgs = append(cmdArgs, "-d", req.Payload, req.Addr, req.Method)
	writeJSON(w, map[string]any{
		"output":  out,
		"ok":      err == nil,
		"ms":      ms,
		"command": shellQuote(cmdArgs),
	})
}

// handlePayloads lists saved bodies on GET, saves one on POST and removes one
// on DELETE, so a working request can be kept without leaving the browser.
func handlePayloads(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		entries, err := os.ReadDir(*payloadDir)
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		type p struct {
			Name string `json:"name"`
			Body string `json:"body"`
		}
		var out []p
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(*payloadDir, e.Name()))
			if err != nil {
				continue
			}
			out = append(out, p{Name: e.Name(), Body: string(b)})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
		writeJSON(w, map[string]any{"payloads": out})

	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		name := strings.TrimSuffix(req.Name, ".json")
		if !safeName.MatchString(name) {
			writeJSON(w, map[string]any{"error": "name must match [A-Za-z0-9._-]{1,64}"})
			return
		}
		// Reject invalid JSON here rather than saving a body that will only
		// fail later, at invoke time, with a worse error.
		var probe any
		if err := json.Unmarshal([]byte(req.Body), &probe); err != nil {
			writeJSON(w, map[string]any{"error": "payload is not valid JSON: " + err.Error()})
			return
		}
		if err := os.WriteFile(filepath.Join(*payloadDir, name+".json"), []byte(req.Body), 0o644); err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, map[string]any{"saved": name + ".json"})

	case http.MethodDelete:
		name := strings.TrimSuffix(r.URL.Query().Get("name"), ".json")
		if !safeName.MatchString(name) {
			writeJSON(w, map[string]any{"error": "bad name"})
			return
		}
		if err := os.Remove(filepath.Join(*payloadDir, name+".json")); err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, map[string]any{"deleted": name + ".json"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Not Fprintf: the page's CSS is full of literal % (100%, calc()), which a
	// format string would misread as verbs.
	fmt.Fprint(w, strings.Replace(indexHTML, "{{ADDR}}", *defaultAddr, 1))
}

// fieldLine matches one field of a `describe <message>` block:
//
//	repeated .pkg.T name = 1;   map<string, .pkg.T> name = 2;   string name = 3;
var fieldLine = regexp.MustCompile(`^\s*(repeated\s+|optional\s+)?(map<\s*\w+\s*,\s*\.?([\w.]+)\s*>|\.?([\w.]+))\s+(\w+)\s*=\s*\d+`)

type field struct {
	Type     string `json:"type"` // fully-qualified for messages/enums, bare for scalars
	Kind     string `json:"kind"` // "any" | "msg" | "scalar" (enums count as scalar: JSON strings)
	Repeated bool   `json:"repeated,omitempty"`
	Map      bool   `json:"map,omitempty"`
}

// parseFields reads the fields of one message from describe text, keyed by
// their JSON (lowerCamel) name, which is what the request body uses.
func parseFields(describe string) map[string]field {
	fields := map[string]field{}
	body := describe
	if i := strings.Index(body, "{"); i >= 0 {
		body = body[i+1:]
	}
	for _, line := range strings.Split(body, "\n") {
		m := fieldLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		f := field{Repeated: strings.HasPrefix(m[1], "repeated"), Map: m[3] != ""}
		f.Type = m[3]
		if !f.Map {
			f.Type = m[4]
		}
		switch {
		case f.Type == "google.protobuf.Any":
			f.Kind = "any"
		case strings.Contains(f.Type, "."):
			f.Kind = "msg" // enums are re-labelled scalar once described
		default:
			f.Kind = "scalar"
		}
		fields[jsonName(m[5])] = f
	}
	return fields
}

func jsonName(snake string) string {
	parts := strings.Split(snake, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

var schemaCache sync.Map // addr|tls|root -> map[string]map[string]field

// handleSchema walks the message graph under ?type= through reflection and
// returns every message's fields. The UI uses it to know which positions in a
// body are google.protobuf.Any -- grpcurl's "Any JSON doesn't have '@type'"
// never says which field, so the tool has to.
func handleSchema(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("type")
	if root == "" {
		writeJSON(w, map[string]any{"error": "type is required"})
		return
	}
	addr, tls := addrOf(r), tlsOf(r)
	key := fmt.Sprintf("%s|%v|%s", addr, tls, root)
	if v, ok := schemaCache.Load(key); ok {
		writeJSON(w, map[string]any{"root": root, "messages": v})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	messages := map[string]map[string]field{}
	enums := map[string]bool{}
	queue := []string{root}
	seen := map[string]bool{root: true}
	for len(queue) > 0 {
		sym := queue[0]
		queue = queue[1:]
		d, _, err := describeSymbol(ctx, addr, tls, sym)
		if err != nil {
			continue
		}
		if strings.Contains(d, " is an enum:") {
			enums[sym] = true
			continue
		}
		fields := parseFields(d)
		messages[sym] = fields
		for _, f := range fields {
			if f.Kind == "msg" && !seen[f.Type] {
				seen[f.Type] = true
				queue = append(queue, f.Type)
			}
		}
	}
	for _, fields := range messages {
		for name, f := range fields {
			if enums[f.Type] {
				f.Kind = "scalar"
				fields[name] = f
			}
		}
	}
	schemaCache.Store(key, messages)
	writeJSON(w, map[string]any{"root": root, "messages": messages})
}
