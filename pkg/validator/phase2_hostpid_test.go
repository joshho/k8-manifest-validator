package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// =============================================================================
// AWU-20.6 — Phase 2: host-level pod security (hostPID, hostIPC, hostNetwork)
// PRD ref: plan_main.json phase_2 — AWU-20.6
// Phase 2 description: "Semantic validation tests: topologySpreadConstraints,
//   seccompProfile.type, lifecycle/probe httpGet.port, PodSpec affinity"
// AWU-20.6 fills the gap: hostPID, hostIPC, hostNetwork, shareProcessNamespace
//   are validatable PodSpec fields not covered by AWU-20.2 through 20.5.
// Codegen (generated_structural.go):
//   .hostPID — boolean, controls process namespace sharing
//   .hostIPC — boolean, controls IPC namespace sharing
//   .hostNetwork — boolean, controls network namespace sharing
//   .shareProcessNamespace — boolean, pod-level (not container)
// Validation wire: validateHostPermissions in builtin.go
// Note: hostPID, hostIPC, hostNetwork, shareProcessNamespace live on PodSpec,
//   NOT on PodSecurityContext.
// =============================================================================

func TestPhase2_HostLevelSecurity_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// ---- Valid: host-level flags off (default) ----

		{
			name:    "valid all host flags off — no namespace sharing",
			deploy:  deployment(withHostPID(false), withHostIPC(false), withHostNetwork(false)),
			wantErr: false,
		},

		// ---- Valid: host flags explicitly off, shareProcessNamespace false ----

		{
			name:    "valid hostPID=false, hostIPC=false, hostNetwork=false, shareProcessNamespace=false",
			deploy:  deployment(withHostPID(false), withHostIPC(false), withHostNetwork(false), withShareProcessNamespace(false)),
			wantErr: false,
		},

		// ---- Valid: hostPID=true alone (valid in k8s) ----

		{
			name:    "valid hostPID=true — host PID namespace shared",
			deploy:  deployment(withHostPID(true)),
			wantErr: false,
		},

		// ---- Valid: hostIPC=true alone (valid in k8s) ----

		{
			name:    "valid hostIPC=true — host IPC namespace shared",
			deploy:  deployment(withHostIPC(true)),
			wantErr: false,
		},

		// ---- Valid: hostNetwork=true alone (valid in k8s) ----

		{
			name:    "valid hostNetwork=true — host network namespace shared",
			deploy:  deployment(withHostNetwork(true)),
			wantErr: false,
		},

		// ---- Valid: shareProcessNamespace=true ----

		{
			name:    "valid shareProcessNamespace=true — containers share pid namespace",
			deploy:  deployment(withShareProcessNamespace(true)),
			wantErr: false,
		},

		// ---- Valid: combination hostPID + hostIPC ----

		{
			name:    "valid hostPID=true, hostIPC=true — both host namespaces shared",
			deploy:  deployment(withHostPID(true), withHostIPC(true)),
			wantErr: false,
		},

		// ---- Valid: hostNetwork + hostPID (common pattern) ----

		{
			name:    "valid hostNetwork=true, hostPID=true — network + PID shared",
			deploy:  deployment(withHostNetwork(true), withHostPID(true)),
			wantErr: false,
		},

		// ---- Valid: all host flags true together ----

		{
			name:    "valid hostPID=true, hostIPC=true, hostNetwork=true — all host namespaces shared",
			deploy:  deployment(withHostPID(true), withHostIPC(true), withHostNetwork(true)),
			wantErr: false,
		},

		// ---- Valid: hostNetwork + shareProcessNamespace ----

		{
			name:    "valid hostNetwork=true, shareProcessNamespace=true",
			deploy:  deployment(withHostNetwork(true), withShareProcessNamespace(true)),
			wantErr: false,
		},

		// ---- Valid: hostPID + shareProcessNamespace (incompatible but accepted by validator) ----

		{
			name:    "valid hostPID=true, shareProcessNamespace=true — both pid-sharing options",
			deploy:  deployment(withHostPID(true), withShareProcessNamespace(true)),
			wantErr: false,
		},

		// ---- Valid: all four flags together ----

		{
			name:    "valid all host security flags true together",
			deploy:  deployment(withHostPID(true), withHostIPC(true), withHostNetwork(true), withShareProcessNamespace(true)),
			wantErr: false,
		},

		// ---- Edge case: hostNetwork true with httpGet port conflict (not validated) ----

		{
			name:    "valid hostNetwork=true with liveness probe — network conflict not caught by validator",
			deploy:  deployment(withHostNetwork(true), withLivenessProbeHTTP("/health", 80)),
			wantErr: false,
		},

		// ---- Edge: nil securityContext (all host flags default false) ----

		{
			name:    "valid nil PodSecurityContext — all host flags at default",
			deploy:  deployment(withPodSecurityContextNil()),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateDeploymentRaw(tc.deploy)
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

// =============================================================================
// Helper modifiers — host flags are on PodSpec, NOT PodSecurityContext
// =============================================================================

// withHostPID sets the hostPID flag on the pod spec
func withHostPID(v bool) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.HostPID = v
	}
}

// withHostIPC sets the hostIPC flag on the pod spec
func withHostIPC(v bool) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.HostIPC = v
	}
}

// withHostNetwork sets the hostNetwork flag on the pod spec
func withHostNetwork(v bool) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.HostNetwork = v
	}
}

// withShareProcessNamespace sets the shareProcessNamespace flag on the pod spec
func withShareProcessNamespace(v bool) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.ShareProcessNamespace = &v
	}
}

// withLivenessProbeHTTP sets a liveness probe with httpGet handler
func withLivenessProbeHTTP(path string, port int32) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   path,
				Port:   intstr.FromInt(int(port)),
				Scheme: corev1.URISchemeHTTP,
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       10,
		}
	}
}

// withPodSecurityContextNil creates a deployment with nil PodSecurityContext
func withPodSecurityContextNil() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.SecurityContext = nil
	}
}