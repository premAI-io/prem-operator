package aideployment

import (
	"context"

	networkv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type MLEngine interface {
	Port() int32
	Deployment(owner metav1.Object) (*appsv1.Deployment, error)
}

func Reconcile(sd v1alpha1.AIDeployment, ctx context.Context, c ctrlClient.Client, mle MLEngine) (bool, error) {
	deployment, err := mle.Deployment(&sd.ObjectMeta)
	if err != nil {
		return false, err
	}
	d := &appsv1.Deployment{}
	// try to find if a deployment already exists
	if err := c.Get(ctx, types.NamespacedName{Namespace: sd.GetNamespace(), Name: sd.GetName()}, d); err != nil {
		if apierrors.IsNotFound(err) { // Create a deployment
			if err := c.Create(ctx, deployment); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else { // Update a deployment
		f := &appsv1.Deployment{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: sd.GetNamespace(), Name: sd.GetName()}, f); err != nil {
			return false, err
		}
		copy := f.DeepCopy()
		copy.Spec = deployment.Spec
		if err := c.Update(ctx, copy); err != nil {
			return false, err
		}

		if d.Status.AvailableReplicas == 0 {
			e := sd.DeepCopy()
			e.Status.Status = "NotReady"
			c.Update(ctx, e)
			return true, nil
		} else {
			e := sd.DeepCopy()
			e.Status.Status = "Ready"
			c.Update(ctx, e)
		}
	}

	svc := resources.DesiredService(
		&sd.ObjectMeta,
		deployment.Name,
		deployment.Namespace,
		deployment.Spec.Template.Labels,
		map[string]string{},
		resources.GenDefaultAnnotation(sd.Name), mle.Port())

	svcK := &v1.Service{}
	// try to find if a svc already exists
	if err := c.Get(ctx, types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name}, svcK); err != nil {
		if apierrors.IsNotFound(err) { // Create a deployment
			if err := c.Create(ctx, svc); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else { // Update a deployment
		f := &v1.Service{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: sd.GetNamespace(), Name: sd.GetName()}, f); err != nil {
			return false, err
		}
		copy := f.DeepCopy()
		copy.Spec = svcK.Spec
		if err := c.Update(ctx, copy); err != nil {
			return false, err
		}
	}

	if len(sd.Spec.Endpoint) == 0 {
		return false, nil
	}

	domains := []string{}
	for _, e := range sd.Spec.Endpoint {
		domains = append(domains, e.Domain)
	}

	ingress := resources.DesiredIngress(
		&sd.ObjectMeta,
		deployment.Name,
		deployment.Namespace,
		domains,
		deployment.Name,
		"",
		int(mle.Port()),
		map[string]string{}, resources.GenDefaultAnnotation(sd.Name))

	ingressK := &networkv1.Ingress{}
	// try to find if an ingress already exists
	if err := c.Get(ctx, types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name}, ingressK); err != nil {
		if apierrors.IsNotFound(err) { // Create a deployment
			if err := c.Create(ctx, ingress); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else { // Update a deployment
		f := &networkv1.Ingress{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: sd.GetNamespace(), Name: sd.GetName()}, f); err != nil {
			return false, err
		}
		copy := f.DeepCopy()
		copy.Spec = f.Spec
		if err := c.Update(ctx, copy); err != nil {
			return false, err
		}
	}

	return false, nil
}
