package engines

import (
	"fmt"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/aideployment"
	"github.com/premAI-io/saas-controller/controllers/aimodelmap"
	"github.com/premAI-io/saas-controller/controllers/constants"
	"github.com/premAI-io/saas-controller/controllers/resources"
	"github.com/premAI-io/saas-controller/pkg/utils"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	vllmContainerVolumePath = "/root/.cache/huggingface"
)

const (
	vllmImageFormat = "%s:%s"
)

var (
	ErrModelsNotSpecified = fmt.Errorf("models not specified")
	ErrorOnlyOneModel     = fmt.Errorf("only one model can be specified")
)

type vllmAi struct {
	// image of the vllm engine
	engineImage string
	// environment variables to pass to the vllm engine
	engineEnvVars []v1.EnvVar

	// name of the k8s resource
	resourceName string
	// k8s namespace
	namespace string

	// used to customize the deployment
	deploymentOptions *a1.AIDeployment
	model             aimodelmap.ResolvedModel
}

func NewVllmAi(ai *a1.AIDeployment, models []aimodelmap.ResolvedModel) (aideployment.MLEngine, error) {
	if len(models) == 0 {
		return nil, ErrModelsNotSpecified
	}

	if len(models) > 1 {
		return nil, ErrorOnlyOneModel
	}

	model := models[0]
	imageTag := "latest"
	imageRepo := constants.ImageRepositoryVllm
	if ai.Spec.Engine.Options[constants.ImageTagKey] != "" {
		imageTag = ai.Spec.Engine.Options[constants.ImageTagKey]
	}
	if ai.Spec.Engine.Options[constants.ImageRepositoryKey] != "" {
		imageRepo = ai.Spec.Engine.Options[constants.ImageRepositoryKey]
	}
	engineImage := fmt.Sprintf(vllmImageFormat, imageRepo, imageTag)

	return &vllmAi{
		engineImage: engineImage,

		resourceName:      ai.Name,
		namespace:         ai.Namespace,
		engineEnvVars:     ai.Spec.Env,
		deploymentOptions: ai,
		model:             model,
	}, nil
}

func (v *vllmAi) Port() int32 {
	return 8000
}

func (v *vllmAi) Deployment(owner metav1.Object) (*appsv1.Deployment, error) {
	log.Info("Creating deployment for vllm engine, model: ", v.model.Name)
	healthProbeHandler := v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path: "/health",
			Port: intstr.FromInt(int(v.Port())),
		},
	}

	container := v1.Container{
		ImagePullPolicy: v1.PullAlways,
		Name:            constants.ContainerEngineName,
		Image:           v.engineImage,
		Env:             v.engineEnvVars,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "models",
				MountPath: vllmContainerVolumePath,
			},
		},
		Args: []string{
			"--model", v.model.Spec.Uri,
		},
		StartupProbe: &v1.Probe{
			InitialDelaySeconds: 3,
			PeriodSeconds:       1,
			FailureThreshold:    120,
			ProbeHandler:        healthProbeHandler,
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

	engineOpts := make(map[string]string)
	if v.model.Spec.DataType != "" {
		engineOpts[constants.DtypeKey] = string(v.model.Spec.DataType)
	}
	if v.model.Spec.Quantization != "" {
		engineOpts[constants.QuantizationKey] = string(v.model.Spec.Quantization)
	}
	v.deploymentOptions.Spec.Engine.Options = utils.MergeMaps(engineOpts, v.deploymentOptions.Spec.Engine.Options)

	if dtype, ok := v.deploymentOptions.Spec.Engine.Options[constants.DtypeKey]; ok {
		if utils.IsAlphanumeric(dtype) {
			container.Args = append(container.Args, "--dtype", dtype)
		} else {
			return nil, fmt.Errorf("dtype must be alphanumeric")
		}
	}

	if quant, ok := v.deploymentOptions.Spec.Engine.Options[constants.QuantizationKey]; ok {
		if utils.IsAlphanumeric(quant) {
			container.Args = append(container.Args, "--quantization", quant)
		} else {
			return nil, fmt.Errorf("quantization must be alphanumeric")
		}
	}

	mergeProbe(v.deploymentOptions.Spec.Deployment.StartupProbe, container.StartupProbe)
	mergeProbe(v.deploymentOptions.Spec.Deployment.ReadinessProbe, container.ReadinessProbe)
	mergeProbe(v.deploymentOptions.Spec.Deployment.LivenessProbe, container.LivenessProbe)

	serviceAccount := false
	replicas := int32(1)
	if v.deploymentOptions.Spec.Deployment.Replicas != nil {
		replicas = *v.deploymentOptions.Spec.Deployment.Replicas
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v.resourceName,
			Namespace:       v.namespace,
			OwnerReferences: resources.GenOwner(owner),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: resources.GenDefaultLabels(v.resourceName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: utils.MergeMaps(
						resources.GenDefaultLabels(v.resourceName),
						v.deploymentOptions.Spec.Deployment.Labels,
					),
					Annotations: utils.MergeMaps(
						v.deploymentOptions.Spec.Deployment.Annotations,
					),
				},
				Spec: v1.PodSpec{
					Containers:                   []v1.Container{},
					AutomountServiceAccountToken: &serviceAccount,
					Volumes: []v1.Volume{
						{
							Name: "models",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, container)
	return deployment, nil
}
