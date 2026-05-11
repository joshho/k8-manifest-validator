package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.6 — Phase 3 Volume vSphere Volume Tests
// Tests deep nested field traversal: Volume → VsphereVirtualDisk → volumePath / fsType
// Codegen confirmed: VsphereVirtualDiskVolumeSourceDeferredFields.volumePath Required:true
// =============================================================================

// TestPhase3_Volume_Vsphere tests Volume.VsphereVirtualDiskVolumeSource fields.
// volumePath is Required:true per codegen.
func TestPhase3_Volume_Vsphere(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — vsphereVolume with volumePath+fsType",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "vsphere",
						VolumeSource: corev1.VolumeSource{
							VsphereVolume: &corev1.VsphereVirtualDiskVolumeSource{
								VolumePath: "[datastore] vm-disk.vmdk",
								FSType:     "ext4",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "vsphere", MountPath: "/mnt/vsphere"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — vsphereVolume.volumePath empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "vsphere",
						VolumeSource: corev1.VolumeSource{
							VsphereVolume: &corev1.VsphereVirtualDiskVolumeSource{
								VolumePath: "",
								FSType:     "ext4",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "volumePath",
		},
		{
			name: "invalid — vsphereVolume.volumePath zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "vsphere",
						VolumeSource: corev1.VolumeSource{
							VsphereVolume: &corev1.VsphereVirtualDiskVolumeSource{
								FSType: "ext4",
								// VolumePath not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "volumePath",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-Vsphere] testing: %s", tc.name)
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
