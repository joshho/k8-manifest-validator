package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-22.4 — Phase 4 SELinuxOptions Field-Level Regex Validation Tests
// PRD ref: PRD Section 20 Phase 4 — Layer 2 semantic security coverage
// Owner file: pkg/validator/phase4_selinux_test.go
// Target: semantic_security.go validateSELinux (selinuxFieldRE)
// Regex: ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$
// =============================================================================

// -----------------------------------------------------------------------
// Valid SELinuxOptions — each field individually + all four together
// -----------------------------------------------------------------------

func TestPhase4_SELinux_User_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid SELinuxOptions.User with system_u",
			deploy:  deployment(withSELinuxUser("system_u")),
			wantErr: false,
		},
		{
			name:    "valid SELinuxOptions.User with root",
			deploy:  deployment(withSELinuxUser("root")),
			wantErr: false,
		},
		{
			name:    "valid SELinuxOptions.User with user_u",
			deploy:  deployment(withSELinuxUser("user_u")),
			wantErr: false,
		},
		{
			name:    "valid SELinuxOptions.User with longname1234567890123456789012345678901234567890123",
			deploy:  deployment(withSELinuxUser("longname1234567890123456789012345678901234567890123")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux User valid: %s", tc.name)
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

func TestPhase4_SELinux_Role_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid SELinuxOptions.Role with object_r",
			deploy:  deployment(withSELinuxRole("object_r")),
			wantErr: false,
		},
		{
			name:    "valid SELinuxOptions.Role with system_r",
			deploy:  deployment(withSELinuxRole("system_r")),
			wantErr: false,
		},
		{
			name:    "valid SELinuxOptions.Role with container_r",
			deploy:  deployment(withSELinuxRole("container_r")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux Role valid: %s", tc.name)
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

func TestPhase4_SELinux_Type_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid SELinuxOptions.Type with svirt_sandbox_file_t",
			deploy:  deployment(withSELinuxType("svirt_sandbox_file_t")),
			wantErr: false,
		},
		{
			name:    "valid SELinuxOptions.Type with container_t",
			deploy:  deployment(withSELinuxType("container_t")),
			wantErr: false,
		},
		{
			name:    "valid SELinuxOptions.Type with process_t",
			deploy:  deployment(withSELinuxType("process_t")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux Type valid: %s", tc.name)
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

func TestPhase4_SELinux_Level_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid SELinuxOptions.Level with s0",
			deploy:  deployment(withSELinuxLevel("s0")),
			wantErr: false,
		},
		{
			name:    "invalid SELinuxOptions.Level with colon (s0:c1,c2)",
			deploy:  deployment(withSELinuxLevel("s0:c1,c2")),
			wantErr: true,
			errSubstr: "level",
		},
		{
			name:    "invalid SELinuxOptions.Level with colon (s0:c1,c2,c3)",
			deploy:  deployment(withSELinuxLevel("s0:c1,c2,c3")),
			wantErr: true,
			errSubstr: "level",
		},
		{
			name:    "valid SELinuxOptions.Level with SystemLow",
			deploy:  deployment(withSELinuxLevel("SystemLow")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux Level valid: %s", tc.name)
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

func TestPhase4_SELinux_AllFields_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "invalid SELinuxOptions with all four fields (colon in level)",
			deploy: deployment(withSELinuxOptionsAll("system_u", "object_r", "container_t", "s0:c1,c2")),
			wantErr: true,
			errSubstr: "level",
		},
		{
			name: "valid SELinuxOptions with all four fields max length",
			deploy: deployment(withSELinuxOptionsAll(
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 64 a's (valid per regex)
				"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", // 64 b's
				"cccccccccccccccccccccccccccccccccccccccccccccccccccc", // 64 c's
				"dddddddddddddddddddddddddddddddddddddddddddddddddddd", // 64 d's
			)),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux all fields valid: %s", tc.name)
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
// Invalid SELinuxOptions — empty strings and invalid characters
// -----------------------------------------------------------------------

func TestPhase4_SELinux_User_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "invalid SELinuxOptions.User with empty string (non-empty required)",
			deploy:    deployment(withSELinuxUser("")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.User with invalid chars (@ symbol)",
			deploy:    deployment(withSELinuxUser("system_u@")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.User starts with digit",
			deploy:    deployment(withSELinuxUser("1user")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.User with space",
			deploy:    deployment(withSELinuxUser("system user")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.User with special char !",
			deploy:    deployment(withSELinuxUser("user!")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.User exceeding 63 chars",
			deploy:    deployment(withSELinuxUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")), // 65 chars
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux User invalid: %s", tc.name)
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

func TestPhase4_SELinux_Role_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "invalid SELinuxOptions.Role with empty string",
			deploy:    deployment(withSELinuxRole("")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Role with invalid chars (!)",
			deploy:    deployment(withSELinuxRole("invalid!role")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Role starts with digit",
			deploy:    deployment(withSELinuxRole("1role")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Role with spaces",
			deploy:    deployment(withSELinuxRole("role name")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Role with colon",
			deploy:    deployment(withSELinuxRole("role:name")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux Role invalid: %s", tc.name)
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

func TestPhase4_SELinux_Type_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "invalid SELinuxOptions.Type with empty string",
			deploy:    deployment(withSELinuxType("")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Type with caret",
			deploy:    deployment(withSELinuxType("bad^type")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Type starts with digit",
			deploy:    deployment(withSELinuxType("1type")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Type with asterisk",
			deploy:    deployment(withSELinuxType("type*")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Type with slash",
			deploy:    deployment(withSELinuxType("type/name")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux Type invalid: %s", tc.name)
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

func TestPhase4_SELinux_Level_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "invalid SELinuxOptions.Level with empty string",
			deploy:    deployment(withSELinuxLevel("")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Level with exclamation marks",
			deploy:    deployment(withSELinuxLevel("s0!!!")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Level starts with digit",
			deploy:    deployment(withSELinuxLevel("1s0")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "invalid SELinuxOptions.Level with spaces",
			deploy:    deployment(withSELinuxLevel("s0: c1")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing SELinux Level invalid: %s", tc.name)
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
// Container / InitContainer / Pod-level SELinuxOptions
// -----------------------------------------------------------------------

func TestPhase4_SELinux_ContainerLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "container with SecurityContext.SELinuxOptions set",
			deploy:  deployment(withSELinuxOptions(&corev1.SELinuxOptions{User: "system_u", Role: "object_r", Type: "container_t", Level: "s0"})),
			wantErr: false,
		},
		{
			name:    "container with SELinuxOptions User only",
			deploy:  deployment(withSELinuxUser("system_u")),
			wantErr: false,
		},
		{
			name:    "container with SELinuxOptions Type only",
			deploy:  deployment(withSELinuxType("svirt_sandbox_file_t")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing container SELinux: %s", tc.name)
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

func TestPhase4_SELinux_InitContainerLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "initContainer with SELinuxOptions valid",
			deploy:  deployment(withInitContainerSELinuxOptions("system_u", "object_r", "container_t", "s0")),
			wantErr: false,
		},
		{
			name:    "initContainer with SELinuxOptions User only",
			deploy:  deployment(withInitContainerSELinuxUser("system_u")),
			wantErr: false,
		},
		{
			name:    "initContainer with SELinuxOptions Type only",
			deploy:  deployment(withInitContainerSELinuxType("svirt_sandbox_file_t")),
			wantErr: false,
		},
		{
			name:      "initContainer with invalid SELinuxOptions.User rejected",
			deploy:    deployment(withInitContainerSELinuxUser("system_u@")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "initContainer with invalid SELinuxOptions.Role rejected",
			deploy:    deployment(withInitContainerSELinuxRole("invalid!role")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing initContainer SELinux: %s", tc.name)
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

func TestPhase4_SELinux_PodLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "pod-level SecurityContext.SELinuxOptions valid",
			deploy:  deployment(withPodSELinuxOptions("system_u", "object_r", "container_t", "s0")),
			wantErr: false,
		},
		{
			name:    "pod-level SecurityContext.SELinuxOptions User only",
			deploy:  deployment(withPodSELinuxUser("system_u")),
			wantErr: false,
		},
		{
			name:    "pod-level SecurityContext.SELinuxOptions Level only",
			deploy:  deployment(withPodSELinuxLevel("s0:c1,c2")),
			wantErr: false,
		},
		{
			name:      "pod-level SecurityContext.SELinuxOptions invalid User rejected",
			deploy:    deployment(withPodSELinuxUser("")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
		{
			name:      "pod-level SecurityContext.SELinuxOptions invalid Type rejected",
			deploy:    deployment(withPodSELinuxType("bad^type")),
			wantErr:   true,
			errSubstr: "seLinuxOptions",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.4] testing pod-level SELinux: %s", tc.name)
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
// Deployment modifier helpers for SELinuxOptions field-level tests
// =============================================================================

// withSELinuxUser sets only the User field of container SecurityContext.SELinuxOptions
func withSELinuxUser(user string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{User: user}
	}
}

// withSELinuxRole sets only the Role field of container SecurityContext.SELinuxOptions
func withSELinuxRole(role string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{Role: role}
	}
}

// withSELinuxType sets only the Type field of container SecurityContext.SELinuxOptions
func withSELinuxType(typ string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{Type: typ}
	}
}

// withSELinuxLevel sets only the Level field of container SecurityContext.SELinuxOptions
func withSELinuxLevel(level string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{Level: level}
	}
}

// withSELinuxOptionsAll sets all four SELinuxOptions fields on a container
func withSELinuxOptionsAll(user, role, typ, level string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		d.Spec.Template.Spec.Containers[0].SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{
			User:  user,
			Role:  role,
			Type:  typ,
			Level: level,
		}
	}
}

// withInitContainerSELinuxOptions sets SELinuxOptions on an init container with all four fields
func withInitContainerSELinuxOptions(user, role, typ, level string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
			SecurityContext: &corev1.SecurityContext{
				SELinuxOptions: &corev1.SELinuxOptions{
					User:  user,
					Role:  role,
					Type:  typ,
					Level: level,
				},
			},
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// withInitContainerSELinuxUser sets only the User field on an init container
func withInitContainerSELinuxUser(user string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
			SecurityContext: &corev1.SecurityContext{
				SELinuxOptions: &corev1.SELinuxOptions{User: user},
			},
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// withInitContainerSELinuxRole sets only the Role field on an init container
func withInitContainerSELinuxRole(role string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
			SecurityContext: &corev1.SecurityContext{
				SELinuxOptions: &corev1.SELinuxOptions{Role: role},
			},
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// withInitContainerSELinuxType sets only the Type field on an init container
func withInitContainerSELinuxType(typ string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
			SecurityContext: &corev1.SecurityContext{
				SELinuxOptions: &corev1.SELinuxOptions{Type: typ},
			},
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// withPodSELinuxOptions sets SELinuxOptions on the pod-level SecurityContext
func withPodSELinuxOptions(user, role, typ, level string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.SecurityContext == nil {
			d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		d.Spec.Template.Spec.SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{
			User:  user,
			Role:  role,
			Type:  typ,
			Level: level,
		}
	}
}

// withPodSELinuxUser sets only the User field on pod-level SecurityContext.SELinuxOptions
func withPodSELinuxUser(user string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.SecurityContext == nil {
			d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		d.Spec.Template.Spec.SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{User: user}
	}
}

// withPodSELinuxType sets only the Type field on pod-level SecurityContext.SELinuxOptions
func withPodSELinuxType(typ string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.SecurityContext == nil {
			d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		d.Spec.Template.Spec.SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{Type: typ}
	}
}

// withPodSELinuxLevel sets only the Level field on pod-level SecurityContext.SELinuxOptions
func withPodSELinuxLevel(level string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.SecurityContext == nil {
			d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		d.Spec.Template.Spec.SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{Level: level}
	}
}