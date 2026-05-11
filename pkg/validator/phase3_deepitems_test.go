package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-17.3 — Phase 3 Deep Nested Volume Items + AzureFile Tests
// Tests deep nested field traversal: Volume → ConfigMap/Secret/AzureFile → items / secretName
// Codegen confirmed: KeyToPathDeferredFields.key Required:true, Items is repeated struct (each item checked),
//                    AzureFileVolumeSourceDeferredFields.secretName Required:true
// =============================================================================

// TestPhase3_DeepNested_VolumeItems tests the deep nested chain:
// Volume.ConfigMap.Items[*].key  — KeyToPathDeferredFields.key Required:true
// Volume.Secret.Items[*].key    — KeyToPathDeferredFields.key Required:true
// Volume.AzureFile.secretName    — AzureFileVolumeSourceDeferredFields.secretName Required:true
func TestPhase3_DeepNested_VolumeItems(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// --- ConfigMap items (key is Required:true per codegen) ---

		{
			name: "valid — configMap.items populated with key+path",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolumes([]corev1.Volume{{
						Name: "cfg",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
								Items:                []corev1.KeyToPath{{Key: "key", Path: "path"}},
							},
						},
					}}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "cfg", MountPath: "/etc/config"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — configMap.items empty (items absent, not validated per repeated-struct rule)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolumes([]corev1.Volume{{
						Name: "cfg",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
								Items:                []corev1.KeyToPath{},
							},
						},
					}}),
				)
			},
			wantErr: false,
		},

		// --- Secret items (key is Required:true per codegen) ---

		{
			name: "valid — secret.items populated with key+path",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolumes([]corev1.Volume{{
						Name: "sec",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "my-secret",
								Items:      []corev1.KeyToPath{{Key: "key", Path: "path"}},
							},
						},
					}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — secret.items empty (items absent, not validated per repeated-struct rule)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolumes([]corev1.Volume{{
						Name: "sec",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "my-secret",
								Items:      []corev1.KeyToPath{},
							},
						},
					}}),
				)
			},
			wantErr: false,
		},

		// --- AzureFile secretName (Required:true per codegen) ---

		{
			name: "valid — azureFile.secretName set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolumes([]corev1.Volume{{
						Name: "az",
						VolumeSource: corev1.VolumeSource{
							AzureFile: &corev1.AzureFileVolumeSource{
								SecretName: "my-az-secret",
								ShareName:  "share-name",
								ReadOnly:   false,
							},
						},
					}}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "az", MountPath: "/mnt/azure"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — azureFile.secretName empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolumes([]corev1.Volume{{
						Name: "az",
						VolumeSource: corev1.VolumeSource{
							AzureFile: &corev1.AzureFileVolumeSource{
								SecretName: "",
								ShareName:  "share-name",
							},
						},
					}}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "az", MountPath: "/mnt/azure"}}),
				)
			},
			wantErr:   true,
			errSubstr: "secretName",
		},
		{
			name: "invalid — azureFile.secretName zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolumes([]corev1.Volume{{
						Name: "az",
						VolumeSource: corev1.VolumeSource{
							AzureFile: &corev1.AzureFileVolumeSource{
								ShareName: "share-name",
								// SecretName not set at all — zero value "" → Required:true → error
							},
						},
					}}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "az", MountPath: "/mnt/azure"}}),
				)
			},
			wantErr:   true,
			errSubstr: "secretName",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepNested-VolumeItems] testing: %s", tc.name)
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