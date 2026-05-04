#!/usr/bin/env bash
set -euo pipefail

# bump-version.sh - Bump k8s dependency versions across the project.
#
# Usage:
#   ./scripts/bump-version.sh [--minor <k8s-version>]   # e.g. --minor 1.32
#   ./scripts/bump-version.sh --patch                  # auto-patch for current minor
#   ./scripts/bump-version.sh --patch --minor 1.31    # specific minor
#
# Behavior:
#   --patch         Fetches latest k8s patch for the current minor from go.mod,
#                   updates all k8s.io/* packages, bumps VERSION patch counter,
#                   commits and tags. No push (CI handles that via tag trigger).
#
#   --minor <ver>   Bumps to a new minor version (1.31 -> 1.32), updates all
#                   k8s.io/* packages to v0.{new_minor}.0, resets VERSION to v{new}-0.
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

# --- Read current state ---
CURRENT_VERSION=$(cat VERSION | tr -d '[:space:]')
CURRENT_K8S_MINOR=$(echo "$CURRENT_VERSION" | sed 's/^v//' | sed 's/-[0-9]*$//' | cut -d. -f1,2)

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
  NEW_TAG="v${TARGET_MINOR}-0"
  NEW_PKG="v0.${MINOR}.0"

  echo "Bumping to k8s ${TARGET_MINOR} (pkg: ${NEW_PKG})"

  # 1. Update VERSION
  echo "$NEW_TAG" > VERSION
  echo "[1/4] VERSION -> ${NEW_TAG}"

  # 2. Update go.mod packages
  sed -i "s|^\([[:space:]]*k8s\.io/[a-zA-Z][a-zA-Z0-9/-]*[[:space:]]*\)v0\.[0-9]*\.[0-9]*$|\1${NEW_PKG}|" go.mod
  echo "[2/4] go.mod -> ${NEW_PKG}"

  # 2b. Add cel-go replace directive if not present (fixes k8s 1.32 cel issue)
  if ! grep -q "cel-go v0.22.0" go.mod; then
    echo "replace github.com/google/cel-go => github.com/google/cel-go v0.22.0" >> go.mod
    echo "[2b] cel-go replace directive added"
  fi

  # 3. Update version const in main.go
  sed -i 's/^\([[:space:]]*const version = "\)v[0-9]*\.[0-9]*-[0-9]*"/\1'"${NEW_TAG}"'"/' cmd/k8-manifest-validator/main.go
  echo "[3/4] main.go version const -> ${NEW_TAG}"

  # 4. Fetch latest patch for target minor
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

  # Use latest patch version, not .0
  ACTUAL_PKG="v0.${MINOR}.${LATEST_PATCH}"
  echo "Using k8s packages: ${ACTUAL_PKG} (k8s 1.${MINOR}.${LATEST_PATCH})"

  # 5. go get all k8s packages at actual version
  go get k8s.io/api@${ACTUAL_PKG} \
       k8s.io/apimachinery@${ACTUAL_PKG} \
       k8s.io/apiextensions-apiserver@${ACTUAL_PKG}
  echo "[5/6] go get k8s packages -> ${ACTUAL_PKG}"

  # 6. go mod tidy + build test
  go mod tidy
  echo "[6/6] go mod tidy"

  echo "[7/7] go build ./..."
  if ! go build ./...; then
    echo "BUILD FAILED — rolling back"
    git checkout -- VERSION go.mod go.sum cmd/k8-manifest-validator/main.go
    exit 1
  fi

  echo "[8/8] go test ./..."
  if ! go test ./...; then
    echo "TESTS FAILED — rolling back"
    git checkout -- VERSION go.mod go.sum cmd/k8-manifest-validator/main.go
    exit 1
  fi

  git add VERSION go.mod go.sum cmd/k8-manifest-validator/main.go
  git commit -m "release: bump to k8s v${TARGET_MINOR}.${LATEST_PATCH} (v0.${MINOR}.${LATEST_PATCH})"
  git tag "$NEW_TAG"

  echo ""
  echo "=== Minor bump complete ==="
  echo "VERSION: ${NEW_TAG}"
  echo "Tag:     ${NEW_TAG}"
  echo "Packages: ${NEW_PKG}"
  echo ""
  echo "Push with: git push origin --follow-tags"
  exit 2
fi

# === PATCH MODE ===
if [[ "$MODE" == "patch" ]]; then
  # Fetch latest k8s release for current minor
  K8S_MINOR_NUM="$CURRENT_PKG_MINOR"

  echo "Checking latest k8s v${K8S_MINOR_NUM}.x release..."

  # Fetch latest release for this minor from GitHub by listing tags
  LATEST_PATCH=$(curl -sSfL --max-time 15 \
    "https://api.github.com/repos/kubernetes/kubernetes/releases?per_page=50" \
    | python3 -c "
import sys, json
releases = json.load(sys.stdin)
for r in releases:
    tag = r.get('tag_name','')
    if not tag.startswith('v1.${K8S_MINOR_NUM}.'):
        continue
    # skip prereleases/rcs for stable patch detection
    if 'alpha' in tag or 'beta' in tag or 'rc' in tag:
        continue
    # strip 'v' and get patch number
    parts = tag.lstrip('v').split('.')
    if len(parts) != 3:
        continue
    print(parts[2])  # patch number
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
  NEW_K8S_VER="v1.${K8S_MINOR_NUM}.${LATEST_PATCH}"

  # Compute new VERSION patch counter
  # Current VERSION might be v1.31 or v1.31-0 or v1.31-1
  # New VERSION: bump patch counter
  CURRENT_PATCH_NUM=$(echo "$CURRENT_VERSION" | grep -oP '-[0-9]+$' | tr -d '-' || echo "0")
  NEW_PATCH_NUM=$((CURRENT_PATCH_NUM + 1))
  MAJOR=$(echo "$CURRENT_K8S_MINOR" | cut -d. -f1)
  MINOR=$(echo "$CURRENT_K8S_MINOR" | cut -d. -f2)
  NEW_VERSION="v${MAJOR}.${MINOR}-${NEW_PATCH_NUM}"

  echo ""
  echo "Bumping to k8s ${NEW_K8S_VER} (pkg: ${NEW_PKG})"
  echo "VERSION: ${CURRENT_VERSION} -> ${NEW_VERSION}"
  echo "K8s version: v1.${K8S_MINOR_NUM}.${LATEST_PATCH}"

  # 1. Update go.mod packages
  go get k8s.io/api@${NEW_PKG} \
         k8s.io/apimachinery@${NEW_PKG} \
         k8s.io/apiextensions-apiserver@${NEW_PKG} \
         k8s.io/api/core/v1@${NEW_PKG} \
         k8s.io/apimachinery/pkg/api@${NEW_PKG} \
         k8s.io/apimachinery/pkg/apis/meta/v1@${NEW_PKG} \
         k8s.io/kubernetes@${NEW_K8S_VER}
  echo "[1/4] go get k8s packages -> ${NEW_PKG}"

  # 2. go mod tidy
  go mod tidy
  echo "[2/4] go mod tidy"

  # 3. Update VERSION
  echo "$NEW_VERSION" > VERSION
  echo "[3/4] VERSION -> ${NEW_VERSION}"

  # 4. Update version const in main.go
  sed -i 's/^\([[:space:]]*const version = "\)v[0-9]*\.[0-9]*-[0-9]*"/\1'"${NEW_VERSION}"'"/' cmd/k8-manifest-validator/main.go
  echo "[4/4] main.go version const -> ${NEW_VERSION}"

  # 5. Commit and tag
  git add VERSION go.mod go.sum cmd/k8-manifest-validator/main.go
  git commit -m "release: patch k8s to v${NEW_K8S_VER}"
  git tag "${NEW_VERSION}"

  echo ""
  echo "=== Patch bump complete ==="
  echo "VERSION: ${NEW_VERSION}"
  echo "Tag:     ${NEW_VERSION}"
  echo "Packages: ${NEW_PKG}"
  echo ""
  echo "Push with: git push origin --follow-tags"
  exit 2
fi