# Delivery State — AWU-21.1

## Status: COMPLETE

## Branch
- **Branch:** `swe-k8-manifest-validator`
- **Remote:** `https://github.com/joshho/k8-manifest-validator.git`

## Context
- AWU-21.1 is the **first AWU in phase_3 wave** (cross-struct and complex validation)
- Phase_2 (AWUs 20.2–20.6) and Phase_2 consolidation (AWU-19.1) are all COMPLETE ✅
- AWU-21.1 owner file: `pkg/validator/phase3_initcontainers_volumes_test.go` (new file)

## Scope (AWU-21.1)
initContainers × volumes cross-struct interaction:
- `volumeMounts` in init containers (mounting volumes into initContainers)
- `configMap` volume refs in init containers
- `secret` volume refs in init containers
- `emptyDir` volume in init containers
- `hostPath` volume type in init containers
- Cross-check: init container volume mounts must reference existing volumes in pod spec

## File Ownership Check

| AWU | File | Status |
|-----|------|--------|
| AWU-19.1 | `phase3_deepinit_test.go` | COMPLETE ✅ (initContainers Required path coverage) |
| AWU-20.2 | `phase2_topologyspread_test.go` | COMPLETE ✅ (topologySpreadConstraints) |
| AWU-20.3 | `phase2_seccomp_test.go` | COMPLETE ✅ (seccompProfile.type) |
| AWU-20.4 | `phase2_lifecycleprobe_test.go` | COMPLETE ✅ (lifecycle/probe httpGet.port) |
| AWU-20.5 | `phase2_affinity_test.go` | COMPLETE ✅ (PodSpec affinity) |
| AWU-20.6 | `phase2_hostpid_test.go` | COMPLETE ✅ (hostPID/hostIPC/hostNetwork) |
| **AWU-21.1** | **`phase3_initcontainers_volumes_test.go`** | **NEW FILE — no conflict** ✅ |

## No Prior File Conflicts
- `phase3_deepinit_test.go` — owned by AWU-19.1, covers initContainers Required paths (NOT volume interactions)
- `phase2_*` files — owned by AWUs 20.2–20.6, Phase 2 semantic layer (different scope)
- `phase3_volume_*` files — volume type deep tests (initContainers × volumes is cross-struct, not volume-type-specific)
- AWU-21.1 is a **cross-struct** test combining initContainers + volumes, distinct from all prior AWUs

## Deliverable
- Owner file: `pkg/validator/phase3_initcontainers_volumes_test.go`
- Type: table-driven Go tests
- Test cases cover initContainers with various volume types (configMap, secret, emptyDir, hostPath)
- Expected status: COMPLETE ✅