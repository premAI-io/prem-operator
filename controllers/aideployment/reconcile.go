package aideployment

import (
	"context"
	"fmt"
	"github.com/premAI-io/saas-controller/controllers/constants"
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

	// Generate a Deployment from the Engine
	deployment, err := mle.Deployment(&sd.ObjectMeta)
	if err != nil {
		return 0, err
	}

	container := findContainerEngine(deployment)
	if container != nil {
		container.Args = append(container.Args, sd.Spec.Args...)
	}

	// Add generic Scheduling properties
	err = AddSchedulingProperties(deployment, sd.Spec)
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

	re, err := UpdateAIDeploymentStatus(ctx, c, &sd, d, "")
	if err != nil {
		return 0, err
	}
	requeue = re

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
		annotations,
		mle.Port(),
	)

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
		deployment.Name,
		domains,
		int(mle.Port()),
		sd.Spec.Ingress.Labels,
		annotations,
		tls,
		sd.Spec.RateLimit,
		sd.Spec.Authentication,
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

	log.Debug(
		"Reconcile completed: ", sd.Name, " in namespace: ", sd.Namespace,
	)

	return requeue, nil
}

// UpdateAIDeploymentStatus updates the status of the AI deployment
func UpdateAIDeploymentStatus(
	ctx context.Context,
	c ctrlClient.Client,
	aiDeployment *v1alpha1.AIDeployment,
	deployment *appsv1.Deployment,
	errMsg string,
) (int, error) {
	aiDep := aiDeployment.DeepCopy()

	if errMsg != "" {
		aiDep.Status.Status = constants.Failed
		aiDep.Status.ErrMsg = errMsg
		if err := c.Status().Update(ctx, aiDep); err != nil {
			return 0, fmt.Errorf("failed to update AI deployment status: %w", err)
		}

		return 0, nil
	}

	// The status of the Deployment might not be immediately available after the
	// Deployment resource is created or updated, requeue to check the deployment
	// status again after 3 seconds
	requeue := 3
	aiDep.Status.Status = constants.NotReady
	if deployment.Status.AvailableReplicas > 0 {
		aiDep.Status.Status = constants.Ready
		requeue = 0
	}

	if err := c.Status().Update(ctx, aiDep); err != nil {
		return requeue, fmt.Errorf("failed to update AI deployment status: %w", err)
	}
	return requeue, nil
}
