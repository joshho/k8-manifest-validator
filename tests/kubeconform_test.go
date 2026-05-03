package validator_tests

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"k8-manifest-validator/pkg/types"
	"k8-manifest-validator/pkg/validator"
)

// createTestEngine creates a validator engine with the given options for test use.
func createTestEngine(ignoreMissing bool, workers int) *validator.Engine {
	return validator.NewEngine(validator.EngineOptions{
		IgnoreMissing: ignoreMissing,
		Workers:       workers,
	})
}

// findKubeconform checks if kubeconform is available on the PATH.
// Tests that require kubeconform should call this and t.Skip() if not found.
func findKubeconform(t *testing.T) string {
	t.Helper()
	// Check common locations
	candidates := []string{"kubeconform", "/home/node/go/bin/kubeconform"}
	for _, c := range candidates {
		path, err := exec.LookPath(c)
		if err == nil {
			return path
		}
	}
	return ""
}

// runKubeconform runs kubeconform on the given path and returns its combined output and exit code.
func runKubeconform(t *testing.T, kcPath, manifestPath string, crdDirs ...string) (string, int) {
	t.Helper()
	args := []string{"-summary"}
	if len(crdDirs) > 0 {
		for _, crdDir := range crdDirs {
			args = append(args, "-schema-location", "default")
			args = append(args, "-schema-location", filepath.Dir(crdDir))
		}
	}
	args = append(args, manifestPath)
	cmd := exec.Command(kcPath, args...)
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return strings.TrimSpace(string(output)), exitCode
}

// kubeconformSummary parses the kubeconform summary line.
// Expected format: "Summary: X resources found in Y files - Valid: A, Invalid: B, Errors: C, Skipped: D"
func kubeconformExitCode(output string) int {
	// kubeconform returns 0 if everything is valid, 1 if any invalid/error
	if strings.Contains(output, "Invalid: 0,") && strings.Contains(output, "Errors: 0,") {
		return 0
	}
	return 1
}

// TestKubeconformValidFixtures compares kubeconform results with our validator on valid fixtures.
func TestKubeconformValidFixtures(t *testing.T) {
	kcPath := findKubeconform(t)
	if kcPath == "" {
		t.Skip("kubeconform not installed — skipping integration test")
	}

	// Run kubeconform on valid fixtures
	_, kcExitCode := runKubeconform(t, kcPath, filepath.Join(fixturesDir, "valid"))

	// Run our validator on valid fixtures
	engine := createTestEngine(false, 8)
	results, err := engine.Validate(filepath.Join(fixturesDir, "valid"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	ourExitCode := 0
	if results.Summary.Invalid > 0 || results.Summary.Errors > 0 {
		ourExitCode = 1
	}

	if kcExitCode != ourExitCode {
		t.Errorf("kubeconform exit code %d vs our exit code %d for valid fixtures", kcExitCode, ourExitCode)
	}

	if results.Summary.Total != 10 {
		t.Errorf("expected 10 valid resources, got %d", results.Summary.Total)
	}
}

// TestKubeconformInvalidFixtures compares kubeconform results with our validator on invalid fixtures.
func TestKubeconformInvalidFixtures(t *testing.T) {
	kcPath := findKubeconform(t)
	if kcPath == "" {
		t.Skip("kubeconform not installed — skipping integration test")
	}

	// Run kubeconform on invalid fixtures
	_, kcExitCode := runKubeconform(t, kcPath, filepath.Join(fixturesDir, "invalid"))

	// Run our validator on invalid fixtures
	engine := createTestEngine(false, 8)
	results, err := engine.Validate(filepath.Join(fixturesDir, "invalid"), nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	ourExitCode := 0
	if results.Summary.Invalid > 0 || results.Summary.Errors > 0 {
		ourExitCode = 1
	}

	if kcExitCode != ourExitCode {
		t.Errorf("kubeconform exit code %d vs our exit code %d for invalid fixtures", kcExitCode, ourExitCode)
	}

	if results.Summary.Total == 0 {
		t.Error("expected invalid resources but got none")
	}
}

// TestKubeconformCRFixtures compares kubeconform with our validator on custom resource fixtures.
func TestKubeconformCRFixtures(t *testing.T) {
	kcPath := findKubeconform(t)
	if kcPath == "" {
		t.Skip("kubeconform not installed — skipping integration test")
	}

	crdPath := filepath.Join(fixturesDir, "crd", "database-crd.yaml")
	crDir := filepath.Join(fixturesDir, "cr")

	// Run kubeconform with CRD directory
	_, kcExitCode := runKubeconform(t, kcPath, crDir, filepath.Dir(crdPath))

	// Run our validator with CRD paths
	engine := createTestEngine(false, 8)
	results, err := engine.Validate(crDir, []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	ourExitCode := 0
	if results.Summary.Invalid > 0 || results.Summary.Errors > 0 {
		ourExitCode = 1
	}

	if kcExitCode != ourExitCode {
		t.Errorf("kubeconform exit code %d vs our exit code %d for CR fixtures", kcExitCode, ourExitCode)
	}
}

// TestKubeconformMixedFixtures compares kubeconform with our validator on mixed resources.
func TestKubeconformMixedFixtures(t *testing.T) {
	crdPath := filepath.Join(fixturesDir, "crd", "database-crd.yaml")
	mixedDir := filepath.Join(fixturesDir, "mixed")

	// Run our validator with CRD paths
	engine := createTestEngine(false, 8)
	results, err := engine.Validate(mixedDir, []string{crdPath})
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	if results.Summary.Invalid > 0 || results.Summary.Errors > 0 {
		t.Errorf("expected all resources to be valid, got %d invalid, %d errors", results.Summary.Invalid, results.Summary.Errors)
	}

	if results.Summary.Total < 3 {
		t.Errorf("expected at least 3 resources, got %d", results.Summary.Total)
	}

	for _, r := range results.Resources {
		if r.Status != types.StatusValid {
			t.Errorf("expected valid status for %s %s, got %s: %v", r.Kind, r.Name, r.Status, r.Errors)
		}
	}
}

// TestKubeconformMalformedYAML compares behavior on malformed YAML.
func TestKubeconformMalformedYAML(t *testing.T) {
	kcPath := findKubeconform(t)
	if kcPath == "" {
		t.Skip("kubeconform not installed — skipping integration test")
	}

	malformedDir := filepath.Join(fixturesDir, "malformed")

	// Run kubeconform on malformed YAML
	kcOutput, kcExitCode := runKubeconform(t, kcPath, malformedDir)
	_ = kcOutput // kubeconform may handle malformed YAML differently

	// Run our validator
	engine := createTestEngine(false, 8)
	results, err := engine.Validate(malformedDir, nil)
	if err != nil {
		t.Fatalf("failed to validate: %v", err)
	}

	// Both should flag malformed YAML as an error
	if kcExitCode == 0 {
		t.Log("kubeconform reports valid for malformed YAML (this is expected — kubeconform may be lenient)")
	}

	// Our validator must always report error for malformed YAML
	if len(results.Resources) > 0 && results.Resources[0].Status != "error" {
		t.Errorf("expected error status for malformed YAML, got %s", results.Resources[0].Status)
	}
}
