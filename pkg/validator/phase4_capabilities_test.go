package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-22.5 — Phase 4 Capabilities Enum Additional Coverage Tests
// PRD ref: PRD Section 20 Phase 4 — Layer 2 semantic security coverage
// Owner file: pkg/validator/phase4_capabilities_test.go
// Target: semantic_security.go validateCapabilities (ValidCapabilities map)
// Note: validateCapabilities only validates .add capabilities, NOT drops.
// =============================================================================

// -----------------------------------------------------------------------
// Valid capabilities — each listed capability should be accepted
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_ValidAdd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "NET_ADMIN accepted",
			deploy:  deployment(withCapAdd("NET_ADMIN")),
			wantErr: false,
		},
		{
			name:    "SYS_ADMIN accepted",
			deploy:  deployment(withCapAdd("SYS_ADMIN")),
			wantErr: false,
		},
		{
			name:    "CHOWN accepted",
			deploy:  deployment(withCapAdd("CHOWN")),
			wantErr: false,
		},
		{
			name:    "DAC_OVERRIDE accepted",
			deploy:  deployment(withCapAdd("DAC_OVERRIDE")),
			wantErr: false,
		},
		{
			name:    "NET_BIND_SERVICE accepted",
			deploy:  deployment(withCapAdd("NET_BIND_SERVICE")),
			wantErr: false,
		},
		{
			name:    "SETGID accepted",
			deploy:  deployment(withCapAdd("SETGID")),
			wantErr: false,
		},
		{
			name:    "SETUID accepted",
			deploy:  deployment(withCapAdd("SETUID")),
			wantErr: false,
		},
		{
			name:    "SYS_CHROOT accepted",
			deploy:  deployment(withCapAdd("SYS_CHROOT")),
			wantErr: false,
		},
		{
			name:    "KILL accepted",
			deploy:  deployment(withCapAdd("KILL")),
			wantErr: false,
		},
		{
			name:    "FOWNER accepted",
			deploy:  deployment(withCapAdd("FOWNER")),
			wantErr: false,
		},
		{
			name:    "FSETID accepted",
			deploy:  deployment(withCapAdd("FSETID")),
			wantErr: false,
		},
		{
			name:    "IPC_LOCK accepted",
			deploy:  deployment(withCapAdd("IPC_LOCK")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing valid add capability: %s", tc.name)
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
// Invalid capabilities — unknown names not in ValidCapabilities map
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_InvalidAdd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "completely unknown capability rejected",
			deploy:    deployment(withCapAdd("TOTALLY_INVALID_CAP")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "invalid capability ADMIN capability rejected",
			deploy:    deployment(withCapAdd("ADMIN")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "invalid capability NET_ADMIN_EXTRA rejected",
			deploy:    deployment(withCapAdd("NET_ADMIN_EXTRA")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "invalid capability SYS_ADMIN_LOCAL rejected",
			deploy:    deployment(withCapAdd("SYS_ADMIN_LOCAL")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "invalid capability empty string (not in map)",
			deploy:    deployment(withCapAdd("")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing invalid add capability: %s", tc.name)
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
// Case sensitivity — capabilities are UPPER_SNAKE_CASE; variations should be rejected
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_CaseSensitivity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "lowercase net_admin rejected",
			deploy:    deployment(withCapAdd("net_admin")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "mixed case NetAdmin rejected",
			deploy:    deployment(withCapAdd("NetAdmin")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "lowercase sys_admin rejected",
			deploy:    deployment(withCapAdd("sys_admin")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "lowercase chown rejected",
			deploy:    deployment(withCapAdd("chown")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "valid NET_ADMIN still accepted",
			deploy:    deployment(withCapAdd("NET_ADMIN")),
			wantErr:   false,
		},
		{
			name:      "valid SYS_ADMIN still accepted",
			deploy:    deployment(withCapAdd("SYS_ADMIN")),
			wantErr:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing case sensitivity: %s", tc.name)
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
// Drop capabilities — not validated (validator only checks Add)
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_DropOnly(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "drop NET_ADMIN accepted (not validated)",
			deploy:  deployment(withCapDrop("NET_ADMIN")),
			wantErr: false,
		},
		{
			name:    "drop SYS_ADMIN accepted (not validated)",
			deploy:  deployment(withCapDrop("SYS_ADMIN")),
			wantErr: false,
		},
		{
			name:    "drop CHOWN accepted (not validated)",
			deploy:  deployment(withCapDrop("CHOWN")),
			wantErr: false,
		},
		{
			name:    "drop completely unknown capability accepted (not validated)",
			deploy:  deployment(withCapDrop("TOTALLY_INVALID_CAP")),
			wantErr: false,
		},
		{
			name:    "drop lowercase net_admin accepted (not validated)",
			deploy:  deployment(withCapDrop("net_admin")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing drop capability: %s", tc.name)
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
// Combined add + drop — add must be valid even when drop is also present
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_CombinedAddDrop(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid add + valid drop accepted",
			deploy:  deployment(withCapAdd("NET_ADMIN"), withCapDrop("SYS_ADMIN")),
			wantErr: false,
		},
		{
			name:    "valid add + invalid drop accepted (drop not validated)",
			deploy:  deployment(withCapAdd("NET_ADMIN"), withCapDrop("TOTALLY_INVALID_CAP")),
			wantErr: false,
		},
		{
			name:      "invalid add + valid drop rejected (add is validated)",
			deploy:    deployment(withCapAdd("INVALID_CAP"), withCapDrop("NET_ADMIN")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:      "invalid add + invalid drop rejected (add is validated)",
			deploy:    deployment(withCapAdd("BAD_CAP"), withCapDrop("ALSO_BAD")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:    "add NET_ADMIN + add SYS_ADMIN both valid",
			deploy:  deployment(withCapAddMulti("NET_ADMIN", "SYS_ADMIN")),
			wantErr: false,
		},
		{
			name:      "add NET_ADMIN + add INVALID_CAP rejected",
			deploy:    deployment(withCapAddMulti("NET_ADMIN", "INVALID_CAP")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing combined add/drop: %s", tc.name)
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
// Empty add/drop lists — should be valid (nil/empty slice)
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_EmptyAddDrop(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "empty capabilities (nil add/drop) accepted",
			deploy:  deployment(withCapEmpty()),
			wantErr: false,
		},
		{
			name:    "empty add list with valid drop accepted",
			deploy:  deployment(withCapAdd(""), withCapDrop("NET_ADMIN")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing empty add/drop: %s", tc.name)
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
// Container level capabilities
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_ContainerLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "container with valid NET_ADMIN add accepted",
			deploy:  deployment(withCapAdd("NET_ADMIN")),
			wantErr: false,
		},
		{
			name:    "container with valid SYS_ADMIN add accepted",
			deploy:  deployment(withCapAdd("SYS_ADMIN")),
			wantErr: false,
		},
		{
			name:    "container with valid CHOWN add accepted",
			deploy:  deployment(withCapAdd("CHOWN")),
			wantErr: false,
		},
		{
			name:    "container with invalid cap add rejected",
			deploy:  deployment(withCapAdd("BAD_CAP")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:    "container with multiple valid caps add accepted",
			deploy:  deployment(withCapAddMulti("NET_ADMIN", "SYS_ADMIN", "CHOWN")),
			wantErr: false,
		},
		{
			name:      "container with one invalid cap among valid caps rejected",
			deploy:    deployment(withCapAddMulti("NET_ADMIN", "INVALID_CAP", "CHOWN")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing container level capabilities: %s", tc.name)
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
// InitContainer level capabilities
// -----------------------------------------------------------------------

func TestPhase4_Capabilities_InitContainerLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "initContainer with valid NET_ADMIN add accepted",
			deploy:  deployment(withInitContainerCapAdd("NET_ADMIN")),
			wantErr: false,
		},
		{
			name:    "initContainer with valid SYS_ADMIN add accepted",
			deploy:  deployment(withInitContainerCapAdd("SYS_ADMIN")),
			wantErr: false,
		},
		{
			name:    "initContainer with valid CHOWN add accepted",
			deploy:  deployment(withInitContainerCapAdd("CHOWN")),
			wantErr: false,
		},
		{
			name:      "initContainer with invalid cap add rejected",
			deploy:    deployment(withInitContainerCapAdd("BAD_CAP")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
		{
			name:    "initContainer with multiple valid caps add accepted",
			deploy:  deployment(withInitContainerCapAddMulti("NET_ADMIN", "SYS_ADMIN")),
			wantErr: false,
		},
		{
			name:      "initContainer with one invalid cap among valid rejected",
			deploy:    deployment(withInitContainerCapAddMulti("NET_ADMIN", "BAD_CAP")),
			wantErr:   true,
			errSubstr: "capability not in the supported set",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.5] testing initContainer level capabilities: %s", tc.name)
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
// Deployment modifier helpers for Capabilities enum tests
// =============================================================================

// withCapAdd sets a single capability add on container SecurityContext
func withCapAdd(cap string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{
			Add: []corev1.Capability{corev1.Capability(cap)},
		}
	}
}

// withCapAddMulti sets multiple capability adds on container SecurityContext
func withCapAddMulti(caps ...string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		var capList []corev1.Capability
		for _, c := range caps {
			capList = append(capList, corev1.Capability(c))
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{
			Add: capList,
		}
	}
}

// withCapDrop sets a single capability drop on container SecurityContext
func withCapDrop(cap string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{
			Drop: []corev1.Capability{corev1.Capability(cap)},
		}
	}
}

// withCapEmpty sets an empty Capabilities (nil add/drop slices)
func withCapEmpty() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{}
	}
}

// withInitContainerCapAdd sets a single capability add on initContainer SecurityContext
func withInitContainerCapAdd(cap string) func(*appsv1.Deployment) {
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

// withInitContainerCapAddMulti sets multiple capability adds on initContainer SecurityContext
func withInitContainerCapAddMulti(caps ...string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		var capList []corev1.Capability
		for _, c := range caps {
			capList = append(capList, corev1.Capability(c))
		}
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: capList,
				},
			},
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// withPodCapAdd sets capability add on pod-level SecurityContext
func withPodCapAdd(cap string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.SecurityContext == nil {
			d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		d.Spec.Template.Spec.SecurityContext.Capabilities = &corev1.Capabilities{
			Add: []corev1.Capability{corev1.Capability(cap)},
		}
	}
}

// withPodCapDrop sets capability drop on pod-level SecurityContext
func withPodCapDrop(cap string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.SecurityContext == nil {
			d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		d.Spec.Template.Spec.SecurityContext.Capabilities = &corev1.Capabilities{
			Drop: []corev1.Capability{corev1.Capability(cap)},
		}
	}
}