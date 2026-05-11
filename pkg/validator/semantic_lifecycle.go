package validator

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// =============================================================================
// AWU-12.1 — Semantic Layer: lifecycle hooks, initContainer probes, restartPolicy
// PRD ref: PRD Section 20 Phase 2 — Layer 2 semantic
// =============================================================================

// validateInitContainerProbes validates that no initContainer has any probe.
// Kubernetes rejects init containers with probes at admission time.
// See: k8s.io/kubernetes/pkg/apis/core/validation/validation.go — validateContainerProbes
func validateInitContainerProbes(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for i := range podSpec.InitContainers {
		containerPath := path.Child("initContainers").Index(i)
		if podSpec.InitContainers[i].LivenessProbe != nil {
			allErrs = append(allErrs, field.Forbidden(containerPath.Child("livenessProbe"), "init container must not have liveness probe"))
		}
		if podSpec.InitContainers[i].ReadinessProbe != nil {
			allErrs = append(allErrs, field.Forbidden(containerPath.Child("readinessProbe"), "init container must not have readiness probe"))
		}
		if podSpec.InitContainers[i].StartupProbe != nil {
			allErrs = append(allErrs, field.Forbidden(containerPath.Child("startupProbe"), "init container must not have startup probe"))
		}
	}

	return allErrs
}

// validateLifecycleHookHandlerCount checks that a lifecycle hook has exactly one handler.
// Valid handlers are: httpGet, tcpSocket, exec — not zero, not multiple.
func validateLifecycleHookHandlerCount(hook *corev1.LifecycleHandler, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if hook == nil {
		return allErrs
	}

	handlerCount := 0
	if hook.HTTPGet != nil {
		handlerCount++
	}
	if hook.TCPSocket != nil {
		handlerCount++
	}
	if hook.Exec != nil {
		handlerCount++
	}

	if handlerCount == 0 {
		allErrs = append(allErrs, field.Invalid(path, hook, "must specify exactly one of httpGet, tcpSocket, or exec"))
	} else if handlerCount > 1 {
		allErrs = append(allErrs, field.Invalid(path, hook, "may not specify more than one handler"))
	}

	return allErrs
}

// validateLifecycleHooks validates lifecycle preStop and postStart handlers on all containers.
// Each hook must have exactly one handler (httpGet, tcpSocket, exec, or grpc).
// See: k8s.io/kubernetes/pkg/apis/core/validation/validation.go — validateLifecycle
func validateLifecycleHooks(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Check regular containers
	for i := range podSpec.Containers {
		containerPath := path.Child("containers").Index(i)
		if podSpec.Containers[i].Lifecycle != nil {
			allErrs = append(allErrs, validateLifecycleHookHandlerCount(podSpec.Containers[i].Lifecycle.PreStop, containerPath.Child("lifecycle").Child("preStop"))...)
			allErrs = append(allErrs, validateLifecycleHookHandlerCount(podSpec.Containers[i].Lifecycle.PostStart, containerPath.Child("lifecycle").Child("postStart"))...)
		}
	}

	// Check init containers
	for i := range podSpec.InitContainers {
		containerPath := path.Child("initContainers").Index(i)
		if podSpec.InitContainers[i].Lifecycle != nil {
			allErrs = append(allErrs, validateLifecycleHookHandlerCount(podSpec.InitContainers[i].Lifecycle.PreStop, containerPath.Child("lifecycle").Child("preStop"))...)
			allErrs = append(allErrs, validateLifecycleHookHandlerCount(podSpec.InitContainers[i].Lifecycle.PostStart, containerPath.Child("lifecycle").Child("postStart"))...)
		}
	}

	return allErrs
}

// ValidRestartPolicyValues is the set of valid restartPolicy values.
var ValidRestartPolicyValues = map[corev1.RestartPolicy]bool{
	corev1.RestartPolicyAlways:    true,
	corev1.RestartPolicyOnFailure: true,
	corev1.RestartPolicyNever:     true,
}

// validateRestartPolicyForJob validates that Job/CronJob pods have restartPolicy Never or OnFailure.
// Kubernetes requires Job pods to not restart automatically (restartPolicy must not be Always).
// See: k8s.io/kubernetes/pkg/apis/batch/validation/validation.go — ValidateJobSpec
func validateRestartPolicyForJob(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if podSpec.RestartPolicy == corev1.RestartPolicyAlways {
		allErrs = append(allErrs, field.NotSupported(path.Child("restartPolicy"), podSpec.RestartPolicy, []string{
			string(corev1.RestartPolicyNever), string(corev1.RestartPolicyOnFailure),
		}))
	}

	return allErrs
}

// maxResourceRequests computes the maximum resource request for a given resource name
// across all containers in the slice. Returns nil if no request is set.
func maxResourceRequests(containers []corev1.Container, resourceName corev1.ResourceName) *resource.Quantity {
	var max *resource.Quantity
	for _, c := range containers {
		if req, ok := c.Resources.Requests[resourceName]; ok {
			if max == nil || req.Cmp(*max) > 0 {
				max = &req
			}
		}
	}
	return max
}

// validateInitContainerResources validates that init container resource requests do not exceed
// the maximum request of any regular container for each resource type.
// Kubernetes enforces this at admission time to ensure init containers don't reserve more
// than regular containers will have available.
// See: k8s.io/kubernetes/pkg/apis/core/validation/validation.go — validateInitContainersResourceRequirements
func validateInitContainerResources(podSpec *corev1.PodSpec, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if len(podSpec.Containers) == 0 || len(podSpec.InitContainers) == 0 {
		return allErrs
	}

	// Pre-compute max requests per resource type for regular containers
	maxRequests := make(map[corev1.ResourceName]*resource.Quantity)
	for _, c := range podSpec.Containers {
		for k, v := range c.Resources.Requests {
			if existing, ok := maxRequests[k]; ok {
				if v.Cmp(*existing) > 0 {
					maxRequests[k] = &v
				}
			} else {
				maxRequests[k] = &v
			}
		}
	}

	// Check each init container's requests against the max
	for i := range podSpec.InitContainers {
		initPath := path.Child("initContainers").Index(i)
		for k, initReq := range podSpec.InitContainers[i].Resources.Requests {
			if maxReq, exists := maxRequests[k]; exists {
				if initReq.Cmp(*maxReq) > 0 {
					allErrs = append(allErrs, field.Invalid(
						initPath.Child("resources").Child("requests").Key(string(k)),
						initReq.String(),
						"init container must not have resource request greater than container maximum",
					))
				}
			}
		}
	}

	return allErrs
}