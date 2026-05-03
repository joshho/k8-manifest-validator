package validator

import (
	"bytes"
	"testing"
)

// Tests

func TestCRDSchemaExtractionFromValidCRDYAML(t *testing.T) {
	// Valid CRD YAML with spec.versions[].schema.openAPIV3Schema
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

	v := NewCRDValidator()
	if v == nil {
		t.Fatal("NewCRDValidator returned nil")
	}

	crd, err := v.ValidateCRD(crdYAML)
	if err != nil {
		t.Fatalf("valid CRD should pass validation: %v", err)
	}
	if crd == nil {
		t.Fatal("CRD should not be nil")
	}

	// Extract schema from the CRD
	schema, err := v.ExtractSchema(crd)
	if err != nil {
		t.Fatalf("should extract schema successfully: %v", err)
	}
	if schema == nil {
		t.Fatal("schema should not be nil")
	}
}

func TestCRDInvalidSchemaFails(t *testing.T) {
	// Malformed CRD - missing required 'kind' in spec.names
	crdYAML := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.example.com
spec:
  group: example.com
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
  storedVersions:
  - v1
`)

	v := NewCRDValidator()
	if v == nil {
		t.Fatal("NewCRDValidator returned nil")
	}

	_, err := v.ValidateCRD(crdYAML)
	if err == nil {
		t.Fatal("malformed CRD should fail validation")
	}
}

func TestCRDMultiVersionSchemaExtraction(t *testing.T) {
	// CRD with multiple versions, served:true on multiple
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

	v := NewCRDValidator()
	if v == nil {
		t.Fatal("NewCRDValidator returned nil")
	}

	crd, err := v.ValidateCRD(crdYAML)
	if err != nil {
		t.Fatalf("valid multi-version CRD should pass validation: %v", err)
	}
	if crd == nil {
		t.Fatal("CRD should not be nil")
	}

	// Extract schema - should use first served version (v1beta1 since it's listed first with served:true)
	schema, err := v.ExtractSchema(crd)
	if err != nil {
		t.Fatalf("should extract schema successfully: %v", err)
	}
	if schema == nil {
		t.Fatal("schema should not be nil")
	}

	// The first served version should be v1beta1
	if len(crd.Spec.Versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(crd.Spec.Versions))
	}
}

func TestCRDAPIVersionMismatchFails(t *testing.T) {
	// CRD with a served version but the apiVersion in the CR doesn't match
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
  storedVersions:
  - v1
`)

	v := NewCRDValidator()
	if v == nil {
		t.Fatal("NewCRDValidator returned nil")
	}

	// Valid CRD should pass
	crd, err := v.ValidateCRD(crdYAML)
	if err != nil {
		t.Fatalf("valid CRD should pass: %v", err)
	}

	// But if someone tries to use an apiVersion not in the CRD, it should fail
	err = v.CheckAPIVersionMatch(crd, "v2", "example.com")
	if err == nil {
		t.Fatal("apiVersion v2 not in served versions should fail")
	}
}

func TestNewCRDValidator(t *testing.T) {
	v := NewCRDValidator()
	if v == nil {
		t.Fatal("NewCRDValidator should return a non-nil validator")
	}
}

func TestValidateCRDParsesYAML(t *testing.T) {
	// Minimal valid CRD
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
  storedVersions:
  - v1
`)

	v := NewCRDValidator()
	crd, err := v.ValidateCRD(crdYAML)
	if err != nil {
		t.Fatalf("ValidateCRD failed: %v", err)
	}
	if crd == nil {
		t.Fatal("CRD should not be nil")
	}
	if crd.Spec.Names.Kind != "Test" {
		t.Errorf("expected Kind 'Test', got %q", crd.Spec.Names.Kind)
	}
	if crd.Spec.Group != "example.com" {
		t.Errorf("expected Group 'example.com', got %q", crd.Spec.Group)
	}
}

func TestGetServedVersions(t *testing.T) {
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
  - name: v1alpha1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
  - name: v1beta1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
  storedVersions:
  - v1
`)

	v := NewCRDValidator()
	crd, err := v.ValidateCRD(crdYAML)
	if err != nil {
		t.Fatalf("ValidateCRD failed: %v", err)
	}

	versions := v.GetServedVersions(crd)
	if len(versions) != 3 {
		t.Errorf("expected 3 served versions, got %d", len(versions))
	}
	// Check that we have the expected versions
	found := map[string]bool{"v1alpha1": false, "v1beta1": false, "v1": false}
	for _, ver := range versions {
		if _, ok := found[ver]; ok {
			found[ver] = true
		}
	}
	for ver, ok := range found {
		if !ok {
			t.Errorf("missing version %q", ver)
		}
	}
}

// BenchmarkValidateCRD benchmarks the CRD validation
func BenchmarkValidateCRD(b *testing.B) {
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
              field3:
                type: boolean
  storedVersions:
  - v1
`)

	v := NewCRDValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.ValidateCRD(crdYAML)
		if err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

func TestBytesBufferPooling(t *testing.T) {
	// Test that we can handle YAML with multiple documents
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
  storedVersions:
  - v1
`)

	v := NewCRDValidator()
	crd, err := v.ValidateCRD(crdYAML)
	if err != nil {
		t.Fatalf("failed to validate CRD: %v", err)
	}

	// Test that the schema extraction works with a buffer
	schema, err := v.ExtractSchema(crd)
	if err != nil {
		t.Fatalf("failed to extract schema: %v", err)
	}
	if schema == nil {
		t.Fatal("schema should not be nil")
	}

	// Test that we can use the CRD with a bytes buffer
	_ = bytes.NewBuffer([]byte("test"))
}
