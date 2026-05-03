package validator

import (
	"context"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// CRDValidator validates CustomResourceDefinition manifests
type CRDValidator struct {
}

// CRD represents a parsed CustomResourceDefinition
type CRD struct {
	Spec   CRDSpec
	Status CRDStatus
}

// CRDSpec is the spec portion of a CRD
type CRDSpec struct {
	Group    string
	Names    CRDNames
	Scope    string
	Versions []CRDVersion
}

// CRDNames contains the names for a CRD
type CRDNames struct {
	Kind   string
	Plural string
}

// CRDVersion represents a single version in a CRD
type CRDVersion struct {
	Name    string
	Served  bool
	Storage bool
	Schema  *apiextensions.JSONSchemaProps
}

// CRDStatus is the status portion of a CRD
type CRDStatus struct {
	StoredVersions []string
}

// NewCRDValidator creates a new CRD validator
func NewCRDValidator() *CRDValidator {
	return &CRDValidator{}
}

// ValidateCRD validates a CRD YAML and returns the parsed CRD
func (v *CRDValidator) ValidateCRD(crdYAML []byte) (*CRD, error) {
	// Set up a scheme for v1 CRDs
	scheme := runtime.NewScheme()
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add to scheme: %w", err)
	}

	// Decode the CRD
	codecFactory := serializer.NewCodecFactory(scheme)
	decoder := codecFactory.UniversalDeserializer()

	obj, _, err := decoder.Decode(crdYAML, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CRD: %w", err)
	}

	// Convert to internal type using the v1 scheme
	crdV1, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
	if !ok {
		return nil, fmt.Errorf("decoded object is not a CustomResourceDefinition")
	}

	// Apply defaults to ensure storedVersions is populated
	apiextensionsv1.SetDefaults_CustomResourceDefinition(crdV1)

	// Convert to internal representation for validation
	crdInternal := &apiextensions.CustomResourceDefinition{}
	if err := scheme.Convert(crdV1, crdInternal, nil); err != nil {
		return nil, fmt.Errorf("failed to convert CRD to internal type: %w", err)
	}

	// Validate the CRD against the meta-schema
	allErrs := validation.ValidateCustomResourceDefinition(context.Background(), crdInternal)
	if len(allErrs) > 0 {
		errMsg := ""
		for _, e := range allErrs {
			errMsg += fmt.Sprintf("%s: %s; ", e.Field, e.Detail)
		}
		return nil, fmt.Errorf("CRD validation failed: %s", errMsg)
	}

	// Build our CRD representation
	return v.buildCRD(crdInternal)
}

// buildCRD constructs our CRD type from the internal type
func (v *CRDValidator) buildCRD(crd *apiextensions.CustomResourceDefinition) (*CRD, error) {
	result := &CRD{
		Spec: CRDSpec{
			Group: crd.Spec.Group,
			Names: CRDNames{
				Kind:   crd.Spec.Names.Kind,
				Plural: crd.Spec.Names.Plural,
			},
			Scope: string(crd.Spec.Scope),
		},
	}

	// Check if schema is at top-level (all versions have identical schemas)
	topLevelSchema := crd.Spec.Validation

	// Convert versions
	for _, ver := range crd.Spec.Versions {
		crdVer := CRDVersion{
			Name:    ver.Name,
			Served:  ver.Served,
			Storage: ver.Storage,
		}
		// Schema can be either per-version or at top-level (if all versions share the same schema)
		if ver.Schema != nil && ver.Schema.OpenAPIV3Schema != nil {
			crdVer.Schema = ver.Schema.OpenAPIV3Schema
		} else if topLevelSchema != nil && topLevelSchema.OpenAPIV3Schema != nil {
			crdVer.Schema = topLevelSchema.OpenAPIV3Schema
		}
		result.Spec.Versions = append(result.Spec.Versions, crdVer)
	}

	// Convert status stored versions
	result.Status.StoredVersions = crd.Status.StoredVersions

	return result, nil
}

// ExtractSchema extracts the OpenAPIV3Schema from the first served version
func (v *CRDValidator) ExtractSchema(crd *CRD) (*apiextensions.JSONSchemaProps, error) {
	for _, version := range crd.Spec.Versions {
		if version.Served {
			if version.Schema == nil {
				return nil, fmt.Errorf("version %s is served but has no schema", version.Name)
			}
			return version.Schema, nil
		}
	}
	return nil, fmt.Errorf("no served version found in CRD")
}

// CheckAPIVersionMatch verifies the requested apiVersion matches a served version
func (v *CRDValidator) CheckAPIVersionMatch(crd *CRD, apiVersion, group string) error {
	for _, ver := range crd.Spec.Versions {
		if ver.Served {
			// Check if this version matches the requested apiVersion
			expectedGV := fmt.Sprintf("%s/%s", group, ver.Name)
			if apiVersion == expectedGV {
				return nil // Match found
			}
		}
	}
	return fmt.Errorf("apiVersion %q does not match any served version in CRD", apiVersion)
}

// GetServedVersions returns all served version names
func (v *CRDValidator) GetServedVersions(crd *CRD) []string {
	var versions []string
	for _, ver := range crd.Spec.Versions {
		if ver.Served {
			versions = append(versions, ver.Name)
		}
	}
	return versions
}
