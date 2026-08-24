# Contributing

grpc-lab is deliberately small: stdlib Go, one HTML file, and `grpcurl` doing
the gRPC work. Please keep it that way.

- **No new dependencies.** If it needs a module, it probably belongs in grpcurl.
- **No build step.** The UI is a Go string constant; edit `ui.go` directly.
- Run `go vet ./... && go test ./...` before opening a PR.
- To try changes against a real server: `go run . -addr host:port`. For a
  reflection-enabled test target with streaming methods, grpcurl ships one:
  `cd $(go env GOMODCACHE)/github.com/fullstorydev/grpcurl@*/ && go run ./internal/testing/cmd/testserver -p 9555`.
- Bug reports: include the method's `grpcurl describe` output and the request
  body (redact anything sensitive). Most "it doesn't work" reports are a
  `google.protobuf.Any` without `@type` — the UI should have told you which
  field; if it didn't, that's the bug.
