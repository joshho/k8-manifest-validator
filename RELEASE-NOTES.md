# Release Notes — CI/CD Pipeline Overhaul

## Summary

Consolidated the old two-workflow CI setup (`test.yml` + `release.yml`) into a single `ci.yml` with three sequential jobs: **build → test → release**.

## What Changed

### Workflow

| Before | After |
|---|---|
| `test.yml` + `release.yml` (race condition) | Single `ci.yml` with `build → test → release` |
| `goreleaser release --clean` (created GitHub release) | `goreleaser build --snapshot` (local artifacts only) |
| test job pushed tags; release job created releases | release job owns everything: release + stable + tag push |

### Release Naming Convention
- **Stable**: `v{k8s-minor}` e.g. `v1.32`
- **Prerelease**: `v{k8s-minor}-{N}` e.g. `v1.32-0`, `v1.32-1`, ...

### Asset Names (goreleaser v2)
```
k8-manifest-validator-linux-amd64-v1
k8-manifest-validator-linux-arm64-v8.0
k8-manifest-validator-darwin-amd64-v1
k8-manifest-validator-darwin-arm64-v8.0
```

## Bugs Fixed

- **`local` keyword outside function** — removed; plain variable assignment in `bash -e` mode
- **PID in temp file paths** (`/tmp/release_$$_file`) — GitHub was using PID-prefixed filename as asset name (`release_2437_k8-manifest-validator-linux-amd64-v1`)
- **HTTP 404 on multi-file `gh release create`** — create release bare, then `gh release upload` each asset individually
- **HTTP 422 `Release.tag_name already exists`** — `set +e` around create, check error, proceed to upload if already exists
- **GHA job output sharing** — `echo "name=value" >> $GITHUB_OUTPUT` + `${{ steps.*.outputs.name }}`
- **`gh release create --latest` failing on recreate** — delete before recreate
- **`git tag` "already exists"** — `git tag -f` + `git push -f`
- **`git tag` "empty ident name"** — set `user.email` and `user.name` before tagging
- **Goreleaser v2 path suffixes** — `_v1` (amd64) / `_v8.0` (arm64) in dist/ directory names

## Triggering CI

To trigger CI without visible README changes, use an HTML comment:
```markdown
<!-- trigger ci -->
```

## Branches
All 5 branches (main + release/v1.32 through v1.35) are in sync with these fixes.