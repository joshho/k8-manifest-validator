package types

import (
	"encoding/json"
	"testing"
)

func TestResultStructJSONMarshaling(t *testing.T) {
	// Create a Result with all fields populated
	result := Result{
		File:       "test.yaml",
		Kind:       "Deployment",
		Name:       "my-app",
		Namespace:  "default",
		APIVersion: "apps/v1",
		Status:     StatusValid,
		Errors:     []ErrorItem{},
	}

	// Marshal to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(result) returned error: %v", err)
	}

	// Verify we got valid JSON (non-empty)
	if len(data) == 0 {
		t.Fatal("json.Marshal produced empty output")
	}

	// Unmarshal back to verify it's valid JSON that reconstructs the struct
	var decoded Result
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed on valid JSON: %v", err)
	}

	// Verify key fields
	if decoded.File != result.File {
		t.Errorf("File mismatch: got %q, want %q", decoded.File, result.File)
	}
	if decoded.Kind != result.Kind {
		t.Errorf("Kind mismatch: got %q, want %q", decoded.Kind, result.Kind)
	}
	if decoded.Name != result.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, result.Name)
	}
	if decoded.APIVersion != result.APIVersion {
		t.Errorf("APIVersion mismatch: got %q, want %q", decoded.APIVersion, result.APIVersion)
	}
	if decoded.Status != result.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, result.Status)
	}
}

func TestResultsStructJSONMarshaling(t *testing.T) {
	// Create a Results struct with multiple resources
	results := Results{
		Summary: Summary{
			Total:   1,
			Valid:   1,
			Invalid: 0,
			Errors:  0,
			Skipped: 0,
		},
		Resources: []Result{
			{
				File:       "test.yaml",
				Kind:       "Deployment",
				Name:       "my-app",
				Namespace:  "default",
				APIVersion: "apps/v1",
				Status:     StatusValid,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("json.Marshal(results) returned error: %v", err)
	}

	// Verify it's valid JSON
	if len(data) == 0 {
		t.Fatal("json.Marshal produced empty output")
	}

	// Unmarshal back
	var decoded Results
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed on valid JSON: %v", err)
	}

	// Verify Summary fields
	if decoded.Summary.Total != results.Summary.Total {
		t.Errorf("Summary.Total mismatch: got %d, want %d", decoded.Summary.Total, results.Summary.Total)
	}
	if decoded.Summary.Valid != results.Summary.Valid {
		t.Errorf("Summary.Valid mismatch: got %d, want %d", decoded.Summary.Valid, results.Summary.Valid)
	}

	// Verify resource count
	if len(decoded.Resources) != len(results.Resources) {
		t.Errorf("Resource count mismatch: got %d, want %d", len(decoded.Resources), len(results.Resources))
	}
}

func TestErrorItemJSONMarshaling(t *testing.T) {
	item := ErrorItem{
		Field:   "spec.replicas",
		Message: "must be greater than or equal to 1",
		Code:   "Forbidden",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal(ErrorItem) returned error: %v", err)
	}

	var decoded ErrorItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Field != item.Field {
		t.Errorf("Field mismatch: got %q, want %q", decoded.Field, item.Field)
	}
	if decoded.Message != item.Message {
		t.Errorf("Message mismatch: got %q, want %q", decoded.Message, item.Message)
	}
	if decoded.Code != item.Code {
		t.Errorf("Code mismatch: got %q, want %q", decoded.Code, item.Code)
	}
}

func TestStatusConstants(t *testing.T) {
	// Verify Status constants have expected string values
	if StatusValid != "valid" {
		t.Errorf("StatusValid = %q, want %q", StatusValid, "valid")
	}
	if StatusInvalid != "invalid" {
		t.Errorf("StatusInvalid = %q, want %q", StatusInvalid, "invalid")
	}
	if StatusSkipped != "skipped" {
		t.Errorf("StatusSkipped = %q, want %q", StatusSkipped, "skipped")
	}
	if StatusError != "error" {
		t.Errorf("StatusError = %q, want %q", StatusError, "error")
	}
}