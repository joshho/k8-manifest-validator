package validator_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8-manifest-validator/pkg/types"
	"k8-manifest-validator/pkg/validator"
)

const scenariosDir = "scenarios"

// exitCodeFromResults determines the exit code from validation results.
func exitCodeFromResults(results *types.Results) int {
	if results.Summary.Invalid > 0 || results.Summary.Errors > 0 {
		return 1
	}
	return 0
}

// TestScenarioKubeconformAligned runs our validator and kubeconform on the same scenario inputs
// and asserts the same pass/fail result for each.
func TestScenarioKubeconformAligned(t *testing.T) {
	kcPath := findKubeconform(t)
	if kcPath == "" {
		t.Skip("kubeconform not installed — skipping alignment test")
	}

	// Each scenario with expected statuses from our validator.
	// Our validator uses k8s Go validation (strict type checks, error on unknown version,
	// strips unknown fields by default). This differs from kubeconform's JSON Schema approach.
	type scenario struct {
		name              string
		file              string
		statuses          []types.Status
		expectKcAlignment bool
	}
	scenarios := []scenario{
		{
			name:              "valid multi-resource",
			file:              "valid-multi-resource.yaml",
			statuses:          []types.Status{types.StatusValid, types.StatusValid, types.StatusValid},
			expectKcAlignment: true,
		},
		{
			name: "missing selector deployment",
			file: "missing-selector-deployment.yaml",
			// k8s Go struct doesn't enforce selector as required (JSON Schema does)
			// This is a known, valid divergence
			statuses:          []types.Status{types.StatusValid},
			expectKcAlignment: false,
		},
		{
			name:              "wrong type value",
			file:              "wrong-type-value.yaml",
			statuses:          []types.Status{types.StatusError},
			expectKcAlignment: true,
		},
		{
			name:              "deprecated API version",
			file:              "deprecated-api-version.yaml",
			statuses:          []types.Status{types.StatusError},
			expectKcAlignment: true,
		},
		{
			name:              "unknown field deployment",
			file:              "unknown-field-deployment.yaml",
			statuses:          []types.Status{types.StatusValid},
			expectKcAlignment: true,
		},
	}

	engine := validator.NewEngine(validator.EngineOptions{})

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			scenarioPath := filepath.Join(scenariosDir, sc.file)

			// Run our validator
			results, err := engine.Validate(scenarioPath, nil)
			if err != nil {
				t.Fatalf("our validator failed: %v", err)
			}

			// Check resource counts match expected
			if len(results.Resources) != len(sc.statuses) {
				t.Errorf("expected %d resources, got %d", len(sc.statuses), len(results.Resources))
				return
			}

			// Check each resource status matches expected
			for i, expected := range sc.statuses {
				if i >= len(results.Resources) {
					break
				}
				r := results.Resources[i]
				if r.Status != expected {
					t.Errorf("resource %d (%s): expected status %s, got %s: %v",
						i, r.Name, expected, r.Status, r.Errors)
				}
			}

			ourExitCode := exitCodeFromResults(results)

			// Run kubeconform on the same scenario
			kcOutput, kcExitCode := runKubeconform(t, kcPath, scenarioPath)

			if sc.expectKcAlignment {
				if (ourExitCode == 0) != (kcExitCode == 0) {
					t.Logf("kubeconform output: %s", kcOutput)
					t.Logf("our resources: %+v", results.Resources)
					t.Errorf("exit code mismatch: our=%d, kubeconform=%d for %s",
						ourExitCode, kcExitCode, sc.name)
				}
			} else {
				if (ourExitCode == 0) != (kcExitCode == 0) {
					t.Logf("known divergence for %s: our=%d, kubeconform=%d (%s)",
						sc.name, ourExitCode, kcExitCode, strings.TrimSpace(kcOutput))
				} else {
					t.Logf("surprising alignment for %s: both %d", sc.name, ourExitCode)
				}
			}
		})
	}
}

// TestScenarioEdgeCases tests edge cases that our validator handles independently.
func TestScenarioEdgeCases(t *testing.T) {
	type edgeCase struct {
		name     string
		file     string
		validate func(*testing.T, *types.Results)
	}
	scenarios := []edgeCase{
		{
			name: "empty file produces 1 resource (YAML comment entry)",
			file: "empty-file.yaml",
			validate: func(t *testing.T, results *types.Results) {
				if results.Summary.Total != 1 {
					t.Errorf("expected 1 resource (YAML comment entry), got %d", results.Summary.Total)
				}
			},
		},
	}

	engine := validator.NewEngine(validator.EngineOptions{})

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			scenarioPath := filepath.Join(scenariosDir, sc.file)
			if _, err := os.Stat(scenarioPath); os.IsNotExist(err) {
				t.Skipf("scenario file not found: %s", sc.file)
			}

			results, err := engine.Validate(scenarioPath, nil)
			if err != nil {
				t.Fatalf("our validator failed: %v", err)
			}

			sc.validate(t, results)
		})
	}
}

// TestKubeconformScenarioConsistency runs all scenario files through both tools
// and reports any divergence for awareness.
func TestKubeconformScenarioConsistency(t *testing.T) {
	kcPath := findKubeconform(t)
	if kcPath == "" {
		t.Skip("kubeconform not installed — skipping consistency test")
	}

	files, err := os.ReadDir(scenariosDir)
	if err != nil {
		t.Fatalf("failed to read scenarios dir: %v", err)
	}

	engine := validator.NewEngine(validator.EngineOptions{})

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".yaml") && !strings.HasSuffix(f.Name(), ".yml") {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			scenarioPath := filepath.Join(scenariosDir, f.Name())

			// Run our validator
			ourResults, err := engine.Validate(scenarioPath, nil)
			if err != nil {
				t.Fatalf("our validator failed: %v", err)
			}

			// Run kubeconform
			kcOutput, kcExitCode := runKubeconform(t, kcPath, scenarioPath)
			ourExitCode := exitCodeFromResults(ourResults)

			t.Logf("%s: our=%d resources, exit=%d; kubeconform: exit=%d",
				f.Name(), ourResults.Summary.Total, ourExitCode, kcExitCode)

			if (ourExitCode == 0) != (kcExitCode == 0) {
				t.Logf("divergence: our exit %d vs kubeconform exit %d — expected due to different validation approaches (Go validation vs JSON Schema)", ourExitCode, kcExitCode)
				t.Logf("kubeconform output: %s", kcOutput)
			}
		})
	}
}
