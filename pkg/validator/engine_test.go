package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8-manifest-validator/pkg/types"
)

func TestEngineValidatesValidDeploymentReturnsValid(t *testing.T) {
	// Create a temp directory with a valid deployment
	tmpDir := t.TempDir()
	deploymentYAML := `apiVersion: apps/v1
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
	if err := os.WriteFile(deploymentPath, []byte(deploymentYAML), 0644); err != nil {
		t.Fatalf("failed to write deployment.yaml: %v", err)
	}

	// Create engine
	engine := NewEngine(EngineOptions{
		IgnoreMissing: false,
		Workers:       1,
	})

	// Validate
	results, err := engine.Validate(tmpDir, nil)
	if err != nil {
		t.Fatalf("Engine.Validate failed: %v", err)
	}

	if results.Summary.Total != 1 {
		t.Errorf("expected 1 resource, got %d", results.Summary.Total)
	}

	if len(results.Resources) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results.Resources))
	}

	if results.Resources[0].Status != types.StatusValid {
		t.Errorf("expected StatusValid, got %s. errors: %v", results.Resources[0].Status, results.Resources[0].Errors)
	}

	if results.Summary.Valid != 1 {
		t.Errorf("expected 1 valid resource, got %d", results.Summary.Valid)
	}
}

func TestEngineValidatesMixedResources(t *testing.T) {
	// Create a temp directory with multiple resources
	tmpDir := t.TempDir()

	// Valid deployment
	deploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: mixed-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: demo
  template:
    metadata:
      labels:
        app: demo
    spec:
      containers:
      - name: demo
        image: demo:1.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "deployment.yaml"), []byte(deploymentYAML), 0644); err != nil {
		t.Fatalf("failed to write deployment.yaml: %v", err)
	}

	// Valid configmap
	configmapYAML := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  DATABASE_HOST: "localhost"
  DATABASE_PORT: "5432"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "configmap.yaml"), []byte(configmapYAML), 0644); err != nil {
		t.Fatalf("failed to write configmap.yaml: %v", err)
	}

	// Valid service
	serviceYAML := `apiVersion: v1
kind: Service
metadata:
  name: app-service
  namespace: default
spec:
  selector:
    app: demo
  ports:
  - port: 80
    targetPort: 8080
`
	if err := os.WriteFile(filepath.Join(tmpDir, "service.yaml"), []byte(serviceYAML), 0644); err != nil {
		t.Fatalf("failed to write service.yaml: %v", err)
	}

	// Create engine
	engine := NewEngine(EngineOptions{
		IgnoreMissing: false,
		Workers:       1,
	})

	// Validate
	results, err := engine.Validate(tmpDir, nil)
	if err != nil {
		t.Fatalf("Engine.Validate failed: %v", err)
	}

	if results.Summary.Total != 3 {
		t.Errorf("expected 3 resources, got %d", results.Summary.Total)
	}

	if results.Summary.Valid != 3 {
		t.Errorf("expected 3 valid resources, got %d (errors: %v)", results.Summary.Valid, results.Resources)
	}
}

func TestEngineUnknownKindSkipped(t *testing.T) {
	// Create a temp directory with an unknown kind
	tmpDir := t.TempDir()

	// Unknown kind manifest
	unknownYAML := `apiVersion: example.com/v1
kind: Widget
metadata:
  name: unknown-widget
spec:
  size: large
`
	if err := os.WriteFile(filepath.Join(tmpDir, "widget.yaml"), []byte(unknownYAML), 0644); err != nil {
		t.Fatalf("failed to write widget.yaml: %v", err)
	}

	// Create engine with ignoreMissing=true
	engine := NewEngine(EngineOptions{
		IgnoreMissing: true,
		Workers:       1,
	})

	// Validate
	results, err := engine.Validate(tmpDir, nil)
	if err != nil {
		t.Fatalf("Engine.Validate failed: %v", err)
	}

	if results.Summary.Total != 1 {
		t.Errorf("expected 1 resource, got %d", results.Summary.Total)
	}

	if results.Resources[0].Status != types.StatusSkipped {
		t.Errorf("expected StatusSkipped, got %s", results.Resources[0].Status)
	}

	if results.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", results.Summary.Skipped)
	}
}

func TestEngineUnknownKindErrors(t *testing.T) {
	// Create a temp directory with an unknown kind
	tmpDir := t.TempDir()

	// Unknown kind manifest
	unknownYAML := `apiVersion: example.com/v1
kind: Widget
metadata:
  name: unknown-widget
spec:
  size: large
`
	if err := os.WriteFile(filepath.Join(tmpDir, "widget.yaml"), []byte(unknownYAML), 0644); err != nil {
		t.Fatalf("failed to write widget.yaml: %v", err)
	}

	// Create engine with ignoreMissing=false (default)
	engine := NewEngine(EngineOptions{
		IgnoreMissing: false,
		Workers:       1,
	})

	// Validate
	results, err := engine.Validate(tmpDir, nil)
	if err != nil {
		t.Fatalf("Engine.Validate failed: %v", err)
	}

	if results.Summary.Total != 1 {
		t.Errorf("expected 1 resource, got %d", results.Summary.Total)
	}

	if results.Resources[0].Status != types.StatusError {
		t.Errorf("expected StatusError, got %s", results.Resources[0].Status)
	}

	if results.Summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", results.Summary.Errors)
	}
}

func TestEngineInvalidResource(t *testing.T) {
	// Create a temp directory with an invalid deployment
	tmpDir := t.TempDir()

	// Invalid deployment - missing required selector
	invalidDeploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: bad-deployment
spec:
  replicas: -1
  selector:
    matchLabels: {}
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid-deployment.yaml"), []byte(invalidDeploymentYAML), 0644); err != nil {
		t.Fatalf("failed to write invalid-deployment.yaml: %v", err)
	}

	// Create engine
	engine := NewEngine(EngineOptions{
		IgnoreMissing: false,
		Workers:       1,
	})

	// Validate
	results, err := engine.Validate(tmpDir, nil)
	if err != nil {
		t.Fatalf("Engine.Validate failed: %v", err)
	}

	if results.Summary.Total != 1 {
		t.Errorf("expected 1 resource, got %d", results.Summary.Total)
	}

	if results.Resources[0].Status != types.StatusInvalid {
		t.Errorf("expected StatusInvalid, got %s. errors: %v", results.Resources[0].Status, results.Resources[0].Errors)
	}

	if results.Summary.Invalid != 1 {
		t.Errorf("expected 1 invalid, got %d", results.Summary.Invalid)
	}
}
func TestValidateNonexistentFile(t *testing.T) {
	engine := NewEngine(EngineOptions{})
	results, err := engine.Validate("/nonexistent/path/to/file.yaml", nil)
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Fatalf("expected 'file not found' in error, got: %v", err)
	}
	if results != nil {
		t.Fatal("expected nil results on error")
	}
}
