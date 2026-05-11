package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.4 — Phase 3 Volume Cinder Tests
// Tests deep nested field traversal: Volume → Cinder → volumeID / fsType / readOnly
// Codegen confirmed: CinderVolumeSourceDeferredFields.volumeID Required:true
// =============================================================================

// TestPhase3_Volume_Cinder tests Volume.CinderVolumeSource fields.
// volumeID is Required:true per codegen.
func TestPhase3_Volume_Cinder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — cinder with volumeID+fsType+readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cinder",
						VolumeSource: corev1.VolumeSource{
							Cinder: &corev1.CinderVolumeSource{
								VolumeID: "cinder-volume-12345",
								FSType:   "ext4",
								ReadOnly: false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "cinder", MountPath: "/mnt/cinder"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — cinder.volumeID empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cinder",
						VolumeSource: corev1.VolumeSource{
							Cinder: &corev1.CinderVolumeSource{
								VolumeID: "",
								FSType:   "ext4",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "volumeID",
		},
		{
			name: "invalid — cinder.volumeID zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "cinder",
						VolumeSource: corev1.VolumeSource{
							Cinder: &corev1.CinderVolumeSource{
								FSType: "ext4",
								// VolumeID not set at all — zero value "" → Required:true → error
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
			t.Logf("[Phase3-Volume-Cinder] testing: %s", tc.name)
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
