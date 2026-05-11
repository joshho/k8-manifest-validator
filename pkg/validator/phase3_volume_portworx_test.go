package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.19 — Phase 3 Volume Portworx Tests
// Tests deep nested field traversal: Volume → PortworxVolume → volumeID
// Codegen confirmed: PortworxVolumeSourceDeferredFields.volumeID Required:true
// =============================================================================

// TestPhase3_Volume_Portworx tests Volume.PortworxVolumeSource fields.
// volumeID is Required:true per codegen.
func TestPhase3_Volume_Portworx(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — portworxVolume with volumeID set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "portworx",
						VolumeSource: corev1.VolumeSource{
							PortworxVolume: &corev1.PortworxVolumeSource{
								VolumeID: "px-volume-001",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "portworx", MountPath: "/mnt/portworx"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — portworxVolume with volumeID + fsType + readOnly set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "portworx",
						VolumeSource: corev1.VolumeSource{
							PortworxVolume: &corev1.PortworxVolumeSource{
								VolumeID: "px-volume-002",
								FSType:   "ext4",
								ReadOnly: true,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "portworx", MountPath: "/mnt/portworx"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — portworxVolume missing volumeID (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "portworx",
						VolumeSource: corev1.VolumeSource{
							PortworxVolume: &corev1.PortworxVolumeSource{
								// VolumeID not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "volumeID",
		},
		{
			name: "invalid — portworxVolume with empty volumeID string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "portworx",
						VolumeSource: corev1.VolumeSource{
							PortworxVolume: &corev1.PortworxVolumeSource{
								VolumeID: "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "volumeID",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-Portworx] testing: %s", tc.name)
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