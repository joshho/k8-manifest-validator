package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.9 — Phase 3 Volume Cephfs Tests
// Tests deep nested field traversal: Volume → Cephfs → monitors (array)
// Codegen confirmed: CephFSVolumeSource.monitors Required:true (array)
// =============================================================================

// TestPhase3_Volume_Cephfs tests Volume.CephFSVolumeSource fields.
// monitors is Required:true (array) per codegen.
func TestPhase3_Volume_Cephfs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — cephfs with monitors+path+user+readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cephfs",
						VolumeSource: corev1.VolumeSource{
							CephFS: &corev1.CephFSVolumeSource{
								Monitors:  []string{"192.168.1.1:6789", "192.168.1.2:6789"},
								Path:      "/",
								User:      "admin",
								ReadOnly:  false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "cephfs", MountPath: "/mnt/cephfs"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — cephfs.monitors empty array",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cephfs",
						VolumeSource: corev1.VolumeSource{
							CephFS: &corev1.CephFSVolumeSource{
								Monitors: []string{},
								Path:     "/",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "monitors",
		},
		{
			name: "invalid — cephfs.monitors nil (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cephfs",
						VolumeSource: corev1.VolumeSource{
							CephFS: &corev1.CephFSVolumeSource{
								Path: "/",
								// Monitors not set at all — nil slice → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "monitors",
		},
		{
			name: "valid — cephfs with single monitor",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cephfs",
						VolumeSource: corev1.VolumeSource{
							CephFS: &corev1.CephFSVolumeSource{
								Monitors: []string{"192.168.1.1:6789"},
								Path:     "/data",
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — cephfs with readOnly true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cephfs",
						VolumeSource: corev1.VolumeSource{
							CephFS: &corev1.CephFSVolumeSource{
								Monitors: []string{"192.168.1.1:6789"},
								Path:     "/",
								ReadOnly: true,
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
			t.Logf("[Phase3-Volume-Cephfs] testing: %s", tc.name)
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