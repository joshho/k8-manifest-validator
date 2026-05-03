#!/usr/bin/env bash
set -euo pipefail

# bump-version.sh - Bump k8s dependency versions across the project.
#
# Usage: ./scripts/bump-version.sh <k8s-version> [--build]
#
# Examples:
#   ./scripts/bump-version.sh 1.32        # bump to k8s 1.32, commit, tag
#   ./scripts/bump-version.sh 1.33 --build # bump, commit, tag, and build

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

# --- Validation ---
if [ $# -lt 1 ]; then
  echo "Usage: $0 <k8s-version> [--build]"
  echo "Example: $0 1.32"
  exit 1
fi

K8S_VER="$1"
BUILD_FLAG="${2:-}"

# Validate version format (should be like "1.32" or "1.31")
if ! echo "$K8S_VER" | grep -qE '^[0-9]+\.[0-9]+$'; then
  echo "Error: version must be in format MAJOR.MINOR (e.g., 1.32)"
  exit 1
fi

MAJOR=$(echo "$K8S_VER" | cut -d. -f1)
MINOR=$(echo "$K8S_VER" | cut -d. -f2)
NEW_TAG="v${K8S_VER}-0"

# --- Pre-checks ---
if [ ! -f VERSION ]; then
  echo "Error: VERSION file not found at $(pwd)/VERSION"
  exit 1
fi

CURRENT_VERSION=$(cat VERSION | tr -d '[:space:]')
echo "Current version: ${CURRENT_VERSION}"
echo "Target version:  ${NEW_TAG}"

# Ensure clean working tree
if [ -n "$(git status --porcelain)" ]; then
  echo "Error: working tree has uncommitted changes. Commit or stash first."
  exit 1
fi

# --- 1. Update VERSION file ---
echo "$NEW_TAG" > VERSION
echo "[1/5] Updated VERSION → ${NEW_TAG}"

# --- 2. Update k8s.io/* deps in go.mod ---
# Only update packages versioned as v0.XX.0 (matches real k8s modules).
# Skips klog/v2 (v2.XX.0), kube-openapi (v0.0.0-*), utils (v0.0.0-*), etc.
NEW_K8S_VER="v0.${MINOR}.0"
sed -i "s|^\([[:space:]]*k8s\.io/[a-zA-Z][a-zA-Z0-9/-]*[[:space:]]*\)v0\.[0-9]*\.[0-9]*$|\1${NEW_K8S_VER}|" go.mod
echo "[2/5] Updated k8s.io/* deps in go.mod → ${NEW_K8S_VER}"

# --- 3. Update version const in main.go ---
MAIN_GO="cmd/k8-manifest-validator/main.go"
sed -i 's/^\([[:space:]]*const version = "\)v[0-9]*\.[0-9]*-[0-9]*"/\1'"${NEW_TAG}"'"/' "$MAIN_GO"
echo "[3/5] Updated version const in ${MAIN_GO} → ${NEW_TAG}"

# --- 4. Run go mod tidy ---
echo "[4/5] Running go mod tidy..."
go mod tidy
echo "[4/5] go mod tidy completed"

# --- 5. Commit and tag ---
git add VERSION go.mod go.sum "$MAIN_GO"
git commit -m "release: bump to k8s v${K8S_VER}"
git tag "$NEW_TAG"

echo "[5/5] Committed and tagged ${NEW_TAG}"

# --- Optional: Build ---
if [ "$BUILD_FLAG" = "--build" ]; then
  echo ""
  echo "--- Building with goreleaser (snapshot) ---"
  if command -v goreleaser &>/dev/null; then
    goreleaser build --snapshot --clean
    echo "Build complete."
  else
    echo "Warning: goreleaser not found in PATH. Skipping build."
  fi
fi

echo ""
echo "Done! To push: git push origin --follow-tags"
