#!/usr/bin/env bash
# generate_proto.sh: run protoc to generate Go (and optional other) code into api/generated/
set -euo pipefail
cd "$(dirname "$0")/.."
# TODO: add protoc invocations for proto/identity, proto/user, proto/organization, proto/membership, proto/session, proto/device, proto/policy, proto/audit
