package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-21.4 — Phase 3 WindowsSecurityContext Tests
// Tests WindowsSecurityContext field traversal on PodSpec containers.
// =============================================================================

func strPtr(s string) *string { return &s }

func windowsSecurityContext() *corev1.WindowsSecurityContextOptions {
	return &corev1.WindowsSecurityContextOptions{
		GMSACredentialSpecName: strPtr("my-gmsa-credspec"),
		GMSACredentialSpec:     strPtr("credential-spec-content"),
		RunAsUserName:          strPtr("ContainerAdministrator"),
	}
}

func withWindowsSecurityContext(ctx *corev1.WindowsSecurityContextOptions) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if len(d.Spec.Template.Spec.Containers) > 0 {
			d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		d.Spec.Template.Spec.Containers[0].SecurityContext.WindowsOptions = ctx
		}
	}
}

func withInitContainerWindowsSecurity(ctx *corev1.WindowsSecurityContextOptions) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if len(d.Spec.Template.Spec.InitContainers) > 0 {
			d.Spec.Template.Spec.InitContainers[0].SecurityContext = &corev1.SecurityContext{}
		d.Spec.Template.Spec.InitContainers[0].SecurityContext.WindowsOptions = ctx
		}
	}
}

// TestPhase3_WindowsSecurityContext_GMSA tests GMSA credential spec on containers.
func TestPhase3_WindowsSecurityContext_GMSA(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		deployFn func() *appsv1.Deployment
		wantErr bool
		errSubstr string
	}

	cases := []testCase{
		{
			name: "valid — container with GMSA credential spec name",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withWindowsSecurityContext(windowsSecurityContext()),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — container with GMSA credential spec full content",
			deployFn: func() *appsv1.Deployment {
				ctx := &corev1.WindowsSecurityContextOptions{
					GMSACredentialSpecName: strPtr("gmsa-spec"),
					GMSACredentialSpec:     strPtr("<?xml version=\"1.0\"?>..."),
				}
				return deployment(withWindowsSecurityContext(ctx))
			},
			wantErr: false,
		},
		{
			name: "valid — container with runAsUserName",
			deployFn: func() *appsv1.Deployment {
				ctx := &corev1.WindowsSecurityContextOptions{
					RunAsUserName: strPtr("ContainerAdministrator"),
				}
				return deployment(withWindowsSecurityContext(ctx))
			},
			wantErr: false,
		},
		{
			name: "valid — container with empty WindowsSecurityContext (all nil/empty)",
			deployFn: func() *appsv1.Deployment {
				return deployment(withWindowsSecurityContext(&corev1.WindowsSecurityContextOptions{}))
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with WindowsSecurityContext",
			deployFn: func() *appsv1.Deployment {
				initCtx := windowsSecurityContext()
				return deployment(
					func(d *appsv1.Deployment) {
						d.Spec.Template.Spec.InitContainers = []corev1.Container{
							{
								Name:  "init-win",
								Image: "mcr.microsoft.com/windows/servercore:ltsc2022",
								SecurityContext: &corev1.SecurityContext{
									WindowsOptions: initCtx,
								},
							},
						}
					},
				)
			},
			wantErr: false,
		},
		{
			name: "valid — WindowsSecurityContext with GMSA + runAsUserName combined",
			deployFn: func() *appsv1.Deployment {
				ctx := &corev1.WindowsSecurityContextOptions{
					GMSACredentialSpecName: strPtr("combined-gmsa"),
					RunAsUserName:          strPtr("ContainerAdmin"),
				}
				return deployment(withWindowsSecurityContext(ctx))
			},
			wantErr: false,
		},
		{
			name: "valid — WindowsSecurityContext on pod spec (not container-level)",
			deployFn: func() *appsv1.Deployment {
				return deployment(func(d *appsv1.Deployment) {
					d.Spec.Template.Spec.SecurityContext.WindowsOptions = &corev1.WindowsSecurityContextOptions{
						GMSACredentialSpecName: strPtr("pod-level-gmsa"),
					}
				})
			},
			wantErr: false,
		},
		{
			name: "valid — multi-container pod, first container with WindowsSecurityContext",
			deployFn: func() *appsv1.Deployment {
				return deployment(func(d *appsv1.Deployment) {
					d.Spec.Template.Spec.Containers[0].SecurityContext.WindowsOptions = windowsSecurityContext()
					d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, corev1.Container{
						Name:  "secondary",
						Image: "mcr.microsoft.com/windows/nanoserver:1809",
					})
				})
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateDeploymentRaw(tc.deployFn())
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v", tc.wantErr, gotErr)
			}
			if tc.wantErr && tc.errSubstr != "" {
				found := false
				for _, e := range result.Errors {
					es := fieldMessage(e)
					if contains(es, tc.errSubstr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q", tc.errSubstr)
				}
			}
		})
	}
}


func fieldMessage(e ErrorItem) string {
	return e.Field + ": " + e.Message
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
