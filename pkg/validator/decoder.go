package validator

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

// Manifest represents a decoded Kubernetes manifest
type Manifest struct {
	APIVersion string
	Kind       string
	Metadata   ManifestMetadata
}

// ManifestMetadata contains the metadata from a Kubernetes manifest
type ManifestMetadata struct {
	Name      string
	Namespace string
}

// Decoder decodes YAML/JSON manifests into structured Manifest objects
type Decoder struct{}

// NewDecoder creates a new Decoder instance
func NewDecoder() *Decoder {
	return &Decoder{}
}

// DecodeManifest decodes a YAML or JSON byte slice into a Manifest struct
func (d *Decoder) DecodeManifest(data []byte) (Manifest, error) {
	// First, convert YAML to JSON using sigs.k8s.io/yaml
	// This handles both YAML and JSON input uniformly
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to parse YAML/JSON: %w", err)
	}

	// Check if we got null (empty document)
	if string(jsonData) == "null" {
		return Manifest{}, fmt.Errorf("empty or invalid manifest document")
	}

	// Now decode the JSON into a generic map[string]interface{}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(jsonData, &doc); err != nil {
		return Manifest{}, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	manifest := Manifest{}

	// Extract apiVersion
	if apiVersion, ok := doc["apiVersion"].(string); ok {
		manifest.APIVersion = apiVersion
	}

	// Extract kind
	if kind, ok := doc["kind"].(string); ok {
		manifest.Kind = kind
	}

	// Extract metadata
	if metadata, ok := doc["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			manifest.Metadata.Name = name
		}
		if namespace, ok := metadata["namespace"].(string); ok {
			manifest.Metadata.Namespace = namespace
		}
	}

	// Basic validation
	if manifest.Kind == "" {
		return Manifest{}, fmt.Errorf("manifest is missing 'kind' field")
	}

	return manifest, nil
}

// IsYAML checks if the data appears to be YAML (contains document separators or indented structures)
func IsYAML(data []byte) bool {
	// If it starts with --- it's definitely YAML
	if len(data) >= 3 && data[0] == '-' && data[1] == '-' && data[2] == '-' {
		return true
	}
	// If it contains : but isn't valid JSON, it's likely YAML
	return true // Most Kubernetes manifests are YAML; let the decoder figure it out
}
