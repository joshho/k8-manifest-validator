package validator

import (
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =============================================================================
// AWU-21.3 — Phase 3 Projected Volume Tests
// Tests deep nested field traversal: Volume → ProjectedVolumeSource
// Codegen confirmed: ProjectedVolumeSourceDeferredFields.sources (Required:false)
// Codegen confirmed: ProjectedVolumeSourceDeferredFields.defaultMode (Required:false)
// Note: projected volumes can combine configMap, secret, downwardAPI, and
// serviceAccountToken sources in a single Projected volume type.
// =============================================================================

// projVol is a deployFn modifier that attaches a single projected volume to the deployment.
func projVol(vol *corev1.Volume) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, *vol)
	}
}

// projVolMounts is a deployFn modifier that appends volumeMounts to the first container.
func projVolMounts(mounts []corev1.VolumeMount) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, mounts...)
	}
}

// TestPhase3_Volume_Projected tests Volume.ProjectedVolumeSource fields.
func TestPhase3_Volume_Projected(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — projected with configMap source",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-configmap",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "my-configmap"},
									},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-configmap", MountPath: "/etc/config"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected with secret source",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-secret",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
									},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-secret", MountPath: "/etc/secrets"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected with downwardAPI (labels and namespace)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-downwardapi",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										DownwardAPI: &corev1.DownwardAPIProjection{
											Items: []corev1.DownwardAPIVolumeFile{
												{
													Path: "labels",
													FieldRef: &corev1.ObjectFieldSelector{
														APIVersion: "v1",
														FieldPath: "metadata.labels",
													},
												},
												{
													Path: "namespace",
													FieldRef: &corev1.ObjectFieldSelector{
														APIVersion: "v1",
														FieldPath: "metadata.namespace",
													},
												},
											},
										},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-downwardapi", MountPath: "/etc/downward"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected with serviceAccountToken",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-sa-token",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
											Audience:          "api",
											ExpirationSeconds: int64Ptr(3600),
											Path:              "token",
										},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-sa-token", MountPath: "/var/run/secrets/tokens"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected combining multiple sources (configMap, secret, downwardAPI)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-multi",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"},
									},
									},
									{
										Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "app-secret"},
									},
									},
									{
										DownwardAPI: &corev1.DownwardAPIProjection{
											Items: []corev1.DownwardAPIVolumeFile{
												{
													Path: "labels",
													FieldRef: &corev1.ObjectFieldSelector{
														APIVersion: "v1",
														FieldPath: "metadata.labels",
													},
												},
											},
										},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-multi", MountPath: "/etc/combined"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected combining all four source types",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-all-types",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "cfg"},
									},
									},
									{
										Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "sec"},
									},
									},
									{
										DownwardAPI: &corev1.DownwardAPIProjection{
											Items: []corev1.DownwardAPIVolumeFile{
												{
													Path: "podname",
													FieldRef: &corev1.ObjectFieldSelector{
														APIVersion: "v1",
														FieldPath: "metadata.name",
													},
												},
											},
										},
									},
									{
										ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
											Audience:          "api",
											ExpirationSeconds: int64Ptr(3600),
											Path:              "token",
										},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-all-types", MountPath: "/etc/all"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with projected volume",
			deployFn: func() *appsv1.Deployment {
				initC := corev1.Container{
					Name:  "init-sidecar",
					Image: "busybox:1.36",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "projected-init", MountPath: "/init-data"},
					},
				}
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-init",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "init-config"},
									},
									},
								},
							},
						},
					}),
					withInitContainers([]corev1.Container{initC}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected volume with defaultMode set",
			deployFn: func() *appsv1.Deployment {
				defaultMode := int32(0644)
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-mode",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								DefaultMode: &defaultMode,
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "cfg-mode"},
									},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-mode", MountPath: "/cfg"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected with empty sources (minimal valid, sources Required:false)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-empty",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected with items (readOnly mode via items field)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-readonly-items",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
											Name: "readonly-cfg",
											Items: []corev1.KeyToPath{
												{Key: "key1", Path: "file1", Mode: int32(0444)},
											},
										},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-readonly-items", MountPath: "/ro", ReadOnly: true}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — multiple containers each with different projected volumes",
			deployFn: func() *appsv1.Deployment {
				container1 := corev1.Container{
					Name:  "app-a",
					Image: "nginx:1.25",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "projected-cfg-a", MountPath: "/cfg-a"},
					},
				}
				container2 := corev1.Container{
					Name:  "app-b",
					Image: "redis:7.2",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "projected-sec-b", MountPath: "/sec-b"},
					},
				}
				return deployment(
					withContainerName("app-a"),
					func(d *appsv1.Deployment) {
						d.Spec.Template.Spec.Containers = []corev1.Container{container1, container2}
					},
					projVol(&corev1.Volume{
						Name: "projected-cfg-a",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "cfg-a"},
									},
									},
								},
							},
						},
					}),
					projVol(&corev1.Volume{
						Name: "projected-sec-b",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "sec-b"},
									},
									},
								},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected secret with optional field items (specific keys)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-secret-keys",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										Secret: &corev1.SecretProjection{
											Name: "multi-key-secret",
											Items: []corev1.KeyToPath{
												{Key: "username", Path: "user.txt"},
												{Key: "password", Path: "pass.txt"},
											},
										},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-secret-keys", MountPath: "/secrets"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — projected downwardAPI with resource field (cpu limit)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					projVol(&corev1.Volume{
						Name: "projected-resource",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										DownwardAPI: &corev1.DownwardAPIProjection{
											Items: []corev1.DownwardAPIVolumeFile{
												{
													Path: "cpu-limit",
													ResourceFieldRef: &corev1.ResourceFieldSelector{
														Resource: "limits.cpu",
													},
												},
											},
										},
									},
								},
							},
						},
					}),
					projVolMounts([]corev1.VolumeMount{{Name: "projected-resource", MountPath: "/resource"}}),
				)
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-Projected] testing: %s", tc.name)
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

// hasErrErrItems checks if any error in the list contains the given substring
