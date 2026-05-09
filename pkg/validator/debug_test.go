package validator

import (
	"reflect"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestDebugWalkStruct(t *testing.T) {
	manifestYAML := []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-structural
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: ""
        whenUnsatisfiable: DoNotSchedule
        labelSelector:
          matchLabels:
            app: test
      containers:
      - name: app
        image: nginx:1.21
`)

	bv := NewBuiltinValidator()
	obj, _, err := bv.decode(manifestYAML)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	deploy := obj.(*appsv1.Deployment)
	basePath := field.NewPath("")

	// Ensure init is called
	deferredFieldInit.Do(func() {
		if deferredFieldIndex == nil {
			deferredFieldIndex = make(map[string]FieldMetadata)
		}
	})

	t.Logf("=== Walking Deployment struct ===")
	debugWalkStruct(reflect.ValueOf(deploy).Elem(), basePath, t)
}

func debugWalkStruct(val reflect.Value, path *field.Path, t *testing.T) {
	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		fieldVal := val.Field(i)

		if !sf.IsExported() {
			continue
		}

		lower := strings.ToLower(sf.Name[:1]) + sf.Name[1:]
		fieldPath := path.Key(lower)

		normalized := normalizePath(fieldPath)
		_, found := LookupDeferredField(normalized)

		metaStr := ""
		if found {
			metaStr = " [FOUND]"
		}
		t.Logf("  field=%15s path=%s normalized=%s%s", sf.Name, fieldPath.String(), normalized, metaStr)

		switch fieldVal.Kind() {
		case reflect.Struct:
			debugWalkStruct(fieldVal, fieldPath, t)
		case reflect.Ptr:
			if !fieldVal.IsNil() && fieldVal.Elem().Kind() == reflect.Struct {
				debugWalkStruct(fieldVal.Elem(), fieldPath, t)
			}
		case reflect.Slice:
			for j := 0; j < fieldVal.Len(); j++ {
				elem := fieldVal.Index(j)
				idxPath := fieldPath.Index(j)
				t.Logf("    [%d] path=%s", j, idxPath.String())
				if elem.Kind() == reflect.Struct {
					debugWalkStruct(elem, idxPath, t)
				}
			}
		}
	}
}

func TestDebugTopologyKeyLookup(t *testing.T) {
	basePath := field.NewPath("")

	// Simulate what walkStruct would do for topologySpreadConstraints[0].topologyKey
	p1 := basePath.Key("spec").Key("template").Key("spec").Key("topologySpreadConstraints").Index(0).Key("topologyKey")
	t.Logf("Manual path: %s", p1.String())
	t.Logf("Normalized:  %s", normalizePath(p1))

	meta, found := LookupDeferredField(normalizePath(p1))
	t.Logf("Lookup result: found=%v meta=%+v", found, meta)

	// Index dump
	t.Logf("=== deferredFieldIndex keys ===")
	for k, v := range deferredFieldIndex {
		t.Logf("  %q: type=%s required=%v", k, v.Type, v.Required)
	}
}

func TestDebugTopologyKeyEmpty(t *testing.T) {
	// Build a Deployment with empty topologyKey programmatically (no YAML decode)
	deploy := &appsv1.Deployment{}
	deploy.Name = "test"
	r := int32(3)
	deploy.Spec.Replicas = &r

	t.Logf("Running RunStructuralValidation directly on Deployment with empty TopologySpreadConstraints...")
	errs := RunStructuralValidation(deploy, field.NewPath(""))
	t.Logf("Errors count: %d", len(errs))
	for _, e := range errs {
		t.Logf("  Error: field=%s detail=%s", e.Field, e.Detail)
	}
}
