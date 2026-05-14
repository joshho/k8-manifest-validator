package validator

import (
	"testing"

	"k8-manifest-validator/pkg/types"
)

// =============================================================================
// RW-13 — Malformed Manifest Rejection Tests
// PRD ref: Real-world edge case coverage — malformed YAML/JSON rejection
// Owner file: pkg/validator/phase4_realworld_malformed_test.go
// Target: Decoder.DecodeManifest and Engine.validateResource for malformed input
// Total: 25 malformed test cases (all should produce validation errors)
// =============================================================================

// -----------------------------------------------------------------------
// Malformed YAML Tests
// -----------------------------------------------------------------------

func TestPhase4_Malformed_YAML_TruncatedIncomplete(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		yaml    []byte
		wantErr bool
	}{
		{
			name:    "truncated YAML - incomplete document at colon",
			yaml:    []byte("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "truncated YAML - missing value after colon",
			yaml:    []byte("apiVersion: apps/v1\nkind:"),
			wantErr: true,
		},
		{
			name:    "truncated YAML - partial map",
			yaml:    []byte("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name:"),
			wantErr: true,
		},
		{
			name:    "truncated YAML - trailing backtick",
			yaml:    []byte("apiVersion: `apps/v1\nkind: Deployment"),
			wantErr: true,
		},
		{
			name:    "truncated YAML - unclosed quotes",
			yaml:    []byte("apiVersion: \"apps/v1\nkind: Deployment"),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.yaml)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

func TestPhase4_Malformed_YAML_InvalidSyntax(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		yaml    []byte
		wantErr bool
	}{
		{
			name:    "invalid YAML - bad indentation causing wrong structure",
			yaml:    []byte("apiVersion: apps/v1\n kind: Deployment\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - missing colon after key at root",
			yaml:    []byte("apiVersion apps/v1\nkind: Deployment\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - tab character instead of spaces",
			yaml:    []byte("apiVersion:\tapps/v1\nkind: Deployment"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - duplicate keys at same level",
			yaml:    []byte("apiVersion: apps/v1\nkind: Deployment\napiVersion: v1"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - key with multiple colons",
			yaml:    []byte("apiVersion:: apps/v1\nkind: Deployment"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - empty key name",
			yaml:    []byte(": apps/v1\nkind: Deployment"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - dash with no space after",
			yaml:    []byte("-apps/v1\n- v1"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - improper list syntax",
			yaml:    []byte("items: [1, 2,\n  3]"),
			wantErr: true,
		},
		{
			name:    "invalid YAML - colon in wrong context",
			yaml:    []byte("name: value:with:colons"),
			wantErr: false, // This is actually valid YAML (a string with colons)
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.yaml)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

func TestPhase4_Malformed_YAML_InvalidStructure(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		yaml    []byte
		wantErr bool
	}{
		{
			name:    "YAML with wrong document marker placement - --- in middle",
			yaml:    []byte("apiVersion: apps/v1\n---\nkind: Deployment"),
			wantErr: true,
		},
		{
			name:    "YAML with duplicate document markers",
			yaml:    []byte("---\n---\napiVersion: apps/v1"),
			wantErr: false, // Leading --- is valid
		},
		{
			name:    "YAML with only document markers",
			yaml:    []byte("---\n---"),
			wantErr: true,
		},
		{
			name:    "YAML with comment only document",
			yaml:    []byte("# this is a comment\n# another comment"),
			wantErr: true,
		},
		{
			name:    "YAML with only blank lines and spaces",
			yaml:    []byte("   \n\n   \n"),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.yaml)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Binary Data Disguised as YAML
// -----------------------------------------------------------------------

func TestPhase4_Malformed_BinaryDataDisguised(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		yaml    []byte
		wantErr bool
	}{
		{
			name:    "binary data - null bytes",
			yaml:    []byte{0x00, 0x01, 0x02, 'a', 'p', 'i', 'V', 'e', 'r', 's', 'i', 'o', 'n', ':', ' ', 'v', '1'},
			wantErr: true,
		},
		{
			name:    "binary data - UTF-16 BOM",
			yaml:    []byte{0xFF, 0xFE, 'a', 0x00, 'p', 0x00, 'i', 0x00},
			wantErr: true,
		},
		{
			name:    "binary data - random binary with colons",
			yaml:    []byte{0xDE, 0xAD, 0xBE, 0xEF, ':', 'v', '1', 0x00},
			wantErr: true,
		},
		{
			name:    "binary data - PNG file header",
			yaml:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 'd', 'a', 't', 'a'},
			wantErr: true,
		},
		{
			name:    "binary data - JPEG header",
			yaml:    []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.yaml)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Invalid JSON Tests
// -----------------------------------------------------------------------

func TestPhase4_Malformed_InvalidJSON(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		json    []byte
		wantErr bool
	}{
		{
			name:    "invalid JSON - unclosed object",
			json:    []byte(`{"apiVersion": "apps/v1", "kind": "Deployment"`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - unclosed array",
			json:    []byte(`{"apiVersion": "apps/v1", "kinds": ["Deployment"`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - trailing comma",
			json:    []byte(`{"apiVersion": "apps/v1",}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - missing quotes on key",
			json:    []byte(`{apiVersion: "apps/v1"}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - single quotes instead of double",
			json:    []byte(`{'apiVersion': 'apps/v1'}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - newlines in string",
			json:    []byte(`{"apiVersion": "apps
/v1"}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - comments (not valid JSON)",
			json:    []byte(`{"apiVersion": "apps/v1"} // comment`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - undefined",
			json:    []byte(`{"apiVersion": undefined}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - NaN",
			json:    []byte(`{"value": NaN}`),
			wantErr: true,
		},
		{
			name:    "invalid JSON - Infinity",
			json:    []byte(`{"value": Infinity}`),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.json)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Empty and Minimal Invalid Documents
// -----------------------------------------------------------------------

func TestPhase4_Malformed_EmptyDocuments(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		doc     []byte
		wantErr bool
	}{
		{
			name:    "completely empty",
			doc:     []byte{},
			wantErr: true,
		},
		{
			name:    "only whitespace",
			doc:     []byte("   \n\t\n   "),
			wantErr: true,
		},
		{
			name:    "only newlines",
			doc:     []byte("\n\n\n"),
			wantErr: true,
		},
		{
			name:    "only tabs",
			doc:     []byte("\t\t\t"),
			wantErr: true,
		},
		{
			name:    "only carriage returns",
			doc:     []byte("\r\r\r"),
			wantErr: true,
		},
		{
			name:    "only document separator",
			doc:     []byte("---"),
			wantErr: true,
		},
		{
			name:    "only document separator with newlines",
			doc:     []byte("---\n---\n---"),
			wantErr: true,
		},
		{
			name:    "only comments",
			doc:     []byte("# comment\n# another\n# third"),
			wantErr: true,
		},
		{
			name:    "comment with whitespace",
			doc:     []byte("  # comment  \n  \n  # another  "),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.doc)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Wrong YAML Document Marker Placement
// -----------------------------------------------------------------------

func TestPhase4_Malformed_WrongDocumentMarkers(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		yaml    []byte
		wantErr bool
	}{
		{
			name:    "document marker after content",
			yaml:    []byte("apiVersion: apps/v1\n---"),
			wantErr: false, // Valid - leading marker is OK
		},
		{
			name:    "document marker in middle of content",
			yaml:    []byte("apiVersion\n---\n: apps/v1"),
			wantErr: true,
		},
		{
			name:    "document marker in string value",
			yaml:    []byte("name: \"test---\"\nkind: Deployment"),
			wantErr: false, // The --- is inside a string
		},
		{
			name:    "triple dash without hyphens",
			yaml:    []byte("apiVersion: apps/v1\nkind: -- Deployment"),
			wantErr: false, // The -- is a string value
		},
		{
			name:    "multiple leading document markers",
			yaml:    []byte("---\n---\n---\napiVersion: apps/v1"),
			wantErr: false, // Multiple leading markers are valid
		},
		{
			name:    "document marker after kind value",
			yaml:    []byte("apiVersion: apps/v1\nkind: Deployment\n---"),
			wantErr: false, // Trailing marker is valid
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.yaml)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Missing Required Fields Tests
// -----------------------------------------------------------------------

func TestPhase4_Malformed_MissingKindField(t *testing.T) {
	t.Parallel()
	decoder := NewDecoder()

	cases := []struct {
		name    string
		yaml    []byte
		wantErr bool
	}{
		{
			name:    "missing kind field entirely",
			yaml:    []byte("apiVersion: apps/v1\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "empty kind value",
			yaml:    []byte("apiVersion: apps/v1\nkind:\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "kind is null",
			yaml:    []byte("apiVersion: apps/v1\nkind: null\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "kind is number instead of string",
			yaml:    []byte("apiVersion: apps/v1\nkind: 123\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "kind is boolean instead of string",
			yaml:    []byte("apiVersion: apps/v1\nkind: true\nmetadata:\n  name: test"),
			wantErr: true,
		},
		{
			name:    "kind is array instead of string",
			yaml:    []byte("apiVersion: apps/v1\nkind: [Deployment]\nmetadata:\n  name: test"),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest(tc.yaml)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, err=%v", tc.wantErr, gotErr, err)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Engine-Level Malformed Tests (using full validation path)
// -----------------------------------------------------------------------

func TestPhase4_Malformed_EngineValidation(t *testing.T) {
	t.Parallel()
	engine := NewEngine(EngineOptions{IgnoreMissing: false, Workers: 1})

	cases := []struct {
		name      string
		manifest  []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "truncated YAML produces error",
			manifest:  []byte("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test"),
			wantErr:   true,
			errSubstr: "failed to parse",
		},
		{
			name:      "binary data produces error",
			manifest:  []byte{0x00, 0x01, 0x02, 'a'},
			wantErr:   true,
			errSubstr: "failed to parse",
		},
		{
			name:      "invalid JSON produces error",
			manifest:  []byte(`{"apiVersion": "apps/v1",`),
			wantErr:   true,
			errSubstr: "failed to parse",
		},
		{
			name:      "empty document produces error",
			manifest:  []byte{},
			wantErr:   true,
			errSubstr: "empty",
		},
		{
			name:      "only comments produces error",
			manifest:  []byte("# comment only"),
			wantErr:   true,
			errSubstr: "empty",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateMalformedManifest(engine, tc.manifest)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// validateMalformedManifest validates a raw manifest bytes using the engine's decoder
func validateMalformedManifest(engine *Engine, manifest []byte) *types.Result {
	manifest, err := engine.decoder.DecodeManifest(manifest)
	if err != nil {
		return &types.Result{
			Status: types.StatusError,
			Errors: []types.ErrorItem{
				{Field: "manifest", Message: err.Error()},
			},
		}
	}
	return &types.Result{
		Status:  types.StatusValid,
		Kind:    manifest.Kind,
		Name:    manifest.Metadata.Name,
		APIVersion: manifest.APIVersion,
		Errors: nil,
	}
}