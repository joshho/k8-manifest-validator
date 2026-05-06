#!/usr/bin/env bash
set -euo pipefail

# bump-version.sh - Bump k8s dependency versions across the project.
#
# VERSION holds the k8s version (e.g. v1.36) — the k8s minor only.
# The project patch counter lives in the git tag (v1.36-0, v1.36-1...).
#
# Usage:
#   ./scripts/bump-version.sh --minor <k8s-version>   # e.g. --minor 1.32
#   ./scripts/bump-version.sh --patch                # auto-patch for current minor
#
# Exits 0 if no update needed, 1 on error, 2 if update was applied.

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

MODE=""
TARGET_MINOR=""

# --- Parse arguments ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --patch)
      MODE="patch"
      shift
      ;;
    --minor)
      MODE="minor"
      TARGET_MINOR="$2"
      shift 2
      ;;
    *)
      echo "Usage: $0 {--patch | --minor <k8s-version>}"
      exit 1
      ;;
  esac
done

if [[ -z "$MODE" ]]; then
  echo "Usage: $0 {--patch | --minor <k8s-version>}"
  exit 1
fi

# Find go
GO="${GO:-$(command -v go || /home/node/.local/go/go/bin/go)}"

# --- Read current state ---
CURRENT_VERSION=$(grep -oP '^v\K[0-9]+\.[0-9]+' VERSION | head -1)
CURRENT_K8S_MINOR=$(echo "$CURRENT_VERSION" | sed 's/^v//')

# Read current k8s.io package minor from go.mod (e.g. "v0.31.0" -> "31")
CURRENT_PKG_MINOR=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.\K[0-9]+')
CURRENT_PKG_PATCH=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.[0-9]+\.\K[0-9]+')

echo "=== k8s version bumper ==="
echo "VERSION:       ${CURRENT_VERSION}"
echo "Current pkg:   v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH}"
echo ""

# === MINOR MODE ===
if [[ "$MODE" == "minor" ]]; then
  MAJOR=$(echo "$TARGET_MINOR" | cut -d. -f1)
  MINOR=$(echo "$TARGET_MINOR" | cut -d. -f2)
  NEW_PKG="v0.${MINOR}.0"

  echo "Bumping to k8s ${TARGET_MINOR} (pkg: ${NEW_PKG})"

  # 1. Update VERSION (holds k8s version only: v{minor})
  echo "v${TARGET_MINOR}" > VERSION
  echo "[1/4] VERSION -> v${TARGET_MINOR}"

  # 2. Update go.mod packages (placeholder — will be replaced with actual patch below)
  sed -i "s|^\([[:space:]]*k8s\.io/[a-zA-Z][a-zA-Z0-9/-]*[[:space:]]*\)v0\.[0-9]*\.[0-9]*$|\1${NEW_PKG}|" go.mod
  echo "[2/4] go.mod placeholder -> ${NEW_PKG}"

  # 2b. Add cel-go replace directive for k8s 1.32 (cel-go incompatibility in apiserver)
  if [[ "$MINOR" == "32" ]] && ! grep -q "cel-go v0.22.0" go.mod; then
    echo "replace github.com/google/cel-go => github.com/google/cel-go v0.22.0" >> go.mod
    echo "[2b] cel-go replace added for k8s 1.32"
  fi

  # 3. Fetch latest patch for target minor from GitHub
  LATEST_PATCH=$(curl -sSfL --max-time 15 \
    "https://api.github.com/repos/kubernetes/kubernetes/releases?per_page=50" \
    | python3 -c "
import sys, json
releases = json.load(sys.stdin)
for r in releases:
    tag = r.get('tag_name','')
    if not tag.startswith('v1.${MINOR}.'):
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
    echo "Could not determine latest patch for k8s 1.${MINOR}. Using 0."
    LATEST_PATCH="0"
  fi

  ACTUAL_PKG="v0.${MINOR}.${LATEST_PATCH}"
  echo "Using k8s packages: ${ACTUAL_PKG} (k8s 1.${MINOR}.${LATEST_PATCH})"

  # 4. go get all k8s packages at actual version
  "$GO" get k8s.io/api@${ACTUAL_PKG} \
       k8s.io/apimachinery@${ACTUAL_PKG} \
       k8s.io/apiextensions-apiserver@${ACTUAL_PKG}
  echo "[4/6] go get k8s packages -> ${ACTUAL_PKG}"

  # 5. go mod tidy
  "$GO" mod tidy
  echo "[5/6] go mod tidy"

  # 6. Build
  echo "[6/6] go build ./..."
  if ! "$GO" build ./...; then
    echo "BUILD FAILED — rolling back"
    git checkout -- VERSION go.mod go.sum
    exit 1
  fi

  # 7. Commit (VERSION only — no tag; CI handles that)
  git add VERSION go.mod go.sum
  git commit -m "release: k8s 1.${MINOR}.${LATEST_PATCH} (v0.${MINOR}.${LATEST_PATCH})"

  echo ""
  echo "=== Minor bump complete ==="
  echo "VERSION: v${TARGET_MINOR}"
  echo "Packages: ${ACTUAL_PKG}"
  echo ""
  echo "Push with: git push origin main"
  exit 2
fi

# === PATCH MODE ===
if [[ "$MODE" == "patch" ]]; then
  K8S_MINOR_NUM="$CURRENT_PKG_MINOR"

  echo "Checking latest k8s v${K8S_MINOR_NUM}.x release..."

  LATEST_PATCH=$(curl -sSfL --max-time 15 \
    "https://api.github.com/repos/kubernetes/releases?per_page=50" \
    | python3 -c "
import sys, json
releases = json.load(sys.stdin)
for r in releases:
    tag = r.get('tag_name','')
    if not tag.startswith('v1.${K8S_MINOR_NUM}.'):
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
    echo "Could not determine latest patch for k8s ${K8S_MINOR_NUM}. Exiting (no update)."
    exit 0
  fi

  CURRENT_PATCH="$CURRENT_PKG_PATCH"
  echo "Current patch: ${CURRENT_PATCH}, Latest patch: ${LATEST_PATCH}"

  if [[ "$LATEST_PATCH" -le "$CURRENT_PATCH" ]]; then
    echo "Already on latest patch (v0.${K8S_MINOR_NUM}.${CURRENT_PATCH}). No update needed."
    exit 0
  fi

  NEW_PKG="v0.${K8S_MINOR_NUM}.${LATEST_PATCH}"

  echo ""
  echo "Bumping to k8s ${NEW_PKG}"
  echo "Updating k8s packages: v0.${K8S_MINOR_NUM}.${CURRENT_PATCH} -> ${NEW_PKG}"

  "$GO" get k8s.io/api@${NEW_PKG} \
         k8s.io/apimachinery@${NEW_PKG} \
         k8s.io/apiextensions-apiserver@${NEW_PKG}
  echo "[1/4] go get k8s packages -> ${NEW_PKG}"

  "$GO" mod tidy
  echo "[2/4] go mod tidy"

  echo "[3/4] go build ./..."
  if ! "$GO" build ./...; then
    echo "BUILD FAILED — rolling back"
    git checkout -- go.mod go.sum
    exit 1
  fi

  git add go.mod go.sum
  git commit -m "release: patch k8s to v0.${K8S_MINOR_NUM}.${LATEST_PATCH}"

  echo ""
  echo "=== Patch bump complete ==="
  echo "Packages: ${NEW_PKG}"
  echo ""
  echo "Push with: git push origin release/v${K8S_MINOR_NUM}"
  exit 2
fi
