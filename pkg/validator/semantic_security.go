package validator

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// =============================================================================
// AWU-11.1 — Semantic Layer: securityContext (PodSecurityContext + per-container SecurityContext)
// PRD ref: PRD Section 20 Phase 2 — Layer 2 semantic
// Depends on: AWU-11.1 (structural layer)
// =============================================================================

// validatePodSecurityContext validates semantic constraints on the pod-level security context.
// Structural validation of field types is handled by the structural layer.
// This validator handles cross-field semantic constraints (e.g., RunAsNonRoot + RunAsUser).
func validatePodSecurityContext(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if podSpec.SecurityContext == nil {
		return allErrs
	}
	sc := podSpec.SecurityContext

	// RunAsNonRoot=true with RunAsUser=0 is contradictory — k8s rejects this at admission
	if sc.RunAsNonRoot != nil && *sc.RunAsNonRoot && sc.RunAsUser != nil && *sc.RunAsUser == 0 {
		allErrs = append(allErrs, field.Invalid(
			path.Child("runAsUser"),
			*sc.RunAsUser,
			"runAsNonRoot is true but runAsUser is 0 (root). Pods cannot run as root when runAsNonRoot is set.",
		))
	}

	// RunAsGroup must be non-negative if set
	if sc.RunAsGroup != nil && *sc.RunAsGroup < 0 {
		allErrs = append(allErrs, field.Invalid(
			path.Child("runAsGroup"),
			*sc.RunAsGroup,
			"must be >= 0",
		))
	}

	// Sysctls: validate each sysctl name is namespaced and valid
	for i, sysctl := range sc.Sysctls {
		if sysctl.Name == "" {
			allErrs = append(allErrs, field.Required(
				path.Child("sysctls").Index(i).Child("name"),
				"sysctl name must not be empty",
			))
		}
		// Kubernetes sysctls must be namespaced (contain a dot); safe sysctls are prefixed "kernel."
		// Reject obviously invalid names (empty, leading dot without prefix, etc.)
		if sysctl.Name != "" && sysctl.Name[0] == '.' {
			allErrs = append(allErrs, field.Invalid(
				path.Child("sysctls").Index(i).Child("name"),
				sysctl.Name,
				"sysctl name must be namespaced (contain at least one dot); safe sysctls are prefixed kernel.",
			))
		}
	}

	return allErrs
}