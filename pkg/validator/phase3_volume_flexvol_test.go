package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.15 — Phase 3 Volume FlexVolume Tests
// Tests deep nested field traversal: Volume → FlexVolume → driver / fsType
// Codegen confirmed:
//   .volumes[*].flexVolume.driver    Required:true  (string)
//   .volumes[*].flexVolume.fsType    Required:false  (string)
//   .volumes[*].flexVolume.options   Required:false  (object)
//   .volumes[*].flexVolume.readOnly  Required:false  (boolean)
//   .volumes[*].flexVolume.secretRef Required:false
// Note: FSType is string (not *string) in k8s.io/api/core/v1.FlexVolumeSource
// =============================================================================

// TestPhase3_Volume_FlexVol tests Volume.FlexVolumeSource fields.
// driver is Required:true per codegen.
func TestPhase3_Volume_FlexVol(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// ── Valid cases ────────────────────────────────────────────────────────
		{
			name: "valid — flexVolume with driver+fsType",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "flex",
						VolumeSource: corev1.VolumeSource{
							FlexVolume: &corev1.FlexVolumeSource{
								Driver: "k8s.io/smb",
								FSType: "ext4",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "flex", MountPath: "/mnt/flex"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — flexVolume with driver only (minimal valid)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "flex",
						VolumeSource: corev1.VolumeSource{
							FlexVolume: &corev1.FlexVolumeSource{
								Driver: "k8s.io/nfs",
								// fsType, options, readOnly, secretRef all optional
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		// ── Invalid cases ──────────────────────────────────────────────────────
		{
			name: "invalid — flexVolume.driver empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "flex",
						VolumeSource: corev1.VolumeSource{
							FlexVolume: &corev1.FlexVolumeSource{
								Driver: "",
								FSType:  "ext4",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "driver",
		},
		{
			name: "invalid — flexVolume.driver zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "flex",
						VolumeSource: corev1.VolumeSource{
							FlexVolume: &corev1.FlexVolumeSource{
								// Driver not set at all — zero value "" → Required:true → error
								FSType: "ext4",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "driver",
		},
		{
			name: "valid — flexVolume with all optional fields",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "flex",
						VolumeSource: corev1.VolumeSource{
							FlexVolume: &corev1.FlexVolumeSource{
								Driver:   "k8s.io/iscsi",
								FSType:   "xfs",
								ReadOnly: true,
								Options: map[string]string{
									"target": "iqn.2014-05.example.storage:vol1",
									"portal": "192.168.1.100:3260",
								},
								SecretRef: &corev1.LocalObjectReference{Name: "iscsi-secret"},
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
			t.Logf("[Phase3-Volume-FlexVol] testing: %s", tc.name)
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
