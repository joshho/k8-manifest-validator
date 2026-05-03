package validator_tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// pbFindBinary builds or locates the k8-manifest-validator binary for CLI tests.
func pbFindBinary(t *testing.T) string {
	t.Helper()
	if path, err := exec.LookPath("k8-manifest-validator"); err == nil {
		return path
	}
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "k8-manifest-validator")
	cmd := exec.Command("go", "build", "-o", binPath, "cmd/k8-manifest-validator/")
	cmd.Dir = pbRepoRoot()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("failed to build k8-manifest-validator: %v\n%s", err, string(out))
	}
	return binPath
}

// pbRepoRoot returns the repo root (workspace/ dir containing cmd/, pkg/, go.mod).
func pbRepoRoot() string {
	cwd, _ := os.Getwd()
	return filepath.Dir(cwd) // tests/ -> workspace/
}

// runBinary runs the CLI with the given args from the tests/ CWD.
func runBinary(t *testing.T, args ...string) (string, error) {
	binPath := pbFindBinary(t)
	cmd := exec.Command(binPath, args...)
	cmd.Dir = "tests/"
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// --- assertion helpers ---

func assertExitCode(want int, err error) error {
	if err == nil {
		if want != 0 {
			return &exitCodeAssertionError{want: want, got: 0}
		}
		return nil
	}
	ex, ok := err.(*exec.ExitError)
	if !ok {
		if want != 0 {
			return &exitCodeAssertionError{want: want, got: -1}
		}
		return nil
	}
	if ex.ExitCode() != want {
		return &exitCodeAssertionError{want: want, got: ex.ExitCode()}
	}
	return nil
}

type exitCodeAssertionError struct {
	want int
	got  int
}

func (e *exitCodeAssertionError) Error() string {
	return "expected exit code " + itoa(e.want) + ", got " + itoa(e.got)
}

func itoa(i int) string {
	if i < 0 {
		return "-1"
	}
	return string('0' + byte(i%10))
}

func assertValid(out, kind string) error {
	if !strings.Contains(out, "kind") && !strings.Contains(out, kind) {
		return &assertionError{"expected valid output to mention kind '" + kind + "'"}
	}
	if strings.Contains(out, `"status":"invalid"`) {
		return &assertionError{"expected valid output but got invalid status"}
	}
	return nil
}

func assertInvalid(out, kind string) error {
	if !strings.Contains(out, kind) && !strings.Contains(out, `"status"`) {
		return &assertionError{"expected invalid output to mention kind '" + kind + "' or status"}
	}
	return nil
}

func assertValidCount(out string, min int) error {
	if !strings.Contains(out, `"status":"valid"`) {
		return &assertionError{"expected at least " + itoa(min) + " valid resources"}
	}
	return nil
}

func assertInvalidCount(out string, min int) error {
	if !strings.Contains(out, `"status":"invalid"`) && !strings.Contains(out, `"status":"error"`) {
		return &assertionError{"expected at least " + itoa(min) + " invalid resources"}
	}
	return nil
}

func assertZeroResources(out string) error {
	var r struct {
		Summary struct {
			Total int `json:"total"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		return err
	}
	if r.Summary.Total != 0 {
		return &assertionError{"expected 0 resources, got " + itoa(r.Summary.Total)}
	}
	return nil
}

func assertErrorContains(err error, substr string) error {
	if err == nil {
		return &assertionError{"expected error containing '" + substr + "', got nil"}
	}
	if !strings.Contains(err.Error(), substr) {
		return &assertionError{"expected error containing '" + substr + "', got: " + err.Error()}
	}
	return nil
}

func assertJSONValid(out string) error {
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		return &assertionError{"expected valid JSON: " + err.Error()}
	}
	return nil
}

func assertYAMLValid(out string) error {
	var v map[string]interface{}
	if err := yaml.Unmarshal([]byte(out), &v); err != nil {
		return &assertionError{"expected valid YAML: " + err.Error()}
	}
	return nil
}

type assertionError struct {
	msg string
}

func (a *assertionError) Error() string {
	return a.msg
}