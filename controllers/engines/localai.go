package engines

import (
	"fmt"

	deploymentsv1alpha1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type LocalAI struct {
	Name      string
	Namespace string
	Object    metav1.Object
	Options   map[string]string
	Env       []v1.EnvVar
}

const LocalAIEngine = "localai"

func (l *LocalAI) genOwner(kind string) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(l.Object, schema.GroupVersionKind{
			Group:   deploymentsv1alpha1.GroupVersion.Group,
			Version: deploymentsv1alpha1.GroupVersion.Version,
			Kind:    kind,
		}),
	}
}

func (l *LocalAI) Deployment(kind string) (*appsv1.Deployment, error) {
	if kind == "" {
		kind = "SimpleDeployments"
	}
	objMeta := metav1.ObjectMeta{
		Name:            l.Name,
		Namespace:       l.Namespace,
		OwnerReferences: l.genOwner(kind),
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

	expose := v1.Container{
		ImagePullPolicy: v1.PullAlways,
		Name:            l.Name,
		Image:           fmt.Sprintf("quay.io/go-skynet/local-ai:%s", imageTag),
		Env:             v,
	}

	pod := v1.PodSpec{
		Containers:                   []v1.Container{expose},
		AutomountServiceAccountToken: &serviceAccount,
	}

	deploymentLabels := genDeploymentLabel(l.Name)
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

func genDeploymentLabel(s string) map[string]string {
	return map[string]string{
		"ai-workload": s,
	}
}
