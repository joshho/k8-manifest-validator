package types

import (
	"encoding/json"
	"fmt"
)

// Status represents the validation status of a resource
type Status string

const (
	StatusValid   Status = "valid"
	StatusInvalid Status = "invalid"
	StatusSkipped Status = "skipped"
	StatusError   Status = "error"
)

// Result represents the validation result for a single resource
type Result struct {
	File       string      `json:"file" yaml:"file"`
	Kind       string      `json:"kind" yaml:"kind"`
	Name       string      `json:"name" yaml:"name"`
	Namespace  string      `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	APIVersion string      `json:"apiVersion" yaml:"apiVersion"`
	Status     Status      `json:"status" yaml:"status"`
	Errors     []ErrorItem `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// ErrorItem represents a single validation error
type ErrorItem struct {
	Field   string `json:"field" yaml:"field"`
	Message string `json:"message" yaml:"message"`
	Code    string `json:"code" yaml:"code"`
}

// Results represents the complete validation output
type Results struct {
	Summary   Summary  `json:"summary" yaml:"summary"`
	Resources []Result `json:"resources" yaml:"resources"`
}

// Summary contains aggregated validation counts
type Summary struct {
	Total   int `json:"total" yaml:"total"`
	Valid   int `json:"valid" yaml:"valid"`
	Invalid int `json:"invalid" yaml:"invalid"`
	Errors  int `json:"errors" yaml:"errors"`
	Skipped int `json:"skipped" yaml:"skipped"`
}

// NewResult creates a new Result with the given fields
func NewResult(file, kind, name, namespace, apiVersion string, status Status, errors []ErrorItem) Result {
	return Result{
		File:       file,
		Kind:       kind,
		Name:       name,
		Namespace:  namespace,
		APIVersion: apiVersion,
		Status:     status,
		Errors:     errors,
	}
}

// NewResults creates a new Results with the given resources
func NewResults(resources []Result) Results {
	summary := CalculateSummary(resources)
	return Results{
		Summary:   summary,
		Resources: resources,
	}
}

// CalculateSummary computes summary statistics from a list of Results
func CalculateSummary(resources []Result) Summary {
	var summary Summary
	summary.Total = len(resources)
	for _, r := range resources {
		switch r.Status {
		case StatusValid:
			summary.Valid++
		case StatusInvalid:
			summary.Invalid++
			if len(r.Errors) > 0 {
				summary.Errors += len(r.Errors)
			}
		case StatusSkipped:
			summary.Skipped++
		case StatusError:
			summary.Errors++
		}
	}
	return summary
}

// ToJSON serializes Results to JSON bytes
func (r Results) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToYAML serializes Results to YAML bytes
func (r Results) ToYAML() ([]byte, error) {
	return json.Marshal(r) // Use json then convert to yaml-compatible format
}

// String implements fmt.Stringer for Status
func (s Status) String() string {
	return string(s)
}

// IsError returns true if the status indicates any kind of problem
func (s Status) IsError() bool {
	return s == StatusInvalid || s == StatusError
}

// HasErrors returns true if result has validation errors
func (r Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// Error codes for validation errors
const (
	// General errors
	ErrCodeInvalid   = "Invalid"
	ErrCodeForbidden = "Forbidden"
	ErrCodeNotFound  = "NotFound"
	ErrCodeConflict  = "Conflict"

	// Field validation errors
	ErrCodeRequired        = "Required"
	ErrCodeFieldTooLarge   = "FieldTooLarge"
	ErrCodeFieldTooSmall   = "FieldTooSmall"
	ErrCodeInvalidFormat   = "InvalidFormat"
	ErrCodeInvalidValue    = "InvalidValue"
	ErrCodeDuplicate       = "Duplicate"
)

// Validation error codes
var (
	ErrUnknownKind      = fmt.Errorf("unknown kind: no schema registered for this resource type")
	ErrVersionMismatch   = fmt.Errorf("version mismatch: resource apiVersion does not match any CRD served version")
	ErrCRDValidation     = fmt.Errorf("CRD validation failed: the custom resource definition is malformed")
	ErrFileNotFound      = fmt.Errorf("file not found: the specified path does not exist")
	ErrMalformedYAML     = fmt.Errorf("malformed YAML: unable to parse the manifest")
)