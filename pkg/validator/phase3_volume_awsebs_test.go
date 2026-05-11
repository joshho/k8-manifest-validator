package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.2 — Phase 3 Volume AWS Elastic Block Store Tests
// Tests deep nested field traversal: Volume → AWSElasticBlockStore → volumeID
// Codegen confirmed: AWSElasticBlockStoreVolumeSourceDeferredFields.volumeID Required:true
// =============================================================================

// TestPhase3_Volume_AWSEBS tests Volume.AWSElasticBlockStoreVolumeSource fields.
// volumeID is Required:true per codegen.
func TestPhase3_Volume_AWSEBS(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — awsElasticBlockStore with all fields set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ebs",
						VolumeSource: corev1.VolumeSource{
							AWSElasticBlockStore: &corev1.AWSElasticBlockStoreVolumeSource{
								VolumeID:  "vol-1234567890abcdef0",
								FSType:    "ext4",
								ReadOnly:  false,
								Partition: 1,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "ebs", MountPath: "/mnt/ebs"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — awsElasticBlockStore.volumeID empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ebs",
						VolumeSource: corev1.VolumeSource{
							AWSElasticBlockStore: &corev1.AWSElasticBlockStoreVolumeSource{
								VolumeID: "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "volumeID",
		},
		{
			name: "invalid — awsElasticBlockStore.volumeID zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ebs",
						VolumeSource: corev1.VolumeSource{
							AWSElasticBlockStore: &corev1.AWSElasticBlockStoreVolumeSource{
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
			t.Logf("[Phase3-Volume-AWSEBS] testing: %s", tc.name)
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
