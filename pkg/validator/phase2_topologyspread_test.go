package validator

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =============================================================================
// AWU-20.2 — Phase 2: topologySpreadConstraints expanded edge-condition tests
// PRD ref: plan_main.json phase_2 — AWU-20.2
// Codegen (generated_structural.go lines 1099–1105):
//   .topologySpreadConstraints[*].maxSkew         — Required, integer, min=1
//   .topologySpreadConstraints[*].topologyKey    — Required, string, minLength=1
//   .topologySpreadConstraints[*].whenUnsatisfiable — Required, string
// Validation wire: validateSchedulingFields in builtin.go (lines 777, 819, 850, 885, 915, 937)
// Existing coverage: phase2_integration_test.go TestPhase2_StructuralLayer_TopologySpreadConstraints (line 28)
// AWU-20.2 complement: expanded edge conditions not covered by the integration file
// =============================================================================

// -----------------------------------------------------------------------
// AWU-20.2 — TopologySpreadConstraints edge-condition coverage
// -----------------------------------------------------------------------

func TestPhase2_TopologySpreadConstraints_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		// ---- Valid cases ----

		{
			name: "valid topologySpreadConstraints — all 3 required fields present",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr: false,
		},
		{
			name: "valid topologySpreadConstraints — maxSkew=1, ScheduleAnyway",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "topology.kubernetes.io/zone",
					WhenUnsatisfiable: corev1.ScheduleAnyway,
				},
			)),
			wantErr: false,
		},
		{
			name: "valid topologySpreadConstraints — maxSkew > 1 (boundary 2)",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           2,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr: false,
		},

		// ---- maxSkew invalid cases ----

		{
			name: "invalid maxSkew=0 — must be >= 1",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           0,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "maxSkew",
		},
		{
			name: "invalid maxSkew negative (-1)",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           -1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "maxSkew",
		},
		{
			name: "invalid maxSkew negative (-100)",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           -100,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "maxSkew",
		},

		// ---- topologyKey invalid cases ----

		{
			name: "invalid topologyKey empty string",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "topologyKey",
		},


		// ---- whenUnsatisfiable invalid cases ----

		{
			name: "invalid whenUnsatisfiable arbitrary string — not DoNotSchedule or ScheduleAnyway",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: "InvalidValue",
				},
			)),
			wantErr:    true,
			errSubstr:  "whenUnsatisfiable",
		},
		{
			name: "invalid whenUnsatisfiable empty string",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: "",
				},
			)),
			wantErr:    true,
			errSubstr:  "whenUnsatisfiable",
		},
		{
			name: "invalid whenUnsatisfiable lowercase doNotSchedule — must be exact case",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: "doNotSchedule", // lowercase
				},
			)),
			wantErr:    true,
			errSubstr:  "whenUnsatisfiable",
		},
		{
			name: "invalid whenUnsatisfiable ScheduleAnyway with different casing",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: "scheduleAnyway", // all lowercase
				},
			)),
			wantErr:    true,
			errSubstr:  "whenUnsatisfiable",
		},

		// ---- Missing each required field ----

		{
			name: "invalid missing topologyKey — only maxSkew and whenUnsatisfiable set",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "topologyKey",
		},
		{
			name: "invalid missing whenUnsatisfiable — only maxSkew and topologyKey set",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: "",
				},
			)),
			wantErr:    true,
			errSubstr:  "whenUnsatisfiable",
		},
		{
			name: "invalid missing maxSkew — only topologyKey and whenUnsatisfiable set",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           0, // zero is treated as missing / invalid
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "maxSkew",
		},

		// ---- Multiple constraints ----

		{
			name: "valid multiple topologySpreadConstraints — 2 entries",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
				corev1.TopologySpreadConstraint{
					MaxSkew:           2,
					TopologyKey:       "topology.kubernetes.io/zone",
					WhenUnsatisfiable: corev1.ScheduleAnyway,
				},
			)),
			wantErr: false,
		},
		{
			name: "valid multiple topologySpreadConstraints — 3 entries with labelSelector",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "topology.kubernetes.io/zone",
					WhenUnsatisfiable: corev1.DoNotSchedule,
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"prod"}},
						},
					},
				},
				corev1.TopologySpreadConstraint{
					MaxSkew:           2,
					TopologyKey:       "topology.kubernetes.io/region",
					WhenUnsatisfiable: corev1.ScheduleAnyway,
				},
			)),
			wantErr: false,
		},

		// ---- Cross-field: one invalid among valid entries ----

		{
			name: "invalid one of two constraints has bad maxSkew",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
				corev1.TopologySpreadConstraint{
					MaxSkew:           0, // invalid — must be >= 1
					TopologyKey:       "topology.kubernetes.io/zone",
					WhenUnsatisfiable: corev1.ScheduleAnyway,
				},
			)),
			wantErr:    true,
			errSubstr:  "maxSkew",
		},
		{
			name: "invalid one of two constraints has empty topologyKey",
			deploy: deployment(withTopologySpreadConstraintsExpanded(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "", // invalid
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "topologyKey",
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
// Helper modifiers — support multi-constraint test cases
// =============================================================================

// withTopologySpreadConstraintsExpanded sets topologySpreadConstraints (replaces existing)
func withTopologySpreadConstraintsExpanded(constraints ...corev1.TopologySpreadConstraint) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.TopologySpreadConstraints = constraints
	}
}

// Re-use existing helpers from phase2_integration_test.go:
//   - deployment(), baseDeployment()
//   - validateDeploymentRaw()
//   - hasErrErrItems()
//   - boolPtr(), int64Ptr()