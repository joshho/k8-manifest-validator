package validator

import (
	"fmt"
	"strings"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8-manifest-validator/pkg/types"
)

// ValidateNameFunc is the function signature required by k8s validation
type ValidateNameFunc func(name string, prefix bool) []string

// routingEntry holds a validator function and its decoder for a specific kind
type routingEntry struct {
	validateFn func(runtime.Object) field.ErrorList
	decodeFn   func([]byte) (runtime.Object, error)
	apiVersion string
}

// BuiltinValidator validates Kubernetes built-in resource types using
// the k8s.io/apimachinery validation packages
type BuiltinValidator struct {
	routing sync.Map // kind string -> routingEntry
	scheme  *runtime.Scheme
	decoder runtime.Decoder
}

// NewBuiltinValidator creates a new BuiltinValidator with all built-in kinds registered
func NewBuiltinValidator() *BuiltinValidator {
	bv := &BuiltinValidator{
		scheme: runtime.NewScheme(),
	}

	// Register all built-in types with the scheme
	_ = corev1.AddToScheme(bv.scheme)
	_ = appsv1.AddToScheme(bv.scheme)
	_ = networkingv1.AddToScheme(bv.scheme)
	_ = batchv1.AddToScheme(bv.scheme)
	_ = rbacv1.AddToScheme(bv.scheme)
	_ = autoscalingv1.AddToScheme(bv.scheme)

	// Set up decoder using UniversalDeserializer
	codecFactory := serializer.NewCodecFactory(bv.scheme)
	bv.decoder = codecFactory.UniversalDeserializer()

	// apps/v1 kinds
	bv.registerDeployment()
	bv.registerStatefulSet()
	bv.registerDaemonSet()
	bv.registerReplicaSet()

	// core/v1 kinds
	bv.registerPod()
	bv.registerService()
	bv.registerConfigMap()
	bv.registerSecret()
	bv.registerReplicationController()
	bv.registerPVC()
	bv.registerNamespace()
	bv.registerServiceAccount()
	bv.registerEndpoints()

	// networking.k8s.io/v1 kinds
	bv.registerIngress()
	bv.registerNetworkPolicy()

	// batch/v1 kinds
	bv.registerJob()
	bv.registerCronJob()
	bv.registerHorizontalPodAutoscaler()
	bv.registerLimitRange()
	bv.registerList()

	// rbac.authorization.k8s.io/v1 kinds
	bv.registerRole()
	bv.registerClusterRole()
	bv.registerRoleBinding()
	bv.registerClusterRoleBinding()

	return bv
}

// registerKind registers a kind with its validator and decoder
func (v *BuiltinValidator) registerKind(kind string, apiVersion string, validateFn func(runtime.Object) field.ErrorList, decodeFn func([]byte) (runtime.Object, error)) {
	v.routing.Store(kind, routingEntry{
		validateFn: validateFn,
		decodeFn:   decodeFn,
		apiVersion: apiVersion,
	})
}

// CanValidate returns true if the validator can validate the given kind
func (v *BuiltinValidator) CanValidate(kind string) bool {
	_, ok := v.routing.Load(kind)
	return ok
}

// ValidateKind validates an object by kind
func (v *BuiltinValidator) ValidateKind(kind string, obj runtime.Object) field.ErrorList {
	entry, ok := v.routing.Load(kind)
	if !ok {
		return field.ErrorList{
			field.Invalid(field.NewPath("kind"), kind, "unknown kind: no validator registered"),
		}
	}
	return entry.(routingEntry).validateFn(obj)
}

// ValidateResource validates a raw YAML manifest and returns a Result
func (v *BuiltinValidator) ValidateResource(data []byte) *Result {
	decoder := NewDecoder()
	manifest, err := decoder.DecodeManifest(data)
	if err != nil {
		return &Result{
			Status: types.StatusError,
			Errors: []ErrorItem{
				{Field: "manifest", Message: err.Error(), Code: types.ErrCodeInvalid},
			},
		}
	}

	return v.ValidateManifest(manifest, data)
}

// ValidateManifest validates a decoded manifest using built-in validation
func (v *BuiltinValidator) ValidateManifest(manifest Manifest, rawData []byte) *Result {
	if !v.CanValidate(manifest.Kind) {
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusSkipped,
			Errors: []ErrorItem{
				{Field: "kind", Message: fmt.Sprintf("kind %q is not a built-in type or not supported", manifest.Kind), Code: types.ErrCodeInvalid},
			},
		}
	}

	entry, ok := v.routing.Load(manifest.Kind)
	if !ok {
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusError,
			Errors: []ErrorItem{
				{Field: "kind", Message: "internal error: no routing entry", Code: types.ErrCodeInvalid},
			},
		}
	}

	obj, err := entry.(routingEntry).decodeFn(rawData)
	if err != nil {
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusError,
			Errors: []ErrorItem{
				{Field: "spec", Message: fmt.Sprintf("failed to decode object: %v", err), Code: types.ErrCodeInvalid},
			},
		}
	}

	// Apply thin namespace defaulter: set default namespace for namespaced resources
	// that lack one. This mirrors the k8s API server's DefaultNamespace admission
	// plugin. In real operational YAML, namespace-scoped resources often omit the
	// namespace field and rely on the API server to default it.
	if objMeta, ok := obj.(metav1.Object); ok {
		if objMeta.GetNamespace() == "" && isNamespacedKind(manifest.Kind) {
			objMeta.SetNamespace(metav1.NamespaceDefault)
		}
	}

	// Apply admission-level defaults: ReplicationController selector can be
	// defaulted from pod template labels (k8s admission behavior).
	if rc, ok := obj.(*corev1.ReplicationController); ok {
		if len(rc.Spec.Selector) == 0 && rc.Spec.Template != nil && len(rc.Spec.Template.Labels) > 0 {
			rc.Spec.Selector = rc.Spec.Template.Labels
		}
	}

	errs := entry.(routingEntry).validateFn(obj)
	if len(errs) > 0 {
		errorItems := make([]ErrorItem, 0, len(errs))
		for _, e := range errs {
			errorItems = append(errorItems, ErrorItem{
				Field:   e.Field,
				Message: e.Detail,
				Code:    string(e.Type),
			})
		}
		return &Result{
			Kind:       manifest.Kind,
			Name:       manifest.Metadata.Name,
			Namespace:  manifest.Metadata.Namespace,
			APIVersion: manifest.APIVersion,
			Status:     types.StatusInvalid,
			Errors:     errorItems,
		}
	}

	return &Result{
		Kind:       manifest.Kind,
		Name:       manifest.Metadata.Name,
		Namespace:  manifest.Metadata.Namespace,
		APIVersion: manifest.APIVersion,
		Status:     types.StatusValid,
		Errors:     nil,
	}
}

// GetRegisteredKinds returns all registered kind strings
func (v *BuiltinValidator) GetRegisteredKinds() []string {
	kinds := []string{}
	v.routing.Range(func(key, value interface{}) bool {
		kinds = append(kinds, key.(string))
		return true
	})
	return kinds
}

// decode decodes YAML/JSON bytes using the universal decoder
func (v *BuiltinValidator) decode(data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	return v.decoder.Decode(data, nil, nil)
}

// Registration helpers - each creates a decode function using the scheme

func (v *BuiltinValidator) registerDeployment() {
	v.registerKind("Deployment", "apps/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateDeployment(obj.(*appsv1.Deployment))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerStatefulSet() {
	v.registerKind("StatefulSet", "apps/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateStatefulSet(obj.(*appsv1.StatefulSet))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerDaemonSet() {
	v.registerKind("DaemonSet", "apps/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateDaemonSet(obj.(*appsv1.DaemonSet))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerReplicaSet() {
	v.registerKind("ReplicaSet", "apps/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateReplicaSet(obj.(*appsv1.ReplicaSet))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerPod() {
	v.registerKind("Pod", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validatePod(obj.(*corev1.Pod))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerService() {
	v.registerKind("Service", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateService(obj.(*corev1.Service))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerConfigMap() {
	v.registerKind("ConfigMap", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateConfigMap(obj.(*corev1.ConfigMap))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerReplicationController() {
	v.registerKind("ReplicationController", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateReplicationController(obj.(*corev1.ReplicationController))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerSecret() {
	v.registerKind("Secret", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateSecret(obj.(*corev1.Secret))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerPVC() {
	v.registerKind("PersistentVolumeClaim", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validatePVC(obj.(*corev1.PersistentVolumeClaim))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerNamespace() {
	v.registerKind("Namespace", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateNamespace(obj.(*corev1.Namespace))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerServiceAccount() {
	v.registerKind("ServiceAccount", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateServiceAccount(obj.(*corev1.ServiceAccount))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerEndpoints() {
	v.registerKind("Endpoints", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateEndpoints(obj.(*corev1.Endpoints))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerIngress() {
	v.registerKind("Ingress", "networking.k8s.io/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateIngress(obj.(*networkingv1.Ingress))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerNetworkPolicy() {
	v.registerKind("NetworkPolicy", "networking.k8s.io/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateNetworkPolicy(obj.(*networkingv1.NetworkPolicy))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerJob() {
	v.registerKind("Job", "batch/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateJob(obj.(*batchv1.Job))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerCronJob() {
	v.registerKind("CronJob", "batch/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateCronJob(obj.(*batchv1.CronJob))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerRole() {
	v.registerKind("Role", "rbac.authorization.k8s.io/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateRole(obj.(*rbacv1.Role))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerClusterRole() {
	v.registerKind("ClusterRole", "rbac.authorization.k8s.io/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateClusterRole(obj.(*rbacv1.ClusterRole))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerRoleBinding() {
	v.registerKind("RoleBinding", "rbac.authorization.k8s.io/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateRoleBinding(obj.(*rbacv1.RoleBinding))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerClusterRoleBinding() {
	v.registerKind("ClusterRoleBinding", "rbac.authorization.k8s.io/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateClusterRoleBinding(obj.(*rbacv1.ClusterRoleBinding))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

// nameValidator is an adapter to use utilvalidation.IsQualifiedName with k8s validation
func nameValidator(name string, prefix bool) []string {
	return utilvalidation.IsQualifiedName(name)
}

// =============================================================================
// PodSpec validation (AWU-1): shared helper + sub-helpers
// =============================================================================

// validatePodSpec validates all Phase 1 fields of a Kubernetes PodSpec.
// skipProbes, when true, skips liveness/readiness/startup probe validation.
// extraVolumeNames allows passing additional valid volume names (e.g., from StatefulSet volumeClaimTemplates).
func validatePodSpec(podSpec *corev1.PodSpec, path *field.Path, skipProbes bool, extraVolumeNames ...map[string]struct{}) field.ErrorList {
	var allErrs field.ErrorList
	if podSpec == nil {
		return allErrs
	}

	// Validate volumes first to build volume name set for mount validation
	volumeNames, volErrs := validateVolumes(podSpec.Volumes, path.Child("volumes"), extraVolumeNames...)
	allErrs = append(allErrs, volErrs...)

	// Validate containers, collecting name set for initContainer cross-check
	containerNameSet := make(map[string]bool)
	allErrs = append(allErrs, validateContainers(podSpec.Containers, false, volumeNames, containerNameSet, path.Child("containers"), skipProbes)...)

	// Validate init containers against existing container names
	allErrs = append(allErrs, validateContainers(podSpec.InitContainers, true, volumeNames, containerNameSet, path.Child("initContainers"), skipProbes)...)

	// Validate serviceAccountName
	allErrs = append(allErrs, validateServiceAccountName(podSpec.ServiceAccountName, path.Child("serviceAccountName"))...)

	return allErrs
}

// validateContainerPort validates a single ContainerPort.
func validateContainerPort(port corev1.ContainerPort, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if port.ContainerPort < 1 || port.ContainerPort > 65535 {
		allErrs = append(allErrs, field.Invalid(path.Child("containerPort"), port.ContainerPort, "must be between 1 and 65535"))
	}
	if port.HostPort < 0 || port.HostPort > 65535 {
		allErrs = append(allErrs, field.Invalid(path.Child("hostPort"), port.HostPort, "must be between 0 and 65535"))
	}
	if port.Protocol != "" && port.Protocol != corev1.ProtocolTCP && port.Protocol != corev1.ProtocolUDP && port.Protocol != corev1.ProtocolSCTP {
		allErrs = append(allErrs, field.Invalid(path.Child("protocol"), port.Protocol, "must be one of TCP, UDP, SCTP"))
	}
	return allErrs
}

// validateContainers validates spec.containers or spec.initContainers.
// initContainer is true when validating initContainers.
// volumeNames is the set of declared volume names for mount validation.
// containerNameSet tracks seen names for duplicate detection (cross-container + initContainer).
// skipProbes, when true, skips probe validation.
func validateContainers(containers []corev1.Container, initContainer bool, volumeNames map[string]struct{}, containerNameSet map[string]bool, path *field.Path, skipProbes bool) field.ErrorList {
	var allErrs field.ErrorList

	for i, c := range containers {
		// Container name: required, DNS-1123 label, unique across containers+initContainers
		if c.Name == "" {
			allErrs = append(allErrs, field.Required(path.Index(i).Child("name"), ""))
		} else {
			if errs := utilvalidation.IsDNS1123Label(c.Name); len(errs) > 0 {
				allErrs = append(allErrs, field.Invalid(path.Index(i).Child("name"), c.Name, errs[0]))
			} else if containerNameSet[c.Name] {
				allErrs = append(allErrs, field.Duplicate(path.Index(i).Child("name"), c.Name))
			}
			containerNameSet[c.Name] = true
		}

		// Container image: non-empty
		if c.Image == "" {
			allErrs = append(allErrs, field.Required(path.Index(i).Child("image"), ""))
		}

		// Ports
		for j, p := range c.Ports {
			allErrs = append(allErrs, validateContainerPort(p, path.Index(i).Child("ports").Index(j))...)
		}

		// Env vars
		allErrs = append(allErrs, validateEnvVar(c.Env, path.Index(i).Child("env"))...)

		// Resources
		allErrs = append(allErrs, validateResourceRequirements(c.Resources.Limits, c.Resources.Requests, path.Index(i).Child("resources"))...)

		// Volume mounts
		for j, m := range c.VolumeMounts {
			allErrs = append(allErrs, validateVolumeMount(m, volumeNames, path.Index(i).Child("volumeMounts").Index(j))...)
		}

		// Probes (skipped for batch workloads)
		if !skipProbes {
			if c.LivenessProbe != nil {
				allErrs = append(allErrs, validateProbe(c.LivenessProbe, path.Index(i).Child("livenessProbe"))...)
			}
			if c.ReadinessProbe != nil {
				allErrs = append(allErrs, validateProbe(c.ReadinessProbe, path.Index(i).Child("readinessProbe"))...)
			}
			if c.StartupProbe != nil {
				allErrs = append(allErrs, validateProbe(c.StartupProbe, path.Index(i).Child("startupProbe"))...)
			}
		}
	}

	return allErrs
}

// validateEnvVar validates the env field of a container.
func validateEnvVar(env []corev1.EnvVar, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	for i, e := range env {
		if e.Name == "" {
			allErrs = append(allErrs, field.Required(path.Index(i).Child("name"), ""))
		} else if len(utilvalidation.IsCIdentifier(e.Name)) > 0 {
			allErrs = append(allErrs, field.Invalid(path.Index(i).Child("name"), e.Name, "must be a valid C identifier"))
		}
	}
	return allErrs
}

// validateProbe validates that exactly one handler is present in a probe.
func validateProbe(probe *corev1.Probe, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if probe == nil {
		return allErrs
	}

	handlerCount := 0
	if probe.ProbeHandler.HTTPGet != nil {
		handlerCount++
	}
	if probe.ProbeHandler.TCPSocket != nil {
		handlerCount++
	}
	if probe.ProbeHandler.Exec != nil {
		handlerCount++
	}
	if probe.ProbeHandler.GRPC != nil {
		handlerCount++
	}

	if handlerCount == 0 {
		allErrs = append(allErrs, field.Invalid(path, probe, "exactly one of httpGet, tcpSocket, exec, or grpc must be specified"))
	} else if handlerCount > 1 {
		allErrs = append(allErrs, field.Invalid(path, probe, "only one of httpGet, tcpSocket, exec, or grpc can be specified"))
	}

	return allErrs
}

// validateResourceRequirements validates that resource limits and requests contain parseable quantities.
func validateResourceRequirements(limits, requests corev1.ResourceList, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if limits != nil {
		for k, v := range limits {
			if _, err := resource.ParseQuantity(v.String()); err != nil {
				allErrs = append(allErrs, field.Invalid(path.Child("limits").Key(string(k)), v.String(), "unable to parse resource quantity"))
			}
		}
	}

	if requests != nil {
		for k, v := range requests {
			if _, err := resource.ParseQuantity(v.String()); err != nil {
				allErrs = append(allErrs, field.Invalid(path.Child("requests").Key(string(k)), v.String(), "unable to parse resource quantity"))
			}
		}
	}

	return allErrs
}

// validateServiceAccountName validates the serviceAccountName field (DNS-1123 subdomain).
func validateServiceAccountName(sa string, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if sa != "" {
		if errs := utilvalidation.IsDNS1123Subdomain(sa); len(errs) > 0 {
			allErrs = append(allErrs, field.Invalid(path, sa, errs[0]))
		}
	}
	return allErrs
}

// validateVolumeMount validates a single VolumeMount.
func validateVolumeMount(mount corev1.VolumeMount, volumeNameSet map[string]struct{}, path *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// mountPath must be absolute
	if !strings.HasPrefix(mount.MountPath, "/") {
		allErrs = append(allErrs, field.Invalid(path.Child("mountPath"), mount.MountPath, "must be an absolute path (start with /)"))
	}

	// name must reference a declared volume
	if mount.Name != "" && volumeNameSet != nil {
		if _, exists := volumeNameSet[mount.Name]; !exists {
			allErrs = append(allErrs, field.Invalid(path.Child("name"), mount.Name, "not found in volumes"))
		}
	}

	return allErrs
}

// validateVolumes validates volume names (DNS-1123 label, unique) and returns the set of volume names.
// extraNames allows passing additional valid volume names that aren't in the volumes list (e.g., from StatefulSet volumeClaimTemplates).
func validateVolumes(volumes []corev1.Volume, path *field.Path, extraNames ...map[string]struct{}) (map[string]struct{}, field.ErrorList) {
	var allErrs field.ErrorList
	volumeNames := make(map[string]struct{})
	seen := make(map[string]bool)

	for i, vol := range volumes {
		if vol.Name == "" {
			allErrs = append(allErrs, field.Required(path.Index(i).Child("name"), ""))
		} else {
			if errs := utilvalidation.IsDNS1123Label(vol.Name); len(errs) > 0 {
				allErrs = append(allErrs, field.Invalid(path.Index(i).Child("name"), vol.Name, errs[0]))
			} else if seen[vol.Name] {
				allErrs = append(allErrs, field.Duplicate(path.Index(i).Child("name"), vol.Name))
			}
			seen[vol.Name] = true
			volumeNames[vol.Name] = struct{}{}
		}
	}

	// Merge extra volume names (e.g., from StatefulSet volumeClaimTemplates)
	for _, extra := range extraNames {
		for name := range extra {
			volumeNames[name] = struct{}{}
		}
	}

	return volumeNames, allErrs
}

// Validation functions using k8s.io/apimachinery/pkg/api/validation

func (v *BuiltinValidator) validateDeployment(deploy *appsv1.Deployment) field.ErrorList {
	var allErrs field.ErrorList

	// Validate metadata using the standard k8s validation
	allErrs = append(allErrs, validation.ValidateObjectMeta(&deploy.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	// Validate spec
	if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "replicas"), *deploy.Spec.Replicas, "must be >= 0"))
	}

	// Validate selector
	if deploy.Spec.Selector != nil {
		selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), deploy.Spec.Selector, err.Error()))
		} else if selector.Empty() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), deploy.Spec.Selector, "must be specified"))
		}
	}

	// Validate template
	if deploy.Spec.Template.ObjectMeta.Name != "" || deploy.Spec.Template.ObjectMeta.Namespace != "" {
		allErrs = append(allErrs, validation.ValidateObjectMeta(&deploy.Spec.Template.ObjectMeta, false, nameValidator, field.NewPath("spec", "template", "metadata"))...)
	}

	// PodSpec validation
	allErrs = append(allErrs, validatePodSpec(&deploy.Spec.Template.Spec, field.NewPath("spec", "template", "spec"), false)...)

	return allErrs
}

func (v *BuiltinValidator) validateStatefulSet(ss *appsv1.StatefulSet) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&ss.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if ss.Spec.Replicas != nil && *ss.Spec.Replicas < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "replicas"), *ss.Spec.Replicas, "must be >= 0"))
	}

	if ss.Spec.Selector != nil {
		selector, err := metav1.LabelSelectorAsSelector(ss.Spec.Selector)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), ss.Spec.Selector, err.Error()))
		} else if selector.Empty() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), ss.Spec.Selector, "must be specified"))
		}
	}

	// Collect volumeClaimTemplate names as valid volume names for mount validation
	vctNames := make(map[string]struct{})
	for _, vct := range ss.Spec.VolumeClaimTemplates {
		vctNames[vct.Name] = struct{}{}
	}

	// PodSpec validation (pass volumeClaimTemplate names as extra valid volume names)
	allErrs = append(allErrs, validatePodSpec(&ss.Spec.Template.Spec, field.NewPath("spec", "template", "spec"), false, vctNames)...)

	return allErrs
}

func (v *BuiltinValidator) validateDaemonSet(ds *appsv1.DaemonSet) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&ds.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if ds.Spec.Selector != nil {
		selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), ds.Spec.Selector, err.Error()))
		} else if selector.Empty() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), ds.Spec.Selector, "must be specified"))
		}
	}

	// PodSpec validation
	allErrs = append(allErrs, validatePodSpec(&ds.Spec.Template.Spec, field.NewPath("spec", "template", "spec"), false)...)

	return allErrs
}

func (v *BuiltinValidator) validateReplicaSet(rs *appsv1.ReplicaSet) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&rs.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if rs.Spec.Replicas != nil && *rs.Spec.Replicas < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "replicas"), *rs.Spec.Replicas, "must be >= 0"))
	}

	if rs.Spec.Selector != nil {
		selector, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), rs.Spec.Selector, err.Error()))
		} else if selector.Empty() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector"), rs.Spec.Selector, "must be specified"))
		}
	}

	// PodSpec validation
	allErrs = append(allErrs, validatePodSpec(&rs.Spec.Template.Spec, field.NewPath("spec", "template", "spec"), false)...)

	return allErrs
}

func (v *BuiltinValidator) validateReplicationController(rc *corev1.ReplicationController) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&rc.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if rc.Spec.Replicas != nil && *rc.Spec.Replicas < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "replicas"), *rc.Spec.Replicas, "must be >= 0"))
	}

	if len(rc.Spec.Selector) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "selector"), ""))
	}

	// AWU-4: validate PodSpec (skipProbes=false for long-running workload)
	allErrs = append(allErrs, validatePodSpec(&rc.Spec.Template.Spec, field.NewPath("spec", "template", "spec"), false)...)

	return allErrs
}

func (v *BuiltinValidator) validatePod(pod *corev1.Pod) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&pod.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	// Validate pod spec
	allErrs = append(allErrs, validatePodSpec(&pod.Spec, field.NewPath("spec"), false)...)

	return allErrs
}

func (v *BuiltinValidator) validateService(svc *corev1.Service) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&svc.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	// Validate service ports
	for i, port := range svc.Spec.Ports {
		if port.Port < 0 || port.Port > 65535 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "ports").Index(i).Child("port"), port.Port, "must be between 0 and 65535"))
		}
		if port.NodePort != 0 && (port.NodePort < 30000 || port.NodePort > 32767) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "ports").Index(i).Child("nodePort"), port.NodePort, "must be between 30000 and 32767"))
		}
	}

	return allErrs
}

func (v *BuiltinValidator) validateConfigMap(cm *corev1.ConfigMap) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&cm.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	// Validate data keys
	for k := range cm.Data {
		for _, msg := range utilvalidation.IsQualifiedName(k) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("data").Key(k), k, msg))
		}
	}

	return allErrs
}

func (v *BuiltinValidator) validateSecret(s *corev1.Secret) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&s.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	// Validate secret type - empty type is allowed but not recommended
	if s.Type == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("type"), ""))
	}

	return allErrs
}

func (v *BuiltinValidator) validatePVC(pvc *corev1.PersistentVolumeClaim) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&pvc.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if pvc.Spec.Resources.Requests != nil {
		if pvc.Spec.Resources.Requests.Storage() == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "resources", "requests", "storage"), ""))
		}
	}

	return allErrs
}

func (v *BuiltinValidator) validateNamespace(ns *corev1.Namespace) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&ns.ObjectMeta, false, nameValidator, field.NewPath("metadata"))...)
	return allErrs
}

func (v *BuiltinValidator) validateServiceAccount(sa *corev1.ServiceAccount) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&sa.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)
	return allErrs
}

func (v *BuiltinValidator) validateEndpoints(ep *corev1.Endpoints) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&ep.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)
	return allErrs
}

func (v *BuiltinValidator) validateIngress(ing *networkingv1.Ingress) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&ing.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	// Validate ingress class
	if ing.Spec.IngressClassName != nil && *ing.Spec.IngressClassName == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "ingressClassName"), *ing.Spec.IngressClassName, "must be non-empty"))
	}

	return allErrs
}

func (v *BuiltinValidator) validateNetworkPolicy(np *networkingv1.NetworkPolicy) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&np.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if np.Spec.PolicyTypes == nil || len(np.Spec.PolicyTypes) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "policyTypes"), "at least one policy type must be specified"))
	}

	return allErrs
}

func (v *BuiltinValidator) validateJob(job *batchv1.Job) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&job.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if job.Spec.Parallelism != nil && *job.Spec.Parallelism < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "parallelism"), *job.Spec.Parallelism, "must be >= 0"))
	}
	if job.Spec.Completions != nil && *job.Spec.Completions < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "completions"), *job.Spec.Completions, "must be >= 0"))
	}
	if job.Spec.BackoffLimit != nil && *job.Spec.BackoffLimit < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "backoffLimit"), *job.Spec.BackoffLimit, "must be >= 0"))
	}

	// AWU-3: validate PodSpec (skipProbes=true for batch workloads)
	allErrs = append(allErrs, validatePodSpec(&job.Spec.Template.Spec, field.NewPath("spec", "template", "spec"), true)...)

	return allErrs
}

func (v *BuiltinValidator) validateCronJob(cj *batchv1.CronJob) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&cj.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if cj.Spec.Schedule == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "schedule"), ""))
	}

	if cj.Spec.StartingDeadlineSeconds != nil && *cj.Spec.StartingDeadlineSeconds < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "startingDeadlineSeconds"), *cj.Spec.StartingDeadlineSeconds, "must be >= 0"))
	}

	// AWU-3: validate PodSpec (skipProbes=true for batch workloads)
	allErrs = append(allErrs, validatePodSpec(&cj.Spec.JobTemplate.Spec.Template.Spec, field.NewPath("spec", "jobTemplate", "spec", "template", "spec"), true)...)

	return allErrs
}

func (v *BuiltinValidator) registerHorizontalPodAutoscaler() {
	v.registerKind("HorizontalPodAutoscaler", "autoscaling/v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateHorizontalPodAutoscaler(obj.(*autoscalingv1.HorizontalPodAutoscaler))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) registerLimitRange() {
	v.registerKind("LimitRange", "v1",
		func(obj runtime.Object) field.ErrorList {
			return v.validateLimitRange(obj.(*corev1.LimitRange))
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) validateHorizontalPodAutoscaler(hpa *autoscalingv1.HorizontalPodAutoscaler) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&hpa.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if hpa.Spec.MinReplicas != nil && *hpa.Spec.MinReplicas < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "minReplicas"), *hpa.Spec.MinReplicas, "must be >= 0"))
	}

	return allErrs
}

func (v *BuiltinValidator) validateLimitRange(lr *corev1.LimitRange) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&lr.ObjectMeta, false, nameValidator, field.NewPath("metadata"))...)

	return allErrs
}

func (v *BuiltinValidator) registerList() {
	v.registerKind("List", "v1",
		func(obj runtime.Object) field.ErrorList {
			return nil // List is a meta-type; items are validated individually
		},
		func(data []byte) (runtime.Object, error) {
			obj, _, err := v.decode(data)
			return obj, err
		},
	)
}

func (v *BuiltinValidator) validateRole(role *rbacv1.Role) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&role.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	for i, rule := range role.Rules {
		if len(rule.Verbs) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("rules").Index(i).Child("verbs"), ""))
		}
	}

	return allErrs
}

func (v *BuiltinValidator) validateClusterRole(cr *rbacv1.ClusterRole) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&cr.ObjectMeta, false, nameValidator, field.NewPath("metadata"))...)

	for i, rule := range cr.Rules {
		if len(rule.Verbs) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("rules").Index(i).Child("verbs"), ""))
		}
	}

	return allErrs
}

func (v *BuiltinValidator) validateRoleBinding(rb *rbacv1.RoleBinding) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&rb.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)

	if len(rb.Subjects) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("subjects"), "at least one subject must be specified"))
	}

	return allErrs
}

func (v *BuiltinValidator) validateClusterRoleBinding(crb *rbacv1.ClusterRoleBinding) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&crb.ObjectMeta, false, nameValidator, field.NewPath("metadata"))...)

	if len(crb.Subjects) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("subjects"), "at least one subject must be specified"))
	}

	return allErrs
}

// Result represents validation result for a single resource
type Result struct {
	Kind       string
	Name       string
	Namespace  string
	APIVersion string
	Status     types.Status
	Errors     []ErrorItem
}

// ErrorItem represents a validation error
type ErrorItem struct {
	Field   string
	Message string
	Code    string
}

// isNamespacedKind returns true for Kubernetes resource kinds that are namespaced.
// Cluster-scoped kinds return false.
func isNamespacedKind(kind string) bool {
	switch kind {
	case "Namespace", "Node", "PersistentVolume",
		"ClusterRole", "ClusterRoleBinding", "StorageClass",
		"CSIDriver", "CSINode", "PriorityClass",
		"RuntimeClass", "FlowSchema", "PriorityLevelConfiguration",
		"EndpointSlice" /* cluster-scoped for service mesh */ :
		return false
	default:
		return true
	}
}
