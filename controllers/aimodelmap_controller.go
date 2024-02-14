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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/aimodelmap"
	"github.com/premAI-io/saas-controller/controllers/constants"
)

// AIModelMapReconciler reconciles a AIModelMap object
type AIModelMapReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=premlabs.io,resources=aimodelmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=premlabs.io,resources=aimodelmaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=premlabs.io,resources=aimodelmaps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AIModelMap object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AIModelMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lg := log.FromContext(ctx)

	// check if the request is triggered by a configmap
	dep := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, req.NamespacedName, dep)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err == nil {
		// TODO: check if the configmap is owned by an AIModelMap and if it is check it has the correct values
		return ctrl.Result{}, nil
	}

	modelMap := &a1.AIModelMap{}
	err = r.Client.Get(ctx, req.NamespacedName, modelMap)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Find the ConfigMap for this AIModelMap
	configMap := &corev1.ConfigMap{}
	err = r.Client.Get(ctx, req.NamespacedName, configMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			configMap = nil
		} else {
			return ctrl.Result{}, err
		}
	}

	newConfigMap := &corev1.ConfigMap{}
	addedCount := 0

	engines := []struct {
		name     a1.AIEngineName
		variants []a1.AIModelVariant
	}{
		{name: a1.AIEngineNameLocalai, variants: modelMap.Spec.Localai},
		{name: a1.AIEngineNameVLLM, variants: modelMap.Spec.Vllm},
		{name: a1.AIEngineNameGeneric, variants: modelMap.Spec.DeepSpeedMii},
	}

	for _, eng := range engines {
		ac, err := addVariants(newConfigMap, eng.name, eng.variants)
		if err != nil {
			return ctrl.Result{}, err
		}
		addedCount += ac
	}

	if addedCount == 0 {
		return ctrl.Result{}, nil
	}

	if configMap == nil {
		newConfigMap.ObjectMeta = metav1.ObjectMeta{
			Name:      modelMap.Name,
			Namespace: modelMap.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(modelMap, schema.GroupVersionKind{
					Group:   a1.GroupVersion.Group,
					Version: a1.GroupVersion.Version,
					Kind:    "AIModelMap",
				}),
			},
			Annotations: map[string]string{
				constants.AIModelMapDefaultAnnotationKey: modelMap.Name,
			},
			Labels: map[string]string{
				constants.AIModelMapDefaultLabelKey: modelMap.Name,
			},
		}

		lg.Info("Creating ConfigMap for AIModelMap", "Namespace", newConfigMap.Namespace, "Name", newConfigMap.Name)
		return ctrl.Result{}, r.Client.Create(ctx, newConfigMap)
	}

	newConfigMap.ObjectMeta = modelMap.ObjectMeta
	err = r.Client.Update(ctx, newConfigMap)
	if apierrors.IsConflict(err) {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *AIModelMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&a1.AIModelMap{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func addVariants(cm *corev1.ConfigMap, engineName a1.AIEngineName, variants []a1.AIModelVariant) (int, error) {
	addedCount := 0

	for _, v := range variants {
		if v.EngineConfigFile != "" {
			aimodelmap.SetEngineConfigFileData(cm, engineName, v.Name, v.EngineConfigFile)
			addedCount += 1
		}
	}

	return addedCount, nil
}
