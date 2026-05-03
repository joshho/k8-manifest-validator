package validator

import (
	"fmt"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	batchv1 "k8s.io/api/batch/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"

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
	routing     sync.Map // kind string -> routingEntry
	scheme      *runtime.Scheme
	decoder     runtime.Decoder
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
	
	return allErrs
}

func (v *BuiltinValidator) validatePod(pod *corev1.Pod) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMeta(&pod.ObjectMeta, true, nameValidator, field.NewPath("metadata"))...)
	
	// Validate pod spec
	if pod.Spec.TerminationGracePeriodSeconds != nil && *pod.Spec.TerminationGracePeriodSeconds < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "terminationGracePeriodSeconds"), *pod.Spec.TerminationGracePeriodSeconds, "must be >= 0"))
	}
	
	// Validate container names
	for i, c := range pod.Spec.Containers {
		for _, msg := range utilvalidation.IsQualifiedName(c.Name) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "containers").Index(i).Child("name"), c.Name, msg))
		}
	}
	
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
	
	return allErrs
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