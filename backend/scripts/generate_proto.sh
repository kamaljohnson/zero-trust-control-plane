#!/usr/bin/env bash
# generate_proto.sh: run buf or protoc to generate Go and gRPC code into api/generated/
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT="$(pwd)"
PROTO_DIR="$ROOT/proto"
# Output is under api/generated/; module strip makes go_package path relative to ROOT.
OUT_DIR="$ROOT"
MODULE="zero-trust-control-plane/backend"

if command -v buf >/dev/null 2>&1; then
  echo "buf generate: $PROTO_DIR -> $OUT_DIR"
  (cd "$PROTO_DIR" && buf generate)
  echo "buf generate done."
  exit 0
fi

if ! command -v protoc >/dev/null 2>&1; then
  echo "error: neither buf nor protoc found. Install buf (brew install buf) or protoc (brew install protobuf)."
  exit 1
fi

for plugin in protoc-gen-go protoc-gen-go-grpc; do
  if ! command -v "$plugin" >/dev/null 2>&1; then
    echo "error: $plugin not found. Install with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    echo "  (ensure \$GOPATH/bin or \$HOME/go/bin is on your PATH)"
    exit 1
  fi
done

echo "protoc generate: $PROTO_DIR -> $ROOT/api/generated"
mkdir -p "$ROOT/api/generated"
PROTO_FILES=$(find "$PROTO_DIR" -name '*.proto')
if [[ -z "$PROTO_FILES" ]]; then
  echo "error: no .proto files found under $PROTO_DIR"
  exit 1
fi
protoc \
  -I "$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=module="$MODULE" \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=module="$MODULE" \
  $PROTO_FILES
echo "protoc generate done."
