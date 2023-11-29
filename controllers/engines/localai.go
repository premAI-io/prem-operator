package engines

import (
	"fmt"
	"strings"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LocalAI struct {
	AIDeployment *a1.AIDeployment
}

func NewLocalAI(ai *a1.AIDeployment) *LocalAI {
	return &LocalAI{AIDeployment: ai}

}
func (l *LocalAI) Port() int32 {
	// if l.ServicePort != 0 {
	// 	return l.ServicePort
	// }
	return 8080
}

const LocalAIEngine = "localai"

func (l *LocalAI) Deployment(owner metav1.Object) (*appsv1.Deployment, error) {
	objMeta := metav1.ObjectMeta{
		Name:            l.AIDeployment.Name,
		Namespace:       l.AIDeployment.Namespace,
		OwnerReferences: resources.GenOwner(owner),
	}

	imageTag := "latest"
	if l.AIDeployment.Spec.Engine.Options["imageTag"] != "" {
		imageTag = l.AIDeployment.Spec.Engine.Options["imageTag"]
	}

	imageRepository := "quay.io/go-skynet/local-ai"
	if l.AIDeployment.Spec.Engine.Options["imageRepository"] != "" {
		imageRepository = l.AIDeployment.Spec.Engine.Options["imageRepository"]
	}

	deployment := appsv1.Deployment{}
	if l.AIDeployment.Spec.Deployment.PodTemplate != nil {
		deployment.Spec.Template = *l.AIDeployment.Spec.Deployment.PodTemplate
	} else {
		deployment.Spec.Template = v1.PodTemplateSpec{}
	}
	deployment.Spec.Replicas = l.AIDeployment.Spec.Deployment.Replicas

	serviceAccount := false

	v := l.AIDeployment.Spec.Env

	v = append(v, v1.EnvVar{Name: "MODELS_PATH", Value: "/models"})
	image := fmt.Sprintf("%s:%s", imageRepository, imageTag)
	expose := v1.Container{
		ImagePullPolicy: v1.PullAlways,
		Name:            l.AIDeployment.Name,
		Image:           image,
		Env:             v,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "models",
				MountPath: "/models",
			},
		},
	}

	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, expose)
	deployment.Spec.Template.Spec.AutomountServiceAccountToken = &serviceAccount
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, v1.Volume{
		Name: "models",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	// TODO: mount a configmap

	for _, m := range l.AIDeployment.Spec.Models {
		if m.Custom != nil {
			deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, v1.Container{
				ImagePullPolicy: v1.PullAlways,
				Name:            fmt.Sprintf("init-%s", m.Custom.Name),
				Image:           image,
				Command:         []string{"sh", "-c"},
				Args:            []string{"wget -O /models/$MODEL_NAME $MODEL_PATH"},
				Env: []v1.EnvVar{
					{Name: "MODEL_NAME", Value: m.Custom.Name},
					{Name: "MODEL_PATH", Value: m.Custom.Url},
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "models",
						MountPath: "/models",
					},
				},
			})
		} else if len(m.ModelName) > 0 {
			// TODO: support in-built model spec definitions
			var models = map[string]string{
				"llama-7b": "",
			}
			key := strings.ToLower(m.ModelName)
			url, ok := models[key]
			if ok {
				deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, v1.Container{
					ImagePullPolicy: v1.PullAlways,
					Name:            fmt.Sprintf("init-%s", key),
					Image:           image,
					Command:         []string{"sh", "-c"},
					Args:            []string{"wget -O /models/$MODEL_NAME $MODEL_PATH"},
					Env: []v1.EnvVar{
						{Name: "MODEL_NAME", Value: key},
						{Name: "MODEL_PATH", Value: url},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "models",
							MountPath: "/models",
						},
					},
				})
			} else {
				return nil, fmt.Errorf("")
			}
		} else {
			return nil, fmt.Errorf("")
		}
	}

	deploymentLabels := resources.GenDefaultLabels(l.AIDeployment.Name)

	if deployment.Spec.Template.Labels == nil {
		deployment.Spec.Template.Labels = map[string]string{}
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{}
	}

	for k, v := range deploymentLabels {
		deployment.Spec.Template.Labels[k] = v
	}

	for k, v := range l.AIDeployment.Spec.Deployment.Labels {
		deployment.Spec.Template.Labels[k] = v
	}

	for k, v := range l.AIDeployment.Spec.Deployment.Annotations {
		deployment.Spec.Template.Annotations[k] = v
	}

	deployment.ObjectMeta = objMeta
	deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: deploymentLabels}

	return &deployment, nil
}
