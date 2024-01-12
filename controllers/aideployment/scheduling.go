package aideployment

import (
	"fmt"

	"github.com/premAI-io/saas-controller/controllers/constants"
	"github.com/premAI-io/saas-controller/pkg/utils"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addTopologySpread(tmpl *v1.PodTemplateSpec) {
	if tmpl.Labels != nil {
		if _, ok := tmpl.Labels[constants.PremSpreadTopologyLabel]; ok {
			return
		}
	} else {
		tmpl.Labels = map[string]string{}
	}

	tmpl.Labels[constants.PremSpreadTopologyLabel] = "ai-model"
	tmpl.Spec.TopologySpreadConstraints = append(tmpl.Spec.TopologySpreadConstraints, v1.TopologySpreadConstraint{
		MaxSkew:           1,
		TopologyKey:       "kubernetes.io/hostname",
		WhenUnsatisfiable: v1.ScheduleAnyway,
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				constants.PremSpreadTopologyLabel: "ai-model",
			},
		},
	})
}

func neededGPUs(deploy a1.Deployment) (resource.Quantity, error) {
	gpus := resource.MustParse("0")

	if deploy.Accelerator == nil {
		if deploy.Resources.Requests == nil {
			return gpus, nil
		}

		if _, ok := deploy.Resources.Requests[constants.NvidiaGPULabel]; ok {
			return gpus, fmt.Errorf("deployment requests Nvidia GPU but no accelerator is specified")
		} else {
			return gpus, nil
		}
	}

	// If you add non-Nvidia accelerators, remember to set the pod runtime class name
	if deploy.Accelerator.Interface != a1.AcceleratorInterfaceCUDA {
		return gpus, fmt.Errorf("unsupported accelerator interface: %s", deploy.Accelerator.Interface)
	}

	if gpusSpec, ok := deploy.Resources.Requests[constants.NvidiaGPULabel]; ok {
		return gpusSpec, nil
	}

	gpus.Add(resource.MustParse("1"))

	return gpus, nil
}

func findContainerEngine(appDeployment *appsv1.Deployment) (engineContainer *v1.Container) {
	for i, c := range appDeployment.Spec.Template.Spec.Containers {
		if c.Name == constants.ContainerEngineName {
			engineContainer = &appDeployment.Spec.Template.Spec.Containers[i]
			return
		}
	}
	return
}

func AddSchedulingProperties(appDeployment *appsv1.Deployment, AIDeployment a1.AIDeploymentSpec) error {
	addTopologySpread(&appDeployment.Spec.Template)

	pod := &appDeployment.Spec.Template.Spec
	pod.NodeSelector = utils.MergeMaps(pod.NodeSelector, AIDeployment.Deployment.NodeSelector)

	gpus, err := neededGPUs(AIDeployment.Deployment)
	if err != nil {
		return err
	}

	engineContainer := findContainerEngine(appDeployment)
	if engineContainer == nil {
		return fmt.Errorf("no container named %s found in deployment", constants.ContainerEngineName)
	}

	engineContainer.Resources.Requests = utils.MergeMaps(
		engineContainer.Resources.Requests,
		AIDeployment.Deployment.Resources.Requests,
	)

	engineContainer.Resources.Limits = utils.MergeMaps(
		engineContainer.Resources.Limits,
		AIDeployment.Deployment.Resources.Limits,
	)

	if !gpus.IsZero() {
		engineContainer.Resources.Requests[constants.NvidiaGPULabel] = gpus
		engineContainer.Resources.Limits[constants.NvidiaGPULabel] = gpus

		if pod.RuntimeClassName == nil {
			runtimeClassName := "nvidia"
			pod.RuntimeClassName = &runtimeClassName
		}
	}

	return nil
}
