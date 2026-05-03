package validator

import (
	"testing"

	"k8-manifest-validator/pkg/types"
)

// TestCustomResourceValidAgainstCRDSchema tests that a valid CR passes validation against a valid CRD
func TestCustomResourceValidAgainstCRDSchema(t *testing.T) {
	// Valid CRD YAML
	crdYAML := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.example.com
spec:
  group: example.com
  names:
    kind: Test
    plural: tests
    singular: test
    listKind: TestList
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              field1:
                type: string
              field2:
                type: integer
  storedVersions:
  - v1
`)

	// Valid CR YAML matching the CRD schema
	crYAML := []byte(`apiVersion: example.com/v1
kind: Test
metadata:
  name: my-test
  namespace: default
spec:
  field1: hello
  field2: 42
`)

	// Create CRD validator
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	// Register the CRD
	err := crValidator.RegisterCRD(crdYAML)
	if err != nil {
		t.Fatalf("failed to register CRD: %v", err)
	}

	// Validate the CR
	result, err := crValidator.ValidateCR(crYAML)
	if err != nil {
		t.Fatalf("valid CR should not return error: %v", err)
	}
	if result.Status != types.StatusValid {
		t.Errorf("expected types.StatusValid, got %v", result.Status)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
}

// TestCRInvalidAgainstCRD tests that a CR violating the schema returns invalid result
func TestCRInvalidAgainstCRD(t *testing.T) {
	// Valid CRD YAML
	crdYAML := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.example.com
spec:
  group: example.com
  names:
    kind: Test
    plural: tests
    singular: test
    listKind: TestList
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              field1:
                type: string
              field2:
                type: integer
  storedVersions:
  - v1
`)

	// Invalid CR YAML - field2 should be integer, not string
	crYAML := []byte(`apiVersion: example.com/v1
kind: Test
metadata:
  name: my-test
  namespace: default
spec:
  field1: hello
  field2: "not-an-integer"
`)

	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	err := crValidator.RegisterCRD(crdYAML)
	if err != nil {
		t.Fatalf("failed to register CRD: %v", err)
	}

	result, err := crValidator.ValidateCR(crYAML)
	if err != nil {
		t.Fatalf("validation error is expected in result, not as error: %v", err)
	}
	if result.Status != types.StatusInvalid {
		t.Errorf("expected types.StatusInvalid, got %v", result.Status)
	}
	if len(result.Errors) == 0 {
		t.Error("expected validation errors for schema violation")
	}
}

// TestCRAPIVersionMismatch tests that CR with mismatched apiVersion returns error
func TestCRAPIVersionMismatch(t *testing.T) {
	// Valid CRD YAML - only serves v1
	crdYAML := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.example.com
spec:
  group: example.com
  names:
    kind: Test
    plural: tests
    singular: test
    listKind: TestList
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              field1:
                type: string
  storedVersions:
  - v1
`)

	// CR with apiVersion v2 - which is not served
	crYAML := []byte(`apiVersion: example.com/v2
kind: Test
metadata:
  name: my-test
  namespace: default
spec:
  field1: hello
`)

	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	err := crValidator.RegisterCRD(crdYAML)
	if err != nil {
		t.Fatalf("failed to register CRD: %v", err)
	}

	// CR apiVersion doesn't match any served version in CRD → should error
	_, err = crValidator.ValidateCR(crYAML)
	if err == nil {
		t.Fatal("expected error for apiVersion mismatch, got nil")
	}
}

// TestCRUnregisteredCRD tests that CR without registered CRD returns error
func TestCRUnregisteredCRD(t *testing.T) {
	// CR without any registered CRD
	crYAML := []byte(`apiVersion: example.com/v1
kind: Unknown
metadata:
  name: my-unknown
  namespace: default
spec:
  field1: hello
`)

	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	// Don't register any CRD
	_, err := crValidator.ValidateCR(crYAML)
	if err == nil {
		t.Fatal("expected error for unregistered CRD, got nil")
	}
}

// TestCRMultiVersionCRD tests multi-version CRD support
func TestCRMultiVersionCRD(t *testing.T) {
	// CRD with multiple versions
	crdYAML := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.example.com
spec:
  group: example.com
  names:
    kind: Test
    plural: tests
    singular: test
    listKind: TestList
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              betaField:
                type: string
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              v1Field:
                type: integer
  storedVersions:
  - v1
`)

	// CR using v1beta1
	crYAML := []byte(`apiVersion: example.com/v1beta1
kind: Test
metadata:
  name: my-test
  namespace: default
spec:
  betaField: hello
`)

	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	err := crValidator.RegisterCRD(crdYAML)
	if err != nil {
		t.Fatalf("failed to register CRD: %v", err)
	}

	result, err := crValidator.ValidateCR(crYAML)
	if err != nil {
		t.Fatalf("valid CR should not return error: %v", err)
	}
	if result.Status != types.StatusValid {
		t.Errorf("expected types.StatusValid, got %v", result.Status)
	}
}

// TestNewCRValidator tests that NewCRValidator returns a non-nil validator
func TestNewCRValidator(t *testing.T) {
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)
	if crValidator == nil {
		t.Fatal("NewCRValidator should return a non-nil validator")
	}
}
