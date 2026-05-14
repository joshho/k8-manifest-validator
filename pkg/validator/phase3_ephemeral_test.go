package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// =============================================================================
// AWU-21.2 — Phase 3 Ephemeral Volume Tests
// Tests deep nested field traversal: Volume → EphemeralVolumeSource
// Codegen confirmed: EphemeralVolumeSourceDeferredFields.volumeClaimTemplate
// Required:false (volumeClaimTemplate is itself optional per k8s schema)
// Note: ephemeral volumes are referenced by containers via volumeName in
// volumeMounts, but the ephemeral volume source lives in the PodSpec's volumes
// list (not in containers directly). The validation path is:
//   PodSpec → volumes[*] → ephemeral → volumeClaimTemplate
// =============================================================================

// withVolume is a deployFn modifier that attaches a single volume to the deployment.
func withVolume(vol *corev1.Volume) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, *vol)
	}
}

// withInitContainers sets init containers on the pod spec.
func withInitContainers(containers []corev1.Container) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.InitContainers = containers
	}
}


// resourceMustParse is a test helper to parse quantity strings.
func resourceMustParse(s string) *resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return &q
}

// withVolumeMounts is a deployFn modifier that appends volumeMounts to the first container.
func withVolumeMounts(mounts []corev1.VolumeMount) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, mounts...)
	}
}

// TestPhase3_Volume_Ephemeral tests Volume.EphemeralVolumeSource fields.
// volumeClaimTemplate is Required:false per codegen (can be absent/empty).
func TestPhase3_Volume_Ephemeral(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid — ephemeral with volumeClaimTemplate (genericEphemeralVolume)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ephemeral-config",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
									Spec: corev1.PersistentVolumeClaimSpec{
										AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												corev1.ResourceStorage: resourceMustParse("10Mi"),
											},
										},
									},
								},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "ephemeral-config", MountPath: "/etc/config"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — ephemeral with readOnly: true",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ephemeral-readonly",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								ReadOnly: true,
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
									Spec: corev1.PersistentVolumeClaimSpec{
										AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany},
									},
								},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "ephemeral-readonly", MountPath: "/data", ReadOnly: true}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — ephemeral volumeClaimTemplate with empty spec (minimal valid)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ephemeral-minimal",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{},
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — ephemeral absent volumeClaimTemplate (not required per codegen)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ephemeral-no-claim",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								// volumeClaimTemplate is nil — Required:false per codegen
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — ephemeral with readOnly: false explicitly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ephemeral-rw",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								ReadOnly: false,
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
									Spec: corev1.PersistentVolumeClaimSpec{
										AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
									},
								},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "ephemeral-rw", MountPath: "/shared"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer with ephemeral volume",
			deployFn: func() *appsv1.Deployment {
				initC := corev1.Container{
					Name:  "init-sidecar",
					Image: "busybox:1.36",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "ephemeral-init", MountPath: "/init-data"},
					},
				}
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ephemeral-init",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
									Spec: corev1.PersistentVolumeClaimSpec{
										AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
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
			name: "valid — multiple containers each with different ephemeral volume types",
			deployFn: func() *appsv1.Deployment {
				container1 := corev1.Container{
					Name:  "app-a",
					Image: "nginx:1.25",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "ephemeral-a", MountPath: "/data-a"},
					},
				}
				container2 := corev1.Container{
					Name:  "app-b",
					Image: "redis:7.2",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "ephemeral-b", MountPath: "/data-b"},
					},
				}
				return deployment(
					withContainerName("app-a"),
					func(d *appsv1.Deployment) {
						d.Spec.Template.Spec.Containers = []corev1.Container{container1, container2}
					},
					withVolume(&corev1.Volume{
						Name: "ephemeral-a",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								ReadOnly: false,
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
									Spec: corev1.PersistentVolumeClaimSpec{
										AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
									},
								},
							},
						},
					}),
					withVolume(&corev1.Volume{
						Name: "ephemeral-b",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								ReadOnly: true,
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
									Spec: corev1.PersistentVolumeClaimSpec{
										AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany},
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
			name: "valid — ephemeral with volumeClaimTemplate.metadata.name set",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					withVolume(&corev1.Volume{
						Name: "ephemeral-named",
						VolumeSource: corev1.VolumeSource{
							Ephemeral: &corev1.EphemeralVolumeSource{
								VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
									ObjectMeta: metav1.ObjectMeta{
										Name: "my-claim-template",
									},
									Spec: corev1.PersistentVolumeClaimSpec{
										AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
									},
								},
							},
						},
					}),
					withVolumeMounts([]corev1.VolumeMount{{Name: "ephemeral-named", MountPath: "/mnt/named"}}),
				)
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-Volume-Ephemeral] testing: %s", tc.name)
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