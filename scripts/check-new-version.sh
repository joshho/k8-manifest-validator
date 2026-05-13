#!/usr/bin/env bash
set -euo pipefail

# check-new-version.sh - Detect newer k8s minor versions.
#
# Usage:
#   ./scripts/check-new-version.sh --minor     # check for new minor release
#   ./scripts/check-new-version.sh --patch     # check for new patch (current minor)
#
# Exits 0 with no output if already current.
# Exits 0 and prints "NEW_MINOR=v1.XX" if new minor found.
# Exits 0 and prints "NEW_PATCH=v1.XX.Y" if new patch found.

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

MODE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --minor) MODE="minor"; shift ;;
    --patch) MODE="patch"; shift ;;
    *) echo "Usage: $0 {--minor|--patch}"; exit 1 ;;
  esac
done

if [[ -z "$MODE" ]]; then
  echo "Usage: $0 {--minor|--patch}"
  exit 1
fi

# Read current state
CURRENT_VERSION=$(cat VERSION | tr -d '[:space:]')
CURRENT_K8S=$(echo "$CURRENT_VERSION" | sed 's/^v//' | sed 's/-[0-9]*$//')

# Read current k8s package minor from go.mod
CURRENT_PKG_MINOR=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.\K[0-9]+')

echo "=== k8s version checker ==="
echo "MODE:            ${MODE}"
echo "VERSION:         ${CURRENT_VERSION}"
echo "K8s minor:       ${CURRENT_K8S}"
echo "Package minor:   ${CURRENT_PKG_MINOR}"

# === MINOR MODE ===
if [[ "$MODE" == "minor" ]]; then
  # Fetch latest k8s release tag
  LATEST_TAG=$(curl -sSfL --max-time 15 \
    "https://api.github.com/repos/kubernetes/kubernetes/releases/latest" \
    | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(data.get('tag_name',''))
" 2>/dev/null || true)

  if [[ -z "$LATEST_TAG" ]]; then
    echo "Error: could not fetch latest k8s release"
    exit 1
  fi

  LATEST_K8S=$(echo "$LATEST_TAG" | sed 's/^v//' | cut -d. -f1,2)
  LATEST_MAJOR=$(echo "$LATEST_K8S" | cut -d. -f1)
  LATEST_MINOR=$(echo "$LATEST_K8S" | cut -d. -f2)
  CURRENT_MAJOR=$(echo "$CURRENT_K8S" | cut -d. -f1)
  CURRENT_MINOR=$(echo "$CURRENT_K8S" | cut -d. -f2)

  echo "Latest k8s:      v${LATEST_K8S}"

  if [[ "$LATEST_MAJOR" -gt "$CURRENT_MAJOR" ]] || \
     { [[ "$LATEST_MAJOR" -eq "$CURRENT_MAJOR" ]] && [[ "$LATEST_MINOR" -gt "$CURRENT_MINOR" ]]; }; then
    echo "NEW_MINOR=v${LATEST_K8S}"
    exit 0
  else
    echo "Already on latest minor (v${CURRENT_K8S})"
    exit 0
  fi
fi

# === PATCH MODE ===
if [[ "$MODE" == "patch" ]]; then
  # Read current patch from go.mod
  CURRENT_PKG_PATCH=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.[0-9]+\.\K[0-9]+')

  echo "Current patch:   v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH}"

  # Fetch latest patch for current minor
  LATEST_PATCH=$(curl -sSfL --max-time 15 \
    "https://api.github.com/repos/kubernetes/kubernetes/releases?per_page=50" \
    | python3 -c "
import sys, json
releases = json.load(sys.stdin)
for r in releases:
    tag = r.get('tag_name','')
    if not tag.startswith('v1.${CURRENT_PKG_MINOR}.'):
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
    echo "Could not determine latest patch for k8s ${CURRENT_PKG_MINOR}. Assuming current."
    exit 0
  fi

  # IMPORTANT: k8s.io/api, k8s.io/apimachinery, k8s.io/apiextensions-apiserver
  # only tag releases at .0 patch. They do NOT have .1, .2 etc tags.
  # Force patch to 0 — go.get always uses v0.N.0
  LATEST_PATCH="0"

  echo "Latest patch:    v1.${CURRENT_PKG_MINOR}.${LATEST_PATCH} (forced .0 — k8s.io packages only tag .0)"

  if [[ "$LATEST_PATCH" -gt "$CURRENT_PKG_PATCH" ]]; then
    echo "NEW_PATCH=v1.${CURRENT_PKG_MINOR}.${LATEST_PATCH}"
    exit 0
  else
    echo "Already on latest patch"
    exit 0
  fi
fi