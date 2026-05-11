package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.7 — Phase 3 Volume NFS Tests
// Tests deep nested field traversal: Volume → NFS → server / path
// Codegen confirmed: NFSVolumeSource.server Required:true, NFSVolumeSource.path Required:true
// =============================================================================

// TestPhase3_Volume_NFS tests Volume.NFSVolumeSource fields.
// server and path are Required:true per codegen.
func TestPhase3_Volume_NFS(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — nfs with server+path+readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "nfs",
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server:   "nfs.example.com",
								Path:     "/exports",
								ReadOnly: false,
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "nfs", MountPath: "/mnt/nfs"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — nfs.server empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "nfs",
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server: "",
								Path:   "/exports",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "server",
		},
		{
			name: "invalid — nfs.server zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "nfs",
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Path: "/exports",
								// Server not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "server",
		},
		{
			name: "invalid — nfs.path empty string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "nfs",
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server: "nfs.example.com",
								Path:   "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "path",
		},
		{
			name: "invalid — nfs.path zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "nfs",
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server: "nfs.example.com",
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
			name: "valid — nfs with readOnly true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "nfs",
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server:   "nfs.example.com",
								Path:     "/exports",
								ReadOnly: true,
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
			t.Logf("[Phase3-Volume-NFS] testing: %s", tc.name)
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