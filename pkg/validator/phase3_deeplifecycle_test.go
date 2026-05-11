package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// =============================================================================
// AWU-17.4 — Phase 3 Deep Nested Lifecycle Handler sub-fields Tests
// Tests deep nested field traversal: Container.Lifecycle → PostStart/PreStop
// → ExecAction.command / HTTPGetAction.host+path+port+scheme / TCPSocketAction.port
// Codegen confirmed: TCPSocketAction.Port Required:true
// =============================================================================

// withLifecycle is a deployFn modifier that sets the container lifecycle hook.
func withLifecycle(lifecycle *corev1.Lifecycle) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Lifecycle = lifecycle
	}
}

// TestPhase3_DeepNested_Lifecycle tests deep nested Lifecycle handler fields:
// Lifecycle.PostStart.ExecAction.Command (Required:true per ExecAction schema)
// Lifecycle.PreStop.HTTPGetAction (host, path, port, scheme)
// Lifecycle.PreStop.TCPSocketAction.Port (Required:true per TCPSocketAction schema)
func TestPhase3_DeepNested_Lifecycle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		deployFn   func() *appsv1.Deployment
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "valid — postStart.exec.command with element",
			deployFn: func() *appsv1.Deployment {
				return deployment(withLifecycle(&corev1.Lifecycle{
					PostStart: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/sh", "-c", "echo alive"},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "valid — preStop.httpGet with host/path/port/scheme",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(withLifecycle(&corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "localhost",
							Path:   "/shutdown",
							Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							Scheme: "HTTP",
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — preStop.tcpSocket.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(withLifecycle(&corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.IntOrString{}, // empty = zero value = error (Required:true)
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "invalid — preStop.tcpSocket port zero value (field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(withLifecycle(&corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						TCPSocket: &corev1.TCPSocketAction{
							// Port field not set at all (zero value)
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "valid — postStart.httpGet with all fields",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(9090)
				return deployment(withLifecycle(&corev1.Lifecycle{
					PostStart: &corev1.LifecycleHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "0.0.0.0",
							Path:   "/ready",
							Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							Scheme: "HTTPS",
						},
					},
				}))
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepLifecycle] testing: %s", tc.name)
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
