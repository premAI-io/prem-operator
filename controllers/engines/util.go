package engines

import (
	"fmt"
	"strings"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mergeProbe(src *a1.Probe, dst *v1.Probe) {
	if src == nil {
		return
	}

	if src.InitialDelaySeconds != nil {
		dst.InitialDelaySeconds = *src.InitialDelaySeconds
	}
	if src.PeriodSeconds != 0 {
		dst.PeriodSeconds = src.PeriodSeconds
	}
	if src.TimeoutSeconds != 0 {
		dst.TimeoutSeconds = src.TimeoutSeconds
	}
	if src.SuccessThreshold != 0 {
		dst.SuccessThreshold = src.SuccessThreshold
	}
	if src.FailureThreshold != 0 {
		dst.FailureThreshold = src.FailureThreshold
	}
}

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
	mems := resource.MustParse("0")
	if m, ok := deploy.Resources.Requests["memory"]; ok {
		mems = m
	} else {
		return mems, fmt.Errorf("Deployment needs to request memory")
	}
	var memoryPerGPU *resource.Quantity
	gpus := resource.MustParse("0")

	if v, ok := deploy.NodeSelector["nvidia.com/gpu-memory"]; ok {
		q := resource.MustParse(v + "Mi")
		memoryPerGPU = &q
	}

	for k := range deploy.NodeSelector {
		if strings.HasPrefix(k, "nvidia.com") {
			if memoryPerGPU == nil {
				return gpus, fmt.Errorf("Deployment requires GPU but per GPU memory is not specified")
			}
		}
	}

	if memoryPerGPU == nil {
		return gpus, nil
	}

	for mems.Cmp(*memoryPerGPU) > 0 {
		gpus.Add(resource.MustParse("1"))
		mems.Sub(*memoryPerGPU)
	}

	if !mems.IsZero() {
		gpus.Add(resource.MustParse("1"))
	}

	return gpus, nil
}

func addSchedulingProperties(appDeployment *appsv1.Deployment, engineContainer *v1.Container, AIDeployment *a1.AIDeploymentSpec) error {
	addTopologySpread(&appDeployment.Spec.Template)

	pod := appDeployment.Spec.Template.Spec
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
		runtimeClassName := "nvidia"

		engineContainer.Resources.Requests["nvidia.com/gpu"] = gpus
		engineContainer.Resources.Limits["nvidia.com/gpu"] = gpus
		pod.RuntimeClassName = &runtimeClassName
	}

	return nil
}

func mergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	merged := make(map[K]V)
	for _, m := range maps {
		if m == nil {
			continue
		}

		for key, value := range m {
			merged[key] = value
		}
	}
	return merged
}
