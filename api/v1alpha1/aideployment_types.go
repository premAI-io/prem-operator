/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/premAI-io/saas-controller/controllers/constants"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AIDeploymentSpec defines the desired state of AIDeployment
type AIDeploymentSpec struct {
	Endpoint []Endpoint `json:"endpoint,omitempty"`
	Engine   AIEngine   `json:"engine,omitempty"`

	RateLimit      bool `json:"ratelimit,omitempty"`
	Authentication bool `json:"auth,omitempty"`

	// Optionally specify a list of environment variables used to start the engine
	// +optional
	Env []v1.EnvVar `json:"env,omitempty"`

	// Optionally specify a list of args used to start the engine. It is preferred to not use this field but instead rely on the Engines settings and options. This is meant for overrides and development.
	// +optional
	Args []string `json:"args,omitempty"`

	// +optional
	Service Service `json:"service,omitempty"`

	// +optional
	Deployment Deployment `json:"deployment,omitempty"`

	// +optional
	Ingress Ingress `json:"ingress,omitempty"`

	Models []AIModel `json:"models,omitempty"`
}

type Service struct {
	// +optional
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Probe struct {
	// +optional
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
	// +optional
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// +optional
	SuccessThreshold int32 `json:"successThreshold,omitempty"`
	// +optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
}

type AcceleratorInterface string

const (
	AcceleratorInterfaceCUDA   AcceleratorInterface = "CUDA"
	AcceleratorInterfaceROCm   AcceleratorInterface = "ROCm"
	AcceleratorInterfaceOpenCL AcceleratorInterface = "OpenCL"
)

type Version struct {
	Major int32 `json:"major"`
	// +optional
	Minor int32 `json:"minor,omitempty"`
}

type Accelerator struct {
	// The name of something the hardware and software needs to support
	Interface AcceleratorInterface `json:"interface"`
	// The minimum needed version of the interface or whatever
	// (e.g. nvidia compute engine)
	// +optional
	MinVersion *Version `json:"minVersion,omitempty"`
}

type Deployment struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	// What kind of interface (e.g. CUDA) the accelerator hardware
	// should support.
	// +optional
	Accelerator *Accelerator `json:"accelerator,omitempty"`

	// The deployment must request the minimum amount of memory required by the models
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// +optional
	PodTemplate *v1.PodTemplateSpec `json:"template,omitempty"`

	// +optional
	StartupProbe *Probe `json:"startupProbe,omitempty"`
	// +optional
	ReadinessProbe *Probe `json:"readinessProbe,omitempty"`
	// +optional
	LivenessProbe *Probe `json:"livenessProbe,omitempty"`
}

type Ingress struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	TLS *bool `json:"tls,omitempty"`
}

type Endpoint struct {
	Domain string `json:"domain"`
	// +optional
	Port int32 `json:"port,omitempty"`
}

// +enum
type AIEngineName string

const (
	AIEngineNameLocalai      AIEngineName = "localai"
	AIEngineNameVLLM         AIEngineName = "vllm"
	AIEngineNameGeneric      AIEngineName = "generic"
	AIEngineNameDeepSpeedMii AIEngineName = "deepspeed-mii"
)

type AIEngine struct {
	Name AIEngineName `json:"name"`
	// +optional
	Options map[string]string `json:"options,omitempty"`
}

type AIModel struct {
	// +optional
	ModelMapRef *AIModelMapReference `json:"modelMapRef,omitempty"`
	AIModelSpec `json:",inline"`
	// +optional
	Options map[string]string `json:"options,omitempty"`
}

// AIDeploymentStatus defines the observed state of AIDeployment
type AIDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status constants.Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AIDeployment is the Schema for the AIDeployment API
type AIDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AIDeploymentSpec   `json:"spec,omitempty"`
	Status AIDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AIDeploymentList contains a list of AIDeployment
type AIDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AIDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AIDeployment{}, &AIDeploymentList{})
}
