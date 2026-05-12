package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.16 — Phase 3 Volume Host Path Tests
// Tests deep nested field traversal: Volume → HostPath → path
// Codegen confirmed: HostPathVolumeSourceDeferredFields.path Required:true, type Optional
// =============================================================================

// TestPhase3_Volume_HostPath tests Volume.HostPathVolumeSource fields.
// path is Required:true per codegen; type is Optional.
func TestPhase3_Volume_HostPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — hostPath with path set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "hostpath",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/var/log/myapp",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "hostpath", MountPath: "/mnt/hostpath"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — hostPath with path and type set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "hostpath",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/var/data",
								Type: func() *corev1.HostPathType {
									t := corev1.HostPathDirectory
									return &t
								}(),
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "hostpath", MountPath: "/mnt/hostpath"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — hostPath missing path (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "hostpath",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								// Path not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "path",
		},
		{
			name: "invalid — hostPath with empty path string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "hostpath",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "path",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-HostPath] testing: %s", tc.name)
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
