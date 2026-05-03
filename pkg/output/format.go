package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"k8-manifest-validator/pkg/types"
)

// Formatter defines the interface for formatting validation results
type Formatter interface {
	Format(results *types.Results) ([]byte, error)
}

// JSONFormatter produces JSON formatted output
type JSONFormatter struct {
	// Pretty indicates whether to use pretty-printing (indent) or compact output
	Pretty bool
}

// NewJSONFormatter creates a JSONFormatter with pretty-printing enabled by default
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{Pretty: true}
}

// Format serializes validation results to JSON bytes
func (f *JSONFormatter) Format(results *types.Results) ([]byte, error) {
	if results == nil {
		return nil, fmt.Errorf("results cannot be nil")
	}

	if f.Pretty {
		return json.MarshalIndent(results, "", "  ")
	}
	return json.Marshal(results)
}

// YAMLFormatter produces YAML formatted output
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAMLFormatter
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format serializes validation results to YAML bytes
func (f *YAMLFormatter) Format(results *types.Results) ([]byte, error) {
	if results == nil {
		return nil, fmt.Errorf("results cannot be nil")
	}

	// Marshal to YAML using gopkg.in/yaml.v3
	data, err := yaml.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("YAML marshaling failed: %w", err)
	}

	return data, nil
}

// CompactJSONFormatter produces compact (non-pretty-printed) JSON
type CompactJSONFormatter struct{}

// NewCompactJSONFormatter creates a CompactJSONFormatter
func NewCompactJSONFormatter() *CompactJSONFormatter {
	return &CompactJSONFormatter{}
}

// Format serializes validation results to compact JSON bytes
func (f *CompactJSONFormatter) Format(results *types.Results) ([]byte, error) {
	if results == nil {
		return nil, fmt.Errorf("results cannot be nil")
	}
	return json.Marshal(results)
}