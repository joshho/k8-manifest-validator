#!/usr/bin/env bash
set -euo pipefail

# test.sh — Run unit tests and binary smoke test.
#
# Exit codes:
#   0 = all tests passed
#   1 = test failed

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

# Find go — use explicit path (works locally and in CI with setup-go having fixed PATH)
if [[ -x /home/node/.local/go/go/bin/go ]]; then
  GO=/home/node/.local/go/go/bin/go
elif command -v go &>/dev/null; then
  GO=$(command -v go)
else
  echo "ERROR: go not found" >&2
  exit 1
fi

echo "=== test.sh ==="
echo ""
echo "[unit tests] go test ./..."
if ! "$GO" test ./...; then
  echo "UNIT TESTS FAILED"
  exit 1
fi
echo "Unit tests passed"

echo ""
echo "[binary smoke test]"
BINARY="dist/k8-manifest-validator"
if [[ ! -f "$BINARY" ]]; then
  echo "ERROR: Binary not found at $BINARY (run build.sh first)"
  exit 1
fi

# Smoke test: --version
VERSION_OUTPUT=$("$BINARY" --version 2>&1 || true)
if [[ -z "$VERSION_OUTPUT" ]]; then
  echo "ERROR: Binary --version produced no output"
  exit 1
fi
echo "Binary version: $VERSION_OUTPUT"

# Smoke test: --help
HELP_OUTPUT=$("$BINARY" --help 2>&1 || true)
if [[ -z "$HELP_OUTPUT" ]]; then
  echo "ERROR: Binary --help produced no output"
  exit 1
fi
echo "Binary help: $HELP_OUTPUT"

echo ""
echo "All tests passed"
exit 0
