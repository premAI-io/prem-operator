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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	premlabsv1alpha1 "github.com/premAI-io/saas-controller/api/v1alpha1"
)

// AutoNodeLabelerReconciler reconciles a AutoNodeLabeler object
type AutoNodeLabelerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=premlabs.io,resources=autonodelabelers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=premlabs.io,resources=autonodelabelers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=premlabs.io,resources=autonodelabelers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AutoNodeLabeler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AutoNodeLabelerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the Deployment instance
	dep := &corev1.Node{}
	err := r.Client.Get(ctx, req.NamespacedName, dep)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.patchNode(ctx, dep)

	return ctrl.Result{}, nil
}

func (r *AutoNodeLabelerReconciler) patchNode(ctx context.Context, n *corev1.Node) {
	labels := &premlabsv1alpha1.AutoNodeLabelerList{}
	err := r.List(ctx, labels)
	if err != nil {
		log.Log.Error(err, "Failed to get list of AutoNodeLabeler rules")
		return
	}
	for _, l := range labels.Items {

		update := false
		if *l.Spec.MatchExpression.Operator == v1.LabelSelectorOpExists {
			if _, exists := n.Labels[*l.Spec.MatchExpression.Key]; exists {
				update = true
			}
		}

		if *l.Spec.MatchExpression.Operator == v1.LabelSelectorOpIn {
			if v, exists := n.Labels[*l.Spec.MatchExpression.Key]; exists {
				for _, vv := range l.Spec.MatchExpression.Values {
					if vv == v {
						update = true
					}
				}
			}
		}

		if update {
			updateNode := n.DeepCopy()
			// Add labels
			for k, v := range l.Spec.Labels {
				updateNode.Labels[k] = v
			}
			// Update node
			err := r.Update(ctx, updateNode)
			if err != nil {
				log.Log.Error(err, "Failed to update node")
				return
			}
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoNodeLabelerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&premlabsv1alpha1.AutoNodeLabeler{}).
		Watches(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
