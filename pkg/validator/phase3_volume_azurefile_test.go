package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.1 — Phase 3 Volume AzureFile shareName Tests
// Tests deep nested field traversal: Volume → AzureFile → shareName
// Codegen confirmed: AzureFileVolumeSourceDeferredFields.shareName Required:true
// Already covered azureFile.secretName in AWU-17.3.
// =============================================================================

// withVolume is a deployFn modifier that attaches a single volume to the deployment.
func withVolume(vol *corev1.Volume) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, *vol)
	}
}

// TestPhase3_Volume_AzureFile tests Volume.AzureFileVolumeSource fields.
// shareName is Required:true per codegen.
func TestPhase3_Volume_AzureFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — azureFile.shareName set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "az",
						VolumeSource: corev1.VolumeSource{
							AzureFile: &corev1.AzureFileVolumeSource{
								SecretName: "my-az-secret",
								ShareName:  "my-share",
								ReadOnly:   false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "az", MountPath: "/mnt/azure"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — azureFile.shareName empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "az",
						VolumeSource: corev1.VolumeSource{
							AzureFile: &corev1.AzureFileVolumeSource{
								SecretName: "my-az-secret",
								ShareName:  "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "shareName",
		},
		{
			name: "invalid — azureFile.shareName zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "az",
						VolumeSource: corev1.VolumeSource{
							AzureFile: &corev1.AzureFileVolumeSource{
								SecretName: "my-az-secret",
								// ShareName not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "shareName",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-AzureFile] testing: %s", tc.name)
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
