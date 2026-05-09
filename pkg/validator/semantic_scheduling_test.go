package validator

import (
	"net"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// =============================================================================
// AWU-14.1 — Semantic Scheduling: unit tests for all 6 validators
// =============================================================================

// ---------------------------------------------------------------------------
// Helper: minimal PodSpec with just the field under test
// ---------------------------------------------------------------------------

func affinityWithNodeSelectorTerms(terms []corev1.NodeSelectorTerm) *corev1.Affinity {
	if len(terms) == 0 {
		return &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{},
			},
		}
	}
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: terms,
			},
		},
	}
}

func affinityWithPreferredWeight(weight int32) *corev1.Affinity {
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: weight,
					Preference: corev1.NodeSelectorTerm{},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// 1. validateAffinity
// ---------------------------------------------------------------------------

func TestValidateAffinity(t *testing.T) {
	t.Run("nil_affinity", func(t *testing.T) {
		errs := validateAffinity(nil, field.NewPath("affinity"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for nil affinity, got %v", errs)
		}
	})

	t.Run("valid_nodeSelectorTerms", func(t *testing.T) {
		errs := validateAffinity(affinityWithNodeSelectorTerms([]corev1.NodeSelectorTerm{{}}), field.NewPath("affinity"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for valid nodeSelectorTerms, got %v", errs)
		}
	})

	t.Run("empty_nodeSelectorTerms", func(t *testing.T) {
		errs := validateAffinity(affinityWithNodeSelectorTerms([]corev1.NodeSelectorTerm{}), field.NewPath("affinity"))
		if len(errs) == 0 {
			t.Error("expected errors for empty nodeSelectorTerms, got none")
		}
	})

	t.Run("preferred_weight_valid", func(t *testing.T) {
		for _, w := range []int32{1, 50, 100} {
			errs := validateAffinity(affinityWithPreferredWeight(w), field.NewPath("affinity"))
			if len(errs) > 0 {
				t.Errorf("weight %d: expected no errors, got %v", w, errs)
			}
		}
	})

	t.Run("preferred_weight_invalid", func(t *testing.T) {
		for _, w := range []int32{0, 101} {
			errs := validateAffinity(affinityWithPreferredWeight(w), field.NewPath("affinity"))
			if len(errs) == 0 {
				t.Errorf("weight %d: expected errors, got none", w)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. validateTolerations
// ---------------------------------------------------------------------------

func toleration(op corev1.TolerationOperator, key, value, effect string) []corev1.Toleration {
	return []corev1.Toleration{{Operator: op, Key: key, Value: value, Effect: effect}}
}

func TestValidateTolerations(t *testing.T) {
	t.Run("Equal_operator_with_value", func(t *testing.T) {
		errs := validateTolerations(toleration(corev1.TolerationOpEqual, "key", "value", "NoSchedule"), field.NewPath("tolerations"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for Equal with value, got %v", errs)
		}
	})

	t.Run("Equal_operator_no_value", func(t *testing.T) {
		errs := validateTolerations(toleration(corev1.TolerationOpEqual, "key", "", "NoSchedule"), field.NewPath("tolerations"))
		if len(errs) == 0 {
			t.Error("expected errors for Equal with no value, got none")
		}
	})

	t.Run("Exists_operator_no_value", func(t *testing.T) {
		errs := validateTolerations(toleration(corev1.TolerationOpExists, "key", "", "NoSchedule"), field.NewPath("tolerations"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for Exists with no value, got %v", errs)
		}
	})

	t.Run("Exists_operator_with_value", func(t *testing.T) {
		errs := validateTolerations(toleration(corev1.TolerationOpExists, "key", "value", "NoSchedule"), field.NewPath("tolerations"))
		if len(errs) == 0 {
			t.Error("expected errors for Exists with value, got none")
		}
	})

	t.Run("empty_effect", func(t *testing.T) {
		errs := validateTolerations(toleration(corev1.TolerationOpExists, "key", "", ""), field.NewPath("tolerations"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for empty effect, got %v", errs)
		}
	})

	t.Run("invalid_key_format", func(t *testing.T) {
		errs := validateTolerations(toleration(corev1.TolerationOpExists, "Invalid/Key/With/Too/Many/Slashes", "", ""), field.NewPath("tolerations"))
		if len(errs) == 0 {
			t.Error("expected errors for invalid key format, got none")
		}
	})
}

// ---------------------------------------------------------------------------
// 3. validateTopologySpreadConstraints
// ---------------------------------------------------------------------------

func topologyConstraint(maxSkew int32, topologyKey string, whenUnsatisfiable corev1.TopologySpreadConstraintType, labelSelector *metav1.LabelSelector) []corev1.TopologySpreadConstraint {
	return []corev1.TopologySpreadConstraint{{
		MaxSkew:            maxSkew,
		TopologyKey:        topologyKey,
		WhenUnsatisfiable:  whenUnsatisfiable,
		LabelSelector:      labelSelector,
	}}
}

func validLabelSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "nginx"},
	}
}

func TestValidateTopologySpreadConstraints(t *testing.T) {
	t.Run("maxSkew_ge_1", func(t *testing.T) {
		for _, skew := range []int32{1, 2, 10} {
			errs := validateTopologySpreadConstraints(topologyConstraint(skew, "kubernetes.io/hostname", corev1.DoNotSchedule, nil), field.NewPath("topologySpreadConstraints"))
			if len(errs) > 0 {
				t.Errorf("maxSkew %d: expected no errors, got %v", skew, errs)
			}
		}
	})

	t.Run("maxSkew_0", func(t *testing.T) {
		errs := validateTopologySpreadConstraints(topologyConstraint(0, "kubernetes.io/hostname", corev1.DoNotSchedule, nil), field.NewPath("topologySpreadConstraints"))
		if len(errs) == 0 {
			t.Error("expected errors for maxSkew=0, got none")
		}
	})

	t.Run("topologyKey_empty", func(t *testing.T) {
		errs := validateTopologySpreadConstraints(topologyConstraint(1, "", corev1.DoNotSchedule, nil), field.NewPath("topologySpreadConstraints"))
		if len(errs) == 0 {
			t.Error("expected errors for empty topologyKey, got none")
		}
	})

	t.Run("whenUnsatisfiable_valid", func(t *testing.T) {
		for _, a := range []corev1.TopologySpreadConstraintType{corev1.DoNotSchedule, corev1.ScheduleAnyway} {
			errs := validateTopologySpreadConstraints(topologyConstraint(1, "kubernetes.io/hostname", a, nil), field.NewPath("topologySpreadConstraints"))
			if len(errs) > 0 {
				t.Errorf("whenUnsatisfiable=%s: expected no errors, got %v", a, errs)
			}
		}
	})

	t.Run("whenUnsatisfiable_invalid", func(t *testing.T) {
		errs := validateTopologySpreadConstraints(topologyConstraint(1, "kubernetes.io/hostname", "InvalidAction", nil), field.NewPath("topologySpreadConstraints"))
		if len(errs) == 0 {
			t.Error("expected errors for invalid whenUnsatisfiable, got none")
		}
	})

	t.Run("labelSelector_valid", func(t *testing.T) {
		errs := validateTopologySpreadConstraints(topologyConstraint(1, "kubernetes.io/hostname", corev1.DoNotSchedule, validLabelSelector()), field.NewPath("topologySpreadConstraints"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for valid labelSelector, got %v", errs)
		}
	})
}

// ---------------------------------------------------------------------------
// 4. validateDNSConfig
// ---------------------------------------------------------------------------

func dnsConfig(nameservers []string, searches []string, options []corev1.PodDNSConfigOption) *corev1.PodDNSConfig {
	return &corev1.PodDNSConfig{
		Nameservers: nameservers,
		Searches:    searches,
		Options:     options,
	}
}

func dnsOption(name string) []corev1.PodDNSConfigOption {
	return []corev1.PodDNSConfigOption{{Name: name}}
}

func TestValidateDNSConfig(t *testing.T) {
	t.Run("valid_nameserver_IPs", func(t *testing.T) {
		cfg := dnsConfig([]string{"8.8.8.8", "1.1.1.1"}, nil, nil)
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirst, false, field.NewPath("dnsConfig"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for valid IPs, got %v", errs)
		}
	})

	t.Run("invalid_nameserver_IP", func(t *testing.T) {
		cfg := dnsConfig([]string{"not-an-ip"}, nil, nil)
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirst, false, field.NewPath("dnsConfig"))
		if len(errs) == 0 {
			t.Error("expected errors for invalid IP, got none")
		}
	})

	t.Run("empty_string_nameserver", func(t *testing.T) {
		cfg := dnsConfig([]string{""}, nil, nil)
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirst, false, field.NewPath("dnsConfig"))
		if len(errs) == 0 {
			t.Error("expected errors for empty nameserver, got none")
		}
	})

	t.Run("searches_max_6", func(t *testing.T) {
		cfg := dnsConfig([]string{"8.8.8.8"}, []string{"svc1", "svc2", "svc3", "svc4", "svc5", "svc6", "svc7"}, nil)
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirst, false, field.NewPath("dnsConfig"))
		if len(errs) == 0 {
			t.Error("expected errors for >6 searches, got none")
		}
	})

	t.Run("searches_within_limit", func(t *testing.T) {
		cfg := dnsConfig([]string{"8.8.8.8"}, []string{"svc1", "svc2", "svc3", "svc4", "svc5", "svc6"}, nil)
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirst, false, field.NewPath("dnsConfig"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for 6 searches, got %v", errs)
		}
	})

	t.Run("options_name_required", func(t *testing.T) {
		cfg := dnsConfig([]string{"8.8.8.8"}, nil, []corev1.PodDNSConfigOption{{Name: ""}})
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirst, false, field.NewPath("dnsConfig"))
		if len(errs) == 0 {
			t.Error("expected errors for empty option name, got none")
		}
	})

	t.Run("options_name_valid", func(t *testing.T) {
		cfg := dnsConfig([]string{"8.8.8.8"}, nil, dnsOption("ndots"))
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirst, false, field.NewPath("dnsConfig"))
		if len(errs) > 0 {
			t.Errorf("expected no errors for valid option name, got %v", errs)
		}
	})

	t.Run("ClusterFirstWithHostNet_and_dnsConfig", func(t *testing.T) {
		cfg := dnsConfig([]string{"8.8.8.8"}, nil, nil)
		errs := validateDNSConfig(cfg, corev1.DNSClusterFirstWithHostNet, true, field.NewPath("dnsConfig"))
		if len(errs) == 0 {
			t.Error("expected errors for ClusterFirstWithHostNet + dnsConfig, got none")
		}
	})

	t.Run("ClusterFirstWithHostNet_no_dnsConfig", func(t *testing.T) {
		errs := validateDNSConfig(nil, corev1.DNSClusterFirstWithHostNet, true, field.NewPath("dnsConfig"))
		if len(errs) > 0 {
			t.Errorf("expected no errors when dnsConfig is nil, got %v", errs)
		}
	})

	// helper: verify parseIP behaviour for nil/empty
	_ = net.ParseIP("")
}

// ---------------------------------------------------------------------------
// 5. validateHostBooleans (no-op pass-through)
// ---------------------------------------------------------------------------

func TestValidateHostBooleans(t *testing.T) {
	t.Run("always_passes", func(t *testing.T) {
		spec := &corev1.PodSpec{
			HostIPC:     true,
			HostNetwork: true,
			HostPID:     true,
		}
		errs := validateHostBooleans(spec, field.NewPath("spec"))
		if len(errs) > 0 {
			t.Errorf("expected no errors (no-op), got %v", errs)
		}
	})

	t.Run("all_false", func(t *testing.T) {
		spec := &corev1.PodSpec{}
		errs := validateHostBooleans(spec, field.NewPath("spec"))
		if len(errs) > 0 {
			t.Errorf("expected no errors (no-op), got %v", errs)
		}
	})
}

// ---------------------------------------------------------------------------
// 6. validateRestartPolicy
// ---------------------------------------------------------------------------

func TestValidateRestartPolicy(t *testing.T) {
	validPolicies := []corev1.RestartPolicy{
		corev1.RestartPolicyAlways,
		corev1.RestartPolicyOnFailure,
		corev1.RestartPolicyNever,
	}

	t.Run("valid_values", func(t *testing.T) {
		for _, p := range validPolicies {
			errs := validateRestartPolicy(p, field.NewPath("restartPolicy"))
			if len(errs) > 0 {
				t.Errorf("policy %s: expected no errors, got %v", p, errs)
			}
		}
	})

	t.Run("invalid_value", func(t *testing.T) {
		errs := validateRestartPolicy("InvalidPolicy", field.NewPath("restartPolicy"))
		if len(errs) == 0 {
			t.Error("expected errors for invalid policy, got none")
		}
	})
}
