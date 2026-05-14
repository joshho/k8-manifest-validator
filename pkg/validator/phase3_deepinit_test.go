package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// =============================================================================
// AWU-19.1 — Phase 3 Deep Nested initContainers Tests
// Tests deep nested field traversal for initContainers: name, env, ports,
// volumeMounts, volumeDevices, livenessProbe, readinessProbe, startupProbe,
// and lifecycle (postStart/preStop).
// Codegen confirmed per deferredFieldIndex for initContainers:
//   - .initContainers[*].name                   Required:true
//   - .initContainers[*].env[*].name             Required:true
//   - .initContainers[*].ports[*].containerPort Required:true
//   - .initContainers[*].volumeMounts[*].mountPath Required:true
//   - .initContainers[*].volumeMounts[*].name    Required:true
//   - .initContainers[*].volumeDevices[*].devicePath Required:true
//   - .initContainers[*].volumeDevices[*].name   Required:true
//   - .initContainers[*].livenessProbe.tcpSocket.port  Required:true
//   - .initContainers[*].readinessProbe.tcpSocket.port Required:true
//   - .initContainers[*].startupProbe.tcpSocket.port   Required:true
//   - .initContainers[*].livenessProbe.httpGet.port   Required:true
//   - .initContainers[*].readinessProbe.httpGet.port  Required:true
//   - .initContainers[*].startupProbe.httpGet.port    Required:true
//   - .initContainers[*].lifecycle.postStart.httpGet.port  Required:true
//   - .initContainers[*].lifecycle.preStop.httpGet.port    Required:true
//   - .initContainers[*].lifecycle.postStart.tcpSocket.port Required:true
//   - .initContainers[*].lifecycle.preStop.tcpSocket.port   Required:true
// =============================================================================

// withInitContainers is a deployFn modifier that sets initContainers on the pod spec.
func withInitContainers(containers []corev1.Container) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.InitContainers = containers
	}
}

// TestPhase3_DeepInit_Name tests initContainer name field (Required:true).
func TestPhase3_DeepInit_Name(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — initContainer with name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{Name: "init-container", Image: "busybox:1.36"},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainer name empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{Name: "", Image: "busybox:1.36"},
				}))
			},
			wantErr:   true,
			errSubstr: "name",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepInit] testing name: %s", tc.name)
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

// TestPhase3_DeepInit_Env tests initContainer env field traversal.
func TestPhase3_DeepInit_Env(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — initContainer env[*].name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{
						Name:  "init-container",
						Image: "busybox:1.36",
						Env: []corev1.EnvVar{
							{Name: "FOO", Value: "bar"},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainer env[*].name empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{
						Name:  "init-container",
						Image: "busybox:1.36",
						Env: []corev1.EnvVar{
							{Name: "", Value: "bar"},
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "name",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepInit] testing env: %s", tc.name)
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

// TestPhase3_DeepInit_Ports tests initContainer ports field traversal.
func TestPhase3_DeepInit_Ports(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — initContainer ports[*].containerPort set",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{
						Name:  "init-container",
						Image: "busybox:1.36",
						Ports: []corev1.ContainerPort{
							{Name: "http", ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainer ports[*].containerPort zero",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{
						Name:  "init-container",
						Image: "busybox:1.36",
						Ports: []corev1.ContainerPort{
							{Name: "http", ContainerPort: 0, Protocol: corev1.ProtocolTCP},
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "containerPort",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepInit] testing ports: %s", tc.name)
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

// TestPhase3_DeepInit_VolumeMounts tests initContainer volumeMounts field traversal.
func TestPhase3_DeepInit_VolumeMounts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — initContainer volumeMounts[*].mountPath and .name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "data-volume", MountPath: "/data"},
							},
						},
					}),
					withVolumes([]corev1.Volume{{Name: "data-volume"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainer volumeMounts[*].mountPath empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{
							Name:  "init-container",
							Image: "busybox:1.36",
							VolumeMounts: []corev1.VolumeMount{
								{Name: "data-volume", MountPath: ""},
							},
						},
					}),
					withVolumes([]corev1.Volume{{Name: "data-volume"}}),
				)
			},
			wantErr:   true,
			errSubstr: "mountPath",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepInit] testing volumeMounts: %s", tc.name)
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

// TestPhase3_DeepInit_VolumeDevices tests initContainer volumeDevices field traversal.
func TestPhase3_DeepInit_VolumeDevices(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — initContainer volumeDevices[*].devicePath and .name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{
						Name:  "init-container",
						Image: "busybox:1.36",
						VolumeDevices: []corev1.VolumeDevice{
							{Name: "data-device", DevicePath: "/dev/sda1"},
						},
					},
				}))
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainer volumeDevices[*].devicePath empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(withInitContainers([]corev1.Container{
					{
						Name:  "init-container",
						Image: "busybox:1.36",
						VolumeDevices: []corev1.VolumeDevice{
							{Name: "data-device", DevicePath: ""},
						},
					},
				}))
			},
			wantErr:   true,
			errSubstr: "devicePath",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepInit] testing volumeDevices: %s", tc.name)
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

// initContainerLivenessProbe sets livenessProbe on the initContainer.
func initContainerLivenessProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if len(d.Spec.Template.Spec.InitContainers) > 0 {
			d.Spec.Template.Spec.InitContainers[0].LivenessProbe = probe
		}
	}
}

// initContainerReadinessProbe sets readinessProbe on the initContainer.
func initContainerReadinessProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if len(d.Spec.Template.Spec.InitContainers) > 0 {
			d.Spec.Template.Spec.InitContainers[0].ReadinessProbe = probe
		}
	}
}

// initContainerStartupProbe sets startupProbe on the initContainer.
func initContainerStartupProbe(probe *corev1.Probe) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if len(d.Spec.Template.Spec.InitContainers) > 0 {
			d.Spec.Template.Spec.InitContainers[0].StartupProbe = probe
		}
	}
}

// initContainerLifecycle sets lifecycle on the initContainer.
func initContainerLifecycle(lifecycle *corev1.Lifecycle) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if len(d.Spec.Template.Spec.InitContainers) > 0 {
			d.Spec.Template.Spec.InitContainers[0].Lifecycle = lifecycle
		}
	}
}

// TestPhase3_DeepInit_Probes tests initContainer liveness/readiness/startup probe fields.
func TestPhase3_DeepInit_Probes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// LivenessProbe tcpSocket port
		{
			name: "valid — initContainers livenessProbe.tcpSocket.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLivenessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers livenessProbe.tcpSocket.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLivenessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// LivenessProbe httpGet port
		{
			name: "valid — initContainers livenessProbe.httpGet.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLivenessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/health",
								Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers livenessProbe.httpGet.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLivenessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/health",
								Port:   intstr.IntOrString{},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// ReadinessProbe tcpSocket port
		{
			name: "valid — initContainers readinessProbe.tcpSocket.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerReadinessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers readinessProbe.tcpSocket.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerReadinessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// ReadinessProbe httpGet port
		{
			name: "valid — initContainers readinessProbe.httpGet.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerReadinessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/ready",
								Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers readinessProbe.httpGet.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerReadinessProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/ready",
								Port:   intstr.IntOrString{},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// StartupProbe tcpSocket port
		{
			name: "valid — initContainers startupProbe.tcpSocket.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerStartupProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers startupProbe.tcpSocket.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerStartupProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// StartupProbe httpGet port
		{
			name: "valid — initContainers startupProbe.httpGet.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerStartupProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/started",
								Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers startupProbe.httpGet.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerStartupProbe(&corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/started",
								Port:   intstr.IntOrString{},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepInit] testing probes: %s", tc.name)
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

// TestPhase3_DeepInit_Lifecycle tests initContainer lifecycle postStart/preStop fields.
func TestPhase3_DeepInit_Lifecycle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// Lifecycle postStart httpGet port
		{
			name: "valid — initContainers lifecycle.postStart.httpGet.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PostStart: &corev1.LifecycleHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/ready",
								Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers lifecycle.postStart.httpGet.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PostStart: &corev1.LifecycleHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/ready",
								Port:   intstr.IntOrString{},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// Lifecycle preStop httpGet port
		{
			name: "valid — initContainers lifecycle.preStop.httpGet.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/shutdown",
								Port:   intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers lifecycle.preStop.httpGet.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "localhost",
								Path:   "/shutdown",
								Port:   intstr.IntOrString{},
								Scheme: "HTTP",
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// Lifecycle postStart tcpSocket port
		{
			name: "valid — initContainers lifecycle.postStart.tcpSocket.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PostStart: &corev1.LifecycleHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers lifecycle.postStart.tcpSocket.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PostStart: &corev1.LifecycleHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
		// Lifecycle preStop tcpSocket port
		{
			name: "valid — initContainers lifecycle.preStop.tcpSocket.port set",
			deployFn: func() *appsv1.Deployment {
				portInt := int32(8080)
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{Type: intstr.Int, IntVal: portInt},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "invalid — initContainers lifecycle.preStop.tcpSocket.port empty",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withInitContainers([]corev1.Container{
						{Name: "init-container", Image: "busybox:1.36"},
					}),
					initContainerLifecycle(&corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{},
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "port",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-DeepInit] testing lifecycle: %s", tc.name)
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