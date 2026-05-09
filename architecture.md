# Architecture: Builtin Validator Exhaustiveness (Phase 1)

## Overview

Add a shared `validatePodSpec()` helper to `pkg/validator/builtin.go` that validates all Phase 1 fields of a Kubernetes PodSpec, then wire it into 8 existing per-type validators.

## The Shared Helper

### Signature

```go
func validatePodSpec(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList
```

### Design Principles

1. **No new dependencies.** Uses only packages already in `go.mod`: `k8s.io/api/core/v1`, `k8s.io/apimachinery/pkg/util/validation`, `k8s.io/apimachinery/pkg/util/validation/field`, `k8s.io/apimachinery/pkg/api/resource`.
2. **Path prefix is caller-provided.** Each caller passes its own `*field.Path` so error paths correctly reflect where the PodSpec lives (e.g., `spec.template.spec` vs `spec`).
3. **Field ordering matches the PRD field list.** Validation methods are grouped by field for readability and testability.
4. **No panics on nil pointers.** Every pointer, slice, and map access is nil-safe.
5. **Probes skipped for batch workloads.** The helper validates probes unconditionally (per PRD: "Job/CronJob: probes may be skipped"). Callers for batch types skip probe validation via not passing the probe path, or the helper itself can be called with a flag. **Decision:** Use a parameter:

```go
func validatePodSpec(podSpec *corev1.PodSpec, path *field.Path, skipProbes bool) field.ErrorList
```

The `skipProbes` flag is `false` for Deployment/StatefulSet/DaemonSet/ReplicaSet/Pod/ReplicationController and `true` for Job/CronJob.

### Internal Helper Sub-functions (unexported)

| Helper | Purpose |
|--------|---------|
| `validateContainers(containers []corev1.Container, initContainer bool, volumeNames set, path *field.Path) field.ErrorList` | Validates container-level fields: name, image, ports, env, resources, volumeMounts, probes |
| `validateContainerPort(port corev1.ContainerPort, path *field.Path) field.ErrorList` | containerPort range [1,65535], protocol enum, hostPort range [0,65535] |
| `validateEnvVar(env corev1.EnvVar, path *field.Path) field.ErrorList` | env.name is valid C identifier |
| `validateResourceList(rl corev1.ResourceList, path *field.Path) field.ErrorList` | parsable resource quantities |
| `validateVolumeMount(mount corev1.VolumeMount, volumeNames set, path *field.Path) field.ErrorList` | name matches volume, mountPath is absolute |
| `validateVolumes(volumes []corev1.Volume, path *field.Path) (set, field.ErrorList)` | DNS-1123 label name, unique; returns set of volume names for mount validation |
| `validateProbe(probe *corev1.Probe, path *field.Path) field.ErrorList` | one handler present (HTTPGet, TCPSocket, GRPCAction, or Exec) |

### Phase 1 Validation Mapping

| PRD Field | Implementation | Validation Rule |
|-----------|---------------|-----------------|
| `containers[*].name` | `utilvalidation.IsDNS1123Label()` | DNS-1123 label, required, unique across containers + initContainers |
| `containers[*].image` | String length check | Non-empty string |
| `containers[*].ports[*].containerPort` | Range check | `1 <= port <= 65535` |
| `containers[*].ports[*].protocol` | Enum check | Must be TCP, UDP, or SCTP (case-insensitive) |
| `containers[*].ports[*].hostPort` | Range check | `0 <= port <= 65535` |
| `containers[*].env[*].name` | `IsCIdentifier()` from `utilvalidation` | Valid C identifier, required |
| `containers[*].resources.limits` | `resource.ParseQuantity()` | Parseable quantities; on parse error return Invalid |
| `containers[*].resources.requests` | `resource.ParseQuantity()` | Same |
| `containers[*].volumeMounts[*].name` | Lookup in volume name set | Must match a declared volume name |
| `containers[*].volumeMounts[*].mountPath` | `strings.HasPrefix("/")` | Absolute path |
| `volumes[*].name` | `utilvalidation.IsDNS1123Label()` | DNS-1123 label, required, unique |
| `restartPolicy` | Enum check | Always/OnFailure/Never (Pod only) |
| `serviceAccountName` | `utilvalidation.IsDNS1123Subdomain()` | DNS-1123 subdomain format |
| `nodeName` | `utilvalidation.IsDNS1123Subdomain()` | DNS-1123 subdomain format |
| `containers[*].livenessProbe` | Handler presence check | One of HTTPGet/TCPSocket/Exec/GRPCAction non-nil |
| `containers[*].readinessProbe` | Same | Same |
| `containers[*].startupProbe` | Same | Same |
| `volumes[*].configMap.name` | Non-empty string | Required when configMap is present |
| `volumes[*].secret.secretName` | Non-empty string | Required when secret is present |
| `initContainers` | Same as containers | Same field set via `validateContainers` |

## Wire-In Strategy

### Type → PodSpec Access Path

| Type | Access Expression | Error Path Prefix | Probes? |
|------|------------------|-------------------|---------|
| Deployment | `deploy.Spec.Template.Spec` | `spec.template.spec` | Yes |
| StatefulSet | `ss.Spec.Template.Spec` | `spec.template.spec` | Yes |
| DaemonSet | `ds.Spec.Template.Spec` | `spec.template.spec` | Yes |
| ReplicaSet | `rs.Spec.Template.Spec` | `spec.template.spec` | Yes |
| Pod | `pod.Spec` | `spec` | Yes |
| ReplicationController | `rc.Spec.Template.Spec` | `spec.template.spec` | Yes |
| Job | `job.Spec.Template.Spec` | `spec.template.spec` | No |
| CronJob | `cj.Spec.JobTemplate.Spec.Template.Spec` | `spec.jobTemplate.spec.template.spec` | No |

### Validation Order per Type (after AWU completion)

All 8 types follow this pattern:

```go
func (v *BuiltinValidator) validateDeployment(deploy *appsv1.Deployment) field.ErrorList {
    var allErrs field.ErrorList
    
    // 1. ObjectMeta validation (existing, unchanged)
    allErrs = append(allErrs, validation.ValidateObjectMeta(...)...)
    
    // 2. Type-specific spec fields (existing, unchanged)
    //    - replicas, selector, etc.
    
    // 3. PodSpec validation (new)
    allErrs = append(allErrs, validatePodSpec(&deploy.Spec.Template.Spec,
        field.NewPath("spec", "template", "spec"),
        false, // skipProbes = false for workloads
    )...)
    
    return allErrs
}
```

### Template Metadata Validation

The existing code validates `deploy.Spec.Template.ObjectMeta` for Deployment. The PRD does not mandate changes here. **Keep existing template metadata validation as-is.** The new `validatePodSpec` does not touch metadata.

## Deferred Fields (Phase 2)

These fields are explicitly **out of scope** for Phase 1 and will be added in Phase 2:

| Deferred Field | Reason |
|---------------|--------|
| `securityContext` (pod-level & container-level) | SELinux/Seccomp/Capabilities require complex enum validation |
| `affinity` | Recursive node/ pod affinity selectors, nested expressions |
| `tolerations` | Multi-field matching logic (key, operator, value, effect, tolerationSeconds) |
| `lifecycle` hooks | PreStop/PostStart exec and HTTP handler validation |
| `topologySpreadConstraints` | Label selector + topologyKey + whenUnsatisfiable enum |
| `dnsConfig` / `hostname` / `subdomain` | Less commonly used, lower priority |
| `readinessGates` | ConditionType validation |
| `containers[*].env[*].valueFrom` | FieldRef/ResourceFieldRef/ConfigMapKeyRef/SecretKeyRef resolution |
| `containers[*].ports[*].hostIP` | IP address validation |

## File Structure After Phase 1

Only `pkg/validator/builtin.go` is modified. No new files, no new packages, no new dependencies.

### Estimated Size

- Current `builtin.go`: ~730 lines
- After Phase 1: ~1200–1400 lines
  - `validatePodSpec`: ~50 lines
  - `validateContainers`: ~80 lines
  - `validateContainerPort`: ~40 lines
  - `validateEnvVar`: ~15 lines
  - `validateResourceList`: ~30 lines
  - `validateVolumeMount`: ~25 lines
  - `validateVolumes`: ~40 lines
  - `validateProbe`: ~35 lines
  - Wire-in calls (8 × 2 lines): ~16 lines
  - Comments/whitespace: ~100 lines

No cyclomatic complexity target is enforced, but each sub-helper should remain under 40 lines of pure validation logic (excluding whitespace/comments).

## Container Name Uniqueness

Container names must be unique across both `containers` and `initContainers` combined. The implementation:

1. Validates all `containers` entries, collecting names in a set
2. Validates all `initContainers` entries, checking against the existing set

## Volume Mount Validation

Volume mount names are validated against the set of declared volume names. This requires validating `spec.volumes` before `spec.containers[*].volumeMounts`. The implementation processes volumes first to build the name set, then uses it during container validation.

## Test Strategy

See `dev_guidelines.md` for detailed testing conventions.

Per AWU:
- Each AWU adds a TDD failing test in `pkg/validator/builtin_test.go`
- Tests use YAML-based resources validated via `BuiltinValidator.ValidateResource()`
- No changes to `tests/fixtures_test.go` (integration tests) are needed in Phase 1
