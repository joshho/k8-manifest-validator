package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-17.1 — Phase 3 Deep Nested EnvVar.valueFrom Tests
// Tests deep nested field traversal: EnvVar → valueFrom → SecretKeyRef / ConfigMapKeyRef → key
// Codegen confirmed: SecretKeySelectorDeferredFields.key Required:true,
//                    ConfigMapKeySelectorDeferredFields.key Required:true
// =============================================================================

var (
	boolFalse = false
	boolTrue  = true
)

// TestPhase3_DeepNested_EnvVar_ValueFrom tests the deep nested chain:
// EnvVar.ValueFrom.SecretKeyRef.key and EnvVar.ValueFrom.ConfigMapKeyRef.key
// Both are Required:true per codegen.
func TestPhase3_DeepNested_EnvVar_ValueFrom(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		deployFn   func() *appsv1.Deployment
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "valid — secretKeyRef complete",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnv([]corev1.EnvVar{
					{
						Name: "SECRET_NAME",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
								Key:                  "username",
								Optional:             &boolFalse,
							},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — secretKeyRef missing key (empty string)",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnv([]corev1.EnvVar{
					{
						Name: "SECRET_NAME",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
								Key:                  "", // empty string — Required:true → error
								Optional:             &boolFalse,
							},
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "key",
		},
		{
			name: "invalid — secretKeyRef key field absent (zero value)",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnv([]corev1.EnvVar{
					{
						Name: "SECRET_NAME",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
								// Key field not set at all — zero value "" → Required:true → error
								Optional: &boolFalse,
							},
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "key",
		},
		{
			name: "valid — configMapKeyRef complete",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnv([]corev1.EnvVar{
					{
						Name: "CONFIG_NAME",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
								Key:                  "config-key",
								Optional:             &boolFalse,
							},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — configMapKeyRef missing key (empty string)",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnv([]corev1.EnvVar{
					{
						Name: "CONFIG_NAME",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
								Key:                  "", // empty string — Required:true → error
								Optional:             &boolFalse,
							},
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "key",
		},
		{
			name: "valid — secretKeyRef optional name empty but key is set",
			deployFn: func() *appsv1.Deployment {
				return deployment(withEnv([]corev1.EnvVar{
					{
						Name: "SECRET_NAME",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: ""}, // name Optional — empty is okay
								Key:                  "username",                           // key Required:true and set → passes
								Optional:             &boolTrue,
							},
						},
					},
				}))
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepEnv] testing: %s", tc.name)
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