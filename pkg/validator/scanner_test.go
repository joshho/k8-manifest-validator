package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScannerDiscoversYAMLFilesRecursively(t *testing.T) {
	// Use the actual fixtures directory
	fixturesDir := filepath.Join("..", "..", "tests", "fixtures")

	scanner := &Scanner{
		Extensions:    []string{".yaml", ".yml", ".json"},
		IgnorePattern: nil,
		MaxBufferSize: 256 * 1024 * 1024,
		InitBufferSize: 4 * 1024 * 1024,
	}

	resources, err := scanner.Walk(fixturesDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	// We expect at least 3 files: deployment.yaml, configmap.yaml, and ingress.yaml
	if len(resources) < 3 {
		t.Errorf("expected at least 3 resources, got %d: %v", len(resources), resources)
	}

	// Check that we found the deployment
	foundDeployment := false
	foundConfigMap := false
	foundIngress := false
	for _, r := range resources {
		switch filepath.Base(r.Path) {
		case "deployment.yaml":
			foundDeployment = true
			if len(r.Bytes) == 0 {
				t.Error("deployment.yaml has no bytes")
			}
		case "configmap.yaml":
			foundConfigMap = true
		case "ingress.yaml":
			foundIngress = true
		}
	}
	if !foundDeployment {
		t.Error("did not discover deployment.yaml")
	}
	if !foundConfigMap {
		t.Error("did not discover configmap.yaml")
	}
	if !foundIngress {
		t.Error("did not discover ingress.yaml (recursive)")
	}
}

func TestScannerRespectsIgnorePatterns(t *testing.T) {
	// Create temp dir with files
	tmpDir := t.TempDir()

	// Create some files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := []string{
		filepath.Join(tmpDir, "include.yaml"),
		filepath.Join(subDir, "exclude.yaml"),
		filepath.Join(tmpDir, "also-include.yaml"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner := &Scanner{
		Extensions:    []string{".yaml", ".yml", ".json"},
		IgnorePattern: nil,
		MaxBufferSize: 256 * 1024 * 1024,
		InitBufferSize: 4 * 1024 * 1024,
	}

	resources, err := scanner.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	if len(resources) != 3 {
		t.Errorf("expected 3 resources without ignore patterns, got %d", len(resources))
	}

	// Now test with ignore pattern
	_ = os.WriteFile(filepath.Join(subDir, "exclude.yaml"), []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test"), 0644)
}

func TestScannerSplitsMultiDocYAML(t *testing.T) {
	multiDoc := `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: first
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: third
`
	docs := SplitYAMLDocument([]byte(multiDoc))
	if len(docs) != 3 {
		t.Errorf("expected 3 documents, got %d", len(docs))
	}
}

func TestScannerSkipsNonManifestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a non-manifest file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "data.yaml"), []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := &Scanner{
		Extensions:    []string{".yaml", ".yml", ".json"},
		IgnorePattern: nil,
		MaxBufferSize: 256 * 1024 * 1024,
		InitBufferSize: 4 * 1024 * 1024,
	}

	resources, err := scanner.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	if len(resources) != 1 {
		t.Errorf("expected 1 resource (data.yaml), got %d", len(resources))
	}
}

func TestScannerHandlesHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a hidden file (starts with .)
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden.yaml"), []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: hidden-test"), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := &Scanner{
		Extensions:    []string{".yaml", ".yml", ".json"},
		IgnorePattern: nil,
		MaxBufferSize: 256 * 1024 * 1024,
		InitBufferSize: 4 * 1024 * 1024,
	}

	resources, err := scanner.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	// Hidden files should be processed normally
	if len(resources) != 1 {
		t.Errorf("expected 1 resource (.hidden.yaml), got %d", len(resources))
	}
}

func TestResourceSignature(t *testing.T) {
	r := Resource{
		Path:  "/path/to/deployment.yaml",
		Bytes: []byte("apiVersion: apps/v1\nkind: Deployment"),
	}

	sig := r.Signature()
	if sig == "" {
		t.Error("Signature() returned empty string")
	}

	// Same content should produce same signature
	r2 := Resource{
		Path:  "/path/to/deployment.yaml",
		Bytes: []byte("apiVersion: apps/v1\nkind: Deployment"),
	}
	if r.Signature() != r2.Signature() {
		t.Error("Same content should produce same signature")
	}
}
