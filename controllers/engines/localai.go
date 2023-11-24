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
	Name        string
	Namespace   string
	ServicePort int32
	Options     map[string]string
	Endpoint    []a1.Endpoint
	Env         []v1.EnvVar
	Models      []a1.AIModel
}

func (l *LocalAI) Port() int32 {
	if l.ServicePort != 0 {
		return l.ServicePort
	}
	return 8080
}

const LocalAIEngine = "localai"

func (l *LocalAI) Deployment(owner metav1.Object) (*appsv1.Deployment, error) {
	objMeta := metav1.ObjectMeta{
		Name:            l.Name,
		Namespace:       l.Namespace,
		OwnerReferences: resources.GenOwner(owner),
	}

	imageTag := "latest"
	if l.Options["imageTag"] != "" {
		imageTag = l.Options["imageTag"]
	}
	serviceAccount := false

	v := l.Env

	v = append(v, v1.EnvVar{Name: "MODELS_PATH", Value: "/models"})

	expose := v1.Container{
		ImagePullPolicy: v1.PullAlways,
		Name:            l.Name,
		Image:           fmt.Sprintf("quay.io/go-skynet/local-ai:%s", imageTag),
		Env:             v,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "models",
				MountPath: "/models",
			},
		},
	}

	pod := v1.PodSpec{
		Containers:                   []v1.Container{expose},
		AutomountServiceAccountToken: &serviceAccount,
	}

	pod.Volumes = append(pod.Volumes, v1.Volume{
		Name: "models",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	// mount a configmap

	for _, m := range l.Models {
		if m.Custom != nil {
			pod.InitContainers = append(pod.InitContainers, v1.Container{
				ImagePullPolicy: v1.PullAlways,
				Name:            fmt.Sprintf("init-%s", m.Custom.Name),
				Image:           fmt.Sprintf("quay.io/go-skynet/local-ai:%s", imageTag),
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
				pod.InitContainers = append(pod.InitContainers, v1.Container{
					ImagePullPolicy: v1.PullAlways,
					Name:            fmt.Sprintf("init-%s", key),
					Image:           fmt.Sprintf("quay.io/go-skynet/local-ai:%s", imageTag),
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

	deploymentLabels := resources.GenDefaultLabels(l.Name)
	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: objMeta,
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: deploymentLabels},
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
				},
				Spec: pod,
			},
		},
	}, nil
}
