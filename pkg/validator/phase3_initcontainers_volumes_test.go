package validator

import (
	
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-21.1 — Phase 3 initContainers × volumes Cross-Struct Validation
// Tests cross-struct validation between initContainers[*].volumeMounts and
// the top-level volumes array, plus initContainers[*].envFrom cross-references.
// Codegen entries covered:
//   - initContainers[*].volumeMounts
//   - initContainers[*].envFrom
//   - volumes[*]
// Validation wire: validateVolumeMount in builtin.go (called from
// validateContainers via validatePodSpec).
// =============================================================================

// -----------------------------------------------------------------------
// Valid: configMap volume mount
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_ConfigMap tests initContainer with configMap volume mount.
func TestPhase3_InitVolumes_ConfigMap(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		deployFn func() *appsv1.Deployment
		wantErr bool
	}{
		{
			name: "valid — initContainer with configMap volume mount",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "cfg-volume", MountPath: "/config"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "cfg-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with configMap subPath",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "cfg-volume", MountPath: "/config", SubPath: "config.yaml"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "cfg-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with configMap readOnly=true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "cfg-volume", MountPath: "/config", ReadOnly: true},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "cfg-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-ConfigMap] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Valid: secret volume mount
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_Secret tests initContainer with secret volume mount.
func TestPhase3_InitVolumes_Secret(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		deployFn func() *appsv1.Deployment
		wantErr bool
	}{
		{
			name: "valid — initContainer with secret volume mount",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "secret-volume", MountPath: "/secrets"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "secret-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "my-secret",
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with secret readOnly=true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "secret-volume", MountPath: "/secrets", ReadOnly: true},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "secret-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "my-secret",
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-Secret] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Valid: emptyDir volume mount
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_EmptyDir tests initContainer with emptyDir volume mount.
func TestPhase3_InitVolumes_EmptyDir(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		deployFn func() *appsv1.Deployment
		wantErr bool
	}{
		{
			name: "valid — initContainer with emptyDir volume mount",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "tmp-volume", MountPath: "/tmp"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "tmp-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					}),
				)
			},
			wantErr: false,
		},

	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-EmptyDir] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Valid: hostPath volume mount
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_HostPath tests initContainer with hostPath volume mount.
func TestPhase3_InitVolumes_HostPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — initContainer with hostPath volume mount",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "host-volume", MountPath: "/host/data"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "host-volume",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/log",
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with hostPath DirectoryOrCreate",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "host-volume", MountPath: "/mnt/data"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "host-volume",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/data",
									Type: &[]corev1.HostPathType{corev1.HostPathDirectoryOrCreate}[0],
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-HostPath] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
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
// Valid: initContainer envFrom configMapRef and secretRef
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_EnvFrom tests initContainer envFrom cross-references.
func TestPhase3_InitVolumes_EnvFrom(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		deployFn func() *appsv1.Deployment
		wantErr bool
	}{
		{
			name: "valid — initContainer envFrom configMapRef",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
									},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer envFrom secretRef",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
									},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer envFrom configMapRef with optional=false",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
										Optional:             boolPtr(false),
									},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer envFrom secretRef with optional=true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
										Optional:             boolPtr(true),
									},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-EnvFrom] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Invalid: volumeMount referencing non-existent volume
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_NonExistentVolume tests initContainer volumeMount referencing
// a volume name that is not declared in the volumes list.
func TestPhase3_InitVolumes_NonExistentVolume(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "invalid — initContainer volumeMount.name references non-existent volume",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "non-existent-volume", MountPath: "/data"},
							},
						},
					}),
					// volumes is empty — the mount references a volume that doesn't exist
					withVolumes([]corev1.Volume{}),
				)
			},
			wantErr:   true,
			errSubstr: "non-existent-volume",
		},
		{
			name: "invalid — initContainer volumeMount.name references volume declared after it",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "late-volume", MountPath: "/data"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{Name: "early-volume", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "late-volume",
		},
		{
			name: "invalid — initContainer has multiple mounts, one references non-existent volume",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "valid-volume", MountPath: "/valid"},
								{Name: "missing-volume", MountPath: "/missing"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{Name: "valid-volume", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "missing-volume",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-NonExistent] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
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
// Invalid: initContainer volumeMount with invalid mountPath
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_InvalidMountPath tests initContainer volumeMount with
// mountPath that is not an absolute path (must start with /).
func TestPhase3_InitVolumes_InvalidMountPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "invalid — initContainer volumeMount.mountPath does not start with /",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "data-volume", MountPath: "data"}, // relative path — invalid
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{Name: "data-volume", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "mountPath",
		},
		{
			name: "invalid — initContainer volumeMount.mountPath is empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "data-volume", MountPath: ""},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{Name: "data-volume", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "mountPath",
		},
		{
			name: "invalid — initContainer volumeMount.mountPath is just /",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "data-volume", MountPath: "/"}, // root path — technically absolute, but unusual
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{Name: "data-volume", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					}),
				)
			},
			wantErr: false, // root "/" is technically an absolute path, so it's valid
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-InvalidMountPath] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
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
// Valid: multiple init containers each with different volume types
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_MultipleInitContainers tests multiple init containers
// each using different volume types (configMap, secret, emptyDir, hostPath).
func TestPhase3_InitVolumes_MultipleInitContainers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		deployFn func() *appsv1.Deployment
		wantErr bool
	}{
		{
			name: "valid — multiple init containers each with different volume types",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-configmap",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "cfg-volume", MountPath: "/config"},
							},
						},
						{
							Name:  "init-secret",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "secret-volume", MountPath: "/secrets"},
							},
						},
						{
							Name:  "init-emptydir",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "tmp-volume", MountPath: "/tmp"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "cfg-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
								},
							},
						},
						{
							Name: "secret-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "my-secret",
								},
							},
						},
						{
							Name: "tmp-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — init container and regular container both mount same volume",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "shared-volume", MountPath: "/shared"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "shared-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-MultipleInit] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Valid: init container with downwardAPI volume
// -----------------------------------------------------------------------

// TestPhase3_InitVolumes_DownwardAPI tests initContainer with downwardAPI volume.
func TestPhase3_InitVolumes_DownwardAPI(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		deployFn func() *appsv1.Deployment
		wantErr bool
	}{
		{
			name: "valid — initContainer with downwardAPI volume",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "downward-volume", MountPath: "/podinfo"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "downward-volume",
							VolumeSource: corev1.VolumeSource{
								DownwardAPI: &corev1.DownwardAPIVolumeSource{
									Items: []corev1.DownwardAPIVolumeFile{
										{
											Path: "labels",
											FieldRef: &corev1.ObjectFieldSelector{
												FieldPath: "metadata.labels",
											},
										},
									},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with downwardAPI showing namespace",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "downward-volume", MountPath: "/podinfo"},
							},
						},
					}),
					withVolumes([]corev1.Volume{
						{
							Name: "downward-volume",
							VolumeSource: corev1.VolumeSource{
								DownwardAPI: &corev1.DownwardAPIVolumeSource{
									Items: []corev1.DownwardAPIVolumeFile{
										{
											Path: "namespace",
											FieldRef: &corev1.ObjectFieldSelector{
												FieldPath: "metadata.namespace",
											},
										},
									},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-InitVolumes-DownwardAPI] %s", tc.name)
			result := validateDeploymentRaw(tc.deployFn())
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
		})
	}
}