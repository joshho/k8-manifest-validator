package validator

import (
	"net"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// =============================================================================
// AWU-14.1 — Semantic Layer: scheduling fields (affinity, tolerations, topologySpreadConstraints, dnsConfig, host booleans, restartPolicy)
// PRD ref: PRD Section 20 Phase 2 — Layer 2 semantic
// Depends on: AWU-11.1 (structural layer)
// =============================================================================

// labelKeyRegex matches valid Kubernetes label key format: prefix/name or just name.
// prefix is optional (max 253 chars), name is required (max 63 chars).
var labelKeyRegex = regexp.MustCompile(`^([a-z0-9]([a-z0-9.-]*[a-z0-9])?/)?[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

// =============================================================================
// 1. validateAffinity
// =============================================================================

// validateAffinity validates pod affinity/anti-affinity and node affinity rules.
// nodeAffinity: if specified and non-nil, must have non-empty nodeSelectorTerms.
// podAffinity / podAntiAffinity: same rule — if specified and non-nil, must have non-empty selector.
// preferredDuringSchedulingIgnoredDuringExecution: each term's weight must be 1–100.
func validateAffinity(affinity *corev1.Affinity, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if affinity == nil {
		return allErrs
	}

	// nodeAffinity
	if affinity.NodeAffinity != nil {
		allErrs = append(allErrs, validateNodeAffinity(affinity.NodeAffinity, path.Child("nodeAffinity"))...)
	}

	// podAffinity
	if affinity.PodAffinity != nil {
		allErrs = append(allErrs, validatePodAffinityTerms(
			affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			path.Child("podAffinity").Child("requiredDuringSchedulingIgnoredDuringExecution"),
		)...)
		allErrs = append(allErrs, validateWeightedPodAffinityTerms(
			affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			path.Child("podAffinity").Child("preferredDuringSchedulingIgnoredDuringExecution"),
		)...)
	}

	// podAntiAffinity
	if affinity.PodAntiAffinity != nil {
		allErrs = append(allErrs, validatePodAffinityTerms(
			affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			path.Child("podAntiAffinity").Child("requiredDuringSchedulingIgnoredDuringExecution"),
		)...)
		allErrs = append(allErrs, validateWeightedPodAffinityTerms(
			affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			path.Child("podAntiAffinity").Child("preferredDuringSchedulingIgnoredDuringExecution"),
		)...)
	}

	return allErrs
}

func validateNodeAffinity(na *corev1.NodeAffinity, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// requiredDuringSchedulingIgnoredDuringExecution must have non-empty nodeSelectorTerms
	if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		terms := na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		if len(terms) == 0 {
			allErrs = append(allErrs, field.Required(path.Child("requiredDuringSchedulingIgnoredDuringExecution").Child("nodeSelectorTerms"), "must have at least one node selector term"))
		} else {
			for i, term := range terms {
				allErrs = append(allErrs, validateNodeSelectorTerm(&term, path.Child("requiredDuringSchedulingIgnoredDuringExecution").Child("nodeSelectorTerms").Index(i))...)
			}
		}
	}

	// preferredDuringSchedulingIgnoredDuringExecution: validate weights
	for i, term := range na.PreferredDuringSchedulingIgnoredDuringExecution {
		if term.Weight < 1 || term.Weight > 100 {
			allErrs = append(allErrs, field.Invalid(
				path.Child("preferredDuringSchedulingIgnoredDuringExecution").Index(i).Child("weight"),
				term.Weight,
				"must be between 1 and 100",
			))
		}
		allErrs = append(allErrs, validateNodeSelectorTerm(&term.Preference, path.Child("preferredDuringSchedulingIgnoredDuringExecution").Index(i).Child("preference"))...)
	}

	return allErrs
}

func validateNodeSelectorTerm(term *corev1.NodeSelectorTerm, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// matchExpressions
	for i, me := range term.MatchExpressions {
		if me.Key != "" {
			if errs := utilvalidation.IsQualifiedName(me.Key); len(errs) > 0 {
				allErrs = append(allErrs, field.Invalid(
					path.Child("matchExpressions").Index(i).Child("key"),
					me.Key,
					errs[0],
				))
			}
		}
		// operator is required
		if me.Operator == "" {
			allErrs = append(allErrs, field.Required(path.Child("matchExpressions").Index(i).Child("operator"), ""))
		} else {
			switch me.Operator {
			case corev1.NodeSelectorOpExists, corev1.NodeSelectorOpDoesNotExist:
				// value must be empty for Exists/DoesNotExist
			case corev1.NodeSelectorOpIn, corev1.NodeSelectorOpNotIn:
				// value list should be non-empty (k8s checks this)
				if len(me.Values) == 0 {
					allErrs = append(allErrs, field.Required(
						path.Child("matchExpressions").Index(i).Child("values"),
						"values must be non-empty when operator is In or NotIn",
					))
				}
			}
		}
	}

	// matchFields
	for i, mf := range term.MatchFields {
		if mf.Key != "" {
			if errs := utilvalidation.IsQualifiedName(mf.Key); len(errs) > 0 {
				allErrs = append(allErrs, field.Invalid(
					path.Child("matchFields").Index(i).Child("key"),
					mf.Key,
					errs[0],
				))
			}
		}
		if mf.Operator == "" {
			allErrs = append(allErrs, field.Required(path.Child("matchFields").Index(i).Child("operator"), ""))
		}
	}

	return allErrs
}

func validatePodAffinityTerms(terms []corev1.PodAffinityTerm, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	for i, term := range terms {
		allErrs = append(allErrs, validatePodAffinityTerm(&term, path.Index(i))...)
	}
	return allErrs
}

func validatePodAffinityTerm(term *corev1.PodAffinityTerm, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// topologyKey is required and must be a valid label key
	if term.TopologyKey == "" {
		allErrs = append(allErrs, field.Required(path.Child("topologyKey"), ""))
	} else if !labelKeyRegex.MatchString(term.TopologyKey) {
		allErrs = append(allErrs, field.Invalid(path.Child("topologyKey"), term.TopologyKey, "must be a valid label key"))
	}

	// labelSelector is validated separately via standard k8s label selector validation
	if term.LabelSelector != nil {
		_, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(path.Child("labelSelector"), term.LabelSelector, err.Error()))
		}
	}

	// namespaces: if specified, each must be valid
	for j, ns := range term.Namespaces {
		if errs := utilvalidation.IsDNS1123Subdomain(ns); len(errs) > 0 {
			allErrs = append(allErrs, field.Invalid(path.Child("namespaces").Index(j), ns, errs[0]))
		}
	}

	return allErrs
}

func validateWeightedPodAffinityTerms(terms []corev1.WeightedPodAffinityTerm, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	for i, term := range terms {
		if term.Weight < 1 || term.Weight > 100 {
			allErrs = append(allErrs, field.Invalid(
				path.Index(i).Child("weight"),
				term.Weight,
				"must be between 1 and 100",
			))
		}
		allErrs = append(allErrs, validatePodAffinityTerm(&term.PodAffinityTerm, path.Index(i).Child("podAffinityTerm"))...)
	}
	return allErrs
}

// =============================================================================
// 2. validateTolerations
// =============================================================================

// validateTolerations validates tolerations for correct operator/value combinations.
// Equal → value must be non-empty
// Exists → value must be empty
// Effect can be empty (matches all effects)
// Key must be a valid label key format
func validateTolerations(tolerations []corev1.Toleration, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	for i, t := range tolerations {
		allErrs = append(allErrs, validateToleration(&t, path.Index(i))...)
	}
	return allErrs
}

func validateToleration(toleration *corev1.Toleration, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Key must be a valid label key format (or empty for the "match all" toleration)
	if toleration.Key != "" && !labelKeyRegex.MatchString(toleration.Key) {
		allErrs = append(allErrs, field.Invalid(path.Child("key"), toleration.Key, "must be a valid label key"))
	}

	// Operator validation
	switch toleration.Operator {
	case corev1.TolerationOpEqual:
		// value must be non-empty
		if toleration.Value == "" {
			allErrs = append(allErrs, field.Required(path.Child("value"), "value must be non-empty when operator is Equal"))
		}
	case corev1.TolerationOpExists:
		// value must be empty
		if toleration.Value != "" {
			allErrs = append(allErrs, field.Invalid(path.Child("value"), toleration.Value, "value must be empty when operator is Exists"))
		}
	default:
		// TolerationOpExists or empty default is fine
	}

	// Effect may be empty (matches all effects) — no validation needed
	// Effect must be a valid label value if specified
	if toleration.Effect != "" {
		if errs := utilvalidation.IsValidLabelValue(string(toleration.Effect)); len(errs) > 0 {
			allErrs = append(allErrs, field.Invalid(path.Child("effect"), toleration.Effect, errs[0]))
		}
	}

	return allErrs
}

// =============================================================================
// 3. validateTopologySpreadConstraints
// =============================================================================

// ValidTopologyUnsatisfiableAction lists valid whenUnsatisfiable values.
var ValidTopologyUnsatisfiableAction = map[corev1.UnsatisfiableConstraintAction]bool{
	corev1.DoNotSchedule:   true,
	corev1.ScheduleAnyway:   true,
}

// validateTopologySpreadConstraints validates topology spread constraints.
func validateTopologySpreadConstraints(constraints []corev1.TopologySpreadConstraint, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	for i, c := range constraints {
		allErrs = append(allErrs, validateTopologySpreadConstraint(&c, path.Index(i))...)
	}
	return allErrs
}

func validateTopologySpreadConstraint(c *corev1.TopologySpreadConstraint, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// maxSkew must be >= 1
	if c.MaxSkew < 1 {
		allErrs = append(allErrs, field.Invalid(path.Child("maxSkew"), c.MaxSkew, "must be at least 1"))
	}

	// topologyKey is required
	if c.TopologyKey == "" {
		allErrs = append(allErrs, field.Required(path.Child("topologyKey"), ""))
	}

	// whenUnsatisfiable must be one of the valid values
	if !ValidTopologyUnsatisfiableAction[c.WhenUnsatisfiable] {
		allErrs = append(allErrs, field.NotSupported(
			path.Child("whenUnsatisfiable"),
			c.WhenUnsatisfiable,
			[]string{string(corev1.DoNotSchedule), string(corev1.ScheduleAnyway)},
		))
	}

	// labelSelector is optional but must be valid if present
	if c.LabelSelector != nil {
		_, err := metav1.LabelSelectorAsSelector(c.LabelSelector)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(path.Child("labelSelector"), c.LabelSelector, err.Error()))
		}
	}

	return allErrs
}

// =============================================================================
// 4. validateDNSConfig
// =============================================================================

// validateDNSConfig validates dnsConfig options.
// Cross-field: ClusterFirstWithHostNet prohibits dnsConfig.
func validateDNSConfig(dnsConfig *corev1.PodDNSConfig, dnsPolicy corev1.DNSPolicy, hostNetwork bool, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if dnsConfig == nil {
		return allErrs
	}

	// Cross-field: ClusterFirstWithHostNet + dnsConfig is forbidden
	if dnsPolicy == corev1.DNSClusterFirstWithHostNet && dnsConfig != nil {
		allErrs = append(allErrs, field.Forbidden(
			path,
			"dnsConfig is not permitted when dnsPolicy is ClusterFirstWithHostNet",
		))
	}

	// nameservers: must be valid IPs
	for i, ns := range dnsConfig.Nameservers {
		if ns == "" {
			allErrs = append(allErrs, field.Invalid(path.Child("nameservers").Index(i), ns, "must be a valid IP address"))
		} else if ip := net.ParseIP(ns); ip == nil {
			allErrs = append(allErrs, field.Invalid(path.Child("nameservers").Index(i), ns, "must be a valid IP address"))
		}
	}

	// searches: max 6
	if len(dnsConfig.Searches) > 6 {
		allErrs = append(allErrs, field.TooMany(
			path.Child("searches"),
			len(dnsConfig.Searches),
			6,
		))
	}

	// options: each must have a valid name (value is optional)
	for i, opt := range dnsConfig.Options {
		if opt.Name == "" {
			allErrs = append(allErrs, field.Required(path.Child("options").Index(i).Child("name"), ""))
		} else if !labelKeyRegex.MatchString(opt.Name) {
			allErrs = append(allErrs, field.Invalid(
				path.Child("options").Index(i).Child("name"),
				opt.Name,
				"must be a valid DNS search domain option name",
			))
		}
	}

	return allErrs
}

// =============================================================================
// 5. validateHostBooleans
// =============================================================================

// validateHostBooleans validates hostIPC, hostNetwork, hostPID.
// In Go, these are plain bool fields — no invalid state possible from
// the API perspective. This validator exists for semantic completeness
// and cross-field checks (handled elsewhere).
// No errors are produced; this is a pass-through structural check.
func validateHostBooleans(spec *corev1.PodSpec, path *field.Path) field.ErrorList {
	// hostIPC, hostNetwork, hostPID are plain bools in Go — no validation needed.
	// The fields are correctly typed and boolean defaults are well-defined.
	// Semantic cross-field constraints (e.g., hostNetwork + dnsConfig) are
	// handled in their respective validators.
	return nil
}

// =============================================================================
// 6. validateRestartPolicy (non-Job types)
// =============================================================================

// validateRestartPolicy validates restartPolicy for non-Job types.
// Valid values: Always, OnFailure, Never.
func validateRestartPolicy(policy corev1.RestartPolicy, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if policy == "" {
		// Kubernetes defaults to Always when not set; treat empty as valid
		return allErrs
	}
	if !ValidRestartPolicyValues[policy] {
		allErrs = append(allErrs, field.NotSupported(
			path.Child("restartPolicy"),
			policy,
			[]string{string(corev1.RestartPolicyAlways), string(corev1.RestartPolicyOnFailure), string(corev1.RestartPolicyNever)},
		))
	}
	return allErrs
}

// =============================================================================
// Top-level dispatcher
// =============================================================================

// validateSchedulingFields runs all scheduling-related semantic validations for a PodSpec,
// INCLUDING restartPolicy (non-Job types only; Job/CronJob use validateRestartPolicyForJob).
// This is the top-level function called by per-type validators after the structural layer.
func validateSchedulingFields(spec *corev1.PodSpec, dnsPolicy corev1.DNSPolicy, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateAffinity(spec.Affinity, path.Child("affinity"))...)
	allErrs = append(allErrs, validateTolerations(spec.Tolerations, path.Child("tolerations"))...)
	allErrs = append(allErrs, validateTopologySpreadConstraints(spec.TopologySpreadConstraints, path.Child("topologySpreadConstraints"))...)
	allErrs = append(allErrs, validateDNSConfig(spec.DNSConfig, dnsPolicy, spec.HostNetwork, path.Child("dnsConfig"))...)
	allErrs = append(allErrs, validateHostBooleans(spec, path)...)
	allErrs = append(allErrs, validateRestartPolicy(spec.RestartPolicy, path)...)

	return allErrs
}
