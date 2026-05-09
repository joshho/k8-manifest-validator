package validator

import (
	"testing"

	"k8-manifest-validator/pkg/types"
)

func TestLibraryNewReturnsValidEngine(t *testing.T) {
	v := New(Options{})
	if v == nil {
		t.Fatal("New() returned nil")
	}
	if v.engine == nil {
		t.Error("New().engine is nil")
	}
}

func TestLibraryNewWithOptions(t *testing.T) {
	v := New(Options{
		IgnoreMissingSchemas: true,
		Workers:              4,
	})
	if v == nil {
		t.Fatal("New() with options returned nil")
	}
	if v.engine == nil {
		t.Error("New().engine is nil")
	}
}

func TestLibraryValidateValidManifest(t *testing.T) {
	v := New(Options{})

	validDeployment := `apiVersion: apps/v1
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

	results, err := v.Validate([]byte(validDeployment))
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Status != types.StatusValid {
		t.Errorf("expected StatusValid, got %s. errors: %v", results[0].Status, results[0].Errors)
	}
}

func TestLibraryValidateInvalidManifest(t *testing.T) {
	v := New(Options{})

	invalidDeployment := `apiVersion: apps/v1
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

	results, err := v.Validate([]byte(invalidDeployment))
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Status != types.StatusInvalid {
		t.Errorf("expected StatusInvalid, got %s", results[0].Status)
	}
}

func TestLibraryValidateEmptyManifest(t *testing.T) {
	v := New(Options{})

	results, err := v.Validate([]byte(""))
	if err != nil {
		t.Fatalf("Validate() with empty manifest failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty manifest, got %d", len(results))
	}
}

func TestLibraryValidateUnknownKindWithIgnoreMissing(t *testing.T) {
	v := New(Options{
		IgnoreMissingSchemas: true,
	})

	unknownKind := `apiVersion: example.com/v1
kind: Widget
metadata:
  name: unknown-widget
spec:
  size: large
`

	results, err := v.Validate([]byte(unknownKind))
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Status != types.StatusSkipped {
		t.Errorf("expected StatusSkipped for unknown kind with IgnoreMissingSchemas=true, got %s", results[0].Status)
	}
}

func TestLibraryValidateUnknownKindWithoutIgnore(t *testing.T) {
	v := New(Options{
		IgnoreMissingSchemas: false,
	})

	unknownKind := `apiVersion: example.com/v1
kind: Widget
metadata:
  name: unknown-widget
spec:
  size: large
`

	results, err := v.Validate([]byte(unknownKind))
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Status != types.StatusError {
		t.Errorf("expected StatusError for unknown kind with IgnoreMissingSchemas=false, got %s", results[0].Status)
	}
}
