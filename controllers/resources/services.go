package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	// KubeGenericLabelPrefix is the prefix of kubernetes generic label key
	KubeGenericLabelPrefix = "app.kubernetes.io"
)

func DesiredService(owner metav1.Object, name, namespace string, selector, labels, annotations map[string]string, port int32) *corev1.Service {
	ports := []corev1.ServicePort{
		{
			Port: port, TargetPort: intstr.FromInt(int(port)),
		},
	}

	if labels == nil {
		labels = map[string]string{}
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: GenOwner(owner),
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			Annotations:     annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: selector,
		},
	}
}
