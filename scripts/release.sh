#!/usr/bin/env bash
set -euo pipefail

# release.sh - Build and dual-release k8-manifest-validator binaries.
#
# Triggered by release.yml (build job runs first and produces dist/ artifacts).
# Expects: dist/k8-manifest-validator_{os}_{arch}/k8-manifest-validator for
# linux/darwin × amd64/arm64.
#
# Reads k8s minor from VERSION file (e.g., "v1.31").
# Queries gh release list for last prerelease tag under that minor to determine
# next patch number.
#
# Creates two GitHub releases:
#   - v{VERSION}       (stable/latest, overwrites previous)
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
    echo "Error: GITHUB_TOKEN environment variable is not set"
    exit 1
fi

if [ ! -d dist ]; then
    echo "Error: dist/ directory not found. Run goreleaser build first."
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
LAST_TAG=$(gh release list --repo "$GITHUB_REPO" --limit 50 2>/dev/null | \
    awk -v minor="$K8S_MINOR" '$2 ~ "^v" minor "-" { print $2 }' | \
    sort -V | tail -1 || true)

if [ -z "$LAST_TAG" ]; then
    PATCH_NUM=0
else
    PATCH_NUM="${LAST_TAG##*-}"
    PATCH_NUM=$((PATCH_NUM + 1))
fi

PRERELEASE_TAG="v${K8S_MINOR}-${PATCH_NUM}"
STABLE_TAG="v${K8S_MINOR}"

echo "Release plan:"
echo "  Stable:  $STABLE_TAG (overwrites previous)"
echo "  Pre:     $PRERELEASE_TAG (new)"
echo ""

if $DRY_RUN; then
    echo "[dry-run] Would create releases and upload artifacts"
    exit 0
fi

# --- Collect artifacts (goreleaser output layout) ---
ARTIFACTS=(
    "dist/k8-manifest-validator_linux_amd64_v1/k8-manifest-validator"
    "dist/k8-manifest-validator_linux_arm64_v8.0/k8-manifest-validator"
    "dist/k8-manifest-validator_darwin_amd64_v1/k8-manifest-validator"
    "dist/k8-manifest-validator_darwin_arm64_v8.0/k8-manifest-validator"
)

for artifact in "${ARTIFACTS[@]}"; do
    if [ ! -f "$artifact" ]; then
        echo "Error: artifact not found: $artifact"
        exit 1
    fi
done

# --- Helper: create a release, or upload --clobber if tag already exists ---
upload_release() {
    local tag="$1"
    local notes="$2"
    local is_prerelease="$3"   # "--prerelease" or ""

    # Try create first (without inline assets — we'll upload them one by one)
    if gh release create "$tag" \
        --repo "$GITHUB_REPO" \
        --title "$tag" \
        --notes "$notes" \
        ${is_prerelease:+"--prerelease"} \
        2>&1; then
        echo "[$tag] created successfully"
    else
        echo "[$tag] already exists — uploading assets (--clobber)"
    fi

    # Upload each artifact individually under a unique name so --clobber only
    # replaces within its own platform group (not across all 4 platforms).
    for artifact in "${ARTIFACTS[@]}"; do
        local src="$artifact"
        # Derive a unique, path-free name from the artifact path that encodes
        # the platform: e.g. "k8-manifest-validator-linux-amd64"
        local unique_name="k8-manifest-validator-$(echo "$artifact" | sed 's|dist/k8-manifest-validator_||' | sed 's|/k8-manifest-validator||' | tr '_' '-')"
        local tmpfile="/tmp/release_$$_$unique_name"
        cp "$src" "$tmpfile"
        echo "  Uploading $unique_name to $tag"
        if gh release upload "$tag" "$tmpfile" \
            --repo "$GITHUB_REPO" \
            --clobber \
            2>&1; then
            echo "  Uploaded $unique_name"
        else
            echo "  FAILED to upload $unique_name"
        fi
        rm -f "$tmpfile"
    done
}

# --- Stable release (latest) ---
echo "Creating stable release: $STABLE_TAG"
upload_release "$STABLE_TAG" \
    "Stable release for k8s $K8S_MINOR" \
    ""

# Promote stable to "Latest" explicitly
gh release edit "$STABLE_TAG" --repo "$GITHUB_REPO" --latest true 2>/dev/null || true

# --- Prerelease ---
echo "Creating prerelease: $PRERELEASE_TAG"
upload_release "$PRERELEASE_TAG" \
    "Prerelease $PRERELEASE_TAG for k8s $K8S_MINOR" \
    "--prerelease"

echo "Done. Released $PRERELEASE_TAG and updated $STABLE_TAG as latest."