package resources

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkv1 "k8s.io/api/networking/v1"
)

// DesiredIngress generates the desired ingress
// XXX: Probably doesn't do the correct thing for now
func DesiredIngress(owner metav1.Object, name, namespace, hostname, svcName, clusterIssuer string, port int, labels, annotations map[string]string) *networkv1.Ingress {
	t := networkv1.PathType("Prefix")
	rules := []networkv1.IngressRule{

		networkv1.IngressRule{
			Host: hostname,
			IngressRuleValue: networkv1.IngressRuleValue{
				HTTP: &networkv1.HTTPIngressRuleValue{
					Paths: []networkv1.HTTPIngressPath{{
						PathType: &t,
						Path:     "/",
						Backend: networkv1.IngressBackend{
							Service: &networkv1.IngressServiceBackend{
								Name: svcName,
								Port: networkv1.ServiceBackendPort{Number: int32(port)},
							},
						},
					}},
				},
			},
		},
	}

	spec := networkv1.IngressSpec{
		Rules: rules,
	}
	if labels == nil {
		labels = map[string]string{}
	}
	if annotations == nil {
		annotations = map[string]string{}
	}

	if clusterIssuer != "" {
		tlsEntry := []networkv1.IngressTLS{
			networkv1.IngressTLS{
				Hosts:      []string{hostname},
				SecretName: fmt.Sprintf("%s-tls", svcName),
			}}

		annotations["cert-manager.io/cluster-issuer"] = clusterIssuer
		spec.TLS = tlsEntry
	}

	return &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: GenOwner(owner),
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			Annotations:     annotations,
		},
		Spec: spec,
	}
}
