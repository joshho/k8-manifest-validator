package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-21.5 — Phase 3 ISCSI Volume Cross-Struct Validation
// Tests deep nested field traversal: PodSpec → volumes → ISCSI → fields
// Also tests cross-struct reference: PodSpec → containers → volumeMounts → volumes
// Codegen confirmed:
//   ISCSIVolumeSource.TargetPortal  (string, Required)
//   ISCSIVolumeSource.Portals      ([]string, optional)
//   ISCSIVolumeSource.IQN          (string, Required)
//   ISCSIVolumeSource.Lun          (int32, Required)
//   ISCSIVolumeSource.ISCSIInterface (string, optional)
//   ISCSIVolumeSource.SecretRef    (*LocalObjectReference, optional)
//   ISCSIVolumeSource.ChapAuthDiscovery (bool, optional)
//   ISCSIVolumeSource.ChapAuthSession (bool, optional)
//   ISCSIVolumeSource.InitiatorName (*string, optional)
// =============================================================================

// addVol is a deployFn modifier that appends a volume to the deployment's PodSpec.
func iscsiVol(vol *corev1.Volume) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, *vol)
	}
}

// volMounts is a deployFn modifier that appends volumeMounts to the first container.
func iscsiMounts(mounts []corev1.VolumeMount) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].VolumeMounts = append(d.Spec.Template.Spec.Containers[0].VolumeMounts, mounts...)
	}
}


// TestPhase3_ISCSI_Volume tests Volume.ISCSIVolumeSource fields via cross-struct
// traversal: PodSpec → volumes[*] → ISCSI → targetPortal/iqn/lun and optional fields.
func TestPhase3_ISCSI_Volume(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deployFn  func() *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// --- Valid cases ---

		{
			name: "valid — iscsi with required fields only (targetPortal+iqn+lun)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-required",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
							},
						},
					}),
					iscsiMounts([]corev1.VolumeMount{{Name: "iscsi-required", MountPath: "/mnt/iscsi"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — iscsi with all optional fields (portals, chapAuth, secret, interface)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-full",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal:      "192.168.1.100:3260",
								Portals:           []string{"192.168.1.101:3260", "192.168.1.102:3260"},
								IQN:               "iqn.2024-05.com.example:storage.target1",
								Lun:               1,
								ISCSIInterface:    "default",
								SecretRef:         &corev1.LocalObjectReference{Name: "iscsi-secret"},
								CHAPAuthDiscovery: true,
								CHAPAuthSession:   true,
							},
						},
					}),
					iscsiMounts([]corev1.VolumeMount{{Name: "iscsi-full", MountPath: "/mnt/iscsi-full"}}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — iscsi with initiatorName set",
			deployFn: func() *appsv1.Deployment {
				initiatorName := "iqn.2024-05.com.example:initiator.node1"
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-initiator",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal:   "192.168.1.100:3260",
								IQN:            "iqn.2024-05.com.example:storage.target1",
								Lun:            1,
								InitiatorName:  &initiatorName,
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — initContainer referencing ISCSI volume",
			deployFn: func() *appsv1.Deployment {
				initC := corev1.Container{
					Name:  "init-sidecar",
					Image: "busybox:1.36",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "iscsi-init", MountPath: "/init-data"},
					},
				}
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-init",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
							},
						},
					}),
					withInitContainers([]corev1.Container{initC}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — multiple containers each referencing ISCSI volume",
			deployFn: func() *appsv1.Deployment {
				container1 := corev1.Container{
					Name:  "app-a",
					Image: "nginx:1.25",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "iscsi-shared", MountPath: "/data-a"},
					},
				}
				container2 := corev1.Container{
					Name:  "app-b",
					Image: "redis:7.2",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "iscsi-shared", MountPath: "/data-b"},
					},
				}
				return deployment(
					func(d *appsv1.Deployment) {
						d.Spec.Template.Spec.Containers = []corev1.Container{container1, container2}
					},
					iscsiVol(&corev1.Volume{
						Name: "iscsi-shared",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
								ReadOnly:    true,
							},
						},
					}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid — iscsi with fsType and readOnly",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-fstype",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
								FSType:       "xfs",
								ReadOnly:     true,
							},
						},
					}),
					iscsiMounts([]corev1.VolumeMount{{Name: "iscsi-fstype", MountPath: "/mnt/xfs"}}),
				)
			},
			wantErr: false,
		},

		// --- Invalid cases ---

		{
			name: "invalid — iscsi missing targetPortal (empty string)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-no-portal",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "targetPortal",
		},
		{
			name: "invalid — iscsi missing targetPortal (zero value, field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-no-portal-zero",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								IQN: "iqn.2024-05.com.example:storage.target1",
								Lun: 1,
								// TargetPortal not set — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "targetPortal",
		},
		{
			name: "invalid — iscsi missing iqn (empty string)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-no-iqn",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "",
								Lun:          1,
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "iqn",
		},
		{
			name: "invalid — iscsi missing iqn (zero value, field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-no-iqn-zero",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								Lun:          1,
								// IQN not set at all — zero value "" → Required:true → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "iqn",
		},
		{
			name: "invalid — iscsi missing lun (zero value, field absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-no-lun",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								IQN:          "iqn.2024-05.com.example:storage.target1",
								// Lun not set — zero value 0 → Required:true (integer) → error
							},
						},
					}),
				)
			},
			wantErr:   true,
			errSubstr: "lun",
		},
		{
			name: "invalid — iscsi with empty portals slice (different from absent)",
			deployFn: func() *appsv1.Deployment {
				return deployment(
					iscsiVol(&corev1.Volume{
						Name: "iscsi-empty-portals",
						VolumeSource: corev1.VolumeSource{
							ISCSI: &corev1.ISCSIVolumeSource{
								TargetPortal: "192.168.1.100:3260",
								Portals:      []string{},
								IQN:          "iqn.2024-05.com.example:storage.target1",
								Lun:          1,
							},
						},
					}),
				)
			},
			wantErr: false, // Portals is optional, empty slice should be valid
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[Phase3-ISCSI-Volume] testing: %s", tc.name)
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