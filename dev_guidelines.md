# Developer Guidelines — k8-manifest-validator (swe-k8-manifest-validator)

> Sourced from `main-swe-k8-manifest-validator/workspace/dev_guidelines.md` — Phase 3.1 deep-nested field coverage AWUs.

## 1. Code Standards

### 1.1 Go Conventions

- **Go version:** 1.22+ (matching k8s v1.31 dependency requirements)
- **Formatting:** `gofumpt` (stricter than `gofmt`) — run before every commit
- **Linting:** `golangci-lint` with default config plus:
  - `errcheck` (all errors checked)
  - `govet`
  - `staticcheck`
  - `ineffassign`
- **Imports:** Grouped in three blocks separated by blank lines:
  1. Standard library
  2. Third-party (k8s.io, sigs.k8s.io, github.com)
  3. Internal project imports (`k8-manifest-validator/...`)

### 1.2 Naming

| Thing | Convention | Example |
|---|---|---|
| Package names | Lowercase, single word, no underscores | `validator`, `output`, `types` |
| File names | snake_case matching package role | `builtin.go`, `crd.go`, `engine.go` |
| Exported types | PascalCase | `Engine`, `Result`, `Options` |
| Exported functions | PascalCase | `New()`, `Validate()` |
| Unexported | camelCase | `kindRouter`, `schemaRegistry` |
| Constants | PascalCase when exported | `StatusValid`, `StatusInvalid` |
| Error vars | `Err` prefix | `ErrUnknownKind`, `ErrVersionMismatch` |
| Interface names | `-er` suffix | `Validator`, `Formatter`, `ResultWriter` |

### 1.3 Error Handling

- **Always check errors from k8s library calls.** The `field.ErrorList` return from validation functions must be checked and propagated, never discarded.
- **Wrap errors with context.** Use `fmt.Errorf("loading CRD %s: %w", path, err)`.
- **Sentinel errors for known cases.** Define package-level `var Err*` for:
  - `ErrUnknownKind` — no schema registered for kind
  - `ErrVersionMismatch` — CR apiVersion doesn't match any CRD served version
  - `ErrCRDValidation` — CRD failed self-validation
  - `ErrFileNotFound` — input path doesn't exist
- **Panic is forbidden in library code.** Only `main.go` may call `log.Fatal` or `os.Exit`.

### 1.4 Testing

- **Table-driven tests are mandatory** for all validator functions. Every test file must use `[]struct{...}` test tables.
- **Test file naming:** `*_test.go` in the same package (white-box) or `*_external_test.go` for black-box.
- **Test fixtures:** Place YAML test data in `tests/fixtures/` organized by category.
- **Each test table entry must include:**
  - `name string` — descriptive test name
  - `deployFn func() *appsv1.Deployment` — deployment factory
  - `wantErr bool` — whether validation should produce an error
  - `errSubstr string` — substring to match in error message (empty if `wantErr: false`)

---

## 2. Phase 3 Deep-Nested Test Pattern (AWUs 17.1–17.5)

### 2.1 File Layout

Each deep-nested AWU produces one test file under `pkg/validator/`:

| AWU | File | Covers |
|---|---|---|
| 17.1 | `phase3_deepenv_test.go` | `env[].valueFrom.secretKeyRef` / `configMapKeyRef` |
| 17.2 | `phase3_deepenvfrom_test.go` | `envFrom[].configMapRef` / `secretRef` |
| 17.3 | `phase3_deepitems_test.go` | `configMap.items`, `secret.items`, `azureFile.secretName` |
| 17.4 | `phase3_deeplifecycle_test.go` | `lifecycle.postStart` / `preStop` handler sub-fields |
| **17.5** | **`phase3_deepprobe_test.go`** | **`liveness/readiness/startup probe sub-fields`** |

### 2.2 Test Structure (Reference: `phase3_deeplifecycle_test.go`)

```go
package validator

import (
    "testing"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
)

// with<SubType> modifier — attaches the sub-type to container[0]
func withLivenessProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
    return func(d *appsv1.Deployment) {
        d.Spec.Template.Spec.Containers[0].LivenessProbe = probe
    }
}

// TestPhase3_DeepNested_<SubType> covers the deep nested chain:
// Container.<SubType> → <Handler> → <Action> → sub-fields
func TestPhase3_DeepNested_<SubType>(t *testing.T) {
    t.Parallel()
    cases := []struct {
        name      string
        deployFn  func() *appsv1.Deployment
        wantErr   bool
        errSubstr string
    }{
        // cases...
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            result := validateDeploymentRaw(tc.deployFn())
            gotErr := len(result.Errors) > 0
            if gotErr != tc.wantErr {
                t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
            }
            if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
                t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
            }
        })
    }
}
```

### 2.3 Key Codegen Facts (Reference)

**Lifecycle handler sub-fields (`deferredFieldIndex` confirmed):**
- `.lifecycle.postStart.exec.command` — `Type:"array"`, `Required:false`
- `.lifecycle.preStop.httpGet.path` — `Type:"string"`, `Required:false`
- `.lifecycle.preStop.httpGet.port` — `Type:""`, `Required:true`
- `.lifecycle.preStop.tcpSocket.port` — `Type:""`, `Required:true`

**Probe sub-fields (`deferredFieldIndex` confirmed):**

Liveness, Readiness, Startup probes share the same sub-field structure:

| Path | Type | Required |
|---|---|---|
| `.{probe}.tcpSocket.port` | `""` (IntOrString) | **true** |
| `.{probe}.httpGet.port` | `""` (IntOrString) | **true** |
| `.{probe}.httpGet.host` | `string` | false |
| `.{probe}.httpGet.path` | `string` | false |
| `.{probe}.httpGet.scheme` | `string` | false |
| `.{probe}.httpGet.httpHeaders` | `array` | false |
| `.{probe}.exec.command` | `array` | false |
| `.{probe}.failureThreshold` | `integer` | false |
| `.{probe}.periodSeconds` | `integer` | false |
| `.{probe}.timeoutSeconds` | `integer` | false |
| `.{probe}.initialDelaySeconds` | `integer` | false |
| `.{probe}.successThreshold` | `integer` | false |
| `.{probe}.terminationGracePeriodSeconds` | `integer` | false |

### 2.4 Modifier Pattern for Probes

```go
// AWU-17.5 modifiers (to be implemented in phase3_deepprobe_test.go)

// withLivenessProbe sets the liveness probe on container[0]
func withLivenessProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
    return func(d *appsv1.Deployment) {
        d.Spec.Template.Spec.Containers[0].LivenessProbe = probe
    }
}

// withReadinessProbe sets the readiness probe on container[0]
func withReadinessProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
    return func(d *appsv1.Deployment) {
        d.Spec.Template.Spec.Containers[0].ReadinessProbe = probe
    }
}

// withStartupProbe sets the startup probe on container[0]
func withStartupProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
    return func(d *appsv1.Deployment) {
        d.Spec.Template.Spec.Containers[0].StartupProbe = probe
    }
}
```

---

## 3. AWU-17.5 Implementation Notes

### 3.1 Test File
`pkg/validator/phase3_deepprobe_test.go`

### 3.2 Required Test Cases (≥5)

Per PRD Review R2, the following 5+ cases must be covered:

1. **Valid:** `livenessProbe.tcpSocket.port` set (Required:true → no error)
2. **Invalid:** `livenessProbe.tcpSocket.port` empty/zero (Required:true → error "port")
3. **Valid:** `readinessProbe.httpGet` with host + path + port + scheme set
4. **Valid:** `startupProbe.exec.command` with array of strings
5. **Invalid:** `livenessProbe.httpGet.port` empty/zero (Required:true → error "port")
6. **Valid:** probe with `failureThreshold`, `periodSeconds`, `timeoutSeconds` set (all Optional → no error)

### 3.3 Pattern

Follow `phase3_deeplifecycle_test.go` exactly:
- One `*_test.go` file per AWU
- Table-driven test with `name`, `deployFn`, `wantErr`, `errSubstr`
- `withProbe()` modifiers for each probe type
- Call `validateDeploymentRaw()` and assert `result.Errors`
- Parallel test via `t.Parallel()`
- Test name prefix: `[Phase3-DeepProbe]`

### 3.4 Probe Construction

Use `corev1.Probe` with `ProbeHandler` discriminated union:

```go
// TCPSocket probe
probe := corev1.Probe{
    ProbeHandler: corev1.ProbeHandler{
        TCPSocket: &corev1.TCPSocketAction{
            Port: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
        },
    },
}

// HTTPGet probe
probe := corev1.Probe{
    ProbeHandler: corev1.ProbeHandler{
        HTTPGet: &corev1.HTTPGetAction{
            Host:   "localhost",
            Path:   "/health",
            Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
            Scheme: "HTTP",
        },
    },
}

// Exec probe
probe := corev1.Probe{
    ProbeHandler: corev1.ProbeHandler{
        Exec: &corev1.ExecAction{
            Command: []string{"cat", "/tmp/healthy"},
        },
    },
}
```

### 3.5 Validation Command

```bash
go test ./pkg/validator/... -run Phase3Deep
```

All 5 AWU test files (17.1–17.5) should pass together.

---

## 4. Architecture Notes

### 4.1 `normalizePath` Collision is Intentional

`normalizePath` strips all array indices, producing bare field names:
```
".spec.containers[0].livenessProbe.tcpSocket.port"
→ ".port"
```

The flat `deferredFieldIndex` uses shared keys across all owning types. This is by design — the union index lets any path ending in `.port` resolve based on struct context. Tests must use the correct Go type to ensure reflection traversal reaches the right struct.

### 4.2 Codegen is the Source of Truth

Never infer `Required` status from intuition. Always check `generated_structural.go`:
```bash
grep "\.livenessProbe\|\.readinessProbe\|\.startupProbe" pkg/validator/generated_structural.go
```

---

## 5. Commit Convention

```
<type>(<scope>): <short description>
```

**Types:** `feat`, `fix`, `test`, `docs`, `refactor`, `chore`
**Scopes:** `validator`, `builtin`, `test`, `phase3`

Example:
```
test(phase3): add AWU-17.5 probe sub-fields deep traversal tests
```

---

## 6. General Rules

- Do NOT modify `generated_structural.go` or `tools/codegen/main.go`
- Do NOT add new k8s types — all needed types already in `io.k8s.api.core.v1`
- Do NOT modify `walkStruct` / `visitFields` behavior
- Run `gofumpt` before commit
- Run `go test ./pkg/validator/... -run Phase3Deep` and confirm all pass
