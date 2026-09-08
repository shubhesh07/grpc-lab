#!/bin/sh
# Builds grpc-lab and the test gRPC server, starts both, runs e2e.js, tears down.
# PW_CORE / PW_CHROME may point at playwright-core and a chromium binary.
set -e
cd "$(dirname "$0")"
PAY=$(mktemp -d)
(cd testsrv && go build -o testsrv .) && (cd .. && go build -o grpc-lab .)
./testsrv/testsrv >"$PAY/testsrv.log" 2>&1 & T=$!
../grpc-lab -port 8097 -addr 127.0.0.1:50077 -payloads "$PAY/payloads" >"$PAY/grpc-lab.log" 2>&1 & L=$!
trap 'kill $T $L 2>/dev/null; rm -rf "$PAY"' EXIT
sleep 1
PAYLOADS="$PAY/payloads" node e2e.js
