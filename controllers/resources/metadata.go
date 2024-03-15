package resources

import (
	"github.com/premAI-io/prem-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	DefaultAnnotation = "mlcontroller.premlabs.io/ai-deployment"
	DefaultLabel      = "mlcontroller.premlabs.io/ai-deployment"
)

func GenDefaultAnnotation(s string) map[string]string {
	return map[string]string{
		DefaultAnnotation: s,
	}
}
func GenDefaultLabels(s string) map[string]string {
	return map[string]string{
		DefaultLabel: s,
	}
}

func GenOwner(obj metav1.Object) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(obj, schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    v1alpha1.ResourceName,
		}),
	}
}
