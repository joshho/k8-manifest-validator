package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.20 — Phase 3 Volume FC (Fibre Channel) Tests
// Tests deep nested field traversal: Volume → FC → targetWWNs / lun / fsType
// Codegen confirmed: All FC fields are Optional (Required:false)
//   .volumes[*].fc.targetWWNs  — array of strings
//   .volumes[*].fc.lun          — integer (int32)
//   .volumes[*].fc.fsType       — string
//   .volumes[*].fc.readOnly     — boolean
//   .volumes[*].fc.wwids        — array of strings
// This is a passthrough AWU — all fields optional → all valid configs pass.
// =============================================================================

// TestPhase3_Volume_FC tests Volume.FCVolumeSource fields.
// All fields are Optional per codegen — no field-level validation errors possible.
func TestPhase3_Volume_FC(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — fc with targetWWNs + lun + fsType",
			deployFn: func() *appsv1.Deployment {
				var lun int32 = 1
				return deployment(
					withVolume(&corev1.Volume{
						Name: "fc",
						VolumeSource: corev1.VolumeSource{
							FC: &corev1.FCVolumeSource{
								TargetWWNs: []string{"50060e801046a4c2", "50060e801046a4c3"},
								Lun:        &lun,
								FSType:     "ext4",
								ReadOnly:   false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "fc", MountPath: "/mnt/fc"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — fc with all fields set",
			deployFn: func() *appsv1.Deployment {
				var lun int32 = 2
				return deployment(
					withVolume(&corev1.Volume{
						Name: "fc",
						VolumeSource: corev1.VolumeSource{
							FC: &corev1.FCVolumeSource{
								TargetWWNs: []string{"50060e801046a4c2", "50060e801046a4c3"},
								Lun:        &lun,
								FSType:     "xfs",
								ReadOnly:   true,
								WWIDs:      []string{"wwid1", "wwid2"},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "fc", MountPath: "/mnt/fc"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — fc with only wwids set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "fc",
						VolumeSource: corev1.VolumeSource{
							FC: &corev1.FCVolumeSource{
								WWIDs: []string{"eui.00a0d70000b4d5c8"},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "fc", MountPath: "/mnt/fc"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — fc with no fields set (empty/zero)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "fc",
						VolumeSource: corev1.VolumeSource{
							FC: &corev1.FCVolumeSource{},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "fc", MountPath: "/mnt/fc"}}),
				)
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-FC] testing: %s", tc.name)
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