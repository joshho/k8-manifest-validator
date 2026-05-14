package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =============================================================================
// AWU-20.5 — Phase 2: PodSpec affinity expanded edge-condition tests
// PRD ref: plan.json AWU-20.5 — PodSpec affinity validation (phase_2 wave)
// Codegen (generated_structural.go lines 291–299, 830–864):
//   .affinity.nodeAffinity — deferred, non-required
//   .affinity.podAffinity — deferred, non-required
//   .affinity.podAntiAffinity — deferred, non-required
//   .affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution
//   .affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution
//   .affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution
//   .affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution
//   .affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution
//   .affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution
//   .podAffinityTerm.topologyKey — required, must be valid label key
//   .podAffinityTerm.labelSelector — optional, must be valid label selector
// Validation wire: validateAffinity in builtin.go → semantic_scheduling.go
// Existing coverage: phase2_integration_test.go TestPhase2_StructuralLayer_Affinity (basic)
// AWU-20.5 complement: expanded edge conditions not covered by the integration file
// =============================================================================

// -----------------------------------------------------------------------
// AWU-20.5 — Affinity edge-condition coverage
// -----------------------------------------------------------------------

func TestPhase2_Affinity_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// ---- Valid: empty/nil affinity ----

		{
			name:      "valid nil affinity — no affinity field set",
			deploy:    deployment(withAffinity(nil)),
			wantErr:   false,
		},
		{
			name: "valid empty affinity struct — all fields nil",
			deploy: deployment(withAffinity(&corev1.Affinity{})),
			wantErr: false,
		},
		{
			name: "valid empty nodeAffinity — NodeAffinity field set but empty",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{},
			})),
			wantErr: false,
		},
		{
			name: "valid empty podAffinity — PodAffinity field set but empty",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{},
			})),
			wantErr: false,
		},
		{
			name: "valid empty podAntiAffinity — PodAntiAffinity field set but empty",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{},
			})),
			wantErr: false,
		},

		// ---- Valid: nodeAffinity requiredDuringSchedulingIgnoredDuringExecution ----

		{
			name: "valid nodeAffinity requiredDuringSchedulingIgnoredExecution — single term with matchExpression",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity requiredDuringSchedulingIgnoredExecution — single term with matchField",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchFields: []corev1.NodeSelectorRequirement{
									{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-1"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity requiredDuringSchedulingIgnoredExecution — multiple terms",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "topology.kubernetes.io/zone", Operator: corev1.NodeSelectorOpNotIn, Values: []string{"zone-f"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: nodeAffinity requiredDuringSchedulingIgnoredDuringExecution empty terms ----

		{
			name: "invalid nodeAffinity requiredDuringSchedulingIgnoredExecution — empty nodeSelectorTerms",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "nodeSelectorTerms",
		},

		// ---- Valid: nodeAffinity preferredDuringSchedulingIgnoredDuringExecution ----

		{
			name: "valid nodeAffinity preferredDuringSchedulingIgnoredExecution — weight=50",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 50,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity preferredDuringSchedulingIgnoredExecution — weight=100 (max boundary)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 100,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "node-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"compute"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity preferredDuringSchedulingIgnoredExecution — weight=1 (min boundary)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 1,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "gpu", Operator: corev1.NodeSelectorOpExists},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity preferredDuringSchedulingIgnoredExecution — multiple terms",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 50,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
						},
						{
							Weight: 30,
							Preference: corev1.NodeSelectorTerm{
								MatchFields: []corev1.NodeSelectorRequirement{
									{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-1"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: nodeAffinity preferredDuringSchedulingIgnoredDuringExecution weight out of range ----

		{
			name: "invalid nodeAffinity preferred weight=0 — must be 1–100",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 0,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "weight",
		},
		{
			name: "invalid nodeAffinity preferred weight=101 — must be 1–100",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 101,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "weight",
		},
		{
			name: "invalid nodeAffinity preferred weight=-1 (negative)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: -1,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "weight",
		},

		// ---- Valid: podAffinity requiredDuringSchedulingIgnoredDuringExecution ----

		{
			name: "valid podAffinity requiredDuringSchedulingIgnoredExecution — single term with topologyKey",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "test"},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid podAffinity requiredDuringSchedulingIgnoredExecution — topologyKey=topology.kubernetes.io/zone",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "topology.kubernetes.io/zone",
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: podAffinity topologyKey empty ----

		{
			name: "invalid podAffinity term topologyKey empty string",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "", // required, cannot be empty
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "topologyKey",
		},

		// ---- Valid: podAffinity preferredDuringSchedulingIgnoredDuringExecution ----

		{
			name: "valid podAffinity preferredDuringSchedulingIgnoredExecution — weight=50",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 50,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "test"},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid podAffinity preferredDuringSchedulingIgnoredExecution — weight=1 (min boundary)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 1,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid podAffinity preferredDuringSchedulingIgnoredExecution — weight=100 (max boundary)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "topology.kubernetes.io/region",
							},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: podAffinity preferred weight out of range ----

		{
			name: "invalid podAffinity preferred weight=0 — must be 1–100",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 0,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "weight",
		},
		{
			name: "invalid podAffinity preferred weight=101 — must be 1–100",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 101,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "weight",
		},
		{
			name: "invalid podAffinity preferred weight=-50 (negative)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: -50,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "weight",
		},

		// ---- Valid: podAntiAffinity requiredDuringSchedulingIgnoredDuringExecution ----

		{
			name: "valid podAntiAffinity requiredDuringSchedulingIgnoredExecution — single term",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"test"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: podAntiAffinity topologyKey empty ----

		{
			name: "invalid podAntiAffinity term topologyKey empty string",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "", // required, cannot be empty
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "topologyKey",
		},

		// ---- Valid: podAntiAffinity preferredDuringSchedulingIgnoredDuringExecution ----

		{
			name: "valid podAntiAffinity preferredDuringSchedulingIgnoredExecution — weight=50",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 50,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: podAntiAffinity preferred weight out of range ----

		{
			name: "invalid podAntiAffinity preferred weight=0 — must be 1–100",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 0,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "weight",
		},

		// ---- Valid: labelSelector variants ----

		{
			name: "valid podAffinity labelSelector with matchLabels",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "web", "env": "prod"},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid podAffinity labelSelector with matchExpressions",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"web", "api"}},
									{Key: "env", Operator: metav1.LabelSelectorOpNotIn, Values: []string{"dev"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid podAffinity labelSelector nil (no selector, matches all)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: nil, // nil is valid — matches all
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: labelSelector malformed ----

		{
			name: "invalid podAffinity labelSelector — empty matchLabel key (invalid label)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"": "value"}, // empty key is invalid
							},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "labelSelector",
		},

		// ---- Valid: multiple node selectors and terms ----

		{
			name: "valid nodeAffinity with both required and preferred terms",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
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
							Weight: 80,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "gpu", Operator: corev1.NodeSelectorOpExists},
								},
							},
						},
						{
							Weight: 20,
							Preference: corev1.NodeSelectorTerm{
								MatchFields: []corev1.NodeSelectorRequirement{
									{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-1"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "valid nodeAffinity multiple required terms",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
							{
								MatchFields: []corev1.NodeSelectorRequirement{
									{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"node-1", "node-2"}},
								},
							},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Valid: all three affinity types combined ----

		{
			name: "valid affinity with nodeAffinity podAffinity and podAntiAffinity all set",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
								},
							},
						},
					},
				},
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "test"},
							},
						},
					},
				},
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 50,
							PodAffinityTerm: corev1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: topologyKey invalid format ----

		{
			name: "invalid topologyKey — value with invalid label key format (uppercase)",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "Invalid/Key", // must be lowercase alphanumeric + ./-
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "topologyKey",
		},

		// ---- Invalid: podAffinity/podAntiAffinity term with empty topologyKey ----

		{
			name: "invalid podAntiAffinity term topologyKey empty string",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "", // required, cannot be empty
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "topologyKey",
		},

		// ---- Valid: namespaces field ----

		{
			name: "valid podAffinity term with namespaces",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							Namespaces:  []string{"default", "kube-system"},
						},
					},
				},
			})),
			wantErr: false,
		},

		// ---- Invalid: namespace with invalid DNS1123 subdomain ----

		{
			name: "invalid podAffinity term namespace with invalid DNS1123 subdomain",
			deploy: deployment(withAffinity(&corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							TopologyKey: "kubernetes.io/hostname",
							Namespaces:  []string{"_invalid_namespace"},
						},
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "namespaces",
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

// =============================================================================
// Re-use helpers from phase2_integration_test.go:
//   - deployment(), baseDeployment()
//   - validateDeploymentRaw()
//   - hasErrErrItems()
// =============================================================================