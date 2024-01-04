package engines

import (
	"fmt"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addTopologySpread(tmpl *v1.PodTemplateSpec) {
	if tmpl.Labels != nil {
		if _, ok := tmpl.Labels["mlcontroller.premlabs.io/spread-topology"]; ok {
			return
		}
	} else {
		tmpl.Labels = map[string]string{}
	}

	tmpl.Labels["mlcontroller.premlabs.io/spread-topology"] = "ai-model"
	tmpl.Spec.TopologySpreadConstraints = append(tmpl.Spec.TopologySpreadConstraints, v1.TopologySpreadConstraint{
		MaxSkew:           1,
		TopologyKey:       "kubernetes.io/hostname",
		WhenUnsatisfiable: v1.ScheduleAnyway,
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"mlcontroller.premlabs.io/spread-topology": "ai-model",
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

		if _, ok := deploy.Resources.Requests["nvidia.com/gpu"]; ok {
			return gpus, fmt.Errorf("deployment requests Nvidia GPU but no accelerator is specified")
		} else {
			return gpus, nil
		}
	}

	// If you add non-Nvidia accelerators, remember to set the pod runtime class name
	if deploy.Accelerator.Interface != a1.AcceleratorInterfaceCUDA {
		return gpus, fmt.Errorf("unsupported accelerator interface: %s", deploy.Accelerator.Interface)
	}

	if gpusSpec, ok := deploy.Resources.Requests["nvidia.com/gpu"]; ok {
		return gpusSpec, nil
	}

	gpus.Add(resource.MustParse("1"))

	return gpus, nil
}

func addSchedulingProperties(appDeployment *appsv1.Deployment, engineContainer *v1.Container, AIDeployment *a1.AIDeploymentSpec) error {
	addTopologySpread(&appDeployment.Spec.Template)

	pod := &appDeployment.Spec.Template.Spec
	pod.NodeSelector = mergeMaps(pod.NodeSelector, AIDeployment.Deployment.NodeSelector)

	gpus, err := neededGPUs(AIDeployment.Deployment)
	if err != nil {
		return err
	}

	engineContainer.Resources.Requests = mergeMaps(
		map[v1.ResourceName]resource.Quantity{
			"cpu": resource.MustParse("2"),
		},
		engineContainer.Resources.Requests,
		AIDeployment.Deployment.Resources.Requests,
	)

	ramLimit := engineContainer.Resources.Requests["memory"]
	ramLimit.Add(resource.MustParse("4Gi"))
	cpuLimit := engineContainer.Resources.Requests["cpu"]
	cpuLimit.Add(resource.MustParse("2"))

	engineContainer.Resources.Limits = mergeMaps(
		map[v1.ResourceName]resource.Quantity{
			"cpu":    cpuLimit,
			"memory": ramLimit,
		},
		engineContainer.Resources.Limits,
		AIDeployment.Deployment.Resources.Limits,
	)

	if !gpus.IsZero() {
		engineContainer.Resources.Requests["nvidia.com/gpu"] = gpus
		engineContainer.Resources.Limits["nvidia.com/gpu"] = gpus

		if pod.RuntimeClassName == nil {
			runtimeClassName := "nvidia"
			pod.RuntimeClassName = &runtimeClassName
		}
	}

	return nil
}
