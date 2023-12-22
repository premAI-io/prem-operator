package aideployment

import (
	"context"

	log "github.com/sirupsen/logrus"

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

func Reconcile(sd v1alpha1.AIDeployment, ctx context.Context, c ctrlClient.Client, mle MLEngine) (int, error) {
	requeue := 0
	deployment, err := mle.Deployment(&sd.ObjectMeta)
	if err != nil {
		return 0, err
	}
	d := &appsv1.Deployment{}
	// try to find if a deployment already exists
	if err := c.Get(ctx, types.NamespacedName{Namespace: sd.GetNamespace(), Name: sd.GetName()}, d); err != nil {
		if apierrors.IsNotFound(err) { // Create a deployment
			log.Info("Creating deployment", deployment.Namespace, ":", deployment.Name)
			d = deployment.DeepCopy()
			if err := c.Create(ctx, d); err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	} else { // Update a deployment
		deployment.ResourceVersion = d.ResourceVersion
		d = deployment.DeepCopy()

		log.Debug("Updating deployment ", deployment.Namespace, ":", deployment.Name)
		if err := c.Update(ctx, d); err != nil {
			if apierrors.IsConflict(err) {
				log.Info("Deployment changed during update, requeueing")
				return 1, nil
			}
			return 0, err
		}
	}

	if d.Status.AvailableReplicas == 0 {
		log.Debug("Deployment ", deployment.Namespace, ":", deployment.Name, " is not ready, requeueing")
		e := sd.DeepCopy()
		e.Status.Status = "NotReady"
		if err := c.Update(ctx, e); err != nil {
			return 0, err
		}
		requeue = 3
	} else {
		e := sd.DeepCopy()
		e.Status.Status = "Ready"
		if err := c.Update(ctx, e); err != nil {
			return 0, err
		}
	}

	annotations := resources.GenDefaultAnnotation(sd.Name)
	for k, v := range sd.Spec.Service.Annotations {
		annotations[k] = v
	}

	svc := resources.DesiredService(
		&sd.ObjectMeta,
		deployment.Name,
		deployment.Namespace,
		deployment.Spec.Template.Labels,
		sd.Spec.Service.Labels,
		annotations, mle.Port())

	svcK := &v1.Service{}
	// try to find if a svc already exists
	if err := c.Get(ctx, types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name}, svcK); err != nil {
		if apierrors.IsNotFound(err) { // Create a deployment
			log.Debug("Creating service ", svc.Namespace, ":", svc.Name)
			svcK = svc.DeepCopy()
			if err := c.Create(ctx, svcK); err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	} else { // Update a deployment
		svc.ResourceVersion = svcK.ResourceVersion
		svcK = svc.DeepCopy()

		log.Debug("Updating service ", svc.Namespace, ":", svc.Name)
		if err := c.Update(ctx, svcK); err != nil {
			return 0, err
		}
	}

	if len(sd.Spec.Endpoint) == 0 {
		log.Debug("No endpoint specified, skipping ingress creation")
		return 0, nil
	}

	domains := []string{}
	for _, e := range sd.Spec.Endpoint {
		domains = append(domains, e.Domain)
	}

	tls := false
	if sd.Spec.Ingress.TLS != nil {
		tls = *sd.Spec.Ingress.TLS
	}

	annotations = resources.GenDefaultAnnotation(sd.Name)
	for k, v := range sd.Spec.Ingress.Annotations {
		annotations[k] = v
	}
	ingress := resources.DesiredIngress(
		&sd.ObjectMeta,
		deployment.Name,
		deployment.Namespace,
		domains,
		deployment.Name,
		int(mle.Port()),
		sd.Spec.Ingress.Labels,
		annotations,
		tls,
	)

	ingressK := &networkv1.Ingress{}
	// try to find if an ingress already exists
	if err := c.Get(ctx, types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name}, ingressK); err != nil {
		if apierrors.IsNotFound(err) {
			log.Debug("Creating ingress ", ingress.Namespace, ":", ingress.Name)
			ingressK = ingress.DeepCopy()
			if err := c.Create(ctx, ingressK); err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	} else { // Update a deployment
		ingress.ResourceVersion = ingressK.ResourceVersion
		ingressK = ingress.DeepCopy()
		log.Debug("Updating ingress ", ingress.Namespace, ":", ingress.Name)
		if err := c.Update(ctx, ingressK); err != nil {
			return 0, err
		}
	}

	log.Info(
		"Reconcile completed: ", sd.Name, " in namespace: ", sd.Namespace,
	)

	return requeue, nil
}
