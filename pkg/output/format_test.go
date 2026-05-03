package output

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"

	"k8-manifest-validator/pkg/types"
)

func TestJSONFormatterProducesValidJSON(t *testing.T) {
	// Create a valid result with all fields
	results := &types.Results{
		Summary: types.Summary{
			Total:   1,
			Valid:   1,
			Invalid: 0,
			Errors:  0,
			Skipped: 0,
		},
		Resources: []types.Result{
			{
				File:       "test.yaml",
				Kind:       "Deployment",
				Name:       "test-deployment",
				Namespace:  "default",
				APIVersion: "apps/v1",
				Status:     types.StatusValid,
				Errors:     nil,
			},
		},
	}

	// Create JSON formatter
	formatter := NewJSONFormatter()

	// Format results
	data, err := formatter.Format(results)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Verify valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, string(data))
	}

	// Verify summary fields
	summary, ok := parsed["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("missing 'summary' in output")
	}

	if total, ok := summary["total"].(float64); !ok || total != 1 {
		t.Errorf("expected summary.total=1, got %v", summary["total"])
	}

	// Verify resources
	resources, ok := parsed["resources"].([]interface{})
	if !ok || len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %v", resources)
	}

	// Verify resource fields
	res := resources[0].(map[string]interface{})
	if res["kind"] != "Deployment" {
		t.Errorf("expected kind=Deployment, got %v", res["kind"])
	}
	if res["status"] != "valid" {
		t.Errorf("expected status=valid, got %v", res["status"])
	}
}

func TestJSONFormatterContainsAllResourceResults(t *testing.T) {
	// Create multiple resources
	results := &types.Results{
		Summary: types.Summary{
			Total:   3,
			Valid:   2,
			Invalid: 1,
			Errors:  0,
			Skipped: 0,
		},
		Resources: []types.Result{
			{
				File:       "deployment.yaml",
				Kind:       "Deployment",
				Name:       "app-deployment",
				Namespace:  "default",
				APIVersion: "apps/v1",
				Status:     types.StatusValid,
				Errors:     nil,
			},
			{
				File:       "configmap.yaml",
				Kind:       "ConfigMap",
				Name:       "app-config",
				Namespace:  "default",
				APIVersion: "v1",
				Status:     types.StatusValid,
				Errors:     nil,
			},
			{
				File:       "service.yaml",
				Kind:       "Service",
				Name:       "app-service",
				Namespace:  "default",
				APIVersion: "v1",
				Status:     types.StatusInvalid,
				Errors: []types.ErrorItem{
					{Field: "spec.ports[0].port", Message: "must be between 0 and 65535", Code: "Invalid"},
				},
			},
		},
	}

	// Create JSON formatter
	formatter := NewJSONFormatter()

	// Format results
	data, err := formatter.Format(results)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify all 3 resources are present
	resources := parsed["resources"].([]interface{})
	if len(resources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(resources))
	}

	// Verify each resource kind is present
	kinds := make(map[string]bool)
	for _, r := range resources {
		res := r.(map[string]interface{})
		kinds[res["kind"].(string)] = true
	}

	expectedKinds := []string{"Deployment", "ConfigMap", "Service"}
	for _, kind := range expectedKinds {
		if !kinds[kind] {
			t.Errorf("expected resource kind %q not found", kind)
		}
	}
}

func TestYAMLFormatterProducesValidYAML(t *testing.T) {
	// Create a valid result with all fields
	results := &types.Results{
		Summary: types.Summary{
			Total:   1,
			Valid:   1,
			Invalid: 0,
			Errors:  0,
			Skipped: 0,
		},
		Resources: []types.Result{
			{
				File:       "test.yaml",
				Kind:       "Deployment",
				Name:       "test-deployment",
				Namespace:  "default",
				APIVersion: "apps/v1",
				Status:     types.StatusValid,
				Errors:     nil,
			},
		},
	}

	// Create YAML formatter
	formatter := NewYAMLFormatter()

	// Format results
	data, err := formatter.Format(results)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Verify it's not empty
	if len(data) == 0 {
		t.Fatal("YAML output is empty")
	}

	// Verify it parses correctly as YAML
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid parseable YAML: %v", err)
	}

	// Verify structure
	if _, ok := parsed["summary"]; !ok {
		t.Error("YAML output missing 'summary' key")
	}
	if _, ok := parsed["resources"]; !ok {
		t.Error("YAML output missing 'resources' key")
	}
}

func TestYAMLFormatterWithErrors(t *testing.T) {
	// Create a result with errors
	results := &types.Results{
		Summary: types.Summary{
			Total:   1,
			Valid:   0,
			Invalid: 1,
			Errors:  1,
			Skipped: 0,
		},
		Resources: []types.Result{
			{
				File:       "bad.yaml",
				Kind:       "Deployment",
				Name:       "bad-deployment",
				Namespace:  "default",
				APIVersion: "apps/v1",
				Status:     types.StatusInvalid,
				Errors: []types.ErrorItem{
					{Field: "spec.replicas", Message: "must be >= 0", Code: "Invalid"},
				},
			},
		},
	}

	formatter := NewYAMLFormatter()
	data, err := formatter.Format(results)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Verify it parses correctly
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid YAML: %v", err)
	}

	// Verify error field is present
	resources := parsed["resources"].([]interface{})
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	res := resources[0].(map[string]interface{})
	if res["status"] != "invalid" {
		t.Errorf("expected status=invalid, got %v", res["status"])
	}
}

func TestCompactJSONFormatter(t *testing.T) {
	results := &types.Results{
		Summary: types.Summary{
			Total: 1,
			Valid: 1,
		},
		Resources: []types.Result{
			{
				File:       "test.yaml",
				Kind:       "Deployment",
				Name:       "test-deployment",
				APIVersion: "apps/v1",
				Status:     types.StatusValid,
			},
		},
	}

	formatter := NewCompactJSONFormatter()
	data, err := formatter.Format(results)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}
}

func TestJSONFormatterNilResults(t *testing.T) {
	formatter := NewJSONFormatter()
	_, err := formatter.Format(nil)
	if err == nil {
		t.Error("expected error for nil results")
	}
}

func TestYAMLFormatterNilResults(t *testing.T) {
	formatter := NewYAMLFormatter()
	_, err := formatter.Format(nil)
	if err == nil {
		t.Error("expected error for nil results")
	}
}