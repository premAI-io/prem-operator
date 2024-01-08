package engines

import (
	"fmt"

	"github.com/premAI-io/saas-controller/controllers/aideployment"
	"github.com/premAI-io/saas-controller/controllers/constants"
	"github.com/premAI-io/saas-controller/pkg/utils"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Generic struct {
	AIDeployment *a1.AIDeployment
}

func NewGeneric(ai *a1.AIDeployment) aideployment.MLEngine {
	return &Generic{AIDeployment: ai}

}

func (l *Generic) Port() int32 {
	if len(l.AIDeployment.Spec.Endpoint) > 0 {
		return l.AIDeployment.Spec.Endpoint[0].Port
	} else {
		return int32(8000)
	}
}

const GenericEngine = "generic"

func (l *Generic) Deployment(owner metav1.Object) (*appsv1.Deployment, error) {
	objMeta := metav1.ObjectMeta{
		Name:            l.AIDeployment.Name,
		Namespace:       l.AIDeployment.Namespace,
		OwnerReferences: resources.GenOwner(owner),
	}

	deployment := appsv1.Deployment{}
	if l.AIDeployment.Spec.Deployment.PodTemplate != nil {
		deployment.Spec.Template = *l.AIDeployment.Spec.Deployment.PodTemplate.DeepCopy()
	} else {
		return nil, fmt.Errorf("Generic AI deployment %s:%s requires a pod template", objMeta.Namespace, objMeta.Name)
	}
	deployment.Spec.Replicas = l.AIDeployment.Spec.Deployment.Replicas
	pod := &deployment.Spec.Template.Spec

	serviceAccount := false

	expose := &pod.Containers[0]
	expose.Name = constants.ContainerEngineName

	if len(expose.Env) > 0 {
		return nil, fmt.Errorf("Generic AI deployment %s:%s: Specify env vars for the first container in the AIDeployment, not the container", objMeta.Namespace, objMeta.Name)
	}
	expose.Env = l.AIDeployment.Spec.Env

	if len(expose.Ports) > 0 {
		return nil, fmt.Errorf("Generic AI deployment %s:%s: Specify ports in AIDeployment.Spec.Endpoint not the container", objMeta.Namespace, objMeta.Name)
	}
	expose.Ports = []v1.ContainerPort{{ContainerPort: l.AIDeployment.Spec.Endpoint[0].Port}}

	mergeProbe(l.AIDeployment.Spec.Deployment.StartupProbe, expose.StartupProbe)
	mergeProbe(l.AIDeployment.Spec.Deployment.ReadinessProbe, expose.ReadinessProbe)
	mergeProbe(l.AIDeployment.Spec.Deployment.LivenessProbe, expose.LivenessProbe)

	pod.AutomountServiceAccountToken = &serviceAccount

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
