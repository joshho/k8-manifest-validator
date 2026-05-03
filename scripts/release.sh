#!/usr/bin/env bash
set -euo pipefail

# release.sh - Build and dual-release k8-manifest-validator binaries.
#
# Reads VERSION file (e.g., "v1.31") to determine k8s minor version.
# Queries GitHub API for last release tag under that minor to determine
# next patch number (e.g., last tag was v1.31-0 → next is v1.31-1).
#
# Creates two GitHub releases:
#   - v{VERSION}       (stable, overwrites previous stable)
#   - v{VERSION}-{N}  (prerelease, new)
#
# Both receive the same binary artifacts.
#
# Usage: ./scripts/release.sh [--dry-run]
# Environment: GITHUB_TOKEN must be set.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_DIR"

# --- Validate ---
if [ ! -f VERSION ]; then
    echo "Error: VERSION file not found at $REPO_DIR/VERSION"
    exit 1
fi

if [ -z "${GITHUB_TOKEN:-}" ]; then
    echo "Error: GITHUB_TOKEN environment variable not set"
    exit 1
fi

VERSION_FILE_CONTENT=$(cat VERSION)
K8S_MINOR="${VERSION_FILE_CONTENT#v}"   # "v1.31" → "1.31"

GITHUB_REPO="joshho/k8-manifest-validator"

# Parse optional --dry-run flag
DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
    DRY_RUN=true
fi

# --- Determine release tags ---
# Query GitHub API for last release under this minor version
LAST_TAG=$(gh release list --repo "$GITHUB_REPO" --limit 50 2>/dev/null | \
    awk -v minor="$K8S_MINOR" '$2 ~ "^v" minor "-" { print $2 }' | \
    sort -V | tail -1 || true)

if [ -z "$LAST_TAG" ]; then
    # No existing prerelease under this minor — patch starts at 0
    PATCH_NUM=0
else
    # Extract patch number: v1.31-5 → 5
    PATCH_NUM="${LAST_TAG##*-}"
    PATCH_NUM=$((PATCH_NUM + 1))
fi

PRERELEASE_TAG="v${K8S_MINOR}-${PATCH_NUM}"
STABLE_TAG="v${K8S_MINOR}"

echo "Release plan:"
echo "  Stable:  $STABLE_TAG (overwrites previous)"
echo "  Pre:    $PRERELEASE_TAG (new)"
echo ""

if $DRY_RUN; then
    echo "[dry-run] Would create releases and upload artifacts"
    exit 0
fi

# --- Build ---
echo "Building with goreleaser..."
goreleaser build --snapshot --clean --id k8-manifest-validator

# --- Create stable release (overwrites previous) ---
echo "Creating stable release: $STABLE_TAG"
gh release create "$STABLE_TAG" \
    --repo "$GITHUB_REPO" \
    --title "$STABLE_TAG" \
    --notes "Stable release for k8s $K8S_MINOR" \
    dist/k8-manifest-validator_linux_amd64/k8-manifest-validator \
    dist/k8-manifest-validator_linux_arm64/k8-manifest-validator \
    dist/k8-manifest-validator_darwin_amd64/k8-manifest-validator \
    dist/k8-manifest-validator_darwin_arm64/k8-manifest-validator \
    "k8-manifest-validator_${PRERELEASE_TAG}_checksums.txt" \
    2>/dev/null || \
    gh release edit "$STABLE_TAG" \
    --repo "$GITHUB_REPO" \
    --title "$STABLE_TAG" \
    --addAsset "dist/k8-manifest-validator_linux_amd64/k8-manifest-validator" \
    --addAsset "dist/k8-manifest-validator_linux_arm64/k8-manifest-validator" \
    --addAsset "dist/k8-manifest-validator_darwin_amd64/k8-manifest-validator" \
    --addAsset "dist/k8-manifest-validator_darwin_arm64/k8-manifest-validator" \
    --addAsset "k8-manifest-validator_${PRERELEASE_TAG}_checksums.txt"

# --- Create prerelease ---
echo "Creating prerelease: $PRERELEASE_TAG"
gh release create "$PRERELEASE_TAG" \
    --repo "$GITHUB_REPO" \
    --title "$PRERELEASE_TAG" \
    --notes "Prerelease $PRERELEASE_TAG for k8s $K8S_MINOR" \
    --prerelease \
    dist/k8-manifest-validator_linux_amd64/k8-manifest-validator \
    dist/k8-manifest-validator_linux_arm64/k8-manifest-validator \
    dist/k8-manifest-validator_darwin_amd64/k8-manifest-validator \
    dist/k8-manifest-validator_darwin_arm64/k8-manifest-validator \
    "k8-manifest-validator_${PRERELEASE_TAG}_checksums.txt"

echo "Done. Released $PRERELEASE_TAG and updated $STABLE_TAG"
