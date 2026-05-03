#!/usr/bin/env bash
set -euo pipefail

# check-new-version.sh - Check if a newer k8s version is available.
#
# Fetches the latest k8s release from GitHub, compares with current
# VERSION file, and optionally creates a release branch.
#
# Usage: ./scripts/check-new-version.sh
#
# Output:
#   "New version v{new}: created branch release/v{new}"  — on upgrade
#   "Already on latest v{current}"                        — no update needed

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

# --- Read current version ---
if [ ! -f VERSION ]; then
  echo "Error: VERSION file not found at $(pwd)/VERSION"
  exit 1
fi

CURRENT_VERSION=$(cat VERSION | tr -d '[:space:]')
CURRENT_K8S=$(echo "$CURRENT_VERSION" | sed 's/^v//' | sed 's/-[0-9]*$//')   # e.g., "1.31"

CURRENT_MAJOR=$(echo "$CURRENT_K8S" | cut -d. -f1)
CURRENT_MINOR=$(echo "$CURRENT_K8S" | cut -d. -f2)

# --- Fetch latest k8s release from GitHub ---
LATEST_TAG=$(curl -sSfL --max-time 10 \
  "https://api.github.com/repos/kubernetes/kubernetes/releases/latest" \
  | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')

if [ -z "$LATEST_TAG" ]; then
  echo "Error: could not fetch latest k8s release from GitHub API"
  exit 1
fi

# Parse tag like "v1.32.3" → "1.32" (drop patch)
LATEST_K8S=$(echo "$LATEST_TAG" | sed 's/^v//' | cut -d. -f1-2)
LATEST_MAJOR=$(echo "$LATEST_K8S" | cut -d. -f1)
LATEST_MINOR=$(echo "$LATEST_K8S" | cut -d. -f2)

echo "Current k8s base: v${CURRENT_K8S}"
echo "Latest k8s base:  v${LATEST_K8S}"

# --- Compare versions (simple numeric comparison) ---
if [ "$LATEST_MAJOR" -gt "$CURRENT_MAJOR" ] || \
   { [ "$LATEST_MAJOR" -eq "$CURRENT_MAJOR" ] && [ "$LATEST_MINOR" -gt "$CURRENT_MINOR" ]; }; then
  NEW_VERSION="v${LATEST_K8S}-0"
  BRANCH_NAME="release/v${LATEST_K8S}"

  # Ensure clean working tree
  if [ -n "$(git status --porcelain)" ]; then
    echo "Error: working tree has uncommitted changes. Commit or stash first."
    exit 1
  fi

  # Create branch from current HEAD
  git checkout -b "$BRANCH_NAME"

  # Update VERSION file
  echo "$NEW_VERSION" > VERSION
  git add VERSION
  git commit -m "release: bump to k8s v${LATEST_K8S}"

  # Checkout original branch so user can review
  git checkout -
  echo ""
  echo "New version v${LATEST_K8S}: created branch ${BRANCH_NAME}"
  echo "Review the branch, then merge and push:"
  echo "  git diff ${BRANCH_NAME}"
  echo "  git merge ${BRANCH_NAME}"
  echo "  git push origin --follow-tags"
else
  echo "Already on latest v${CURRENT_K8S}"
fi
