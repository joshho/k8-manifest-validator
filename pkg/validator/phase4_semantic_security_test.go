package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-22.1 — Phase 4 Semantic Security Tests
// PRD ref: PRD Section 20 Phase 4 — Layer 2 semantic security coverage
// Targets: semantic_security.go validateCapabilities, validateSeccomp, validateSELinux
// =============================================================================

// -----------------------------------------------------------------------
// Capabilities tests (validateCapabilities)
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "container with NET_ADMIN capability valid",
			deploy: deployment(withCapabilitiesAdd("NET_ADMIN")),
			wantErr: false,
		},
		{
			name: "container with SYS_ADMIN capability valid",
			deploy: deployment(withCapabilitiesAdd("SYS_ADMIN")),
			wantErr: false,
		},
		{
			name: "container with Capabilities.Drop valid",
			deploy: deployment(withCapabilitiesDrop("SYS_ADMIN")),
			wantErr: false,
		},
		{
			name: "initContainer with Capabilities.Add valid",
			deploy: deployment(withInitContainerCapabilitiesAdd("NET_ADMIN")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.1] testing capabilities: %s", tc.name)
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

func TestPhase4_Capabilities_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "container with unknown capability rejected",
			deploy:    deployment(withCapabilitiesAdd("NOT_A_REAL_CAPABILITY")),
			wantErr:   true,
			errSubstr: "capability",
		},
		{
			name:      "container with invalid capability string rejected",
			deploy:    deployment(withCapabilitiesAdd("foobarbaz")),
			wantErr:   true,
			errSubstr: "capability",
		},
		{
			name:      "initContainer with unknown capability rejected",
			deploy:    deployment(withInitContainerCapabilitiesAdd("INVALID_CAP")),
			wantErr:   true,
			errSubstr: "capability",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.1] testing capabilities invalid: %s", tc.name)
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

// -----------------------------------------------------------------------
// SELinux tests (validateSELinux)
// -----------------------------------------------------------------------

func TestPhase4_SELinux_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "SELinuxOptions with user, role, type, level all valid",
			deploy: deployment(withSELinuxOptions(&corev1.SELinuxOptions{
				User:  "system_u",
				Role:  "system_r",
				Type:  "container_t",
				Level: "s0c123c456",
			})),
			wantErr: false,
		},
		{
			name: "SELinuxOptions with only type set",
			deploy: deployment(withSELinuxOptions(&corev1.SELinuxOptions{
				Type: "container_t",
			})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.1] testing SELinux valid: %s", tc.name)
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

func TestPhase4_SELinux_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "SELinuxOptions with invalid characters in user field",
			deploy: deployment(withSELinuxOptions(&corev1.SELinuxOptions{
				User: "system_u@", // @ is not alphanumeric or separator
			})),
			wantErr:    true,
			errSubstr: "seLinuxOptions",
		},
		{
			name: "SELinuxOptions with invalid characters in role field",
			deploy: deployment(withSELinuxOptions(&corev1.SELinuxOptions{
				Role: "invalid!role",
			})),
			wantErr:    true,
			errSubstr: "seLinuxOptions",
		},
		{
			name: "SELinuxOptions with invalid characters in type field",
			deploy: deployment(withSELinuxOptions(&corev1.SELinuxOptions{
				Type: "bad^type",
			})),
			wantErr:    true,
			errSubstr: "seLinuxOptions",
		},
		{
			name: "SELinuxOptions with invalid characters in level field",
			deploy: deployment(withSELinuxOptions(&corev1.SELinuxOptions{
				Level: "s0!!!",
			})),
			wantErr:    true,
			errSubstr: "seLinuxOptions",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.1] testing SELinux invalid: %s", tc.name)
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

// -----------------------------------------------------------------------
// seccompProfile tests (validateSeccomp)
// -----------------------------------------------------------------------

func TestPhase4_SeccompProfile_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "seccompProfile.type RuntimeDefault valid",
			deploy:  deployment(withSeccompProfile(corev1.SeccompProfileTypeRuntimeDefault, nil)),
			wantErr: false,
		},
		{
			name:    "seccompProfile.type Unconfined valid",
			deploy:  deployment(withSeccompProfile(corev1.SeccompProfileTypeUnconfined, nil)),
			wantErr: false,
		},
		{
			name: "seccompProfile.type Localhost with localhostProfile valid",
			deploy: deployment(withSeccompProfile(corev1.SeccompProfileTypeLocalhost, strPtr("profiles/myprofile.json"))),
			wantErr: false,
		},
		{
			name: "initContainer seccompProfile.type RuntimeDefault valid",
			deploy: deployment(withInitContainerSeccompProfile(corev1.SeccompProfileTypeRuntimeDefault, nil)),
			wantErr: false,
		},
		{
			name: "initContainer seccompProfile.type Localhost with localhostProfile valid",
			deploy: deployment(withInitContainerSeccompProfile(corev1.SeccompProfileTypeLocalhost, strPtr("profiles/init.json"))),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.1] testing seccompProfile valid: %s", tc.name)
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

func TestPhase4_SeccompProfile_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "seccompProfile.type Unknown rejected",
			deploy:    deployment(withSeccompProfile("Unknown", nil)),
			wantErr:   true,
			errSubstr: "seccompProfile",
		},
		{
			name:      "seccompProfile.type Localhost without localhostProfile rejected",
			deploy:    deployment(withSeccompProfile(corev1.SeccompProfileTypeLocalhost, nil)),
			wantErr:   true,
			errSubstr: "localhostProfile",
		},
		{
			name:      "seccompProfile.type Localhost with empty localhostProfile rejected",
			deploy:    deployment(withSeccompProfile(corev1.SeccompProfileTypeLocalhost, strPtr(""))),
			wantErr:   true,
			errSubstr: "localhostProfile",
		},
		{
			name:      "initContainer seccompProfile.type Unknown rejected",
			deploy:    deployment(withInitContainerSeccompProfile("Unknown", nil)),
			wantErr:   true,
			errSubstr: "seccompProfile",
		},
		{
			name:      "initContainer seccompProfile.type Localhost without localhostProfile rejected",
			deploy:    deployment(withInitContainerSeccompProfile(corev1.SeccompProfileTypeLocalhost, nil)),
			wantErr:   true,
			errSubstr: "localhostProfile",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.1] testing seccompProfile invalid: %s", tc.name)
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
// Deployment modifier helpers for security context
// =============================================================================

// withCapabilitiesAdd sets a container's securityContext.capabilities.add
func withCapabilitiesAdd(cap string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		if d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities.Add = []corev1.Capability{corev1.Capability(cap)}
	}
}

// withCapabilitiesDrop sets a container's securityContext.capabilities.drop
func withCapabilitiesDrop(cap string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		if d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities.Drop = []corev1.Capability{corev1.Capability(cap)}
	}
}

// withSELinuxOptions sets a container's securityContext.seLinuxOptions
func withSELinuxOptions(opts *corev1.SELinuxOptions) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SELinuxOptions = opts
	}
}

// withSeccompProfile sets a container's securityContext.seccompProfile
func withSeccompProfile(profileType corev1.SeccompProfileType, localhostProfile *string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SeccompProfile = &corev1.SeccompProfile{
			Type:             profileType,
			LocalhostProfile: localhostProfile,
		}
	}
}

// withInitContainerCapabilitiesAdd sets an init container's securityContext.capabilities.add
func withInitContainerCapabilitiesAdd(cap string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{corev1.Capability(cap)},
				},
			},
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// withInitContainerSeccompProfile sets an init container's securityContext.seccompProfile
func withInitContainerSeccompProfile(profileType corev1.SeccompProfileType, localhostProfile *string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
			SecurityContext: &corev1.SecurityContext{
				SeccompProfile: &corev1.SeccompProfile{
					Type:             profileType,
					LocalhostProfile: localhostProfile,
				},
			},
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// validateDeploymentRaw and hasErrErrItems are shared helpers from phase2_integration_test.go
// baseDeployment, deployment(), boolPtr(), int32Ptr(), int64Ptr(), strPtr() are shared from phase2_integration_test.go