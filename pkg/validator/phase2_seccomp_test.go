package validator

import (
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-20.3 — Phase 2: seccompProfile.type expanded edge-condition tests
// PRD ref: plan_main.json phase_2 — AWU-20.3
// Codegen (generated_structural.go):
//   .securityContext.seccompProfile.type — one of "", "Unconfined", "RuntimeDefault", "LocalDefault"
// Validation wire: validateSeccomp in semantic_security.go (called via validateSecurityContext)
// Existing coverage: semantic_security.go has validateSeccomp unit-level coverage
// AWU-20.3 complement: expanded edge conditions not covered by existing tests
// =============================================================================

func TestPhase2_SeccompProfile_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// ---- Valid: nil seccompProfile (skip) ----

		{
			name: "valid nil seccompProfile — no profile set at all",
			deploy: deployment(withContainerSecurityContextNil()),
			wantErr: false,
		},
		{
			name: "valid nil seccompProfile — securityContext exists but no seccompProfile field",
			deploy: deployment(withContainerSecurityContextNoSeccomp()),
			wantErr: false,
		},

		// ---- Valid: empty seccompProfile (type not emitted in YAML — zero value) ----

		{
			name: "invalid seccompProfile.type=zero — empty struct is not emitted by YAML marshaler but IS a zero-value in Go (errors)",
			deploy: deployment(withSeccompProfileEmpty()),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},

		// ---- Valid: known enum values ----

		{
			name: "valid seccompProfile.type=RuntimeDefault",
			deploy: deployment(withSeccompProfileType("RuntimeDefault")),
			wantErr: false,
		},
		{
			name: "valid seccompProfile.type=Unconfined",
			deploy: deployment(withSeccompProfileType("Unconfined")),
			wantErr: false,
		},
		{
			name: "valid seccompProfile.type=Localhost",
			deploy: deployment(withSeccompProfileTypeLocalhost("profile.json")),
			wantErr: false,
		},

		// ---- Invalid: non-enum string values ----

		{
			name: "invalid seccompProfile.type=LocalDefault — not a recognised value",
			deploy: deployment(withSeccompProfileType("LocalDefault")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},
		{
			name: "invalid seccompProfile.type=UnknownValue",
			deploy: deployment(withSeccompProfileType("UnknownValue")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},
		{
			name: "invalid seccompProfile.type=runtime-default (wrong casing)",
			deploy: deployment(withSeccompProfileType("runtime-default")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},
		{
			name: "invalid seccompProfile.type=UNCONFINED (all caps)",
			deploy: deployment(withSeccompProfileType("UNCONFINED")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},

		// ---- Invalid: whitespace ----

		{
			name: "invalid seccompProfile.type=Runtime Default (space in value)",
			deploy: deployment(withSeccompProfileType("Runtime Default")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},
		{
			name: "invalid seccompProfile.type=   (only spaces)",
			deploy: deployment(withSeccompProfileType("   ")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},

		// ---- Invalid: too long (>128 chars) ----

		{
			name: "invalid seccompProfile.type length=129 chars — not in valid set",
			deploy: deployment(withSeccompProfileType(strings.Repeat("a", 129))),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},
		{
			name: "invalid seccompProfile.type length=200 chars — not in valid set",
			deploy: deployment(withSeccompProfileType(strings.Repeat("x", 200))),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},

		// ---- Valid: valid non-empty strings within enum ----

		{
			name: "valid seccompProfile.type length=128 chars — not in valid set",
			deploy: deployment(withSeccompProfileType(strings.Repeat("a", 128))),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},

		// ---- Valid: Localhost with localhostProfile ----

		{
			name: "valid Localhost with simple profile path",
			deploy: deployment(withSeccompProfileTypeLocalhost("profiles/default.json")),
			wantErr: false,
		},
		{
			name: "valid Localhost with empty localhostProfile — actually should error",
			deploy: deployment(withSeccompProfileTypeLocalhost("")),
			wantErr:    true,
			errSubstr:  "localhostProfile",
		},

		// ---- Pod-level securityContext seccompProfile (container-level only in validateSeccomp) ----

		{
			name: "valid pod-level seccompProfile.type=RuntimeDefault",
			deploy: deployment(withPodSeccompProfileType("RuntimeDefault")),
			wantErr: false,
		},
		{
			name: "pod-level seccompProfile.type=BadValue — seccomp validation skipped for pod-level",
			deploy: deployment(withPodSeccompProfileType("BadValue")),
			wantErr: false,
		},

		// ---- Multiple containers: mixed valid/invalid ----

		{
			name: "invalid second container has bad seccompProfile.type",
			deploy: deployment(withSecondContainerSeccompProfileType("BadValue")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},
		{
			name: "valid second container has RuntimeDefault",
			deploy: deployment(withSecondContainerSeccompProfileType("RuntimeDefault")),
			wantErr: false,
		},

		// ---- Init container seccompProfile ----

		{
			name: "valid init container seccompProfile.type=Unconfined",
			deploy: deployment(withInitContainerSeccompProfileType("Unconfined")),
			wantErr: false,
		},
		{
			name: "invalid init container seccompProfile.type=InvalidType",
			deploy: deployment(withInitContainerSeccompProfileType("InvalidType")),
			wantErr:    true,
			errSubstr:  "seccompProfile",
		},

		// ---- Init container Localhost ----

		{
			name: "valid init container Localhost profile with path",
			deploy: deployment(withInitContainerSeccompProfileTypeLocalhost("profile.json")),
			wantErr: false,
		},
		{
			name: "invalid init container Localhost with empty localhostProfile",
			deploy: deployment(withInitContainerSeccompProfileTypeLocalhost("")),
			wantErr:    true,
			errSubstr:  "localhostProfile",
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
// Deployment modifier helpers for seccompProfile
// =============================================================================

// withContainerSecurityContextNil sets container securityContext to nil
func withContainerSecurityContextNil() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].SecurityContext = nil
	}
}

// withContainerSecurityContextNoSeccomp sets securityContext without seccompProfile
func withContainerSecurityContextNoSeccomp() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
	}
}

// withSeccompProfileEmpty sets seccompProfile with type=zero (empty struct, not emitted in YAML)
func withSeccompProfileEmpty() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		// SeccompProfile with zero-value Type — YAML marshaler omits the field entirely
		d.Spec.Template.Spec.Containers[0].SecurityContext.SeccompProfile = &corev1.SeccompProfile{}
	}
}

// withSeccompProfileType sets seccompProfile.type on container[0]
func withSeccompProfileType(t string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SeccompProfile = &corev1.SeccompProfile{
			Type: corev1.SeccompProfileType(t),
		}
	}
}

// withSeccompProfileTypeLocalhost sets seccompProfile.type=Localhost with localhostProfile
func withSeccompProfileTypeLocalhost(profile string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SeccompProfile = &corev1.SeccompProfile{
			Type:             corev1.SeccompProfileTypeLocalhost,
			LocalhostProfile: &profile,
		}
	}
}

// withPodSeccompProfileType sets seccompProfile on pod-level securityContext
func withPodSeccompProfileType(t string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.SecurityContext == nil {
			d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		d.Spec.Template.Spec.SecurityContext.SeccompProfile = &corev1.SeccompProfile{
			Type: corev1.SeccompProfileType(t),
		}
	}
}

// withSecondContainerSeccompProfileType adds a second container with given seccompProfile.type
func withSecondContainerSeccompProfileType(t string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, corev1.Container{
			Name:  "second",
			Image: "nginx:1.21",
			SecurityContext: &corev1.SecurityContext{
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileType(t),
				},
			},
		})
	}
}

// withInitContainerSeccompProfileType adds an init container with given seccompProfile.type
func withInitContainerSeccompProfileType(t string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:  "init",
				Image: "busybox:1.36",
				SecurityContext: &corev1.SecurityContext{
					SeccompProfile: &corev1.SeccompProfile{
						Type: corev1.SeccompProfileType(t),
					},
				},
			},
		}
	}
}

// withInitContainerSeccompProfileTypeLocalhost adds an init container with Localhost profile
func withInitContainerSeccompProfileTypeLocalhost(profile string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:  "init",
				Image: "busybox:1.36",
				SecurityContext: &corev1.SecurityContext{
					SeccompProfile: &corev1.SeccompProfile{
						Type:             corev1.SeccompProfileTypeLocalhost,
						LocalhostProfile: &profile,
					},
				},
			},
		}
	}
}

// validateDeploymentRaw and hasErrErrItems are shared helpers from phase2_integration_test.go
// baseDeployment, deployment(), boolPtr(), int32Ptr() are also shared from phase2_integration_test.go