package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.11 — Phase 3 Volume RBD Tests
// Tests deep nested field traversal: Volume → RBD → image / monitors (array)
// Codegen confirmed: RBDVolumeSource.image Required:true
//                   RBDVolumeSource.monitors Required:true (array)
// =============================================================================

// TestPhase3_Volume_RBD tests Volume.RBDVolumeSource fields.
// image and monitors are Required:true per codegen.
func TestPhase3_Volume_RBD(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — rbd with image+monitors+fsType+readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "rbd",
						VolumeSource: corev1.VolumeSource{
							RBD: &corev1.RBDVolumeSource{
								RBDImage:     "my-rbd-image",
								CephMonitors: []string{"192.168.1.1:6789", "192.168.1.2:6789"},
								FSType:       "ext4",
								RBDPool:      "rbd",
								ReadOnly:     false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "rbd", MountPath: "/mnt/rbd"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — rbd.image empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "rbd",
						VolumeSource: corev1.VolumeSource{
							RBD: &corev1.RBDVolumeSource{
								RBDImage:     "",
								CephMonitors: []string{"192.168.1.1:6789"},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "image",
		},
		{
			name: "invalid — rbd.image zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "rbd",
						VolumeSource: corev1.VolumeSource{
							RBD: &corev1.RBDVolumeSource{
								CephMonitors: []string{"192.168.1.1:6789"},
								// RBDImage not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "image",
		},
		{
			name: "invalid — rbd.monitors empty array",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "rbd",
						VolumeSource: corev1.VolumeSource{
							RBD: &corev1.RBDVolumeSource{
								RBDImage:     "my-rbd-image",
								CephMonitors: []string{},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "monitors",
		},
		{
			name: "invalid — rbd.monitors nil (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "rbd",
						VolumeSource: corev1.VolumeSource{
							RBD: &corev1.RBDVolumeSource{
								RBDImage: "my-rbd-image",
								// CephMonitors not set at all — nil slice → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "monitors",
		},
		{
			name: "valid — rbd with single monitor",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "rbd",
						VolumeSource: corev1.VolumeSource{
							RBD: &corev1.RBDVolumeSource{
								RBDImage:     "my-rbd-image",
								CephMonitors: []string{"192.168.1.1:6789"},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — rbd with readOnly true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "rbd",
						VolumeSource: corev1.VolumeSource{
							RBD: &corev1.RBDVolumeSource{
								RBDImage:     "my-rbd-image",
								CephMonitors: []string{"192.168.1.1:6789"},
								ReadOnly:     true,
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
			t.Logf("[Phase3-Volume-RBD] testing: %s", tc.name)
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