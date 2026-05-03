package validator

import (
	"testing"
)

func TestDecoderExtractsKindFromValidManifest(t *testing.T) {
	tests := []struct {
		name       string
		manifest   string
		wantKind   string
		wantAPI    string
		wantName   string
		wantNS     string
		wantErr    bool
	}{
		{
			name: "simple deployment",
			manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
spec:
  replicas: 3`,
			wantKind: "Deployment",
			wantAPI:  "apps/v1",
			wantName: "my-deployment",
			wantNS:   "default",
			wantErr:  false,
		},
		{
			name: "configmap without namespace",
			manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value`,
			wantKind: "ConfigMap",
			wantAPI:  "v1",
			wantName: "my-config",
			wantNS:   "",
			wantErr:  false,
		},
		{
			name: "ingress networking",
			manifest: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: ingress-space
spec:
  rules:
  - host: example.com`,
			wantKind: "Ingress",
			wantAPI:  "networking.k8s.io/v1",
			wantName: "my-ingress",
			wantNS:   "ingress-space",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDecoder()
			manifest, err := decoder.DecodeManifest([]byte(tt.manifest))
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if manifest.Kind != tt.wantKind {
					t.Errorf("Kind = %v, want %v", manifest.Kind, tt.wantKind)
				}
				if manifest.APIVersion != tt.wantAPI {
					t.Errorf("APIVersion = %v, want %v", manifest.APIVersion, tt.wantAPI)
				}
				if manifest.Metadata.Name != tt.wantName {
					t.Errorf("Name = %v, want %v", manifest.Metadata.Name, tt.wantName)
				}
				if manifest.Metadata.Namespace != tt.wantNS {
					t.Errorf("Namespace = %v, want %v", manifest.Metadata.Namespace, tt.wantNS)
				}
			}
		})
	}
}

func TestDecoderHandlesMultiDocYAML(t *testing.T) {
	manifest := `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: first-config
data:
  first: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second-config
data:
  second: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 1`

	docs := SplitYAMLDocument([]byte(manifest))
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}

	decoder := NewDecoder()

	expectedNames := []string{"first-config", "second-config", "my-deployment"}
	expectedKinds := []string{"ConfigMap", "ConfigMap", "Deployment"}

	for i, doc := range docs {
		m, err := decoder.DecodeManifest(doc)
		if err != nil {
			t.Errorf("DecodeManifest() for doc[%d] error = %v", i, err)
			continue
		}
		if m.Metadata.Name != expectedNames[i] {
			t.Errorf("doc[%d] Name = %v, want %v", i, m.Metadata.Name, expectedNames[i])
		}
		if m.Kind != expectedKinds[i] {
			t.Errorf("doc[%d] Kind = %v, want %v", i, m.Kind, expectedKinds[i])
		}
	}
}

func TestDecoderReturnsErrorOnMalformedYAML(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
	}{
		{
			name:     "invalid yaml syntax",
			manifest: `apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test\n  invalid: [`,
		},
		{
			name:     "missing colon after key",
			manifest: `apiVersion: apps/v1\nkind Deployment\nmetadata:\n  name: test`,
		},
		{
			name:     "empty document",
			manifest: ``,
		},
		{
			name:     "only whitespace",
			manifest: `   \n  \t  \n`,
		},
	}

	decoder := NewDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decoder.DecodeManifest([]byte(tt.manifest))
			if err == nil {
				t.Errorf("DecodeManifest() expected error for manifest: %q", tt.manifest)
			}
		})
	}
}

func TestDecoderHandlesJSONInput(t *testing.T) {
	manifest := `{
  "apiVersion": "v1",
  "kind": "ConfigMap",
  "metadata": {
    "name": "json-config",
    "namespace": "default"
  },
  "data": {
    "key": "value"
  }
}`

	decoder := NewDecoder()
	m, err := decoder.DecodeManifest([]byte(manifest))
	if err != nil {
		t.Errorf("DecodeManifest() error = %v", err)
		return
	}
	if m.Kind != "ConfigMap" {
		t.Errorf("Kind = %v, want %v", m.Kind, "ConfigMap")
	}
	if m.APIVersion != "v1" {
		t.Errorf("APIVersion = %v, want %v", m.APIVersion, "v1")
	}
	if m.Metadata.Name != "json-config" {
		t.Errorf("Name = %v, want %v", m.Metadata.Name, "json-config")
	}
	if m.Metadata.Namespace != "default" {
		t.Errorf("Namespace = %v, want %v", m.Metadata.Namespace, "default")
	}
}

func TestDecoderHandlesYAMLEquivalentJSON(t *testing.T) {
	// YAML can represent JSON objects as inline tables
	manifest := `{"apiVersion": "v1", "kind": "Secret", "metadata": {"name": "my-secret", "namespace": "secret-ns"}}`

	decoder := NewDecoder()
	m, err := decoder.DecodeManifest([]byte(manifest))
	if err != nil {
		t.Errorf("DecodeManifest() error = %v", err)
		return
	}
	if m.Kind != "Secret" {
		t.Errorf("Kind = %v, want %v", m.Kind, "Secret")
	}
	if m.Metadata.Name != "my-secret" {
		t.Errorf("Name = %v, want %v", m.Metadata.Name, "my-secret")
	}
	if m.Metadata.Namespace != "secret-ns" {
		t.Errorf("Namespace = %v, want %v", m.Metadata.Namespace, "secret-ns")
	}
}
