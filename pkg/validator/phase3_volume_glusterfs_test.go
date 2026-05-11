package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.8 — Phase 3 Volume Glusterfs Tests
// Tests deep nested field traversal: Volume → Glusterfs → endpoints / path
// Codegen confirmed: GlusterfsVolumeSource.endpoints Required:true, path Required:true
// =============================================================================

// TestPhase3_Volume_Glusterfs tests Volume.GlusterfsVolumeSource fields.
// endpoints and path are Required:true per codegen.
func TestPhase3_Volume_Glusterfs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — glusterfs with endpoints+path+readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "glusterfs",
						VolumeSource: corev1.VolumeSource{
							Glusterfs: &corev1.GlusterfsVolumeSource{
								EndpointsName: "glusterfs-cluster.default.svc:24007",
								Path:           "my-volume",
								ReadOnly:       false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "glusterfs", MountPath: "/mnt/glusterfs"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — glusterfs.endpoints empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "glusterfs",
						VolumeSource: corev1.VolumeSource{
							Glusterfs: &corev1.GlusterfsVolumeSource{
								EndpointsName: "",
								Path:           "my-volume",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "endpoints",
		},
		{
			name: "invalid — glusterfs.endpoints zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "glusterfs",
						VolumeSource: corev1.VolumeSource{
							Glusterfs: &corev1.GlusterfsVolumeSource{
								Path: "my-volume",
								// EndpointsName not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "endpoints",
		},
		{
			name: "invalid — glusterfs.path empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "glusterfs",
						VolumeSource: corev1.VolumeSource{
							Glusterfs: &corev1.GlusterfsVolumeSource{
								EndpointsName: "glusterfs-cluster.default.svc:24007",
								Path:           "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "path",
		},
		{
			name: "invalid — glusterfs.path zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "glusterfs",
						VolumeSource: corev1.VolumeSource{
							Glusterfs: &corev1.GlusterfsVolumeSource{
								EndpointsName: "glusterfs-cluster.default.svc:24007",
								// Path not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "path",
		},
		{
			name: "valid — glusterfs with readOnly true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "glusterfs",
						VolumeSource: corev1.VolumeSource{
							Glusterfs: &corev1.GlusterfsVolumeSource{
								EndpointsName: "glusterfs-cluster.default.svc:24007",
								Path:           "my-volume",
								ReadOnly:       true,
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
			t.Logf("[Phase3-Volume-Glusterfs] testing: %s", tc.name)
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