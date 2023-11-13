package engines

import (
	"fmt"

	"github.com/premAI-io/saas-controller/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LocalAI struct {
	Name      string
	Namespace string
	Options   map[string]string
	Env       []v1.EnvVar
	Models    []map[string]string
}

const LocalAIEngine = "localai"

func (l *LocalAI) Port() int32 {
	return 8080
}

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

	// svc := &v1.Service{}
	// if ent.Spec.ServiceRef != nil {
	// 	err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: ent.Namespace, Name: *ent.Spec.ServiceRef}, svc)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	v := l.Env

	v = append(v, v1.EnvVar{Name: "MODELS_PATH", Value: "/models"})

	expose := v1.Container{
		ImagePullPolicy: v1.PullAlways,
		Name:            l.Name,
		Image:           fmt.Sprintf("quay.io/go-skynet/local-ai:%s", imageTag),
		Env:             v,
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
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
		if _, ok := m["url"]; ok {
			if _, ok := m["name"]; ok {
				pod.InitContainers = append(pod.InitContainers, v1.Container{
					ImagePullPolicy: v1.PullAlways,
					Name:            fmt.Sprintf("init-%s", m["name"]),
					Image:           fmt.Sprintf("quay.io/go-skynet/local-ai:%s", imageTag),
					Command:         []string{"/bin/bash", "-c"},
					Args:            []string{"wget -O /models/$MODEL_NAME $MODEL_PATH"},
					Env: []v1.EnvVar{
						{Name: "MODEL_NAME", Value: m["name"]},
						{Name: "MODEL_PATH", Value: m["url"]},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "models",
							MountPath: "/models",
						},
					},
				})
			}
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
