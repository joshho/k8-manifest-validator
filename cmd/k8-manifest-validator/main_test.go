package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// repoRoot returns the absolute path to the repository root.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find repo root (no go.mod)")
		}
		wd = parent
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	binPath := filepath.Join(root, "k8-manifest-validator-test")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/k8-manifest-validator/")
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build CLI: %v", err)
	}
	t.Cleanup(func() { os.Remove(binPath) })
	return binPath
}

// TestCLIExitCodeZeroForValidManifest tests that CLI exits with code 0 for valid manifests
func TestCLIExitCodeZeroForValidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	validDeploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
`
	deploymentPath := filepath.Join(tmpDir, "deployment.yaml")
	if err := os.WriteFile(deploymentPath, []byte(validDeploymentYAML), 0644); err != nil {
		t.Fatalf("failed to write deployment.yaml: %v", err)
	}

	binaryPath := buildBinary(t)

	cliCmd := exec.Command(binaryPath, "-f", deploymentPath)
	var stdout, stderr bytes.Buffer
	cliCmd.Stdout = &stdout
	cliCmd.Stderr = &stderr
	if err := cliCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 0 {
				t.Errorf("expected exit code 0 for valid manifest, got %d", exitErr.ExitCode())
			}
		} else {
			t.Errorf("CLI error: %v", err)
		}
	}
}

// TestCLIExitCodeOneForInvalidManifest tests that CLI exits with code 1 for invalid manifests
func TestCLIExitCodeOneForInvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	invalidDeploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: bad-deployment
spec:
  replicas: -1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
`
	deploymentPath := filepath.Join(tmpDir, "deployment.yaml")
	if err := os.WriteFile(deploymentPath, []byte(invalidDeploymentYAML), 0644); err != nil {
		t.Fatalf("failed to write deployment.yaml: %v", err)
	}

	binaryPath := buildBinary(t)

	cliCmd := exec.Command(binaryPath, "-f", deploymentPath)
	var stdout, stderr bytes.Buffer
	cliCmd.Stdout = &stdout
	cliCmd.Stderr = &stderr
	err := cliCmd.Run()
	if err == nil {
		t.Error("expected CLI to fail for invalid manifest, but it succeeded")
	} else {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1 for invalid manifest, got %d", exitErr.ExitCode())
			}
		}
	}
}

// TestCLIHelpFlag tests that -h prints help text
func TestCLIHelpFlag(t *testing.T) {
	binaryPath := buildBinary(t)

	cliCmd := exec.Command(binaryPath, "-h")
	var stdout, stderr bytes.Buffer
	cliCmd.Stdout = &stdout
	cliCmd.Stderr = &stderr
	err := cliCmd.Run()

	t.Logf("stdout: %s", stdout.String())
	t.Logf("stderr: %s", stderr.String())

	if err != nil {
		t.Errorf("CLI -h should succeed: %v", err)
	}

	output := stdout.String() + stderr.String()
	if !bytes.Contains([]byte(output), []byte("-f")) {
		t.Error("help text should contain -f flag description")
	}
}

// TestCLIVersionFlag tests that --version prints the version string
func TestCLIVersionFlag(t *testing.T) {
	binaryPath := buildBinary(t)

	cliCmd := exec.Command(binaryPath, "--version")
	var stdout, stderr bytes.Buffer
	cliCmd.Stdout = &stdout
	cliCmd.Stderr = &stderr
	err := cliCmd.Run()

	t.Logf("stdout: %s", stdout.String())
	t.Logf("stderr: %s", stderr.String())

	if err != nil {
		t.Errorf("CLI --version should succeed: %v", err)
	}

	output := stdout.String() + stderr.String()
	if !bytes.Contains([]byte(output), []byte("version")) {
		t.Error("version output should contain 'version'")
	}
}

// TestCLIOutputFormat tests that -output yaml produces YAML output
func TestCLIOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	validDeploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
`
	deploymentPath := filepath.Join(tmpDir, "deployment.yaml")
	if err := os.WriteFile(deploymentPath, []byte(validDeploymentYAML), 0644); err != nil {
		t.Fatalf("failed to write deployment.yaml: %v", err)
	}

	binaryPath := buildBinary(t)

	cliCmd := exec.Command(binaryPath, "-f", deploymentPath, "-output", "yaml")
	var stdout, stderr bytes.Buffer
	cliCmd.Stdout = &stdout
	cliCmd.Stderr = &stderr
	err := cliCmd.Run()

	t.Logf("stdout: %s", stdout.String())
	t.Logf("stderr: %s", stderr.String())

	if err != nil {
		t.Errorf("CLI should succeed: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte(" Deployment")) && !bytes.Contains([]byte(output), []byte("test-deployment")) {
		t.Error("YAML output should contain Deployment information")
	}
}

// TestCLIIgnoreMissingSchemas tests that unknown kinds are skipped when flag is set
func TestCLIIgnoreMissingSchemas(t *testing.T) {
	tmpDir := t.TempDir()
	unknownKindYAML := `apiVersion: example.com/v1
kind: Widget
metadata:
  name: unknown-widget
spec:
  size: large
`
	manifestPath := filepath.Join(tmpDir, "widget.yaml")
	if err := os.WriteFile(manifestPath, []byte(unknownKindYAML), 0644); err != nil {
		t.Fatalf("failed to write widget.yaml: %v", err)
	}

	binaryPath := buildBinary(t)

	cliCmd := exec.Command(binaryPath, "-f", manifestPath, "-ignore-missing-schemas")
	var stdout, stderr bytes.Buffer
	cliCmd.Stdout = &stdout
	cliCmd.Stderr = &stderr
	err := cliCmd.Run()

	t.Logf("stdout: %s", stdout.String())
	t.Logf("stderr: %s", stderr.String())

	if err != nil {
		t.Errorf("CLI with -ignore-missing-schemas should succeed for unknown kinds: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("skipped")) {
		t.Error("output should contain 'skipped' status for unknown kind")
	}
}
