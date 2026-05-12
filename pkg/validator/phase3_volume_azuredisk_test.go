package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.7 — Phase 3 Volume Azure Disk Tests
// Tests deep nested field traversal: Volume → AzureDiskVolumeSource → diskName, diskURI
// Per k8s.io/api/core/v1.AzureDiskVolumeSource:
//   - DiskName      — Required (string)
//   - DataDiskURI   — Required (string) — NOTE: field name is DataDiskURI not DiskURI
//   - CachingMode, FSType, ReadOnly, Kind — Optional
// =============================================================================

// TestPhase3_Volume_AzureDisk tests Volume.AzureDiskVolumeSource fields.
// DiskName and DataDiskURI are Required per API spec.
func TestPhase3_Volume_AzureDisk(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — azureDisk with diskName and dataDiskURI set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "azuredisk",
						VolumeSource: corev1.VolumeSource{
							AzureDisk: &corev1.AzureDiskVolumeSource{
								DiskName:    "my-data-disk",
								DataDiskURI: "https://myaccount.blob.core.windows.net/disks/my-data-disk.vhd",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "azuredisk", MountPath: "/mnt/azuredisk"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — azureDisk with all optional fields set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "azuredisk-full",
						VolumeSource: corev1.VolumeSource{
							AzureDisk: &corev1.AzureDiskVolumeSource{
								DiskName:    "my-data-disk",
								DataDiskURI: "https://myaccount.blob.core.windows.net/disks/my-data-disk.vhd",
								FSType:      ptrstring("ext4"),
								ReadOnly:    ptrbool(true),
								CachingMode: ptrAzureDataDiskCachingMode(corev1.AzureDataDiskCachingReadOnly),
								Kind:        ptrAzureDataDiskKind(corev1.AzureManagedDisk),
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "azuredisk-full", MountPath: "/mnt/azuredisk-full"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — azureDisk missing diskName (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "azuredisk",
						VolumeSource: corev1.VolumeSource{
							AzureDisk: &corev1.AzureDiskVolumeSource{
								// DiskName not set — zero value "" → Required → error
								DataDiskURI: "https://myaccount.blob.core.windows.net/disks/my-data-disk.vhd",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "diskName",
		},
		{
			name: "invalid — azureDisk missing dataDiskURI (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "azuredisk",
						VolumeSource: corev1.VolumeSource{
							AzureDisk: &corev1.AzureDiskVolumeSource{
								DiskName: "my-data-disk",
								// DataDiskURI not set — zero value "" → Required → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "diskURI",
		},
		{
			name: "invalid — azureDisk with empty diskName string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "azuredisk",
						VolumeSource: corev1.VolumeSource{
							AzureDisk: &corev1.AzureDiskVolumeSource{
								DiskName:    "",
								DataDiskURI:  "https://myaccount.blob.core.windows.net/disks/my-data-disk.vhd",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "diskName",
		},
		{
			name: "invalid — azureDisk with empty dataDiskURI string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "azuredisk",
						VolumeSource: corev1.VolumeSource{
							AzureDisk: &corev1.AzureDiskVolumeSource{
								DiskName:   "my-data-disk",
								DataDiskURI: "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "diskURI",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-AzureDisk] testing: %s", tc.name)
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

// Helper functions for pointer construction

func ptrstring(s string) *string {
	return &s
}

func ptrbool(b bool) *bool {
	return &b
}

func ptrAzureDataDiskCachingMode(m corev1.AzureDataDiskCachingMode) *corev1.AzureDataDiskCachingMode {
	return &m
}

func ptrAzureDataDiskKind(k corev1.AzureDataDiskKind) *corev1.AzureDataDiskKind {
	return &k
}