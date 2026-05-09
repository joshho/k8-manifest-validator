package validator

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// =============================================================================
// AWU-12.1 — Semantic Layer Tests: lifecycle hooks, initContainer probes, restartPolicy
// =============================================================================

func TestValidateInitContainerProbesRejectsLivenessProbe(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   *corev1.Probe
		expectError bool
	}{
		{
			name:        "liveness probe on init container",
			initProbe:   &corev1.Probe{},
			expectError: true,
		},
		{
			name:        "no probe",
			initProbe:   nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podSpec := &corev1.PodSpec{
				InitContainers: []corev1.Container{
					{
						Name:          "init-container",
						LivenessProbe: tt.initProbe,
					},
				},
			}
			path := field.NewPath("spec")
			errs := validateInitContainerProbes(podSpec, path)
			if tt.expectError && len(errs) == 0 {
				t.Errorf("validateInitContainerProbes() expected error, got nil")
			}
			if !tt.expectError && len(errs) > 0 {
				t.Errorf("validateInitContainerProbes() expected no error, got %d errors", len(errs))
			}
		})
	}
}

func TestValidateInitContainerProbesRejectsReadinessProbe(t *testing.T) {
	podSpec := &corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:          "init-container",
				ReadinessProbe: &corev1.Probe{},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerProbes(podSpec, path)
	if len(errs) == 0 {
		t.Errorf("validateInitContainerProbes() expected error for readiness probe, got nil")
	}
}

func TestValidateInitContainerProbesRejectsStartupProbe(t *testing.T) {
	podSpec := &corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:          "init-container",
				StartupProbe: &corev1.Probe{},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerProbes(podSpec, path)
	if len(errs) == 0 {
		t.Errorf("validateInitContainerProbes() expected error for startup probe, got nil")
	}
}

func TestValidateInitContainerProbesMultipleInitContainers(t *testing.T) {
	podSpec := &corev1.PodSpec{
		InitContainers: []corev1.Container{
			{Name: "init-1"},
			{
				Name:          "init-2",
				LivenessProbe: &corev1.Probe{},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerProbes(podSpec, path)
	if len(errs) == 0 {
		t.Errorf("validateInitContainerProbes() expected error, got nil")
	}
}

func TestValidateInitContainerProbesNoInitContainers(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{Name: "main", LivenessProbe: &corev1.Probe{}},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerProbes(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateInitContainerProbes() expected no error for regular containers, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHookHandlerCountZeroHandlers(t *testing.T) {
	hook := &corev1.LifecycleHandler{}
	path := field.NewPath("spec", "lifecycle", "preStop")
	errs := validateLifecycleHookHandlerCount(hook, path)
	if len(errs) == 0 {
		t.Errorf("validateLifecycleHookHandlerCount() expected error for zero handlers, got nil")
	}
}

func TestValidateLifecycleHookHandlerCountMultipleHandlers(t *testing.T) {
	hook := &corev1.LifecycleHandler{
		HTTPGet:  &corev1.HTTPGetAction{},
		Exec:     &corev1.ExecAction{},
	}
	path := field.NewPath("spec", "lifecycle", "preStop")
	errs := validateLifecycleHookHandlerCount(hook, path)
	if len(errs) == 0 {
		t.Errorf("validateLifecycleHookHandlerCount() expected error for multiple handlers, got nil")
	}
}

func TestValidateLifecycleHookHandlerCountOneHandler(t *testing.T) {
	hook := &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: []string{"/bin/sh"}},
	}
	path := field.NewPath("spec", "lifecycle", "preStop")
	errs := validateLifecycleHookHandlerCount(hook, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHookHandlerCount() expected no error for one handler, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHookHandlerCountNilHook(t *testing.T) {
	path := field.NewPath("spec", "lifecycle", "preStop")
	errs := validateLifecycleHookHandlerCount(nil, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHookHandlerCount(nil) expected no error, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHookHandlerCountHTTPGet(t *testing.T) {
	hook := &corev1.LifecycleHandler{
		HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
	}
	path := field.NewPath("spec", "lifecycle", "postStart")
	errs := validateLifecycleHookHandlerCount(hook, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHookHandlerCount(HTTPGet) expected no error, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHookHandlerCountTCPSocket(t *testing.T) {
	hook := &corev1.LifecycleHandler{
		TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(8080)},
	}
	path := field.NewPath("spec", "lifecycle", "preStop")
	errs := validateLifecycleHookHandlerCount(hook, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHookHandlerCount(TCPSocket) expected no error, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHooksRegularContainer(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "main",
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "sleep 5"}},
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateLifecycleHooks(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHooks() expected no error for valid lifecycle, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHooksMultipleContainers(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{Name: "main"},
			{
				Name: "sidecar",
				Lifecycle: &corev1.Lifecycle{
					PostStart: &corev1.LifecycleHandler{
						HTTPGet: &corev1.HTTPGetAction{Path: "/start"},
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateLifecycleHooks(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHooks() expected no error, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHooksNoLifecycle(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{Name: "main"},
		},
	}
	path := field.NewPath("spec")
	errs := validateLifecycleHooks(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHooks() expected no error for no lifecycle, got %d errors", len(errs))
	}
}

func TestValidateLifecycleHooksInitContainer(t *testing.T) {
	podSpec := &corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name: "init",
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "cleanup"}},
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateLifecycleHooks(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateLifecycleHooks() expected no error for valid init container lifecycle, got %d errors", len(errs))
	}
}

func TestValidateRestartPolicyForJobRejectsAlways(t *testing.T) {
	podSpec := &corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyAlways,
	}
	path := field.NewPath("spec")
	errs := validateRestartPolicyForJob(podSpec, path)
	if len(errs) == 0 {
		t.Errorf("validateRestartPolicyForJob() expected error for RestartPolicyAlways, got nil")
	}
}

func TestValidateRestartPolicyForJobAcceptsNever(t *testing.T) {
	podSpec := &corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
	}
	path := field.NewPath("spec")
	errs := validateRestartPolicyForJob(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateRestartPolicyForJob() expected no error for RestartPolicyNever, got %d errors", len(errs))
	}
}

func TestValidateRestartPolicyForJobAcceptsOnFailure(t *testing.T) {
	podSpec := &corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyOnFailure,
	}
	path := field.NewPath("spec")
	errs := validateRestartPolicyForJob(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateRestartPolicyForJob() expected no error for RestartPolicyOnFailure, got %d errors", len(errs))
	}
}

func TestValidateInitContainerResourcesNoInitContainers(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "main",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("100m"),
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerResources(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateInitContainerResources() expected no error when no init containers, got %d errors", len(errs))
	}
}

func TestValidateInitContainerResourcesNoRegularContainers(t *testing.T) {
	podSpec := &corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name: "init",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("100m"),
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerResources(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateInitContainerResources() expected no error when no regular containers, got %d errors", len(errs))
	}
}

func TestValidateInitContainerResourcesExceedsMaxCPU(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "main",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("100m"),
					},
				},
			},
		},
		InitContainers: []corev1.Container{
			{
				Name: "init",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("200m"),
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerResources(podSpec, path)
	if len(errs) == 0 {
		t.Errorf("validateInitContainerResources() expected error when init exceeds container max, got nil")
	}
}

func TestValidateInitContainerResourcesWithinLimit(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "main",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("100m"),
					},
				},
			},
		},
		InitContainers: []corev1.Container{
			{
				Name: "init",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("50m"),
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerResources(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateInitContainerResources() expected no error when init within limit, got %d errors", len(errs))
	}
}

func TestValidateInitContainerResourcesMultipleContainers(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "main",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("100m"),
						corev1.ResourceMemory: *parseQuantity("128Mi"),
					},
				},
			},
			{
				Name: "sidecar",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: *parseQuantity("50m"),
						corev1.ResourceMemory: *parseQuantity("64Mi"),
					},
				},
			},
		},
		InitContainers: []corev1.Container{
			{
				Name: "init",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						// init container uses max CPU from main (100m), but max memory from sidecar (128Mi)
						corev1.ResourceCPU:    *parseQuantity("100m"),
						corev1.ResourceMemory: *parseQuantity("64Mi"),
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerResources(podSpec, path)
	if len(errs) > 0 {
		t.Errorf("validateInitContainerResources() expected no error when init within multi-container max, got %d errors", len(errs))
	}
}

func TestValidateInitContainerResourcesInitExceedsMemory(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "main",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: *parseQuantity("128Mi"),
					},
				},
			},
		},
		InitContainers: []corev1.Container{
			{
				Name: "init",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: *parseQuantity("256Mi"),
					},
				},
			},
		},
	}
	path := field.NewPath("spec")
	errs := validateInitContainerResources(podSpec, path)
	if len(errs) == 0 {
		t.Errorf("validateInitContainerResources() expected error for memory exceed, got nil")
	}
}

// parseQuantity is a test helper to parse resource quantities.
func parseQuantity(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}