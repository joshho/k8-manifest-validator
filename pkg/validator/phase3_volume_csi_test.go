package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.14 — Phase 3 Volume CSI Tests
// Tests deep nested field traversal: Volume → CSIVolumeSource → driver
// Codegen confirmed: CSIVolumeSourceDeferredFields.driver Required:true
// Note: volumeHandle/controllerPublishSecretRef/nodeStageSecretRef are NOT in
// k8s.io/api@v0.36.0 CSIVolumeSource — the struct only has:
//   Driver, ReadOnly (*bool), FSType (*string), VolumeAttributes, NodePublishSecretRef
// =============================================================================

// TestPhase3_Volume_CSI tests Volume.CSIVolumeSource fields.
// driver is Required:true per codegen.
func TestPhase3_Volume_CSI(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — csi with driver set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "csi",
						VolumeSource: corev1.VolumeSource{
							CSI: &corev1.CSIVolumeSource{
								Driver: "ebs.csi.aws.com",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "csi", MountPath: "/mnt/csi"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — csi with all optional fields set",
			deployFn: func() *appsv1.Deployment {
				readOnly := true
				fsType := "ext4"
				return deployment(
					withVolume(&corev1.Volume{
						Name: "csi-full",
						VolumeSource: corev1.VolumeSource{
							CSI: &corev1.CSIVolumeSource{
								Driver:                  "ebs.csi.aws.com",
								FSType:                  &fsType,
								ReadOnly:                &readOnly,
								NodePublishSecretRef:    &corev1.LocalObjectReference{Name: "node-publish-secret"},
								VolumeAttributes:        map[string]string{"storage.kubernetes.io/csiProvisionerIdentity": "adsf"},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "csi-full", MountPath: "/mnt/csi-full"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — csi with nodePublishSecretRef.name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "csi-secret",
						VolumeSource: corev1.VolumeSource{
							CSI: &corev1.CSIVolumeSource{
								Driver:               "ebs.csi.aws.com",
								NodePublishSecretRef: &corev1.LocalObjectReference{Name: "my-csi-node-secret"},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "csi-secret", MountPath: "/mnt/csi-secret"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — csi.driver empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "csi-no-driver",
						VolumeSource: corev1.VolumeSource{
							CSI: &corev1.CSIVolumeSource{
								Driver: "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "driver",
		},
		{
			name: "invalid — csi.driver zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "csi-no-driver",
						VolumeSource: corev1.VolumeSource{
							CSI: &corev1.CSIVolumeSource{
								// Driver not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "driver",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-CSI] testing: %s", tc.name)
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