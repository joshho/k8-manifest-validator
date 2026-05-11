package validator

import (
	"testing"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestDebugTestcase(t *testing.T) {
	// Replicate the exact test case
	d := &appsv1.Deployment{}
	d.Name = "test"
	d.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}}
	d.Spec.Template.Labels = map[string]string{"app": "test"}
	d.Spec.Template.Spec.Containers = []corev1.Container{{Name: "test", Image: "nginx"}}
	d.Spec.Template.Spec.Volumes = []corev1.Volume{{
		Name: "az",
		VolumeSource: corev1.VolumeSource{
			AzureFile: &corev1.AzureFileVolumeSource{
				SecretName: "",
				ShareName:  "share-name",
			},
		},
	}}
	d.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{Name: "az", MountPath: "/mnt"}}
	
	raw, _ := yaml.Marshal(d)
	fmt.Printf("YAML:\n%s\n", string(raw))
	
	result := validateDeploymentRaw(d)
	fmt.Printf("validateDeploymentRaw: Status=%v, Errors=%v\n", result.Status, result.Errors)
}
