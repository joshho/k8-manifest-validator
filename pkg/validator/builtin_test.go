package validator

import (
	"testing"
)

func TestBuiltInKindRoutingDeploymentReturnsValidator(t *testing.T) {
	manifestYAML := []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
`)

	bv := NewBuiltinValidator()

	if !bv.CanValidate("Deployment") {
		t.Errorf("CanValidate(%q) = false, want true", "Deployment")
	}

	result := bv.ValidateResource(manifestYAML)
	if result == nil {
		t.Errorf("ValidateResource() returned nil, want non-nil Result")
		return
	}

	if result.Status == "error" {
		t.Errorf("ValidateResource() status = error, want valid or invalid, errors: %v", result.Errors)
	}
}

func TestBuiltInKindRoutingUnknownKindReturnsError(t *testing.T) {
	manifestYAML := []byte(`apiVersion: v1
kind: UnknownKind
metadata:
  name: test
`)

	bv := NewBuiltinValidator()

	if bv.CanValidate("UnknownKind") {
		t.Errorf("CanValidate(%q) = true, want false", "UnknownKind")
	}

	result := bv.ValidateResource(manifestYAML)
	if result == nil {
		t.Errorf("ValidateResource() returned nil, want non-nil Result")
		return
	}

	if result.Status != "skipped" {
		t.Errorf("ValidateResource() status = %q, want %q for unknown kind", result.Status, "skipped")
	}
}

func TestBuiltInKindRoutingAllKnownKinds(t *testing.T) {
	bv := NewBuiltinValidator()

	kinds := bv.GetRegisteredKinds()
	if len(kinds) == 0 {
		t.Errorf("GetRegisteredKinds() returned empty, want at least some kinds")
	}

	// Test that all registered kinds return non-nil validator from CanValidate
	for _, kind := range kinds {
		if !bv.CanValidate(kind) {
			t.Errorf("CanValidate(%q) = false, want true for registered kind", kind)
		}
	}

	// Verify specific known kinds are registered
	knownKinds := []string{
		"Deployment", "StatefulSet", "DaemonSet", "ReplicaSet",
		"Pod", "Service", "ConfigMap", "Secret", "PersistentVolumeClaim",
		"Namespace", "ServiceAccount", "Endpoints",
		"Ingress", "NetworkPolicy",
		"Job", "CronJob",
		"Role", "ClusterRole", "RoleBinding", "ClusterRoleBinding",
	}

	for _, kind := range knownKinds {
		if !bv.CanValidate(kind) {
			t.Errorf("CanValidate(%q) = false, want true", kind)
		}
	}
}

func TestBuiltinValidatorValidatesDeployment(t *testing.T) {
	// Valid deployment
	validYAML := []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("ValidateResource() status = %q, want %q for valid deployment, errors: %v", result.Status, "valid", result.Errors)
	}
}

func TestBuiltinValidatorDetectsInvalidDeployment(t *testing.T) {
	// Invalid deployment - empty name
	invalidYAML := []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: ""
spec:
  replicas: -1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(invalidYAML)
	if result.Status == "valid" {
		t.Errorf("ValidateResource() status = %q for invalid deployment, want invalid or error", result.Status)
	}
}

func TestBuiltinValidatorValidatesConfigMap(t *testing.T) {
	validYAML := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  key: value
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("ValidateResource() status = %q, want %q for valid ConfigMap", result.Status, "valid")
	}
}

func TestBuiltinValidatorValidatesService(t *testing.T) {
	validYAML := []byte(`apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
spec:
  selector:
    app: my-app
  ports:
  - port: 80
    targetPort: 8080
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("ValidateResource() status = %q, want %q for valid Service", result.Status, "valid")
	}
}

func TestBuiltinValidatorValidatesIngress(t *testing.T) {
	validYAML := []byte(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("ValidateResource() status = %q, want %q for valid Ingress", result.Status, "valid")
	}
}

func TestBuiltinValidatorValidatesJob(t *testing.T) {
	validYAML := []byte(`apiVersion: batch/v1
kind: Job
metadata:
  name: my-job
  namespace: default
spec:
  parallelism: 1
  completions: 1
  backoffLimit: 6
  template:
    spec:
      containers:
      - name: job
        image: busybox
        command: ["echo", "hello"]
      restartPolicy: OnFailure
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("ValidateResource() status = %q, want %q for valid Job", result.Status, "valid")
	}
}