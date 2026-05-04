#!/usr/bin/env bash
set -euo pipefail

echo "DEBUG: build.sh started for commit $(git log -1 --pretty=%H)"

# build.sh — Parse k8s version from VERSION, update go packages, build binary.
#
# VERSION holds the k8s version (e.g. v1.36).
# This script:
#   1. Reads the k8s minor from VERSION
#   2. Fetches the latest k8s patch for that minor
#   3. Updates all k8s.io/* packages
#   4. Builds the binary
#
# Exit codes:
#   0 = build succeeded (no package update needed or update applied)
#   1 = build failed

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

# --- Read k8s minor from VERSION ---
K8S_MINOR=$(grep -oP '^v\K[0-9]+\.[0-9]+' VERSION | head -1)
if [[ -z "$K8S_MINOR" ]]; then
  echo "ERROR: Could not parse k8s version from VERSION (got: $(cat VERSION))"
  exit 1
fi

MINOR_NUM=$(echo "$K8S_MINOR" | cut -d. -f2)   # "36"

echo "=== build.sh ==="
echo "K8s version from VERSION: $K8S_MINOR"

# --- Fetch latest stable patch for this minor ---
LATEST_PATCH=$(curl -sSfL --max-time 15 \
  "https://api.github.com/repos/kubernetes/kubernetes/releases?per_page=50" \
  | python3 -c "
import sys, json
releases = json.load(sys.stdin)
for r in releases:
    tag = r.get('tag_name','')
    if not tag.startswith('v1.${MINOR_NUM}.'):
        continue
    if 'alpha' in tag or 'beta' in tag or 'rc' in tag:
        continue
    parts = tag.lstrip('v').split('.')
    if len(parts) != 3:
        continue
    print(parts[2])
    break
" 2>/dev/null || true)

if [[ -z "$LATEST_PATCH" ]]; then
  echo "ERROR: Could not determine latest patch for k8s 1.${MINOR_NUM}"
  exit 1
fi

K8S_PKG="v0.${MINOR_NUM}.${LATEST_PATCH}"
echo "Latest k8s patch: $K8S_PKG (k8s 1.${MINOR_NUM}.${LATEST_PATCH})"

# --- Check if update needed ---
CURRENT_PKG_MINOR=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.\K[0-9]+')
CURRENT_PKG_PATCH=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.[0-9]+\.\K[0-9]+')

if [[ "$K8S_PKG" == "v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH}" ]]; then
  echo "go.mod already at $K8S_PKG — no update needed"
else
  echo "Updating k8s packages: v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH} -> $K8S_PKG"
  # Add cel-go replace for k8s 1.32 (cel-go incompatibility in apiserver)
  if [[ "$MINOR_NUM" == "32" ]] && ! grep -q "cel-go v0.22.0" go.mod; then
    echo "replace github.com/google/cel-go => github.com/google/cel-go v0.22.0" >> go.mod
    echo "[cel-go replace added for k8s 1.32]"
  fi
  "$GO" get k8s.io/api@${K8S_PKG} \
         k8s.io/apimachinery@${K8S_PKG} \
         k8s.io/apiextensions-apiserver@${K8S_PKG}
  "$GO" mod tidy
fi

# --- Build ---
echo ""
echo "[build] $GO build ./..."
if ! "$GO" build -o dist/k8-manifest-validator ./cmd/k8-manifest-validator/; then
  echo "BUILD FAILED"
  exit 1
fi

echo "Build successful: dist/k8-manifest-validator"

# --- Update binary version const (seen by --version) ---
VERSION_STR=$(grep -oP '^v\K[0-9]+\.[0-9]+' VERSION | head -1)
CURRENT_CONST=$(grep -oP 'const version = "\Kv[^"]+' cmd/k8-manifest-validator/main.go)
if [[ "$VERSION_STR" != "$CURRENT_CONST" ]]; then
  sed -i "s/const version = \"v[0-9.]*[0-9]\"/const version = \"v${VERSION_STR}\"/" cmd/k8-manifest-validator/main.go
  echo "[version] main.go const updated to v${VERSION_STR}"
fi

exit 0
