# Dev Guidelines: Builtin Validator Exhaustiveness (Phase 1)

## Validation Field Conventions

### Error Path Style

All error paths use `field.NewPath()` with human-readable child segments, matching Kubernetes upstream conventions:

| Field | Error Path |
|-------|-----------|
| Container name | `spec.containers[0].name` |
| Container image | `spec.containers[0].image` |
| Container port | `spec.containers[0].ports[0].containerPort` |
| Protocol | `spec.containers[0].ports[0].protocol` |
| Env var name | `spec.containers[0].env[0].name` |
| Resource limit | `spec.containers[0].resources.limits` |
| Resource request | `spec.containers[0].resources.requests` |
| Volume mount name | `spec.containers[0].volumeMounts[0].name` |
| Volume mount path | `spec.containers[0].volumeMounts[0].mountPath` |
| Volume name | `spec.volumes[0].name` |
| ConfigMap name | `spec.volumes[0].configMap.name` |
| Secret name | `spec.volumes[0].secret.secretName` |
| Restart policy | `spec.restartPolicy` |
| ServiceAccount name | `spec.serviceAccountName` |
| Node name | `spec.nodeName` |
| Probe (liveness) | `spec.containers[0].livenessProbe` |

**Important:** The caller's `path` prefix is prepended. If the PodSpec lives at `spec.template.spec`, errors become `spec.template.spec.containers[0].name` — this is correct and matches what kubeconform/kubectl would produce.

### Error Message Style

- Use the same messages as k8s upstream where applicable:
  - Container name: `"a lowercase RFC 1123 label must consist of..."` (from `IsDNS1123Label`)
  - Port range: `"must be between 1 and 65535"`
  - Protocol: `"must be one of TCP, UDP, SCTP"`
  - Required fields: use empty detail string `""` (k8s convention: `field.Required()` sets a default message)
  - Resource quantities: `"unable to parse resource quantity"`
  - Absolute path: `"must be an absolute path (start with /)"`
- Do NOT include the field path in the error message (it's in the `Field` attribute).

### Error Code to Use

| Situation | Code |
|-----------|------|
| Field value out of range | `ErrCodeInvalidValue` |
| Required field empty | `ErrCodeRequired` |
| Invalid format (DNS label, C identifier) | `ErrCodeInvalidFormat` |
| Duplicate name | `ErrCodeDuplicate` |
| General invalid | `ErrCodeInvalid` |

Use `field.Invalid()` for range/format issues, `field.Required()` for missing required fields, `field.Duplicate()` for duplicates.

```go
// Good
field.Invalid(path.Child("containerPort"), port, "must be between 1 and 65535")
field.Required(path.Child("name"), "")
field.Duplicate(path.Child("name"), duplicateName)

// Bad — don't repeat the field path in the message
field.Invalid(path.Child("containerPort"), port, "spec.containers[0].ports[0].containerPort must be between 1 and 65535")
```

### Nil Safety

Every pointer dereference must be guarded:

```go
if container.LivenessProbe != nil {
    // validate liveness probe
}
```

Slices and maps are safe to range over when nil in Go, but the validation function should still be fine with nil/empty input (produce no errors for empty slices).

## Function Organization

### Sub-helper naming

Use `validate<Resource>()`:

```go
func validateContainers(...)
func validateContainerPort(...)
func validateEnvVar(...)
func validateResourceList(...)
func validateVolumeMount(...)
func validateVolumes(...)
func validateProbe(...)
```

### Function placement in builtin.go

Add all new functions **before** the per-type validator methods (before `validateDeployment`). Place `validatePodSpec` first, then its sub-helpers in alphabetical order. This makes the file read top-down from general to specific.

```
// existing imports + type definitions
// existing NewBuiltinValidator + registerKind methods
// --- new code starts here ---
// validatePodSpec
// validateContainerPort
// validateContainers
// validateEnvVar
// validateProbe
// validateResourceList
// validateVolumeMount
// validateVolumes
// --- existing code continues here ---
// validateDeployment (updated)
// validateStatefulSet (updated)
// ...
```

### Comment conventions

```go
// validateContainers validates the spec.containers or spec.initContainers field.
// When initContainer is true, errors reference spec.initContainers.
// volumeNames is the set of declared volume names for mount validation.
// skipProbes, when true, skips liveness/readiness/startup probe validation.
func validateContainers(containers []corev1.Container, initContainer bool, volumeNames map[string]struct{}, path *field.Path, skipProbes bool) field.ErrorList {
```

No doc comments needed for trivial one-liner helpers. Each exported-adjacent (unexported validation) function should have a single-line comment describing its purpose.

## Wire-In Pattern

When calling `validatePodSpec` from a type validator, follow this exact pattern:

```go
// After existing spec field validations:
allErrs = append(allErrs, validatePodSpec(
    &deploy.Spec.Template.Spec,
    field.NewPath("spec", "template", "spec"),
    false, // skipProbes
)...)
```

**Do not** bundle the PodSpec validation into the existing `var allErrs field.ErrorList` line — add it after the existing `allErrs = append(allErrs, validation.ValidateObjectMeta(...)...)` and type-specific field checks.

**Do not** remove or modify any existing validation logic unless the PRD explicitly says so.

## Test Expectations

### Unit Tests (in `pkg/validator/builtin_test.go`)

Each AWU adds one or more failing TDD test cases that pass after implementation. Tests follow the existing pattern:

```go
func TestBuiltinValidatorDetectsInvalidXxx(t *testing.T) {
    invalidYAML := []byte(`...`)
    bv := NewBuiltinValidator()
    result := bv.ValidateResource(invalidYAML)
    if result.Status != "invalid" {
        t.Errorf("expected invalid, got %s: %v", result.Status, result.Errors)
    }
    // Optionally check specific error field
    if len(result.Errors) > 0 && result.Errors[0].Field != "spec.containers[0].name" {
        t.Errorf("expected error on spec.containers[0].name, got %s", result.Errors[0].Field)
    }
}
```

### Test data conventions

- Use inline YAML strings in test functions (not fixture files) for unit tests
- Each test validates exactly one failure mode (single responsibility)
- Include valid-manifest tests to ensure existing behavior is preserved
- Test edge cases: nil pointers, empty slices, boundary values (port 1, port 65535, hostPort 0)

### Valid manifest test pattern

Every AWU must also add or verify that a valid manifest of the wired type passes:

```go
// This test already exists for Deployment. Add similar for each wired type
// if not already present. The existing TestBuiltInKindRoutingDeploymentReturnsValidator
// and TestBuiltinValidatorValidatesDeployment patterns should be replicated
// for StatefulSet, DaemonSet, etc. if missing.
```

### What NOT to test

- Do not test sub-helpers in isolation (they're unexported; test through the public per-type validators)
- Do not test `validatePodSpec` directly — test through the 8 type validators
- Do not add integration tests in `tests/` — unit tests in `pkg/validator/builtin_test.go` only
- Do not add CLI/output tests — these are unchanged

### Existing tests that must remain green

All tests in:
- `pkg/validator/builtin_test.go`
- `pkg/validator/engine_test.go`
- `tests/fixtures_test.go`

Must continue to pass unchanged. New validations may surface errors in previously-valid manifests (e.g., a running Deployment that had invalid container ports was previously not validated). This is expected and correct behavior — update test YAML fixtures if needed.
