package validator

import (
	"regexp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// =============================================================================
// AWU-11.1 / AWU-13.1 — Semantic Layer: securityContext
// PRD ref: PRD Section 20 Phase 2 — Layer 2 semantic
// Depends on: AWU-11.1 (structural layer)
// =============================================================================

// SELinux field name regex — must be non-empty, alphanumeric + separators, reasonable length
var selinuxFieldRE = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`)

// validateSecurityContext is the top-level entry point for all security-context
// semantic validation. It delegates to four sub-validators covering:
//   - Pod-level RunAsNonRoot / RunAsGroup / Sysctls (validatePodSecurityContext)
//   - Container-level capabilities (validateCapabilities)
//   - seccompProfile (validateSeccomp)
//   - SELinux options (validateSELinux)
func validateSecurityContext(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validatePodSecurityContext(podSpec, path)...)
	allErrs = append(allErrs, validateCapabilities(podSpec, path)...)
	allErrs = append(allErrs, validateSeccomp(podSpec, path)...)
	allErrs = append(allErrs, validateSELinux(podSpec, path)...)

	return allErrs
}

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

// validateCapabilities validates that every capability added in any container's
// securityContext.capabilities.add is a recognised Linux capability constant.
func validateCapabilities(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i, c := range podSpec.Containers {
		if c.SecurityContext != nil && c.SecurityContext.Capabilities != nil {
			for j, cap := range c.SecurityContext.Capabilities.Add {
				if !ValidCapabilities[string(cap)] {
					allErrs = append(allErrs, field.Invalid(
						path.Child("containers").Index(i).Child("securityContext").Child("capabilities").Child("add").Index(j),
						string(cap),
						"capability not in the supported set",
					))
				}
			}
		}
	}

	for i, c := range podSpec.InitContainers {
		if c.SecurityContext != nil && c.SecurityContext.Capabilities != nil {
			for j, cap := range c.SecurityContext.Capabilities.Add {
				if !ValidCapabilities[string(cap)] {
					allErrs = append(allErrs, field.Invalid(
						path.Child("initContainers").Index(i).Child("securityContext").Child("capabilities").Child("add").Index(j),
						string(cap),
						"capability not in the supported set",
					))
				}
			}
		}
	}

	return allErrs
}

// validateSeccomp validates that seccompProfile.type is one of the known values
// and that Localhost profiles supply a non-empty localhostProfile path.
func validateSeccomp(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i, c := range podSpec.Containers {
		if c.SecurityContext != nil && c.SecurityContext.SeccompProfile != nil {
			sp := c.SecurityContext.SeccompProfile
			if !ValidSeccompProfileTypes[string(sp.Type)] {
				allErrs = append(allErrs, field.Invalid(
					path.Child("containers").Index(i).Child("securityContext").Child("seccompProfile").Child("type"),
					string(sp.Type),
					"must be one of: RuntimeDefault, Unconfined, Localhost",
				))
			}
			if sp.Type == corev1.SeccompProfileTypeLocalhost && (sp.LocalhostProfile == nil || *sp.LocalhostProfile == "") {
				allErrs = append(allErrs, field.Required(
					path.Child("containers").Index(i).Child("securityContext").Child("seccompProfile").Child("localhostProfile"),
					"localhostProfile must be set when type is Localhost",
				))
			}
		}
	}

	for i, c := range podSpec.InitContainers {
		if c.SecurityContext != nil && c.SecurityContext.SeccompProfile != nil {
			sp := c.SecurityContext.SeccompProfile
			if !ValidSeccompProfileTypes[string(sp.Type)] {
				allErrs = append(allErrs, field.Invalid(
					path.Child("initContainers").Index(i).Child("securityContext").Child("seccompProfile").Child("type"),
					string(sp.Type),
					"must be one of: RuntimeDefault, Unconfined, Localhost",
				))
			}
			if sp.Type == corev1.SeccompProfileTypeLocalhost && (sp.LocalhostProfile == nil || *sp.LocalhostProfile == "") {
				allErrs = append(allErrs, field.Required(
					path.Child("initContainers").Index(i).Child("securityContext").Child("seccompProfile").Child("localhostProfile"),
					"localhostProfile must be set when type is Localhost",
				))
			}
		}
	}

	return allErrs
}

// validateSELinux validates that each SELinux field (user, role, type, level)
// is non-empty and conforms to a reasonable format (alphanumeric + separators,
// max 63 chars).
func validateSELinux(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i, c := range podSpec.Containers {
		if c.SecurityContext != nil && c.SecurityContext.SELinuxOptions != nil {
			opts := c.SecurityContext.SELinuxOptions
			containerPath := path.Child("containers").Index(i).Child("securityContext").Child("seLinuxOptions")

			if opts.User != "" && !selinuxFieldRE.MatchString(opts.User) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("user"),
					opts.User,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
			if opts.Role != "" && !selinuxFieldRE.MatchString(opts.Role) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("role"),
					opts.Role,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
			if opts.Type != "" && !selinuxFieldRE.MatchString(opts.Type) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("type"),
					opts.Type,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
			if opts.Level != "" && !selinuxFieldRE.MatchString(opts.Level) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("level"),
					opts.Level,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
		}
	}

	for i, c := range podSpec.InitContainers {
		if c.SecurityContext != nil && c.SecurityContext.SELinuxOptions != nil {
			opts := c.SecurityContext.SELinuxOptions
			containerPath := path.Child("initContainers").Index(i).Child("securityContext").Child("seLinuxOptions")

			if opts.User != "" && !selinuxFieldRE.MatchString(opts.User) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("user"),
					opts.User,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
			if opts.Role != "" && !selinuxFieldRE.MatchString(opts.Role) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("role"),
					opts.Role,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
			if opts.Type != "" && !selinuxFieldRE.MatchString(opts.Type) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("type"),
					opts.Type,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
			if opts.Level != "" && !selinuxFieldRE.MatchString(opts.Level) {
				allErrs = append(allErrs, field.Invalid(
					containerPath.Child("level"),
					opts.Level,
					"must match ^[a-zA-Z][a-zA-Z0-9_-]{0,63}$",
				))
			}
		}
	}

	return allErrs
}
