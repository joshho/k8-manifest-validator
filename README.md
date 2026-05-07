# k8-manifest-validator

Validate Kubernetes manifests against the k8s API server schema.

## Problem

Writing Kubernetes YAML by hand is error-prone. Missing fields, wrong types, invalid
enum values — these only surface at `kubectl apply` time, not at authoring time.

## Solution

k8-manifest-validator statically validates Kubernetes manifests against the official
k8s API schema for your target k8s minor version. Catch schema violations before
they reach the cluster.

## Usage

```bash
# Validate a manifest
./k8-manifest-validator path/to/deployment.yaml

# Validate all manifests in a directory
./k8-manifest-validator ./manifests/

# Show version
./k8-manifest-validator --version
```

## Build

```bash
./scripts/build.sh
```

## Test

```bash
./scripts/test.sh
```

## Releases

Releases are built automatically on every push to `main`. You get:
- `v{k8s-minor}` — stable release for a given k8s minor (e.g. v1.36)
- `v{k8s-minor}-{N}` — prerelease builds, one per successful CI run
<!-- last updated: 2026-05-05 -->
<!-- test: 2026-05-07T01:49:51Z -->
-e 
---
Multi-platform CI: 2026-05-07
-e 
---
Fix goreleaser paths: 2026-05-07T02:35:15Z
-e 
---
Binary path fix CI trigger: 2026-05-07T02:40:54Z
-e 
---
Debug dist listing: 2026-05-07T02:54:08Z
-e 
---
Upload tarball fix: 2026-05-07T03:02:29Z
-e 
---
Debug ls: 2026-05-07T03:11:55Z
-e 
---
Separate create/upload: 2026-05-07T03:23:06Z
ci: trigger --name fix test 033507
ci: trigger temp file rename 034110
ci: trigger fixed upload 034911
ci: trigger syntax fix 035549
ci: trigger GITHUB_TOKEN fix 040439
