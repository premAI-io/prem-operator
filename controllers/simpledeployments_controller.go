/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"strings"

	appsv1 "k8s.io/api/apps/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	deploymentsv1alpha1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/engines"
)

// SimpleDeploymentsReconciler reconciles a SimpleDeployments object
type SimpleDeploymentsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=deployments.premlabs.io,resources=simpledeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=deployments.premlabs.io,resources=simpledeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=deployments.premlabs.io,resources=simpledeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=ingress,verbs=create;get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=create;get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimpleDeployments object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SimpleDeploymentsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Creates a deployment targeting a service
	// TODO(user): your logic here
	var ent deploymentsv1alpha1.SimpleDeployments
	if err := r.Get(ctx, req.NamespacedName, &ent); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	switch strings.ToLower(ent.Spec.MLEngine) {
	case engines.LocalAIEngine:
		// Create a deployment targeting a service
		e := engines.LocalAI{
			Name:      ent.Name,
			Namespace: ent.Namespace,
			Object:    &ent.ObjectMeta,
			Options:   ent.Spec.Options,
		}
		deployment, err := e.Deployment("SimpleDeployments")
		if err != nil {
			return ctrl.Result{}, err
		}
		d := &appsv1.Deployment{}
		// try to find if a deployment already exists
		if err := r.Get(ctx, types.NamespacedName{Namespace: ent.Namespace, Name: ent.Name}, d); err != nil {
			if apierrors.IsNotFound(err) { // Create a deployment
				if err := r.Create(ctx, deployment); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, err
			}
		} else { // Update a deployment
			if err := r.Update(ctx, deployment); err != nil {
				return ctrl.Result{}, err
			}

			if d.Status.AvailableReplicas == 0 {
				e := ent.DeepCopy()
				e.Status.Status = "NotReady"
				r.Update(ctx, e)
				return ctrl.Result{Requeue: true}, nil
			} else {
				e := ent.DeepCopy()
				e.Status.Status = "Ready"
				r.Update(ctx, e)
			}
		}

		// Create a service
		// Create an ingress
		// Create a secret
	}

	// Reconcile does the following:
	// Create an associated Kubernetes deployment if doesn't exist yet, or update it if already exists
	// Create an associated Kubernetes service if doesn't exist yet, or update it if already exists
	// Create an associated Kubernetes ingress if doesn't exist yet, or update it if already exists

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleDeploymentsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deploymentsv1alpha1.SimpleDeployments{}).
		Complete(r)
}
