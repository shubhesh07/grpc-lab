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
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	defaultAddr = flag.String("addr", "localhost:8082", "default gRPC target")
	uiPort      = flag.Int("port", 8090, "port for this UI")
	payloadDir  = flag.String("payloads", "payloads", "directory of saved request bodies")
	callTimeout = flag.Duration("timeout", 30*time.Second, "per-call timeout")
)

// safeName keeps saved payloads inside payloadDir. Names arrive from the
// browser, so anything with a path separator or dot-dot is rejected outright
// rather than sanitised -- an obvious rule beats a clever one.
var safeName = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func main() {
	flag.Parse()
	if _, err := exec.LookPath("grpcurl"); err != nil {
		log.Fatal("grpcurl not found on PATH. Install: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest")
	}
	if err := os.MkdirAll(*payloadDir, 0o755); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/methods", handleMethods)
	http.HandleFunc("/api/template", handleTemplate)
	http.HandleFunc("/api/invoke", handleInvoke)
	http.HandleFunc("/api/payloads", handlePayloads)

	addr := fmt.Sprintf("127.0.0.1:%d", *uiPort)
	log.Printf("grpc-lab  http://%s   target=%s   payloads=%s", addr, *defaultAddr, *payloadDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// grpcurlRun executes grpcurl and returns combined output. Errors are returned
// alongside the output, never instead of it: grpcurl writes the useful part of
// a failure (status code, message, details) to stderr, so discarding it on a
// non-zero exit would throw away exactly what the developer needs to see.
func grpcurlRun(ctx context.Context, stdin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "grpcurl", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
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

// handleMethods lists every fully-qualified method on the target, grouped by
// service, so the UI never needs a .proto file or a descriptor set.
func handleMethods(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	out, err := grpcurlRun(ctx, "", "-plaintext", addrOf(r), "list")
	if err != nil {
		writeJSON(w, map[string]any{"error": strings.TrimSpace(out)})
		return
	}

	type svc struct {
		Name    string   `json:"name"`
		Methods []string `json:"methods"`
	}
	var services []svc
	for _, name := range strings.Fields(out) {
		// Reflection and health are noise for day-to-day testing.
		if strings.HasPrefix(name, "grpc.reflection.") || name == "grpc.health.v1.Health" {
			continue
		}
		mo, merr := grpcurlRun(ctx, "", "-plaintext", addrOf(r), "list", name)
		if merr != nil {
			continue
		}
		methods := strings.Fields(mo)
		sort.Strings(methods)
		services = append(services, svc{Name: name, Methods: methods})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	writeJSON(w, map[string]any{"services": services})
}

// handleTemplate returns grpcurl's own describe output for the request message,
// which gives a skeleton to edit rather than a blank page.
func handleTemplate(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")
	if method == "" {
		writeJSON(w, map[string]any{"error": "method is required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	out, err := grpcurlRun(ctx, "", "-plaintext", "-msg-template", addrOf(r), "describe", method+"Request")
	if err != nil {
		// Not every request message follows the <Method>Request convention;
		// an empty template is a fine starting point, not an error worth
		// blocking on.
		writeJSON(w, map[string]any{"template": "{}"})
		return
	}
	if i := strings.Index(out, "{"); i >= 0 {
		out = out[i:]
	}
	writeJSON(w, map[string]any{"template": strings.TrimSpace(out)})
}

func handleInvoke(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Addr    string `json:"addr"`
		Method  string `json:"method"`
		Payload string `json:"payload"`
		Token   string `json:"token"`
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

	args := []string{"-plaintext", "-max-time", fmt.Sprintf("%.0f", callTimeout.Seconds())}
	if req.Token != "" {
		args = append(args, "-H", "authorization: Bearer "+req.Token)
	}
	args = append(args, "-d", "@", req.Addr, req.Method)

	started := time.Now()
	out, err := grpcurlRun(ctx, req.Payload, args...)
	writeJSON(w, map[string]any{
		"output": out,
		"ok":     err == nil,
		"ms":     time.Since(started).Milliseconds(),
	})
}

// handlePayloads lists saved bodies on GET and saves one on POST, so a working
// request can be kept without leaving the browser.
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
