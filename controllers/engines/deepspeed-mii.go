package engines

import (
	"fmt"

	"github.com/premAI-io/prem-operator/controllers/aideployment"
	"github.com/premAI-io/prem-operator/controllers/aimodelmap"
	"github.com/premAI-io/prem-operator/controllers/constants"
	"github.com/premAI-io/prem-operator/controllers/resources"
	"github.com/premAI-io/prem-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	a1 "github.com/premAI-io/prem-operator/api/v1alpha1"
)

type DeepSpeedMii struct {
	AIDeployment *a1.AIDeployment
	model        aimodelmap.ResolvedModel
}

func NewDeepSpeedMii(ai *a1.AIDeployment, models []aimodelmap.ResolvedModel) (aideployment.MLEngine, error) {
	if len(models) == 0 {
		return nil, ErrModelsNotSpecified
	}

	if len(models) > 1 {
		return nil, ErrorOnlyOneModel
	}

	return &DeepSpeedMii{AIDeployment: ai, model: models[0]}, nil
}

func (l *DeepSpeedMii) Port() int32 {
	return 8080
}

func (l *DeepSpeedMii) Deployment(owner metav1.Object) (*appsv1.Deployment, error) {
	objMeta := metav1.ObjectMeta{
		Name:            l.AIDeployment.Name,
		Namespace:       l.AIDeployment.Namespace,
		OwnerReferences: resources.GenOwner(owner),
	}

	imageTag := constants.ImageTagLatest
	if l.AIDeployment.Spec.Engine.Options[constants.ImageTagKey] != "" {
		imageTag = l.AIDeployment.Spec.Engine.Options[constants.ImageTagKey]
	}

	imageRepository := constants.ImageRepositoryDeepSpeedMii
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

	if pod.ImagePullSecrets == nil {
		pod.ImagePullSecrets = []v1.LocalObjectReference{
			{Name: constants.RegcredDockerhub},
		}
	}

	serviceAccount := false

	image := fmt.Sprintf("%s:%s", imageRepository, imageTag)

	backendProbeHandler := v1.ProbeHandler{
		// This is infact a gRPC server for the backend so we could use a gRPC probe here
		TCPSocket: &v1.TCPSocketAction{
			Port: intstr.FromInt(int(50051)),
		},
	}

	httpProbeHandler := v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path: "/healthz",
			Port: intstr.FromInt(int(l.Port())),
		},
	}

	container := v1.Container{
		ImagePullPolicy: v1.PullAlways,
		Name:            constants.ContainerEngineName,
		Image:           image,
		Args: []string{
			"--uri", l.model.Spec.Uri,
		},
		StartupProbe: &v1.Probe{
			InitialDelaySeconds: 30,
			PeriodSeconds:       5,
			FailureThreshold:    120,
			ProbeHandler:        backendProbeHandler,
		},
		ReadinessProbe: &v1.Probe{
			FailureThreshold: 3,
			ProbeHandler:     httpProbeHandler,
		},
		LivenessProbe: &v1.Probe{
			PeriodSeconds:    30,
			TimeoutSeconds:   15,
			FailureThreshold: 10,
			ProbeHandler:     httpProbeHandler,
		},
	}

	mergeProbe(l.AIDeployment.Spec.Deployment.StartupProbe, container.StartupProbe)
	mergeProbe(l.AIDeployment.Spec.Deployment.ReadinessProbe, container.ReadinessProbe)
	mergeProbe(l.AIDeployment.Spec.Deployment.LivenessProbe, container.LivenessProbe)

	pod.AutomountServiceAccountToken = &serviceAccount

	pod.Containers = append(pod.Containers, container)
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
