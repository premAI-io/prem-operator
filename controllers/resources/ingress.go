package resources

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkv1 "k8s.io/api/networking/v1"
)

func getEnv(key, def string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return value
}

var (
	envAuthMiddleware      = getEnv("TRAEFIK_AUTH_MIDDLEWARE", "api-gateway-auth-service")
	envRateLimitMiddleware = getEnv("TRAEFIK_RATE_LIMIT_MIDDLEWARE", "api-gateway-rate-limiter")
)

// DesiredIngress generates the desired ingress
// XXX: Probably doesn't do the correct thing for now
func DesiredIngress(
	owner metav1.Object,
	name, namespace, svcName string,
	hostname []string,
	port int,
	labels, annotations map[string]string,
	tls, ratelimit, authentication bool,
) *networkv1.Ingress {
	t := networkv1.PathType("Prefix")
	rules := []networkv1.IngressRule{}
	for _, h := range hostname {
		rules = append(rules, networkv1.IngressRule{
			Host: h,
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
		})
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

	// ensure middleware for traefik
	addMiddlewareIfNeeded(annotations, ratelimit, authentication)

	// for traefik ingress service it should pick up the ingress spec
	if tls {
		tlsEntry := []networkv1.IngressTLS{
			{
				Hosts:      hostname,
				SecretName: fmt.Sprintf("%s-tls", svcName),
			}}
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

func addMiddlewareIfNeeded(annotations map[string]string, ratelimit, authentication bool) {
	var middlewares = ""
	if ratelimit {
		middlewares += envRateLimitMiddleware + "@kubernetescrd"
	}
	if authentication {
		if len(middlewares) > 0 {
			middlewares += ", "
		}
		middlewares += envAuthMiddleware + "@kubernetescrd"
	}
	if len(middlewares) > 0 {
		annotations["traefik.ingress.kubernetes.io/router.middlewares"] = middlewares
	}
}
