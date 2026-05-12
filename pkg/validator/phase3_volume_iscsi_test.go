package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.10 — Phase 3 Volume ISCSI Tests
// Tests deep nested field traversal: Volume → ISCSI → targetPortal / iqn / lun
// Codegen confirmed: ISCSIVolumeSource.targetPortal Required:true
//                   ISCSIVolumeSource.iqn Required:true
//                   ISCSIVolumeSource.lun Required:true
// =============================================================================

// TestPhase3_Volume_ISCSI tests Volume.ISCSIVolumeSource fields.
// targetPortal, iqn, and lun are Required:true per codegen.
func TestPhase3_Volume_ISCSI(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — iscsi with targetPortal+iqn+lun=1+fsType",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "iscsi",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
								FSType:       "ext4",
								ReadOnly:     false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "iscsi", MountPath: "/mnt/iscsi"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — iscsi.targetPortal empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "iscsi",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "targetPortal",
		},
		{
			name: "invalid — iscsi.targetPortal zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "iscsi",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								IQN: "iqn.2024-05.com.example:storage.target1",
								Lun: 0,
								// TargetPortal not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "targetPortal",
		},
		{
			name: "invalid — iscsi.iqn empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "iscsi",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "",
								Lun:          1,
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "iqn",
		},
		{
			name: "invalid — iscsi.iqn zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "iscsi",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								Lun:          1,
								// IQN not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "iqn",
		},
		{
			name: "invalid — iscsi.lun zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "iscsi",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								// Lun not set at all — zero value 0 → Required:true (integer) → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "lun",
		},
		{
			name: "valid — iscsi with lun=1, readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "iscsi",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
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
			t.Logf("[Phase3-Volume-ISCSI] testing: %s", tc.name)
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