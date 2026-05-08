package validator

import (
	"encoding/json"
	"fmt"
	"sync"

	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"

	"k8-manifest-validator/pkg/types"
)

// CRValidator validates Custom Resource manifests against CRD schemas
type CRValidator struct {
	crdValidator   *CRDValidator
	schemaRegistry sync.Map // key: "group/version/kind" → *crdSchemaEntry
}

// crdSchemaEntry holds the CRD and its schema validator
type crdSchemaEntry struct {
	crd            *CRD
	schemaValidator validation.SchemaValidator
}

// NewCRValidator creates a new CR validator
func NewCRValidator(crdValidator *CRDValidator) *CRValidator {
	return &CRValidator{
		crdValidator:   crdValidator,
		schemaRegistry: sync.Map{},
	}
}

// crdKey builds a lookup key from group, version, and kind
func crdKey(group, version, kind string) string {
	return fmt.Sprintf("%s/%s/%s", group, version, kind)
}

// RegisterCRD parses a CRD YAML, validates it, extracts the schema, and stores it in the registry
func (v *CRValidator) RegisterCRD(crdYAML []byte) error {
	// Use the CRDValidator to parse and validate the CRD
	crd, err := v.crdValidator.ValidateCRD(crdYAML)
	if err != nil {
		return fmt.Errorf("failed to validate CRD: %w", err)
	}

	// Register a schema validator for each served version, built from that version's schema
	for _, ver := range crd.Spec.Versions {
		if !ver.Served {
			continue
		}
		schemaValidator, err := v.buildSchemaValidatorForVersion(crd, ver.Name)
		if err != nil {
			return fmt.Errorf("failed to build schema validator for version %s: %w", ver.Name, err)
		}
		key := crdKey(crd.Spec.Group, ver.Name, crd.Spec.Names.Kind)
		v.schemaRegistry.Store(key, &crdSchemaEntry{
			crd:            crd,
			schemaValidator: schemaValidator,
		})
	}

	return nil
}

// buildSchemaValidator creates a validation.SchemaValidator from a CRD's OpenAPI schema
func (v *CRValidator) buildSchemaValidator(crd *CRD) (validation.SchemaValidator, error) {
	// Get the schema from the first served version
	schema, err := v.crdValidator.ExtractSchema(crd)
	if err != nil {
		return nil, err
	}

	// Build a schema validator using NewSchemaValidator
	schemaValidator, _, err := validation.NewSchemaValidator(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema validator: %w", err)
	}

	return schemaValidator, nil
}

// buildSchemaValidatorForVersion creates a validation.SchemaValidator from a specific CRD version's schema
func (v *CRValidator) buildSchemaValidatorForVersion(crd *CRD, versionName string) (validation.SchemaValidator, error) {
	schema, err := v.crdValidator.ExtractVersionSchema(crd, versionName)
	if err != nil {
		return nil, err
	}

	schemaValidator, _, err := validation.NewSchemaValidator(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema validator for version %s: %w", versionName, err)
	}

	return schemaValidator, nil
}

// ValidateCR validates a Custom Resource against its registered CRD schema
func (v *CRValidator) ValidateCR(crYAML []byte) (*Result, error) {
	// Decode the CR to get its identity
	decoder := NewDecoder()
	manifest, err := decoder.DecodeManifest(crYAML)
	if err != nil {
		return &Result{
			Status: types.StatusError,
			Errors: []ErrorItem{
				{Field: "manifest", Message: err.Error(), Code: types.ErrCodeInvalid},
			},
		}, nil
	}

	// Parse the apiVersion to extract group and version
	group, version := parseAPIVersion(manifest.APIVersion)
	if group == "" || version == "" {
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusError,
			Errors: []ErrorItem{
				{Field: "apiVersion", Message: "invalid or missing apiVersion", Code: types.ErrCodeInvalid},
			},
		}, nil
	}

	// Look up the CRD by group/version/kind
	key := crdKey(group, version, manifest.Kind)
	entry, ok := v.schemaRegistry.Load(key)
	if !ok {
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusError,
			Errors: []ErrorItem{
				{Field: "kind", Message: fmt.Sprintf("no CRD registered for %s/%s", manifest.APIVersion, manifest.Kind), Code: types.ErrCodeNotFound},
			},
		}, fmt.Errorf("no CRD registered for %s/%s", manifest.APIVersion, manifest.Kind)
	}

	crdEntry := entry.(*crdSchemaEntry)

	// Parse the CR YAML into a map interface for validation
	// ValidateCustomResource expects a parsed JSON object, not raw bytes
	jsonData, err := yaml.YAMLToJSON(crYAML)
	if err != nil {
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusError,
			Errors: []ErrorItem{
				{Field: "manifest", Message: fmt.Sprintf("failed to parse YAML: %v", err), Code: types.ErrCodeInvalid},
			},
		}, nil
	}

	var crObject interface{}
	if err := json.Unmarshal(jsonData, &crObject); err != nil {
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusError,
			Errors: []ErrorItem{
				{Field: "manifest", Message: fmt.Sprintf("failed to parse JSON: %v", err), Code: types.ErrCodeInvalid},
			},
		}, nil
	}

	// Validate the CR against the schema using Kubernetes validation
	fldPath := field.NewPath("spec")
	allErrs := validation.ValidateCustomResource(fldPath, crObject, crdEntry.schemaValidator)
	if len(allErrs) > 0 {
		errorItems := make([]ErrorItem, 0, len(allErrs))
		for _, e := range allErrs {
			errorItems = append(errorItems, ErrorItem{
				Field:   e.Field,
				Message: e.Detail,
				Code:    string(e.Type),
			})
		}
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusInvalid,
			Errors:     errorItems,
		}, nil
	}

	return &Result{
		Kind:       manifest.Kind,
		Name:       manifest.Metadata.Name,
		Namespace:  manifest.Metadata.Namespace,
		APIVersion: manifest.APIVersion,
		Status:     types.StatusValid,
		Errors:     nil,
	}, nil
}

// parseAPIVersion splits "group/version" into separate group and version strings
func parseAPIVersion(apiVersion string) (group, version string) {
	for i := 0; i < len(apiVersion); i++ {
		if apiVersion[i] == '/' {
			return apiVersion[:i], apiVersion[i+1:]
		}
	}
	// No slash found - the whole string is the version (shouldn't happen for CRs)
	return "", apiVersion
}
