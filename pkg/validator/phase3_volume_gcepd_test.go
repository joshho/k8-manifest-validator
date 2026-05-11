package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.3 — Phase 3 Volume GCE Persistent Disk Tests
// Tests deep nested field traversal: Volume → GCEPersistentDisk → pdName / fsType / readOnly
// Codegen confirmed: GCEPersistentDiskVolumeSourceDeferredFields.pdName Required:true
// =============================================================================

// TestPhase3_Volume_GCEPD tests Volume.GCEPersistentDiskVolumeSource fields.
// pdName is Required:true per codegen.
func TestPhase3_Volume_GCEPD(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — gcePersistentDisk with pdName+fsType+readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "gcepd",
						VolumeSource: corev1.VolumeSource{
							GCEPersistentDisk: &corev1.GCEPersistentDiskVolumeSource{
								PDName:   "my-persistent-disk",
								FSType:   "ext4",
								ReadOnly: true,
								Partition: 0,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "gcepd", MountPath: "/mnt/gcepd"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — gcePersistentDisk.pdName empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "gcepd",
						VolumeSource: corev1.VolumeSource{
							GCEPersistentDisk: &corev1.GCEPersistentDiskVolumeSource{
								PDName: "",
								FSType: "ext4",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "pdName",
		},
		{
			name: "invalid — gcePersistentDisk.pdName zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "gcepd",
						VolumeSource: corev1.VolumeSource{
							GCEPersistentDisk: &corev1.GCEPersistentDiskVolumeSource{
								FSType: "ext4",
								// PDName not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "pdName",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-GCEPD] testing: %s", tc.name)
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
