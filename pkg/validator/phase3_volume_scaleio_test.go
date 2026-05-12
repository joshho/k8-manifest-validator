package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.13 — Phase 3 Volume ScaleIO Tests
// Tests deep nested field traversal: Volume → ScaleIO → gateway/secretRef/system
// Codegen confirmed: ScaleIOVolumeSourceDeferredFields.gateway Required:true
//                     ScaleIOVolumeSourceDeferredFields.secretRef Required:true
//                     ScaleIOVolumeSourceDeferredFields.system Required:true
// =============================================================================

// TestPhase3_Volume_ScaleIO tests Volume.ScaleIOVolumeSource fields.
// gateway, secretRef.name, and system are Required:true per codegen.
func TestPhase3_Volume_ScaleIO(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — scaleIO with gateway + secretRef.name + system set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "sio",
						VolumeSource: corev1.VolumeSource{
							ScaleIO: &corev1.ScaleIOVolumeSource{
								Gateway:          "https://scaleio-gateway.example.com:443",
								SecretRef:        &corev1.LocalObjectReference{Name: "scaleio-secret"},
								System:           "scaleio-system",
								StorageMode:      "thick",
								StoragePool:      "sp1",
								ProtectionDomain: "pd1",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "sio", MountPath: "/mnt/scaleio"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — scaleIO with all optional fields set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "sio",
						VolumeSource: corev1.VolumeSource{
							ScaleIO: &corev1.ScaleIOVolumeSource{
								Gateway:          "https://scaleio-gateway.example.com:443",
								SecretRef:        &corev1.LocalObjectReference{Name: "scaleio-secret"},
								System:           "scaleio-system",
								FSType:           "ext4",
								ProtectionDomain: "pd1",
								ReadOnly:         true,
								SSLEnabled:       true,
								StorageMode:      "thick",
								StoragePool:      "sp1",
								VolumeName:       "my-volume",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "sio", MountPath: "/mnt/scaleio"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — scaleIO missing gateway (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "sio",
						VolumeSource: corev1.VolumeSource{
							ScaleIO: &corev1.ScaleIOVolumeSource{
								SecretRef: &corev1.LocalObjectReference{Name: "scaleio-secret"},
								System:    "scaleio-system",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "gateway",
		},
		{
			name: "invalid — scaleIO missing secretRef (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "sio",
						VolumeSource: corev1.VolumeSource{
							ScaleIO: &corev1.ScaleIOVolumeSource{
								Gateway: "https://scaleio-gateway.example.com:443",
								System:  "scaleio-system",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "secretRef",
		},
		{
			name: "invalid — scaleIO missing system (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "sio",
						VolumeSource: corev1.VolumeSource{
							ScaleIO: &corev1.ScaleIOVolumeSource{
								Gateway:   "https://scaleio-gateway.example.com:443",
								SecretRef: &corev1.LocalObjectReference{Name: "scaleio-secret"},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "system",
		},
		{
			name: "invalid — scaleIO with empty gateway string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "sio",
						VolumeSource: corev1.VolumeSource{
							ScaleIO: &corev1.ScaleIOVolumeSource{
								Gateway:   "",
								SecretRef: &corev1.LocalObjectReference{Name: "scaleio-secret"},
								System:    "scaleio-system",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "gateway",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-ScaleIO] testing: %s", tc.name)
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
