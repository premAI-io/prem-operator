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
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/premAI-io/prem-operator/api/v1alpha1"
	"github.com/premAI-io/prem-operator/controllers/aideployment"
	"github.com/premAI-io/prem-operator/controllers/aimodelmap"
	"github.com/premAI-io/prem-operator/controllers/engines"
)

// AIDeploymentReconciler reconciles a AIDeployment object
type AIDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=premlabs.io,resources=aideployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=premlabs.io,resources=aideployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=premlabs.io,resources=aideployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=create;get;list;update;watch
//+kubebuilder:rbac:groups=premlabs.io,resources=aimodelmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// the AIDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AIDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Creates a deployment targeting a service
	var ent v1alpha1.AIDeployment
	if err := r.Get(ctx, req.NamespacedName, &ent); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var (
		mlEngine aideployment.MLEngine
		err      error
	)

	models, err := aimodelmap.Resolve(&ent, ctx, r.Client)
	if err != nil {
		_, err1 := aideployment.UpdateAIDeploymentStatus(
			ctx, r.Client, &ent, nil, err.Error(),
		)

		return ctrl.Result{}, fmt.Errorf("%w: %v", err, err1)
	}

	switch ent.Spec.Engine.Name {
	case v1alpha1.AIEngineNameTriton:
		mlEngine = engines.NewTriton(&ent, models)
	case v1alpha1.AIEngineNameLocalai:
		mlEngine = engines.NewLocalAI(&ent, models)
	case v1alpha1.AIEngineNameVLLM:
		mlEngine, err = engines.NewVllmAi(&ent, models)
	case v1alpha1.AIEngineNameGeneric:
		mlEngine = engines.NewGeneric(&ent)
	case v1alpha1.AIEngineNameDeepSpeedMii:
		mlEngine, err = engines.NewDeepSpeedMii(&ent, models)
	default:
		err = fmt.Errorf("unknown engine %s", ent.Spec.Engine.Name)
	}
	if err != nil {
		_, err1 := aideployment.UpdateAIDeploymentStatus(
			ctx, r.Client, &ent, nil, err.Error(),
		)

		return ctrl.Result{}, fmt.Errorf("%w: %v", err, err1)
	}

	requeue, err := aideployment.Reconcile(ent, ctx, r.Client, mlEngine)
	if requeue > 0 {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(requeue)}, err
	}
	if err != nil {
		_, err1 := aideployment.UpdateAIDeploymentStatus(
			ctx,
			r.Client,
			&ent,
			nil,
			fmt.Sprintf("Reconciliation error: %v", err),
		)

		return ctrl.Result{}, fmt.Errorf("%w: %v", err, err1)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AIDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AIDeployment{}).
		Complete(r)
}
