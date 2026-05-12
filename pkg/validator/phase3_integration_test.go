package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-16.3 — Phase 3 Integration Tests
// PRD ref: PRD Section 20 Phase 3 sub-type traversal + structural recursion
// Tests container sub-field structural validation: VolumeMount, VolumeDevice,
// EnvVar, ContainerPort, EnvFromSource
// =============================================================================

// -----------------------------------------------------------------------
// A. Structural Layer Tests — Sub-type Traversal
// Validates that container sub-types (EnvVar, VolumeMount, VolumeDevice,
// ContainerPort, EnvFromSource) are correctly traversed and their required
// fields are enforced.
// -----------------------------------------------------------------------

func TestPhase3_Structural_VolumeMount_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "volumeMount with empty mountPath — Required:true",
			deploy: deployment(withVolumeMounts([]corev1.VolumeMount{
				{Name: "data-volume", MountPath: ""},
			}), withVolumes([]corev1.Volume{
				{Name: "data-volume"},
			})),
			wantErr:    true,
			errSubstr:  "mountPath",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3] testing volumeMount invalid: %s", tc.name)
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

func TestPhase3_Structural_VolumeMount_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "volumeMount with valid mountPath passes",
			deploy: deployment(withVolumeMounts([]corev1.VolumeMount{
				{Name: "data-volume", MountPath: "/data"},
			}), withVolumes([]corev1.Volume{
				{Name: "data-volume"},
			})),
			wantErr: false,
		},
		{
			name: "volumeMount with name and readOnly passes",
			deploy: deployment(withVolumeMounts([]corev1.VolumeMount{
				{Name: "config-volume", MountPath: "/config", ReadOnly: true},
			}), withVolumes([]corev1.Volume{
				{Name: "config-volume"},
			})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3] testing volumeMount valid: %s", tc.name)
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

func TestPhase3_Structural_VolumeDevice_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "volumeDevice with empty devicePath — Required:true",
			deploy: deployment(withVolumeDevices([]corev1.VolumeDevice{
				{Name: "data-device", DevicePath: ""},
			})),
			wantErr:    true,
			errSubstr:  "devicePath",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3] testing volumeDevice invalid: %s", tc.name)
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

func TestPhase3_Structural_VolumeDevice_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "volumeDevice with valid devicePath passes",
			deploy: deployment(withVolumeDevices([]corev1.VolumeDevice{
				{Name: "data-device", DevicePath: "/dev/sda1"},
			})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3] testing volumeDevice valid: %s", tc.name)
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

func TestPhase3_Structural_EnvVar_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "envVar with name and value passes",
			deploy: deployment(withEnv([]corev1.EnvVar{
				{Name: "FOO", Value: "bar"},
			})),
			wantErr: false,
		},
		{
			name: "envVar with name and valueFrom passes",
			deploy: deployment(withEnv([]corev1.EnvVar{
				{
					Name: "SECRET_NAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
							Key:                  "api-key",
						},
					},
				},
			})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3] testing envVar valid: %s", tc.name)
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

func TestPhase3_Structural_ContainerPort_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "containerPort with name, containerPort, and protocol passes",
			deploy: deployment(withPorts([]corev1.ContainerPort{
				{Name: "http", ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
			})),
			wantErr: false,
		},
		{
			name: "containerPort with hostPort and hostIP passes",
			deploy: deployment(withPorts([]corev1.ContainerPort{
				{Name: "metrics", ContainerPort: 9090, HostPort: 9090, HostIP: "0.0.0.0", Protocol: corev1.ProtocolTCP},
			})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3] testing containerPort valid: %s", tc.name)
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
// B. Auto-Catches Verification Tests
// Confirms that all 85 schemas are registered, expected deferred fields
// are present with correct Required flags, and flat-index entries resolve
// correctly.
// -----------------------------------------------------------------------

func TestPhase3_AutoCatches_AllSchemasRegistered(t *testing.T) {
	t.Parallel()
	schemas := GetAllSchemas()
	if len(schemas) != 85 {
		t.Errorf("GetAllSchemas() returned %d entries, want exactly 85", len(schemas))
	}
}

func TestPhase3_AutoCatches_ExpectedFields(t *testing.T) {
	t.Parallel()

	// EnvVarDeferredFields: .name (Required:true), .value (Required:false), .valueFrom (Required:false)
	envVarFields := EnvVarDeferredFields
	expectedEnvVar := map[string]bool{
		".name":     true,
		".value":    false,
		".valueFrom": false,
	}
	for field, required := range expectedEnvVar {
		meta, ok := envVarFields[field]
		if !ok {
			t.Errorf("EnvVarDeferredFields missing field %q", field)
			continue
		}
		if meta.Required != required {
			t.Errorf("EnvVarDeferredFields[%q].Required = %v, want %v", field, meta.Required, required)
		}
	}

	// VolumeMountDeferredFields: .mountPath (Required:true), .name (Required:true), .readOnly (Required:false)
	volMountFields := VolumeMountDeferredFields
	expectedVolMount := map[string]bool{
		".mountPath":      true,
		".name":           true,
		".readOnly":       false,
		".mountPropagation": false,
	}
	for field, required := range expectedVolMount {
		meta, ok := volMountFields[field]
		if !ok {
			t.Errorf("VolumeMountDeferredFields missing field %q", field)
			continue
		}
		if meta.Required != required {
			t.Errorf("VolumeMountDeferredFields[%q].Required = %v, want %v", field, meta.Required, required)
		}
	}

	// VolumeDeviceDeferredFields: .devicePath (Required:true), .name (Required:true)
	volDeviceFields := VolumeDeviceDeferredFields
	expectedVolDevice := map[string]bool{
		".devicePath": true,
		".name":       true,
	}
	for field, required := range expectedVolDevice {
		meta, ok := volDeviceFields[field]
		if !ok {
			t.Errorf("VolumeDeviceDeferredFields missing field %q", field)
			continue
		}
		if meta.Required != required {
			t.Errorf("VolumeDeviceDeferredFields[%q].Required = %v, want %v", field, meta.Required, required)
		}
	}
}

func TestPhase3_AutoCatches_FlatIndexEntries(t *testing.T) {
	t.Parallel()

	// .mountPath must be Required:true in flat index
	meta, ok := LookupDeferredField(".mountPath")
	if !ok {
		t.Errorf("LookupDeferredField(\".mountPath\") returned false, want true")
	} else if !meta.Required {
		t.Errorf("LookupDeferredField(\".mountPath\").Required = false, want true")
	}

	// .devicePath must be Required:true in flat index
	meta, ok = LookupDeferredField(".devicePath")
	if !ok {
		t.Errorf("LookupDeferredField(\".devicePath\") returned false, want true")
	} else if !meta.Required {
		t.Errorf("LookupDeferredField(\".devicePath\").Required = false, want true")
	}

	// .name entries should exist (name on EnvVar is Required:true, name on VolumeMount is Required:true)
	nameMeta, ok := LookupDeferredField(".name")
	if !ok {
		t.Errorf("LookupDeferredField(\".name\") returned false, want true")
	}
	_ = nameMeta // name may or may not be required; not the focus of this test
}

// -----------------------------------------------------------------------
// C. Implementation Helpers — Deployment modifiers for Phase 3 sub-types
// -----------------------------------------------------------------------

// withEnv sets env vars on the container.
func withEnv(envs []corev1.EnvVar) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Env = envs
	}
}

// withVolumeMounts sets volume mounts on the container.
func withVolumeMounts(volMounts []corev1.VolumeMount) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].VolumeMounts = volMounts
	}
}

// withPorts sets container ports on the container.
func withPorts(ports []corev1.ContainerPort) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Ports = ports
	}
}

// withVolumeDevices sets volume devices on the container.
func withVolumeDevices(devices []corev1.VolumeDevice) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].VolumeDevices = devices
	}
}

// withEnvFrom sets envFrom on the container.
func withEnvFrom(envFrom []corev1.EnvFromSource) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].EnvFrom = envFrom
	}
}

// withVolumes sets volumes on the pod spec (needed for VolumeMount references).
func withVolumes(volumes []corev1.Volume) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Volumes = volumes
	}
}
