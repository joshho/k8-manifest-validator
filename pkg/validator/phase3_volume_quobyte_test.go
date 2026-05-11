package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.18 — Phase 3 Volume Quobyte Tests
// Tests deep nested field traversal: Volume → QuobyteVolumeSource → registry, volume
// Codegen confirmed:
//   - .volumes[*].quobyte.registry  Required:true, Type:"string"
//   - .volumes[*].quobyte.volume    Required:true, Type:"string"
//   - .volumes[*].quobyte.readOnly  Optional:false, Type:"boolean"
//   - .volumes[*].quobyte.tenant    Optional:false, Type:"string"
//   - .volumes[*].quobyte.group     Optional:false, Type:"string"
//   - .volumes[*].quobyte.user      Optional:false, Type:"string"
// =============================================================================

// TestPhase3_Volume_Quobyte tests Volume.QuobyteVolumeSource fields.
func TestPhase3_Volume_Quobyte(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — quobyte with registry + volume set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "quobyte",
						VolumeSource: corev1.VolumeSource{
							Quobyte: &corev1.QuobyteVolumeSource{
								Registry: "registry.example.com:6789",
								Volume:   "myvolume",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "quobyte", MountPath: "/mnt/quobyte"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — quobyte with registry + volume + readOnly set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "quobyte",
						VolumeSource: corev1.VolumeSource{
							Quobyte: &corev1.QuobyteVolumeSource{
								Registry: "registry.example.com:6789",
								Volume:   "myvolume",
								ReadOnly: true,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "quobyte", MountPath: "/mnt/quobyte"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — quobyte with all optional fields set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "quobyte",
						VolumeSource: corev1.VolumeSource{
							Quobyte: &corev1.QuobyteVolumeSource{
								Registry: "registry.example.com:6789",
								Volume:   "myvolume",
								ReadOnly: true,
								Tenant:   "mytenant",
								Group:    "mygroup",
								User:     "myuser",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "quobyte", MountPath: "/mnt/quobyte"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — quobyte missing registry (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "quobyte",
						VolumeSource: corev1.VolumeSource{
							Quobyte: &corev1.QuobyteVolumeSource{
								Volume: "myvolume",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "registry",
		},
		{
			name: "invalid — quobyte with empty registry string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "quobyte",
						VolumeSource: corev1.VolumeSource{
							Quobyte: &corev1.QuobyteVolumeSource{
								Registry: "",
								Volume:   "myvolume",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "registry",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-Quobyte] testing: %s", tc.name)
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
