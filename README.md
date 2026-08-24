# grpc-lab

A browser UI for exercising gRPC services during development. Point it at a
running server, pick a method, edit JSON, invoke.

```sh
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest   # once
go run . -addr localhost:8082
# open http://127.0.0.1:8090
```

## Why this exists

`grpcui` is the obvious tool and it is good, but it cannot invoke any method
whose request carries a `google.protobuf.Any`. It builds its type registry from
the method's descriptor closure, and an `Any` is opaque to that, so the concrete
type never resolves and the call fails in the browser with
`Unexpected error: error` before reaching the server. That rules out
catalog-service's `CartWorkerService/ExecuteWorker`, where every cart's
`seller_info` is an `Any`.

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
- Generates a request skeleton per method (`grpcurl -msg-template`).
- Saves request bodies to `payloads/` and reloads them in one click.
- Sends a bearer token as `authorization` metadata for authenticated RPCs.
- Shows the raw `grpcurl` output, including status code, message and details on
  failure — the error text is the point, so it is never swallowed.

## Flags

| flag | default | meaning |
|---|---|---|
| `-addr` | `localhost:8082` | default gRPC target (editable in the UI) |
| `-port` | `8090` | port for this UI |
| `-payloads` | `payloads` | directory of saved request bodies |
| `-timeout` | `30s` | per-call timeout |

## Limits

- Unary calls only. Streaming methods will not work.
- `-plaintext` only; no TLS targets yet.
- Requires reflection enabled on the target (go-commons registers it by default).
- Binds to `127.0.0.1` and shells out to `grpcurl` — a local dev tool, not
  something to expose on a network.
