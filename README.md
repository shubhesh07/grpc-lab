# grpc-lab

[![Go](https://img.shields.io/badge/go-stdlib%20only-5eead4)](go.mod) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A browser UI for exercising gRPC services during development. Point it at a
running server, pick a method, edit JSON, invoke.

```sh
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest   # once
go run . -addr localhost:8086
# open http://127.0.0.1:8090
```

## Why this exists

`grpcui` is the obvious tool and it is good, but it cannot invoke any method
whose request carries a `google.protobuf.Any`. It builds its type registry from
the method's descriptor closure, and an `Any` is opaque to that, so the concrete
type never resolves and the call fails in the browser with
`Unexpected error: error` before reaching the server. That rules out any
service whose requests carry an `Any` — the case that motivated this tool was a
cart worker where every cart's `seller_info` is one.

Passing `-protoset` with the concrete types does get the request out of the
browser, but then reflection is disabled and the instance only exposes whatever
is in the descriptor set.

grpc-lab sidesteps the problem: it does not speak gRPC at all. Every call shells
out to `grpcurl`, which resolves `Any` by asking the server's reflection service
for the type on demand. The UI inherits correct behaviour for free and has no
protobuf dependencies — stdlib Go and one HTML file, no build step.

## What it does

- Lists services and methods from **server reflection** — no `.proto` files, no
  descriptor sets, and it cannot drift from what the server actually serves.
  Each method shows its request/response types and a `client stream` /
  `server stream` / `bidi` badge.
- Generates a request skeleton from the method's **real input type** (resolved
  through `describe`, not guessed from the method name).
- **`google.protobuf.Any` resolver** — every `Any` position in the body is
  listed under the editor automatically, with a picker of every message type the server knows (walked via reflection) and
  splices in that type's template with the right `@type` URL.
- Streaming: server-stream responses print in order; client/bidi streams take
  several JSON objects one after another in the editor.
- **Any lint** — before every call (the body is also auto-formatted) the body is checked against the request
  schema (walked via reflection); any `Any`-typed position without `@type`
  blocks the call with its exact path (grpcurl's own error never says which
  field). Delete the field or fill it from the picker.
- The request is a **collapsible JSON tree that is also the editor**:
  click a row and press `Delete` (or the hover `✕`) to remove it, `⌘Z` to
  undo; double-click a value to edit it inline; `✎` on any object/array (or the
  root) opens it as JSON text to replace wholesale — paste a whole body there,
  `//` and `/* */` comments allowed; the `//` toggle on a key disables it
  (kept as `"//key"`, dropped when sent); the editor also offers `copy JSON`
  and `delete`. Selecting text in the tree copies clean JSON. Client streams
  get `+ message`.
- The response renders as the same tree (collapse/expand all, `raw` for text).
- **Request tabs** — `+` opens an independent workspace (method + body);
  tabs persist across reloads.
- **Paste grpcurl** — paste a full `grpcurl … host:port svc/Method` command;
  target, TLS flag, headers, bearer token, body and method are filled in.
- Tab indents in the editor; `⌃⏎` invokes.
- `Describe` tab shows the proto text of the method, request and response.
- `grpcurl` tab gives the exact pasteable command for the last call.
- Metadata headers (`k: v` lines), bearer token, TLS (`-insecure`),
  `-emit-defaults`, `-v` verbose mode, `{{var}}` substitution from a
  variables box, `⌃⏎` to invoke.
- Saves request bodies to `payloads/` (delete with ×), keeps a call history
  (`payloads/.history.jsonl`, last 100) that reloads a request in one click,
  and remembers addr/token/headers/body per method in `localStorage`.
- Errors come back as `-format-error` JSON with the gRPC code name in the
  status pill; nothing is swallowed.

## Flags

| flag | default | meaning |
|---|---|---|
| `-addr` | `localhost:8086` | default gRPC target (editable in the UI) |
| `-port` | `8090` | port for this UI |
| `-payloads` | `payloads` | directory of saved request bodies |
| `-timeout` | `30s` | per-call timeout |
| `-bind` | `127.0.0.1` | interface to listen on (`0.0.0.0` inside Docker) |

## Limits

- `-insecure` TLS only (no CA/client certs yet).
- Requires reflection enabled on the target.
- The `Any` type picker lists every message in files reachable from the
  services. A type that lives in a file nothing references can still be
  typed by hand — the server resolves it at invoke time.
- Binds to `127.0.0.1` and shells out to `grpcurl` — a local dev tool, not
  something to expose on a network.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). MIT licensed.
