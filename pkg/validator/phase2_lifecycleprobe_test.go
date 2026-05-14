package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// =============================================================================
// AWU-20.4 — Phase 2: Lifecycle hooks + Probe httpGet.port validation
// PRD ref: plan_main.json phase_2 — AWU-20.4
// Codegen entries: lifecycle.postStart/exec, lifecycle.preStop/exec,
//   startupProbe/httpGet.port, readinessProbe/httpGet.port, livenessProbe/httpGet.port
// Validation wire: validateProbe in builtin.go, validateLifecycle in builtin.go,
//   validateLifecycleHookHandlerCount in semantic_lifecycle.go
// Existing coverage: phase3_deepprobe_test.go (port empty/zero), semantic_lifecycle_test.go (handler count)
// AWU-20.4 complement: port range (>65535, negative), named port, lifecycle invalid handler types,
//   probe with grpc handler, multiple containers, all 3 probes together
// =============================================================================

// -----------------------------------------------------------------------
// AWU-20.4 — Lifecycle hook + Probe httpGet.port edge-condition coverage
// -----------------------------------------------------------------------

func TestPhase2_LifecycleProbeHttpGetPort_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// ---- Valid: lifecycle handlers (exec, httpGet, tcpSocket) ----

		{
			name: "valid lifecycle preStop exec",
			deploy: deployment(withContainerName("test"), withLifecyclePreStop(&corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "graceful-shutdown"}},
			})),
			wantErr: false,
		},
		{
			name: "valid lifecycle postStart exec",
			deploy: deployment(withContainerName("test"), withLifecyclePostStart(&corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{Command: []string{"echo", "started"}},
			})),
			wantErr: false,
		},
		{
			name: "valid lifecycle preStop httpGet",
			deploy: deployment(withContainerName("test"), withLifecyclePreStop(&corev1.LifecycleHandler{
				HTTPGet: &corev1.HTTPGetAction{Path: "/shutdown", Port: intstr.FromInt(8080)},
			})),
			wantErr: false,
		},
		{
			name: "valid lifecycle postStart httpGet",
			deploy: deployment(withContainerName("test"), withLifecyclePostStart(&corev1.LifecycleHandler{
				HTTPGet: &corev1.HTTPGetAction{Path: "/start", Port: intstr.FromInt(9090)},
			})),
			wantErr: false,
		},
		{
			name: "valid lifecycle preStop tcpSocket",
			deploy: deployment(withContainerName("test"), withLifecyclePreStop(&corev1.LifecycleHandler{
				TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(8080)},
			})),
			wantErr: false,
		},
		{
			name: "valid lifecycle postStart tcpSocket",
			deploy: deployment(withContainerName("test"), withLifecyclePostStart(&corev1.LifecycleHandler{
				TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(9090)},
			})),
			wantErr: false,
		},

		// ---- Valid: httpGet.port with numeric port, named port, valid host and path ----

		{
			name: "valid httpGet.port numeric (int32)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/healthz",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid httpGet.port named string (IntOrString string type)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "my-service",
						Path:   "/healthz",
						Port:   intstr.IntOrString{Type: intstr.String, StrVal: "web"},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid httpGet with all fields: host, path, port, scheme",
			deploy: deployment(withContainerName("test"), withReadinessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "my-app.example.com",
						Path:   "/api/ready",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8443},
						Scheme: "HTTPS",
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid startupProbe httpGet.port numeric",
			deploy: deployment(withContainerName("test"), withStartupProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/startup",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: httpGet.port empty/zero ----

		{
			name: "invalid httpGet.port empty IntOrString (zero value)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/health",
						Port:   intstr.IntOrString{}, // zero value = empty = error (Required:true)
						Scheme: "HTTP",
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "invalid livenessProbe.httpGet.port empty (IntOrString zero)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.IntOrString{},
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "invalid readinessProbe.httpGet.port empty",
			deploy: deployment(withContainerName("test"), withReadinessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.IntOrString{},
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "invalid startupProbe.httpGet.port empty",
			deploy: deployment(withContainerName("test"), withStartupProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.IntOrString{},
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},

		// ---- Invalid: httpGet.port port number > 65535 ----

		{
			name: "invalid httpGet.port int value > 65535",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/health",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 70000},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "invalid httpGet.port int value exactly 65536 (boundary)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.IntOrString{Type: intstr.Int, IntVal: 65536},
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},

		// ---- Invalid: httpGet.port negative port ----

		{
			name: "invalid httpGet.port int value negative (-1)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/health",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: -1},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "invalid httpGet.port int value negative (-100)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/health",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: -100},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr:   true,
			errSubstr: "port",
		},

		// ---- Invalid: lifecycle hook with zero handlers (empty LifecycleHandler {}) ----

		{
			name: "invalid lifecycle preStop with zero handlers",
			deploy: deployment(withContainerName("test"), withLifecyclePreStop(&corev1.LifecycleHandler{})),
			wantErr:   true,
			errSubstr: "exactly one",
		},
		{
			name: "invalid lifecycle postStart with zero handlers",
			deploy: deployment(withContainerName("test"), withLifecyclePostStart(&corev1.LifecycleHandler{})),
			wantErr:   true,
			errSubstr: "exactly one",
		},

		// ---- Invalid: lifecycle hook with multiple handlers ----

		{
			name: "invalid lifecycle preStop with multiple handlers (exec + httpGet)",
			deploy: deployment(withContainerName("test"), withLifecyclePreStop(&corev1.LifecycleHandler{
				Exec:    &corev1.ExecAction{Command: []string{"echo"}},
				HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(8080)},
			})),
			wantErr:   true,
			errSubstr: "more than one",
		},
		{
			name: "invalid lifecycle postStart with multiple handlers (tcpSocket + exec)",
			deploy: deployment(withContainerName("test"), withLifecyclePostStart(&corev1.LifecycleHandler{
				TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(8080)},
				Exec:     &corev1.ExecAction{Command: []string{"echo"}},
			})),
			wantErr:   true,
			errSubstr: "more than one",
		},

		// ---- Valid: probes with all handler types ----

		{
			name: "valid probe with exec handler",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "cat /tmp/healthy"}},
				},
				InitialDelaySeconds: 10,
				PeriodSeconds:       5,
			})),
			wantErr: false,
		},
		{
			name: "valid probe with tcpSocket handler",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(8080)},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid probe with httpGet handler (full fields)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "0.0.0.0",
						Path:   "/health",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
						Scheme: "HTTP",
					},
				},
				InitialDelaySeconds: 5,
				PeriodSeconds:       10,
				TimeoutSeconds:      3,
				SuccessThreshold:    1,
				FailureThreshold:    3,
			})),
			wantErr: false,
		},
		{
			name: "valid probe with grpc handler",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					GRPC: &corev1.GRPCAction{Port: 50051},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: probe with zero handlers ----

		{
			name: "invalid probe with zero handlers",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{},
			})),
			wantErr:   true,
			errSubstr: "exactly one",
		},

		// ---- Invalid: probe with multiple handlers ----

		{
			name: "invalid probe with multiple handlers (httpGet + exec)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(8080)},
					Exec:     &corev1.ExecAction{Command: []string{"echo"}},
				},
			})),
			wantErr:   true,
			errSubstr: "only one",
		},
		{
			name: "invalid probe with all 4 handlers (httpGet + tcpSocket + exec + grpc)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet:  &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(8080)},
					TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(8081)},
					Exec:     &corev1.ExecAction{Command: []string{"echo"}},
					GRPC:     &corev1.GRPCAction{Port: 50051},
				},
			})),
			wantErr:   true,
			errSubstr: "only one",
		},

		// ---- httpGet port boundary: exactly 65535 (max valid) ----

		{
			name: "valid httpGet.port int value exactly 65535 (max valid port)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/health",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 65535},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr: false,
		},

		// ---- httpGet port boundary: exactly 1 (min valid port) ----

		{
			name: "valid httpGet.port int value exactly 1 (min valid port)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Host:   "localhost",
						Path:   "/health",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 1},
						Scheme: "HTTP",
					},
				},
			})),
			wantErr: false,
		},

		// ---- Valid: tcpSocket port with numeric port ----

		{
			name: "valid tcpSocket port (int32 8080)",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid tcpSocket port named string",
			deploy: deployment(withContainerName("test"), withLivenessProbe(&corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{Type: intstr.String, StrVal: "web"},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Valid: all three probe types on same container (liveness, readiness, startup) ----

		{
			name: "valid container with all three probes (liveness, readiness, startup)",
			deploy: deployment(withContainerName("test"),
				withLivenessProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "localhost",
							Path:   "/health",
							Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
							Scheme: "HTTP",
						},
					},
				}),
				withReadinessProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "localhost",
							Path:   "/ready",
							Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8081},
							Scheme: "HTTP",
						},
					},
				}),
				withStartupProbe(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Host:   "localhost",
							Path:   "/startup",
							Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 8082},
							Scheme: "HTTP",
						},
					},
				}),
			),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateDeploymentRaw(tc.deploy)
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