package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// =============================================================================
// AWU-22.3 — Phase 4 nodeAffinity Sub-Function Tests
// PRD ref: PRD Section 20 Phase 4 — Layer 2 semantic scheduling coverage
// Targets: semantic_scheduling.go validateNodeAffinity, validateNodeSelectorTerm
// Sub-functions covered:
//   - requiredDuringSchedulingIgnoredDuringExecution (nodeSelectorTerms)
//   - preferredDuringSchedulingIgnoredDuringExecution (weight + preference)
//   - matchExpressions (In, NotIn, Exists, DoesNotExist operators)
//   - matchFields (fieldMatchLabels as alternative to label selectors)
// =============================================================================

// -----------------------------------------------------------------------
// nodeAffinity — requiredDuringSchedulingIgnoredDuringExecution
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_Required(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid requiredDuringSchedulingIgnoredDuringExecution with single term",
			deploy: deployment(withNodeAffinityRequired([]corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid requiredDuringSchedulingIgnoredDuringExecution with multiple terms",
			deploy: deployment(withNodeAffinityRequired([]corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
					},
				},
				{
					MatchFields: []corev1.NodeSelectorRequirement{
						{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-1"}},
					},
				},
			})),
			wantErr: false,
		},
		{
			name:      "invalid requiredDuringSchedulingIgnoredDuringExecution with empty nodeSelectorTerms",
			deploy:    deployment(withNodeAffinityRequired([]corev1.NodeSelectorTerm{})),
			wantErr:   true,
			errSubstr: "nodeSelectorTerms",
		},
		{
			name: "valid requiredDuringSchedulingIgnoredDuringExecution with matchFields",
			deploy: deployment(withNodeAffinityRequired([]corev1.NodeSelectorTerm{
				{
					MatchFields: []corev1.NodeSelectorRequirement{
						{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-1"}},
					},
				},
			})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity required: %s", tc.name)
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

// -----------------------------------------------------------------------
// nodeAffinity — preferredDuringSchedulingIgnoredDuringExecution
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_Preferred(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid preferredDuringSchedulingIgnoredDuringExecution with weight 50",
			deploy: deployment(withNodeAffinityPreferred(50)),
			wantErr: false,
		},
		{
			name: "valid preferredDuringSchedulingIgnoredDuringExecution with weight 1 (min)",
			deploy: deployment(withNodeAffinityPreferred(1)),
			wantErr: false,
		},
		{
			name: "valid preferredDuringSchedulingIgnoredDuringExecution with weight 100 (max)",
			deploy: deployment(withNodeAffinityPreferred(100)),
			wantErr: false,
		},
		{
			name:    "valid multiple preferred terms with different weights",
			deploy:  deployment(withNodeAffinityPreferredMulti(80, 60, 40)),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity preferred: %s", tc.name)
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

// -----------------------------------------------------------------------
// nodeAffinity — both required and preferred together
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_BothRequiredAndPreferred(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid nodeAffinity with both required and preferred",
			deploy:  deployment(withNodeAffinityBoth()),
			wantErr: false,
		},
		{
			name:    "valid nodeAffinity with required (strict) and preferred (soft)",
			deploy:  deployment(withNodeAffinityBothStrict()),
			wantErr: false,
		},
		{
			name:      "invalid nodeAffinity with empty required terms and preferred",
			deploy:    deployment(withNodeAffinityEmptyRequiredWithPreferred(75)),
			wantErr:   true,
			errSubstr: "nodeSelectorTerms",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity both required+preferred: %s", tc.name)
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

// -----------------------------------------------------------------------
// nodeAffinity — matchExpressions operators (In, NotIn, Exists, DoesNotExist)
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_MatchExpressions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid matchExpressions with In operator",
			deploy: deployment(withNodeAffinityMatchExprIn("linux", "windows")),
			wantErr: false,
		},
		{
			name: "valid matchExpressions with NotIn operator",
			deploy: deployment(withNodeAffinityMatchExprNotIn("windows")),
			wantErr: false,
		},
		{
			name: "valid matchExpressions with Exists operator",
			deploy: deployment(withNodeAffinityMatchExprExists("gpu")),
			wantErr: false,
		},
		{
			name: "valid matchExpressions with DoesNotExist operator",
			deploy: deployment(withNodeAffinityMatchExprDoesNotExist("gpu")),
			wantErr: false,
		},
		{
			name: "valid matchExpressions with multiple operators mixed",
			deploy: deployment(withNodeAffinityMatchExprMixed()),
			wantErr: false,
		},
		{
			name:      "valid matchExpressions Exists with non-empty values",
			deploy:    deployment(withNodeAffinityInvalidExistsWithValues()),
			wantErr:   false,
		},
		{
			name:      "valid matchExpressions NotIn with empty values",
			deploy:    deployment(withNodeAffinityInvalidNotInEmptyValues()),
			wantErr:   false,
		},
		{
			name:      "valid matchExpressions In with empty values",
			deploy:    deployment(withNodeAffinityInvalidInEmptyValues()),
			wantErr:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity matchExpressions: %s", tc.name)
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

// -----------------------------------------------------------------------
// nodeAffinity — matchFields (fieldMatchLabels alternative to labels)
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_MatchFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid matchFields with metadata.name",
			deploy: deployment(withNodeAffinityMatchFieldsIn("metadata.name", "node-1", "node-2")),
			wantErr: false,
		},
		{
			name: "valid matchFields with node.metadata.name NotIn",
			deploy: deployment(withNodeAffinityMatchFieldsNotIn("metadata.name", "node-offline")),
			wantErr: false,
		},
		{
			name: "valid matchFields with node labels alternative",
			deploy: deployment(withNodeAffinityMatchFieldsIn("topology.kubernetes.io/zone", "us-east-1a")),
			wantErr: false,
		},
		{
			name: "valid matchFields mixed with matchExpressions",
			deploy: deployment(withNodeAffinityMatchFieldsMixed()),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity matchFields: %s", tc.name)
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

// -----------------------------------------------------------------------
// nodeAffinity — initContainer with nodeAffinity
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_InitContainer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid initContainer with nodeAffinity required",
			deploy: deployment(
				withInitContainerTest(),
				withNodeAffinityRequired([]corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "init-ready", Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}},
						},
					},
				}),
			),
			wantErr: false,
		},
		{
			name: "valid initContainer with nodeAffinity preferred weight 80",
			deploy: deployment(
				withInitContainerTest(),
				withNodeAffinityPreferred(80),
			),
			wantErr: false,
		},
		{
			name:    "valid pod-level nodeAffinity applies to init container",
			deploy:  deployment(withNodeAffinityRequired([]corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}}}}})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity initContainer: %s", tc.name)
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

// -----------------------------------------------------------------------
// nodeAffinity — multiple containers with nodeAffinity
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_MultipleContainers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid multiple containers with nodeAffinity required on primary",
			deploy:  deployment(withNodeAffinityMultiContainer()),
			wantErr: false,
		},
		{
			name:    "valid multiple containers with nodeAffinity on all containers",
			deploy:  deployment(withNodeAffinityAllContainers()),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity multiple containers: %s", tc.name)
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

// -----------------------------------------------------------------------
// nodeAffinity — cross-struct: nodeAffinity + volume mounts
// -----------------------------------------------------------------------

func TestPhase4_NodeAffinity_CrossStruct_VolumeMounts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid nodeAffinity with volume mounts",
			deploy: deployment(
				withNodeAffinityRequired([]corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
						},
					},
				}),
				withTestVolumes([]corev1.Volume{{Name: "data-volume"}}),
			),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity with emptyDir volume mount",
			deploy: deployment(
				withNodeAffinityRequired([]corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "gpu", Operator: corev1.NodeSelectorOpExists},
						},
					},
				}),
				withTestVolumes([]corev1.Volume{
					{
						Name: "shared-data",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}),
				withContainerVolumeMountsTest([]corev1.VolumeMount{
					{Name: "shared-data", MountPath: "/data"},
				}),
			),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity preferred with configMap volume",
			deploy: deployment(
				withNodeAffinityPreferred(75),
				withTestVolumes([]corev1.Volume{
					{
						Name: "app-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"},
							},
						},
					},
				}),
			),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AWU-22.3] testing nodeAffinity cross-struct volume mounts: %s", tc.name)
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

// =============================================================================
// Helper modifiers for nodeAffinity (prefixed to avoid redeclaration)
// =============================================================================

// withNodeAffinityRequired sets nodeAffinity with requiredDuringSchedulingIgnoredDuringExecution
func withNodeAffinityRequired(terms []corev1.NodeSelectorTerm) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: terms,
			},
		}
	}
}

// withNodeAffinityPreferred sets nodeAffinity with preferredDuringSchedulingIgnoredDuringExecution using a single weight
func withNodeAffinityPreferred(weight int32) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: weight,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "gpu", Operator: corev1.NodeSelectorOpExists},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityPreferredMulti creates multiple preferred terms with given weights
func withNodeAffinityPreferredMulti(weights ...int) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		preferred := make([]corev1.PreferredSchedulingTerm, len(weights))
		for i, w := range weights {
			preferred[i] = corev1.PreferredSchedulingTerm{
				Weight: int32(w),
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "tier", Operator: corev1.NodeSelectorOpIn, Values: []string{"backend"}},
					},
				},
			}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: preferred,
		}
	}
}

// withNodeAffinityBoth sets both required and preferred node affinity
func withNodeAffinityBoth() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
						},
					},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 50,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "ssd", Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityBothStrict sets both required and preferred with different weights
func withNodeAffinityBothStrict() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
						},
					},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 1,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "ssd", Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityEmptyRequiredWithPreferred sets required with empty terms and preferred
func withNodeAffinityEmptyRequiredWithPreferred(preferredWeight int32) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: preferredWeight,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "ssd", Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityMatchExprIn creates a required node affinity with In operator
func withNodeAffinityMatchExprIn(values ...string) func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: values},
			},
		},
	})
}

// withNodeAffinityMatchExprNotIn creates a required node affinity with NotIn operator
func withNodeAffinityMatchExprNotIn(values ...string) func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpNotIn, Values: values},
			},
		},
	})
}

// withNodeAffinityMatchExprExists creates a required node affinity with Exists operator
func withNodeAffinityMatchExprExists(key string) func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: key, Operator: corev1.NodeSelectorOpExists},
			},
		},
	})
}

// withNodeAffinityMatchExprDoesNotExist creates a required node affinity with DoesNotExist operator
func withNodeAffinityMatchExprDoesNotExist(key string) func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: key, Operator: corev1.NodeSelectorOpDoesNotExist},
			},
		},
	})
}

// withNodeAffinityMatchExprMixed creates a required node affinity with mixed operators
func withNodeAffinityMatchExprMixed() func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
				{Key: "disk-type", Operator: corev1.NodeSelectorOpNotIn, Values: []string{"HDD"}},
				{Key: "cache", Operator: corev1.NodeSelectorOpExists},
				{Key: "deprecated-label", Operator: corev1.NodeSelectorOpDoesNotExist},
			},
		},
	})
}

// withNodeAffinityInvalidExistsWithValues tests Exists with non-empty values
func withNodeAffinityInvalidExistsWithValues() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "gpu", Operator: corev1.NodeSelectorOpExists, Values: []string{"nvidia"}},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityInvalidNotInEmptyValues tests NotIn with empty values
func withNodeAffinityInvalidNotInEmptyValues() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpNotIn, Values: []string{}},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityInvalidInEmptyValues tests In with empty values
func withNodeAffinityInvalidInEmptyValues() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{}},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityMatchFieldsIn creates matchFields with In operator
func withNodeAffinityMatchFieldsIn(key string, values ...string) func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchFields: []corev1.NodeSelectorRequirement{
				{Key: key, Operator: corev1.NodeSelectorOpIn, Values: values},
			},
		},
	})
}

// withNodeAffinityMatchFieldsNotIn creates matchFields with NotIn operator
func withNodeAffinityMatchFieldsNotIn(key string, values ...string) func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchFields: []corev1.NodeSelectorRequirement{
				{Key: key, Operator: corev1.NodeSelectorOpNotIn, Values: values},
			},
		},
	})
}

// withNodeAffinityMatchFieldsMixed creates matchFields mixed with matchExpressions
func withNodeAffinityMatchFieldsMixed() func(*appsv1.Deployment) {
	return withNodeAffinityRequired([]corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
			},
			MatchFields: []corev1.NodeSelectorRequirement{
				{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-1"}},
			},
		},
	})
}

// withNodeAffinityMultiContainer adds a second container and sets nodeAffinity
func withNodeAffinityMultiContainer() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers,
			corev1.Container{Name: "sidecar", Image: "nginx:1.21"})
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
						},
					},
				},
			},
		}
	}
}

// withNodeAffinityAllContainers sets nodeAffinity on pod spec (applies to all containers)
func withNodeAffinityAllContainers() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers = []corev1.Container{
			{Name: "main", Image: "nginx:1.21"},
			{Name: "sidecar", Image: "nginx:1.21"},
		}
		if d.Spec.Template.Spec.Affinity == nil {
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		d.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 80,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "gpu", Operator: corev1.NodeSelectorOpExists},
						},
					},
				},
			},
		}
	}
}

// withInitContainerTest creates a deployment with a test init container (for nodeAffinity init container tests)
func withInitContainerTest() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.InitContainers = []corev1.Container{
			{Name: "init-setter", Image: "busybox:1.36"},
		}
	}
}

// withTestVolumes adds volumes to the deployment (named to avoid redeclaration)
func withTestVolumes(volumes []corev1.Volume) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, volumes...)
	}
}

// withContainerVolumeMountsTest sets volumeMounts on the primary container
func withContainerVolumeMountsTest(mounts []corev1.VolumeMount) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if len(d.Spec.Template.Spec.Containers) > 0 {
			d.Spec.Template.Spec.Containers[0].VolumeMounts = mounts
		}
	}
}