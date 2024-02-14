package engines

import (
	"fmt"
	"strings"

	"github.com/premAI-io/saas-controller/controllers/aideployment"
	"github.com/premAI-io/saas-controller/controllers/aimodelmap"
	"github.com/premAI-io/saas-controller/controllers/constants"
	"github.com/premAI-io/saas-controller/pkg/utils"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type LocalAI struct {
	AIDeployment *a1.AIDeployment
	Models       []aimodelmap.ResolvedModel
}

func NewLocalAI(ai *a1.AIDeployment, m []aimodelmap.ResolvedModel) aideployment.MLEngine {
	return &LocalAI{AIDeployment: ai, Models: m}

}
func (l *LocalAI) Port() int32 {
	return 8080
}

func (l *LocalAI) Deployment(owner metav1.Object) (*appsv1.Deployment, error) {
	objMeta := metav1.ObjectMeta{
		Name:            l.AIDeployment.Name,
		Namespace:       l.AIDeployment.Namespace,
		OwnerReferences: resources.GenOwner(owner),
	}

	imageTag := constants.ImageTagLatest
	if l.AIDeployment.Spec.Engine.Options[constants.ImageTagKey] != "" {
		imageTag = l.AIDeployment.Spec.Engine.Options[constants.ImageTagKey]
	}

	imageRepository := constants.ImageRepositoryLocalai
	if l.AIDeployment.Spec.Engine.Options[constants.ImageRepositoryKey] != "" {
		imageRepository = l.AIDeployment.Spec.Engine.Options[constants.ImageRepositoryKey]
	}

	deployment := appsv1.Deployment{}
	if l.AIDeployment.Spec.Deployment.PodTemplate != nil {
		deployment.Spec.Template = *l.AIDeployment.Spec.Deployment.PodTemplate.DeepCopy()
	} else {
		deployment.Spec.Template = v1.PodTemplateSpec{}
	}
	deployment.Spec.Replicas = l.AIDeployment.Spec.Deployment.Replicas
	pod := &deployment.Spec.Template.Spec

	serviceAccount := false

	v := l.AIDeployment.Spec.Env

	v = append(v, v1.EnvVar{Name: "MODELS_PATH", Value: "/models"})
	image := fmt.Sprintf("%s:%s", imageRepository, imageTag)

	healthProbeHandler := v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path: "/healthz",
			Port: intstr.FromInt(int(l.Port())),
		},
	}
	expose := &v1.Container{
		ImagePullPolicy: v1.PullAlways,
		Name:            constants.ContainerEngineName,
		Image:           image,
		Env:             v,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "models",
				MountPath: "/models",
			},
		},
		StartupProbe: &v1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       10,
			FailureThreshold:    120,
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/readyz",
					Port: intstr.FromInt(int(l.Port())),
				},
			},
		},
		ReadinessProbe: &v1.Probe{
			FailureThreshold: 3,
			ProbeHandler:     healthProbeHandler,
		},
		LivenessProbe: &v1.Probe{
			PeriodSeconds:    30,
			TimeoutSeconds:   15,
			FailureThreshold: 10,
			ProbeHandler:     healthProbeHandler,
		},
	}

	mergeProbe(l.AIDeployment.Spec.Deployment.StartupProbe, expose.StartupProbe)
	mergeProbe(l.AIDeployment.Spec.Deployment.ReadinessProbe, expose.ReadinessProbe)
	mergeProbe(l.AIDeployment.Spec.Deployment.LivenessProbe, expose.LivenessProbe)

	pod.AutomountServiceAccountToken = &serviceAccount

	pod.Volumes = append(pod.Volumes, v1.Volume{
		Name: "models",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium: v1.StorageMediumMemory,
			},
		},
	}, v1.Volume{
		Name: "cache",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	configSources := []v1.VolumeProjection{}
	configSourceVolume := "configs"

	for _, m := range l.Models {
		if m.Spec.EngineConfigFile != "" {
			if m.Variant == "" {
				return nil, fmt.Errorf("inline model %s has engine config file, but we haven't implemented generating ConfigMaps for inline configs", m.Name)
			}
			configSources = append(configSources, v1.VolumeProjection{
				ConfigMap: &v1.ConfigMapProjection{
					LocalObjectReference: v1.LocalObjectReference{
						Name: m.Name,
					},
					Items: []v1.KeyToPath{
						{
							Key:  aimodelmap.FmtConfigMapKey(a1.AIEngineNameLocalai, m.Variant, constants.AIModelMapSpecEngineConfig),
							Path: "engine/" + m.HostName + ".yaml",
						},
					},
				},
			})

			// If an engine config is set then specifying the model some other way doesn't make sense
			continue
		}

		if strings.HasPrefix(m.Spec.Uri, "http") {
			pod.InitContainers = append(pod.InitContainers, v1.Container{
				ImagePullPolicy: v1.PullAlways,
				Name:            fmt.Sprintf("init-models-%s", l.AIDeployment.Name),
				Image:           image,
				Command:         []string{"sh", "-c"},
				Args:            []string{"curl -L -v -o /models/$MODEL_NAME $MODEL_PATH"},
				Env: []v1.EnvVar{
					{Name: "MODEL_NAME", Value: m.Name},
					{Name: "MODEL_PATH", Value: m.Spec.Uri},
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "models",
						MountPath: "/models",
					},
				},
			})
		} else {
			// Pass models as args.
			// LocalAI accepts both names and full URLs passed by as Args.
			expose.Args = append(expose.Args, m.Spec.Uri)
		}
	}

	if len(configSources) > 0 {
		pod.Volumes = append(pod.Volumes, v1.Volume{
			Name: configSourceVolume,
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					Sources: configSources,
				},
			},
		})

		pod.InitContainers = append(pod.InitContainers, v1.Container{
			ImagePullPolicy: v1.PullAlways,
			Name:            fmt.Sprintf("init-%s-%s", configSourceVolume, l.AIDeployment.Name),
			Image:           image,
			Command:         []string{"sh", "-c"},
			Args:            []string{fmt.Sprintf("ls /configs/engine && cp -v /%s/engine/* /models", configSourceVolume)},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      configSourceVolume,
					MountPath: "/" + configSourceVolume,
				},
				{
					Name:      "models",
					MountPath: "/models",
				},
			},
		})
	}

	pod.Containers = append(pod.Containers, *expose)
	deploymentLabels := resources.GenDefaultLabels(l.AIDeployment.Name)
	deployment.Spec.Template.Labels = utils.MergeMaps(
		deploymentLabels,
		deployment.Spec.Template.Labels,
		l.AIDeployment.Spec.Deployment.Labels,
	)

	deployment.Spec.Template.Annotations = utils.MergeMaps(
		deployment.Spec.Template.Annotations,
		l.AIDeployment.Spec.Deployment.Annotations,
	)

	deployment.ObjectMeta = objMeta
	deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: deploymentLabels}

	return &deployment, nil
}
