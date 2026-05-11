package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.5 — Phase 3 Volume Photon Persistent Disk Tests
// Tests deep nested field traversal: Volume → PhotonPersistentDisk → pdID / fsType
// Codegen confirmed: PhotonPersistentDiskVolumeSourceDeferredFields.pdID Required:true
// =============================================================================

// TestPhase3_Volume_PhotonPD tests Volume.PhotonPersistentDiskVolumeSource fields.
// pdID is Required:true per codegen.
func TestPhase3_Volume_PhotonPD(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — photonPersistentDisk with pdID+fsType",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "photonpd",
						VolumeSource: corev1.VolumeSource{
							PhotonPersistentDisk: &corev1.PhotonPersistentDiskVolumeSource{
								PdID:   "photon-disk-12345",
								FSType: "ext4",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "photonpd", MountPath: "/mnt/photonpd"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — photonPersistentDisk.pdID empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "photonpd",
						VolumeSource: corev1.VolumeSource{
							PhotonPersistentDisk: &corev1.PhotonPersistentDiskVolumeSource{
								PdID:   "",
								FSType: "ext4",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "pdID",
		},
		{
			name: "invalid — photonPersistentDisk.pdID zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "photonpd",
						VolumeSource: corev1.VolumeSource{
							PhotonPersistentDisk: &corev1.PhotonPersistentDiskVolumeSource{
								FSType: "ext4",
								// PdID not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "pdID",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-PhotonPD] testing: %s", tc.name)
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
