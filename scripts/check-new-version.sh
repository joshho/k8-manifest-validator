#!/usr/bin/env bash
set -euo pipefail

# check-new-version.sh - Detect newer k8s minor versions.
#
# Usage:
#   ./scripts/check-new-version.sh --minor     # check for new minor release
#   ./scripts/check-new-version.sh --patch    # check for new patch (current minor)
#
# Exits 0 with no output if already current.
# Exits 0 and prints "NEW_MINOR=v1.XX" if new minor found.
# Exits 0 and prints "NEW_PATCH=v1.XX.0" if new patch found.

SCRIPT_DIR="$(cd "$(dirname "$0")/" && pwd)"
cd "$SCRIPT_DIR"

MODE=""
CURRENT_VERSION=""
CURRENT_K8S=""
CURRENT_PKG_MINOR=""

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

CURRENT_VERSION=$(cat VERSION | tr -d '[:space:]')
CURRENT_K8S=$(echo "$CURRENT_VERSION" | sed 's/^v//' | sed 's/-[0-9]*$//')
CURRENT_PKG_MINOR=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.\K[0-9]+')
CURRENT_PKG_PATCH=$(grep 'k8s\.io/api ' go.mod | grep -oP 'v0\.[0-9]+\.\K[0-9]+')

echo "=== k8s version checker ==="
echo "MODE:            ${MODE}"
echo "VERSION:         ${CURRENT_VERSION}"
echo "K8s minor:       ${CURRENT_K8S}"
echo "Package minor:   v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH}"

# === MINOR MODE ===
if [[ "$MODE" == "minor" ]]; then
  LATEST_TAG=$(curl -sSfL --max-time 15 \
    "https://api.github.com/repos/kubernetes/kubernetes/releases/latest" \
    | python3 -c "import sys,json;print(json.load(sys.stdin).get('tag_name',''))" 2>/dev/null || true)

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
# For release branches: checks if k8s has released a newer minor than what
# go.mod currently has. k8s.io packages only tag at .0, so a patch bump
# means bumping to a newer {minor}.0 when a new release branch is cut.
if [[ "$MODE" == "patch" ]]; then
  echo "Current:        v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH}"

  K8S_LATEST=$(curl -sSfL --max-time 15 \
    "https://api.github.com/repos/kubernetes/kubernetes/releases/latest" \
    | python3 -c "import sys,json;print(json.load(sys.stdin).get('tag_name',''))" 2>/dev/null || true)

  if [[ -z "$K8S_LATEST" ]]; then
    echo "Could not fetch latest k8s release"
    exit 0
  fi

  K8S_LATEST_MINOR=$(echo "$K8S_LATEST" | sed 's/^v//' | cut -d. -f2)
  K8S_LATEST_PATCH=$(echo "$K8S_LATEST" | sed 's/^v//' | cut -d. -f3)
  echo "Latest k8s:     $K8S_LATEST (minor $K8S_LATEST_MINOR, patch $K8S_LATEST_PATCH)"

  # Only bump if k8s is EXACTLY one minor ahead — release branches are pinned
  # to their own minor. A multi-minor jump (e.g. v0.32.0 → v0.36.0) would break
  # the build because the codebase uses APIs that changed between those versions.
  DIFF=$((K8S_LATEST_MINOR - CURRENT_PKG_MINOR))
  if [[ "$DIFF" -gt 1 ]]; then
    echo "k8s jumped multiple minors (v0.${CURRENT_PKG_MINOR} → v0.${K8S_LATEST_MINOR}); refusing to bump (would break build)"
    exit 0
  elif [[ "$K8S_LATEST_MINOR" -gt "$CURRENT_PKG_MINOR" ]]; then
    echo "New k8s minor detected: v1.${K8S_LATEST_MINOR}.0 (go.mod is v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH})"
    echo "NEW_PATCH=v1.${K8S_LATEST_MINOR}.0"
    exit 0
  elif [[ "$K8S_LATEST_MINOR" -eq "$CURRENT_PKG_MINOR" ]] && \
       [[ "$K8S_LATEST_PATCH" -gt "$CURRENT_PKG_PATCH" ]]; then
    echo "New k8s patch available: v1.${K8S_LATEST_MINOR}.${K8S_LATEST_PATCH} (go.mod is v0.${CURRENT_PKG_MINOR}.${CURRENT_PKG_PATCH})"
    echo "NEW_PATCH=v1.${K8S_LATEST_MINOR}.${K8S_LATEST_PATCH}"
    exit 0
  else
    echo "Already on latest patch"
    exit 0
  fi
fi