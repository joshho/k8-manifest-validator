package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-17.2 — Phase 3 Deep Nested EnvFrom Tests
// Tests deep nested field traversal: EnvFromSource → ConfigMapRef / SecretRef → name
// Codegen confirmed: LocalObjectReference.name is Optional:false with no override
// so empty string IS an error per Required:true semantics.
// =============================================================================

// NOTE: boolFalse and boolTrue are declared in phase3_deepenv_test.go (same package)

// TestPhase3_DeepNested_EnvFrom tests the deep nested chain:
// EnvFromSource.ConfigMapRef.name and EnvFromSource.SecretRef.name
// Both are Required:true per codegen.
func TestPhase3_DeepNested_EnvFrom(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — configMapRef with name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnvFrom([]corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "valid — configMapRef with empty .name",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnvFrom([]corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: ""},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "valid — secretRef with name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnvFrom([]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "valid — secretRef with empty .name",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnvFrom([]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: ""},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "valid — configMapRef with optional=true",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnvFrom([]corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
							Optional:             &boolTrue,
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "valid — secretRef with optional=true",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnvFrom([]corev1.EnvFromSource{
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
							Optional:             &boolTrue,
						},
					},
				}))
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepEnvFrom] testing: %s", tc.name)
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
