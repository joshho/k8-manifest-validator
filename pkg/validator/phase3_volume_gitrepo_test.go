package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-18.17 — Phase 3 Volume GitRepo Tests
// Tests deep nested field traversal: Volume → GitRepo → repository
// Codegen confirmed: GitRepoVolumeSourceDeferredFields.repository Required:true
// =============================================================================

// TestPhase3_Volume_GitRepo tests Volume.GitRepoVolumeSource fields.
// repository is Required:true per codegen.
func TestPhase3_Volume_GitRepo(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — gitRepo with repository set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "git-repo",
						VolumeSource: corev1.VolumeSource{
							GitRepo: &corev1.GitRepoVolumeSource{
								Repository: "https://github.com/example/repo.git",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "git-repo", MountPath: "/mnt/git"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — gitRepo with repository + directory + revision set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "git-repo-full",
						VolumeSource: corev1.VolumeSource{
							GitRepo: &corev1.GitRepoVolumeSource{
								Repository: "https://github.com/example/repo.git",
								Directory:  "subdir",
								Revision:   "main",
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "git-repo-full", MountPath: "/mnt/git-full"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — gitRepo missing repository (Required → error)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "git-repo-missing",
						VolumeSource: corev1.VolumeSource{
							GitRepo: &corev1.GitRepoVolumeSource{
								// Repository not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "repository",
		},
		{
			name: "invalid — gitRepo with empty repository string",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "git-repo-empty",
						VolumeSource: corev1.VolumeSource{
							GitRepo: &corev1.GitRepoVolumeSource{
								Repository: "",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "repository",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-GitRepo] testing: %s", tc.name)
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