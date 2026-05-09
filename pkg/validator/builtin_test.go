package validator

import (
	"testing"
)

// Helper to verify the test file compiles
var _ = NewBuiltinValidator

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

// --- AWU-1: TDD failing tests ---

// TestBuiltinValidatorDetectsInvalidContainerNameInDeployment verifies that
// a Deployment with a container name containing uppercase characters is rejected.
// This is the AWU-1 TDD test: it currently FAILS because validateDeployment
// does not call validatePodSpec yet.
func TestBuiltinValidatorDetectsInvalidContainerNameInDeployment(t *testing.T) {
	invalidYAML := []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: Bad_Name
        image: nginx
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(invalidYAML)
	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s: %v", result.Status, result.Errors)
	}
	if len(result.Errors) > 0 && !containsField(result.Errors[0].Field, "containers[0].name") {
		t.Errorf("expected error on containers[0].name, got %s", result.Errors[0].Field)
	}
}

// TestBuiltinValidatorDetectsDuplicateContainerNameInStatefulSet verifies that
// a StatefulSet with duplicate container names is rejected.
func TestBuiltinValidatorDetectsDuplicateContainerNameInStatefulSet(t *testing.T) {
	invalidYAML := []byte(`apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: test-ss
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
      - name: nginx
        image: nginx:1.22
  serviceName: test
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(invalidYAML)
	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s: %v", result.Status, result.Errors)
	}
}

// TestBuiltinValidatorDetectsEmptyContainerImageInDaemonSet verifies that
// a DaemonSet with an empty container image string is rejected.
func TestBuiltinValidatorDetectsEmptyContainerImageInDaemonSet(t *testing.T) {
	invalidYAML := []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: test-ds
  namespace: default
spec:
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: app
        image: ""
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(invalidYAML)
	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s: %v", result.Status, result.Errors)
	}
}

// TestBuiltinValidatorAcceptsValidDeployment verifies a valid Deployment passes.
func TestBuiltinValidatorAcceptsValidDeployment(t *testing.T) {
	validYAML := []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: valid-deploy
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: valid
  template:
    metadata:
      labels:
        app: valid
    spec:
      containers:
      - name: app
        image: app:1.0
        ports:
        - containerPort: 8080
        env:
        - name: FOO
          value: bar
        resources:
          limits:
            cpu: "1"
            memory: 512Mi
          requests:
            cpu: 100m
            memory: 128Mi
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("expected valid, got %s: %v", result.Status, result.Errors)
	}
}

// TestBuiltinValidatorAcceptsValidStatefulSet verifies a valid StatefulSet passes.
func TestBuiltinValidatorAcceptsValidStatefulSet(t *testing.T) {
	validYAML := []byte(`apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: valid-ss
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: valid
  template:
    metadata:
      labels:
        app: valid
    spec:
      containers:
      - name: app
        image: app:1.0
        ports:
        - containerPort: 8080
  serviceName: valid
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("expected valid, got %s: %v", result.Status, result.Errors)
	}
}

// TestBuiltinValidatorAcceptsValidDaemonSet verifies a valid DaemonSet passes.
func TestBuiltinValidatorAcceptsValidDaemonSet(t *testing.T) {
	validYAML := []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: valid-ds
  namespace: default
spec:
  selector:
    matchLabels:
      app: valid
  template:
    metadata:
      labels:
        app: valid
    spec:
      containers:
      - name: app
        image: app:1.0
        ports:
        - containerPort: 8080
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("expected valid, got %s: %v", result.Status, result.Errors)
	}
}

// --- AWU-2: TDD failing tests for Pod and ReplicaSet ---

// TestBuiltinValidatorDetectsInvalidServiceAccountNameInPod verifies that
// a Pod with an invalid serviceAccountName containing underscores is rejected.
func TestBuiltinValidatorDetectsInvalidServiceAccountNameInPod(t *testing.T) {
	invalidYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
spec:
  serviceAccountName: invalid_account
  containers:
  - name: app
    image: nginx
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(invalidYAML)
	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s: %v", result.Status, result.Errors)
	}
	if len(result.Errors) > 0 && !containsField(result.Errors[0].Field, "serviceAccountName") {
		t.Errorf("expected error on spec.serviceAccountName, got %s", result.Errors[0].Field)
	}
}

// TestBuiltinValidatorDetectsInvalidContainerPortInReplicaSet verifies that
// a ReplicaSet with an out-of-range container port is rejected.
func TestBuiltinValidatorDetectsInvalidContainerPortInReplicaSet(t *testing.T) {
	invalidYAML := []byte(`apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: test-rs
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: app
        image: nginx
        ports:
        - containerPort: 70000
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(invalidYAML)
	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s: %v", result.Status, result.Errors)
	}
	if len(result.Errors) > 0 && !containsField(result.Errors[0].Field, "containers[0].ports[0].containerPort") {
		t.Errorf("expected error on containers[0].ports[0].containerPort, got %s", result.Errors[0].Field)
	}
}

// TestBuiltinValidatorAcceptsValidPod verifies a valid Pod passes validation.
func TestBuiltinValidatorAcceptsValidPod(t *testing.T) {
	validYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: valid-pod
  namespace: default
spec:
  containers:
  - name: app
    image: nginx:1.21
    ports:
    - containerPort: 8080
    env:
    - name: FOO
      value: bar
    resources:
      limits:
        cpu: "1"
        memory: 512Mi
      requests:
        cpu: 100m
        memory: 128Mi
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("expected valid, got %s: %v", result.Status, result.Errors)
	}
}

// TestBuiltinValidatorAcceptsValidReplicaSet verifies a valid ReplicaSet passes validation.
func TestBuiltinValidatorAcceptsValidReplicaSet(t *testing.T) {
	validYAML := []byte(`apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: valid-rs
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: valid
  template:
    metadata:
      labels:
        app: valid
    spec:
      containers:
      - name: app
        image: nginx:1.21
        ports:
        - containerPort: 8080
`)

	bv := NewBuiltinValidator()
	result := bv.ValidateResource(validYAML)
	if result.Status != "valid" {
		t.Errorf("expected valid, got %s: %v", result.Status, result.Errors)
	}
}

// containsField checks if the field string contains the given substring.
func containsField(field, substr string) bool {
	return len(field) >= len(substr) && field[len(field)-len(substr):] == substr
}
