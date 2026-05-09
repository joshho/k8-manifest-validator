# Traceability: PRD Phrases → Acceptance Tests → AWU Mapping

## PRD Phrase Index

Each numbered PRD requirement links to the AWU that implements it and the test that validates it.

### Phase 1 Field Validation

| # | PRD Phrase | AWU | Acceptance Test | Test File |
|---|-----------|-----|-----------------|-----------|
| 1 | `spec.containers[*].name` — DNS-1123 label, required, unique | AWU-1 | `TestBuiltinValidatorDetectsInvalidContainerNameInDeployment` | `builtin_test.go` |
| 2 | `spec.containers[*].image` — non-empty string | AWU-1 | `TestBuiltinValidatorDetectsEmptyContainerImageInDeployment` | `builtin_test.go` |
| 3 | `spec.containers[*].ports[*].containerPort` — range [1, 65535] | AWU-1 | `TestBuiltinValidatorDetectsOutOfRangeContainerPort` | `builtin_test.go` |
| 4 | `spec.containers[*].ports[*].protocol` — TCP/UDP/SCTP | AWU-1 | `TestBuiltinValidatorDetectsInvalidProtocol` | `builtin_test.go` |
| 5 | `spec.containers[*].ports[*].hostPort` — range [0, 65535] | AWU-1 | `TestBuiltinValidatorDetectsOutOfRangeHostPort` | `builtin_test.go` |
| 6 | `spec.containers[*].env[*].name` — valid C identifier, required | AWU-1 | `TestBuiltinValidatorDetectsInvalidEnvName` | `builtin_test.go` |
| 7 | `spec.containers[*].resources.limits` — parseable resource quantities | AWU-1 | `TestBuiltinValidatorDetectsUnparseableResourceLimits` | `builtin_test.go` |
| 8 | `spec.containers[*].resources.requests` — parseable resource quantities | AWU-1 | `TestBuiltinValidatorDetectsUnparseableResourceRequests` | `builtin_test.go` |
| 9 | `spec.containers[*].volumeMounts[*].name` — must match volume name | AWU-1 | `TestBuiltinValidatorDetectsVolumeMountNameMismatch` | `builtin_test.go` |
| 10 | `spec.containers[*].volumeMounts[*].mountPath` — absolute path (starts with /) | AWU-1 | `TestBuiltinValidatorDetectsNonAbsoluteMountPath` | `builtin_test.go` |
| 11 | `spec.volumes[*].name` — DNS-1123 label, required, unique | AWU-1 | `TestBuiltinValidatorDetectsInvalidVolumeName` | `builtin_test.go` |
| 12 | `spec.restartPolicy` — Always/OnFailure/Never (Pod only) | AWU-2 | `TestBuiltinValidatorDetectsInvalidRestartPolicy` | `builtin_test.go` |
| 13 | `spec.serviceAccountName` — DNS-1123 subdomain format | AWU-2 | `TestBuiltinValidatorDetectsInvalidServiceAccountNameInPod` | `builtin_test.go` |
| 14 | `spec.nodeName` — DNS-1123 subdomain format | AWU-2 | `TestBuiltinValidatorDetectsInvalidNodeName` | `builtin_test.go` |
| 15 | `spec.containers[*].livenessProbe/readinessProbe/startupProbe` — one handler present | AWU-1 | `TestBuiltinValidatorDetectsMissingProbeHandler` | `builtin_test.go` |
| 16 | `spec.volumes[*].configMap.name / secret.secretName` — non-empty | AWU-1 | `TestBuiltinValidatorDetectsEmptyConfigMapNameInVolume` | `builtin_test.go` |
| 17 | `spec.initContainers` — same fields as containers | AWU-1 | `TestBuiltinValidatorDetectsInvalidInitContainerName` | `builtin_test.go` |

### Type Wiring

| # | PRD Phrase | AWU | Acceptance Test | Test File |
|---|-----------|-----|-----------------|-----------|
| 18 | Shared helper: `validatePodSpec(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList` | AWU-1 | Function present, called from all 8 validators | `builtin_test.go` |
| 19 | Wire into `validateDeployment` | AWU-1 | `TestBuiltinValidatorDetectsInvalidContainerNameInDeployment` | `builtin_test.go` |
| 20 | Wire into `validateStatefulSet` | AWU-1 | `TestBuiltinValidatorDetectsInvalidContainerNameInStatefulSet` | `builtin_test.go` |
| 21 | Wire into `validateDaemonSet` | AWU-1 | `TestBuiltinValidatorDetectsInvalidContainerNameInDaemonSet` | `builtin_test.go` |
| 22 | Wire into `validateReplicaSet` | AWU-2 | `TestBuiltinValidatorDetectsInvalidContainerNameInReplicaSet` | `builtin_test.go` |
| 23 | Wire into `validatePod` | AWU-2 | `TestBuiltinValidatorDetectsInvalidServiceAccountNameInPod` | `builtin_test.go` |
| 24 | Wire into `validateReplicationController` | AWU-4 | `TestBuiltinValidatorDetectsMissingVolumeNameInReplicationController` | `builtin_test.go` |
| 25 | Wire into `validateJob` | AWU-3 | `TestBuiltinValidatorDetectsInvalidContainerImageInJob` | `builtin_test.go` |
| 26 | Wire into `validateCronJob` | AWU-3 | `TestBuiltinValidatorDetectsInvalidContainerImageInCronJob` | `builtin_test.go` |

### Scope Boundaries

| # | PRD Phrase | AWU | Verifier |
|---|-----------|-----|----------|
| 27 | In-scope types: Deployment, StatefulSet, DaemonSet, ReplicaSet, ReplicationController, Pod, Job, CronJob | All | `plan.json` lists all 8 as wired |
| 28 | Out-of-scope: Service, Ingress, ConfigMap, Secret, PVC, Namespace, etc. (no PodSpec) | — | Zero changes to their validators; tests for them unchanged |
| 29 | Job/CronJob: probes may be skipped for batch workloads | AWU-3 | `skipProbes=true` passed for batch types |
| 30 | Phase 2 (deferred): lifecycle hooks, securityContext, affinity, tolerations, etc. | — | Documented in `architecture.md` deferred fields table |

### Other Constraints

| # | PRD Phrase | AWU | Verifier |
|---|-----------|-----|----------|
| 31 | No new dependencies | All | `go.mod` unchanged; `go mod tidy` produces no diffs |
| 32 | Hand-written per-field validators | All | Code review confirms no imported validation functions from `k8s.io/kubernetes` |
| 33 | Wire into: validateDeployment etc. | All | Each AWU acceptance criteria confirms the wire-in call exists |

## AWU → Test Mapping Summary

| AWU | Primary Tests |
|-----|---------------|
| AWU-1 | All field validation tests (items 1–11, 15–21) |
| AWU-2 | ServiceAccountName, nodeName, restartPolicy tests (items 12–14, 22–23) |
| AWU-3 | Job/CronJob wire-in and skipProbes tests (items 25–26, 29) |
| AWU-4 | ReplicationController wire-in test (item 24) |
| All | Valid manifest regression tests for all 8 types |

## Regression Tests (All AWUs)

These existing tests must remain passing:

| Test | Types Covered | Owner |
|------|---------------|-------|
| `TestBuiltInKindRoutingDeploymentReturnsValidator` | Deployment | AWU-1 |
| `TestBuiltinValidatorValidatesDeployment` | Deployment | AWU-1 |
| `TestBuiltinValidatorDetectsInvalidDeployment` | Deployment | AWU-1 |
| `TestBuiltinValidatorValidatesConfigMap` | ConfigMap (no PodSpec) | — |
| `TestBuiltinValidatorValidatesService` | Service (no PodSpec) | — |
| `TestBuiltinValidatorValidatesIngress` | Ingress (no PodSpec) | — |
| `TestBuiltinValidatorValidatesJob` | Job | AWU-3 |
| `TestPerTypeRouting` (all builtins) | All | All |
| `TestValidBuiltinsAllPass` (fixtures) | All | All |
| `TestInvalidBuiltinsAllFail` (fixtures) | All | All |

## Change Impact

### Files Modified

```
pkg/validator/builtin.go      — +validatePodSpec + wire-in calls (only file changed)
pkg/validator/builtin_test.go — +TDD test cases (tests only, no production logic)
```

### Files NOT Modified

```
pkg/validator/decoder.go       — unchanged
pkg/validator/decoder_test.go  — unchanged
pkg/validator/engine.go        — unchanged
pkg/validator/engine_test.go   — unchanged
pkg/validator/validator.go     — unchanged
pkg/validator/scanner.go       — unchanged
pkg/validator/crd.go           — unchanged
pkg/validator/cr.go            — unchanged
pkg/types/result.go            — unchanged
cmd/k8-manifest-validator/     — unchanged
tests/                         — unchanged (integration tests continue to pass)
```

### Fixtures Impact

If any existing fixture YAML contains invalid PodSpec combinations (e.g., a Deployment that previously passed because PodSpec was unvalidated, but now fails), fixtures must be fixed. This is expected and represents improved validation coverage.
