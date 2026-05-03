package validator_tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"k8-manifest-validator/pkg/types"
	"k8-manifest-validator/pkg/validator"
)

func findBinary(t *testing.T) string {
	t.Helper()
	// First check PATH
	if path, err := exec.LookPath("k8-manifest-validator"); err == nil {
		return path
	}
	// Build from source. Tests run from tests/ dir, so repo root is parent.
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "k8-manifest-validator")
	cmd := exec.Command("go", "build", "-o", binPath, "../cmd/k8-manifest-validator/")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("failed to build k8-manifest-validator: %v\n%s", err, string(out))
	}
	return binPath
}

const fixturesDir = "fixtures"  // relative to tests/ directory (CWD for go test ./tests)

func TestRequiredFixtureCategoriesExist(t *testing.T) {
	categories := map[string][]string{
		"valid":         {"deployment.yaml", "service.yaml", "configmap.yaml", "secret.yaml", "ingress.yaml", "statefulset.yaml", "daemonset.yaml", "job.yaml", "cronjob.yaml", "pvc.yaml"},
		"invalid":       {"deployment_wrong_kind.yaml", "service_missing_selector.yaml", "configmap_missing_name.yaml", "secret_missing_name.yaml", "ingress_missing_rules.yaml", "job_out_of_range.yaml", "cronjob_invalid_schedule.yaml", "pvc_missing_access_modes.yaml"},
		"cr":            {"valid-database.yaml", "invalid-database-missing-version.yaml", "invalid-database-wrong-type.yaml"},
		"crd":           {"database-crd.yaml"},
		"multi-version": {"cache-crd.yaml", "cache-v1.yaml", "cache-v1beta1.yaml", "cache-v2-mismatch.yaml"},
		"malformed":     {"deployment_bad_indent.yaml"},
		"mixed":         {"mixed-resources.yaml"},
		"empty":         {},
		"unknown":       {"unknown-widget.yaml"},
	}

	for category, expectedFiles := range categories {
		categoryPath := filepath.Join(fixturesDir, category)
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			t.Errorf("category directory missing: %s", category)
			continue
		}

		if len(expectedFiles) == 0 {
			continue
		}

		for _, expectedFile := range expectedFiles {
			filePath := filepath.Join(categoryPath, expectedFile)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("expected fixture file missing: %s", filePath)
			}
		}
	}
}

func TestValidBuiltinsAllPass(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "valid"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	for _, r := range results.Resources {
		if r.Status != types.StatusValid {
			t.Errorf("expected valid status for %s (%s), got %s: %v", r.Name, r.Kind, r.Status, r.Errors)
		}
	}
}

func TestInvalidBuiltinsAllFail(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "invalid"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected invalid resources but got none")
	}

	for _, r := range results.Resources {
		if r.Status == types.StatusValid {
			t.Errorf("expected invalid status for %s (%s), got valid", r.Name, r.Kind)
		}
	}
}

func TestCustomResourceValid(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "crd", "database-crd.yaml")
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "cr", "valid-database.yaml"), []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status != types.StatusValid {
		t.Errorf("expected valid status for custom resource, got %s: %v", r.Status, r.Errors)
	}
}

func TestCustomResourceInvalid(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "crd", "database-crd.yaml")
	engine := validator.NewEngine(validator.EngineOptions{})

	results, err := engine.Validate(filepath.Join(fixturesDir, "cr", "invalid-database-missing-version.yaml"), []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status == types.StatusValid {
		t.Errorf("expected invalid status for CR missing required field, got valid")
	}
}

func TestCustomResourceWrongType(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "crd", "database-crd.yaml")
	engine := validator.NewEngine(validator.EngineOptions{})

	results, err := engine.Validate(filepath.Join(fixturesDir, "cr", "invalid-database-wrong-type.yaml"), []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status == types.StatusValid {
		t.Errorf("expected invalid status for CR with wrong enum value, got valid")
	}
}

func TestCRDSelfValidationValid(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "crd", "database-crd.yaml"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status != types.StatusValid {
		t.Errorf("expected valid status for valid CRD, got %s: %v", r.Status, r.Errors)
	}
}

func TestCRDSelfValidationInvalid(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})

	tmpDir := t.TempDir()
	invalidCRD := filepath.Join(tmpDir, "invalid-crd.yaml")
	err := os.WriteFile(invalidCRD, []byte(`
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: invalid.example.com
spec: []
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	results, err := engine.Validate(invalidCRD, nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status == types.StatusValid {
		t.Errorf("expected invalid/error status for malformed CRD, got valid")
	}
}

func TestCRDMissingRequiredFields(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})

	tmpDir := t.TempDir()
	missingCRD := filepath.Join(tmpDir, "missing-required.yaml")
	err := os.WriteFile(missingCRD, []byte(`
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: missing.example.com
spec:
  group: example.com
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	results, err := engine.Validate(missingCRD, nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status == types.StatusValid {
		t.Errorf("expected invalid/error status for CRD missing required fields, got valid")
	}
}

func TestUnknownKindWithIgnore(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{IgnoreMissing: true})
	results, err := engine.Validate(filepath.Join(fixturesDir, "unknown"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status != types.StatusSkipped {
		t.Errorf("expected skipped status for unknown kind with ignore-missing-schemas, got %s", r.Status)
	}
}

func TestUnknownKindWithoutIgnore(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{IgnoreMissing: false})
	results, err := engine.Validate(filepath.Join(fixturesDir, "unknown"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status != types.StatusError {
		t.Errorf("expected error status for unknown kind without ignore-missing-schemas, got %s", r.Status)
	}
}

func TestMalformedYAML(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "malformed"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status != types.StatusError {
		t.Errorf("expected error status for malformed YAML, got %s", r.Status)
	}
}

func TestMultiVersionCRDv1(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "multi-version", "cache-crd.yaml")
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "multi-version", "cache-v1.yaml"), []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status != types.StatusValid {
		t.Errorf("expected valid status for v1 CR, got %s: %v", r.Status, r.Errors)
	}
}

func TestMultiVersionCRDv1beta1(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "multi-version", "cache-crd.yaml")
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "multi-version", "cache-v1beta1.yaml"), []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status != types.StatusValid {
		t.Errorf("expected valid status for v1beta1 CR, got %s: %v", r.Status, r.Errors)
	}
}

func TestMultiVersionMismatch(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "multi-version", "cache-crd.yaml")
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "multi-version", "cache-v2-mismatch.yaml"), []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result but got none")
	}

	r := results.Resources[0]
	if r.Status == types.StatusValid {
		t.Errorf("expected invalid/error status for v2 CR when only v1 and v1beta1 exist, got valid")
	}
}

func TestOutputFormatJSON(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "valid"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	jsonBytes, err := results.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize to JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Errorf("output is not valid parseable JSON: %v", err)
	}
}

func TestOutputFormatYAML(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "valid"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	yamlBytes, err := results.ToYAML()
	if err != nil {
		t.Fatalf("failed to serialize to YAML: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &parsed); err != nil {
		t.Errorf("output is not valid parseable YAML: %v", err)
	}
}

func TestExitCodeZero(t *testing.T) {
	binPath := findBinary(t)
	tmpDir := t.TempDir()
	validManifest := filepath.Join(tmpDir, "valid.yaml")
	err := os.WriteFile(validManifest, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binPath, "-f", validManifest)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 0 {
			return
		}
	}

	t.Errorf("expected exit code 0 for valid manifest, got error: %v, output: %s", err, string(output))
}

func TestExitCodeOne(t *testing.T) {
	binPath := findBinary(t)
	tmpDir := t.TempDir()
	invalidManifest := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(invalidManifest, []byte(`
apiVersion: v1
kind: UnknwonKind
metadata:
  name: test
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binPath, "-f", invalidManifest)
	_, err = cmd.CombinedOutput()

	if err == nil {
		t.Errorf("expected non-zero exit code for invalid manifest, got exit code 0")
		return
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Errorf("expected ExitError, got: %v", err)
	}
}

func TestEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(tmpDir, nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if results.Summary.Total != 0 {
		t.Errorf("expected 0 resources for empty directory, got %d", results.Summary.Total)
	}
}

func TestMixedResources(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "crd", "database-crd.yaml")
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(filepath.Join(fixturesDir, "mixed"), []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if results.Summary.Total < 3 {
		t.Errorf("expected at least 3 resources (ConfigMap, Deployment, Database CR), got %d", results.Summary.Total)
	}

	for _, r := range results.Resources {
		if r.Kind == "ConfigMap" && r.Status != types.StatusValid {
			t.Errorf("ConfigMap should be valid, got %s: %v", r.Status, r.Errors)
		}
		if r.Kind == "Database" && r.Status != types.StatusValid {
			t.Errorf("Database CR should be valid, got %s: %v", r.Status, r.Errors)
		}
	}
}

func TestPerTypeRouting(t *testing.T) {
	builtinKinds := []string{
		"Deployment",
		"Service",
		"ConfigMap",
		"Secret",
		"Ingress",
		"StatefulSet",
		"DaemonSet",
		"Job",
		"CronJob",
		"PersistentVolumeClaim",
	}

	tmpDir := t.TempDir()
	engine := validator.NewEngine(validator.EngineOptions{})

	for _, kind := range builtinKinds {
		manifest := buildValidBuiltinManifest(kind)
		manifestPath := filepath.Join(tmpDir, kind+".yaml")
		if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
			t.Fatal(err)
		}

		results, err := engine.Validate(manifestPath, nil)
		if err != nil {
			t.Fatalf("failed to validate %s: %v", kind, err)
		}

		if len(results.Resources) == 0 {
			t.Errorf("expected result for %s", kind)
			continue
		}

		r := results.Resources[0]
		if r.Kind != kind {
			t.Errorf("expected kind %s, got %s", kind, r.Kind)
		}
	}
}

func buildValidBuiltinManifest(kind string) string {
	switch kind {
	case "Deployment":
		return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:1.21`
	case "Service":
		return `apiVersion: v1
kind: Service
metadata:
  name: test-svc
spec:
  selector:
    app: test
  ports:
  - port: 80`
	case "ConfigMap":
		return `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
data:
  key: value`
	case "Secret":
		return `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
type: Opaque
data:
  key: dmFsdWU=`
	case "Ingress":
		return `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: test-svc
            port:
              number: 80`
	case "StatefulSet":
		return `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: test-ss
  namespace: default
spec:
  serviceName: test
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:1.21`
	case "DaemonSet":
		return `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: test-ds
  namespace: default
spec:
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:1.21`
	case "Job":
		return `apiVersion: batch/v1
kind: Job
metadata:
  name: test-job
spec:
  template:
    spec:
      containers:
      - name: hello
        image: busybox:1.28
        command: ["sh", "-c", "echo hello"]
      restartPolicy: OnFailure`
	case "CronJob":
		return `apiVersion: batch/v1
kind: CronJob
metadata:
  name: test-cj
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox:1.28
            command: ["sh", "-c", "echo hello"]
          restartPolicy: OnFailure`
	case "PersistentVolumeClaim":
		return `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi`
	default:
		return ""
	}
}

func TestCLIFlagsOutputFormat(t *testing.T) {
	binPath := findBinary(t)
	tmpDir := t.TempDir()
	validManifest := filepath.Join(tmpDir, "valid.yaml")
	err := os.WriteFile(validManifest, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binPath, "-f", validManifest, "-output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("JSON output failed: %v, output: %s", err, string(output))
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(output, &jsonData); err != nil {
		t.Errorf("JSON output is not valid JSON: %v", err)
	}

	cmd = exec.Command(binPath, "-f", validManifest, "-output", "yaml")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("YAML output failed: %v, output: %s", err, string(output))
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(output, &yamlData); err != nil {
		t.Errorf("YAML output is not valid YAML: %v", err)
	}
}

func TestCLIFlagsIgnoreMissingSchemas(t *testing.T) {
	binPath := findBinary(t)
	tmpDir := t.TempDir()
	unknownManifest := filepath.Join(tmpDir, "unknown.yaml")
	err := os.WriteFile(unknownManifest, []byte(`
apiVersion: example.com/v1
kind: UnknownKind
metadata:
  name: test-unknown
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binPath, "-f", unknownManifest)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("expected error without -ignore-missing-schemas flag")
	}

	cmd = exec.Command(binPath, "-f", unknownManifest, "-ignore-missing-schemas")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("expected success with -ignore-missing-schemas flag: %v, output: %s", err, string(output))
	}
}

func TestCLIFlagVersion(t *testing.T) {
	binPath := findBinary(t)
	cmd := exec.Command(binPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("version flag failed: %v", err)
	}
	if !strings.Contains(string(output), "version") {
		t.Errorf("expected version output to contain 'version', got: %s", string(output))
	}
}

func TestCRDLegacyFormat(t *testing.T) {
	engine := validator.NewEngine(validator.EngineOptions{})

	tmpDir := t.TempDir()
	legacyCRD := filepath.Join(tmpDir, "legacy-crd.yaml")
	err := os.WriteFile(legacyCRD, []byte(`
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: legacy.example.com
spec:
  group: example.com
  names:
    kind: Legacy
    plural: legacies
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	results, err := engine.Validate(legacyCRD, nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result for legacy CRD")
	}
}

func TestMalformedYAMLParseError(t *testing.T) {
	tmpDir := t.TempDir()
	malformedYAML := filepath.Join(tmpDir, "bad.yaml")
	err := os.WriteFile(malformedYAML, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  - invalid: list
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(malformedYAML, nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if len(results.Resources) == 0 {
		t.Fatal("expected result for malformed YAML")
	}

	r := results.Resources[0]
	if r.Status != types.StatusError {
		t.Errorf("expected error status for malformed YAML, got %s", r.Status)
	}

	if len(r.Errors) == 0 {
		t.Error("expected error details for malformed YAML")
	}
}

func TestEmptyManifestsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	engine := validator.NewEngine(validator.EngineOptions{})
	results, err := engine.Validate(tmpDir, nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if results.Summary.Total != 0 {
		t.Errorf("expected 0 results for empty directory, got %d", results.Summary.Total)
	}
}
