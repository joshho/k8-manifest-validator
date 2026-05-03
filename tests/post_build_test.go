package validator_tests

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIHelp(t *testing.T) {
	binPath := pbFindBinary(t)
	cmd := exec.Command(binPath, "-h")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("help flag failed: %v", err)
	}
	output := string(out)
	if !strings.Contains(output, "-f ") && !strings.Contains(output, "manifest") {
		t.Errorf("help output missing expected content, got: %s", output)
	}
}

func TestCLIVersion(t *testing.T) {
	binPath := pbFindBinary(t)
	cmd := exec.Command(binPath, "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("version flag failed: %v", err)
	}
	if !strings.Contains(string(out), "version") {
		t.Errorf("expected version output to contain 'version', got: %s", string(out))
	}
}

func TestCLIUnknownFlag(t *testing.T) {
	binPath := pbFindBinary(t)
	cmd := exec.Command(binPath, "-unknown-flag")
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("expected error for unknown flag")
	}
}

func TestCLIRequiredFlagMissing(t *testing.T) {
	binPath := pbFindBinary(t)
	cmd := exec.Command(binPath)
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("expected error when -f flag is missing")
	}
}

// --- 2. File Target Tests (single files, valid) ---

func TestCLIValidDeployment(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/deployment.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "Deployment"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidService(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/service.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "Service"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidConfigMap(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/configmap.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "ConfigMap"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidSecret(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/secret.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "Secret"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidIngress(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/ingress.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "Ingress"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidStatefulSet(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/statefulset.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "StatefulSet"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidDaemonSet(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/daemonset.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "DaemonSet"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidJob(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/job.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "Job"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidCronJob(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/cronjob.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "CronJob"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLIValidPVC(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/pvc.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "PersistentVolumeClaim"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

// --- 3. File Target Tests (single files, invalid) ---

func TestCLIInvalidIngress(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/ingress_missing_rules.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "Ingress"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidDeployment(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/deployment_wrong_kind.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "Deployment"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidService(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/service_missing_selector.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "Service"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidConfigMap(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/configmap_missing_name.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "ConfigMap"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidSecret(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/secret_missing_name.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "Secret"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidStatefulSet(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/statefulset_missing_service_name.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "StatefulSet"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidDaemonSet(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/daemonset_missing_selector.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "DaemonSet"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidJob(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/job_out_of_range.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "Job"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidCronJob(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/cronjob_invalid_schedule.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "CronJob"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

func TestCLIInvalidPVC(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid/pvc_missing_access_modes.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "PersistentVolumeClaim"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

// --- 4. Directory Target Tests ---

func TestCLIDirectoryValid(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValidCount(out, 1); gotErr != nil {
		t.Errorf("assertValidCount: %v", gotErr)
	}
}

func TestCLIDirectoryInvalid(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/invalid")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalidCount(out, 1); gotErr != nil {
		t.Errorf("assertInvalidCount: %v", gotErr)
	}
}

func TestCLIDirectoryMixed(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/mixed", "-crd", "fixtures/crd/database-crd.yaml")
	// mixed with CRD: all 3 resources (ConfigMap, Deployment, Database CR) should be valid -> exit 0
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if !strings.Contains(out, `"status"`) {
		t.Errorf("expected output to contain status field")
	}
}

func TestCLIDirectoryEmpty(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/empty")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertZeroResources(out); gotErr != nil {
		t.Errorf("assertZeroResources: %v", gotErr)
	}
}

func TestCLIDirectoryUnknownKind(t *testing.T) {
	// Unknown kind without -ignore-missing-schemas should fail
	out, err := runBinary(t, "-f", "fixtures/unknown")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if !strings.Contains(out, `"status"`) {
		t.Errorf("expected output to contain status field for unknown kind")
	}
}

// --- 5. Non-Existent File ---

func TestCLINonexistentFile(t *testing.T) {
	_, err := runBinary(t, "-f", "/nonexistent/file.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
}

func TestCLINonexistentDirectory(t *testing.T) {
	_, err := runBinary(t, "-f", "/nonexistent/dir")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
}

// --- 6. CR/CRD Tests ---

func TestCLICRDDirectory(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/cr", "-crd", "fixtures/crd/database-crd.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	var r struct {
		Resources []struct {
			Kind string `json:"kind"`
		} `json:"resources"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Errorf("failed to parse output: %v", err)
	}
	if len(r.Resources) == 0 {
		t.Errorf("expected at least one resource in CR directory output")
	}
}

func TestCLICRDValidCustomResource(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/cr/valid-database.yaml", "-crd", "fixtures/crd/database-crd.yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertValid(out, "Database"); gotErr != nil {
		t.Errorf("assertValid: %v", gotErr)
	}
}

func TestCLICRDInvalidCustomResource(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/cr/invalid-database-missing-version.yaml", "-crd", "fixtures/crd/database-crd.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertInvalid(out, "Database"); gotErr != nil {
		t.Errorf("assertInvalid: %v", gotErr)
	}
}

// --- 7. Output Format Tests ---

func TestCLIOutputJSON(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/deployment.yaml", "-output", "json")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertJSONValid(out); gotErr != nil {
		t.Errorf("assertJSONValid: %v", gotErr)
	}
}

func TestCLIOutputYAML(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/deployment.yaml", "-output", "yaml")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if gotErr := assertYAMLValid(out); gotErr != nil {
		t.Errorf("assertYAMLValid: %v", gotErr)
	}
}

func TestCLIUnknownFormat(t *testing.T) {
	_, err := runBinary(t, "-f", "fixtures/valid/deployment.yaml", "-output", "unknown")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
}

// --- 8. --ignore-missing-schemas ---

func TestCLIIgnoreMissingSchemas(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/unknown/unknown-widget.yaml", "-ignore-missing-schemas")
	if gotErr := assertExitCode(0, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	// With ignore, unknown kinds should be skipped (exit 0) and not produce error
	if strings.Contains(out, `"status":"error"`) {
		t.Errorf("expected no error status with -ignore-missing-schemas")
	}
}

func TestCLIUnknownKindWithoutIgnore(t *testing.T) {
	_, err := runBinary(t, "-f", "fixtures/unknown/unknown-widget.yaml")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
}

// --- 9. Malformed YAML ---

func TestCLIMalformedYAML(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/malformed")
	if gotErr := assertExitCode(1, err); gotErr != nil {
		t.Errorf("exit code: %v", gotErr)
	}
	if !strings.Contains(out, `"status"`) {
		t.Errorf("expected output to contain status field for malformed YAML")
	}
}

// --- 10. JSON result structure assertions ---

func TestCLIJSONSummaryFields(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/deployment.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var r struct {
		Summary struct {
			Total   int `json:"total"`
			Valid   int `json:"valid"`
			Invalid int `json:"invalid"`
			Skipped int `json:"skipped"`
			Errors  int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
	if r.Summary.Total == 0 {
		t.Errorf("expected at least 1 resource in output")
	}
	if r.Summary.Valid == 0 {
		t.Errorf("expected at least 1 valid resource")
	}
}

func TestCLIJSONResourceFields(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid/deployment.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var r struct {
		Resources []struct {
			Name   string `json:"name"`
			Kind   string `json:"kind"`
			Status string `json:"status"`
			File   string `json:"file"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"resources"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
	if len(r.Resources) == 0 {
		t.Fatalf("expected at least 1 resource")
	}
	res := r.Resources[0]
	if res.Name == "" {
		t.Errorf("expected resource to have a name")
	}
	if res.Kind == "" {
		t.Errorf("expected resource to have a kind")
	}
	if res.Status == "" {
		t.Errorf("expected resource to have a status")
	}
}

func TestCLIDirectoryMultiResourceSummary(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var r struct {
		Summary struct {
			Total int `json:"total"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if r.Summary.Total < 5 {
		t.Errorf("expected at least 5 resources in valid/ directory, got %d", r.Summary.Total)
	}
}

// --- 11. Workers flag ---

func TestCLIWorkersFlag(t *testing.T) {
	out, err := runBinary(t, "-f", "fixtures/valid", "-n", "2")
	if err != nil {
		t.Fatalf("unexpected error with -n flag: %v", err)
	}
	if !strings.Contains(out, `"total"`) {
		t.Errorf("expected valid output with workers flag set to 2")
	}
}

func TestCLIWorkersFlagInvalid(t *testing.T) {
	_, err := runBinary(t, "-f", "fixtures/valid", "-n", "-1")
	if err == nil {
		t.Errorf("expected error with negative workers")
	}
}
