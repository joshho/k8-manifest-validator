package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// =============================================================================
// AWU-17.5 — Phase 3 Deep Nested Probe sub-fields Tests
// Tests deep nested field traversal: Container.LivenessProbe/ReadinessProbe/StartupProbe
// → exec/httpGet/tcpSocket sub-fields
// Codegen confirmed: tcpSocket.port Required:true, httpGet.port Required:true
// =============================================================================

// withLivenessProbe is a deployFn modifier that sets the container liveness probe.
func withLivenessProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].LivenessProbe = probe
	}
}

// withReadinessProbe is a deployFn modifier that sets the container readiness probe.
func withReadinessProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].ReadinessProbe = probe
	}
}

// withStartupProbe is a deployFn modifier that sets the container startup probe.
func withStartupProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].StartupProbe = probe
	}
}

// TestPhase3_DeepNested_Probe tests deep nested probe handler fields:
// Container.LivenessProbe/ReadinessProbe/StartupProbe → ExecAction.Command / HTTPGetAction.host+path+port+scheme / TCPSocketAction.Port
// Codegen confirmed per deferredFieldIndex:
//   - .livenessProbe.tcpSocket.port   Required:true
//   - .livenessProbe.httpGet.port     Required:true
//   - .livenessProbe.exec.command     Type:array, Required:false
//   - .readinessProbe.tcpSocket.port Required:true
//   - .readinessProbe.httpGet.port   Required:true
//   - .readinessProbe.exec.command   Type:array, Required:false
//   - .startupProbe.tcpSocket.port   Required:true
//   - .startupProbe.httpGet.port      Required:true
//   - .startupProbe.exec.command     Type:array, Required:false
func TestPhase3_DeepNested_Probe(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — livenessProbe.tcpSocket.port set as int",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(withLivenessProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — livenessProbe.tcpSocket.port empty/zero",
			deployFn: func() *appsv1.Deployment {
				return deployment(withLivenessProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
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
			name: "valid — readinessProbe.httpGet with host/path/port/scheme",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(withReadinessProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "localhost",
							Path:   "/ready",
							Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							Scheme: "HTTP",
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "valid — startupProbe.exec.command with element",
			deployFn: func() *appsv1.Deployment {
				return deployment(withStartupProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/sh", "-c", "cat /tmp/healthy"},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — livenessProbe.httpGet.port empty/zero",
			deployFn: func() *appsv1.Deployment {
				return deployment(withLivenessProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "localhost",
							Path:   "/health",
							Port:   intstr.IntOrString{}, // empty = zero value = error (Required:true)
							Scheme: "HTTP",
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "valid — livenessProbe with all optional fields set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(withLivenessProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "0.0.0.0",
							Path:   "/health",
							Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							Scheme: "HTTPS",
						},
					},
					InitialDelaySeconds: 10,
					PeriodSeconds:       5,
					TimeoutSeconds:      3,
					SuccessThreshold:    1,
					FailureThreshold:    3,
				}))
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepProbe] testing: %s", tc.name)
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
