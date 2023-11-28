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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AIDeploymentSpec defines the desired state of AIDeployment
type AIDeploymentSpec struct {
	Endpoint []Endpoint `json:"endpoint,omitempty"`
	Engine   AIEngine   `json:"engine,omitempty"`

	// +optional
	Env []v1.EnvVar `json:"env,omitempty"`

	// +optional
	Services Services `json:"services,omitempty"`

	// +optional
	Deployment Deployment `json:"deployment,omitempty"`

	// +optional
	Ingress Ingress `json:"ingress,omitempty"`

	Models []AIModel `json:"models,omitempty"`
}

type Services struct {
	// +optional
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Deployment struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	// +optional
	DeploymentSpec *appsv1.DeploymentSpec `json:"spec,omitempty"`
}

type Ingress struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Endpoint struct {
	Domain string `json:"domain"`
}

type AIEngine struct {
	Name string `json:"name"`
	// +optional
	Options map[string]string `json:"options,omitempty"`
}

type AIModel struct {
	// +optional
	Custom *AIModelCustom `json:"custom,omitempty"`
	// +optional
	ModelName string `json:"model_name,omitempty"`
	// +optional
	Options map[string]string `json:"options,omitempty"`
}

// +enum
type AIModelFormat string

const (
	AIModelFormatGguf       AIModelFormat = "gguf"
	AIModelFormatCtranslate AIModelFormat = "ctranslate"
	AIModelFormatPickle     AIModelFormat = "picke"
	AIModelFormatSafeTensor AIModelFormat = "safetensor"
)

type AIModelCustom struct {
	Format AIModelFormat `json:"format"`
	Name   string        `json:"name"`
	Url    string        `json:"url"`
	// +optional
	Checksum string `json:"checksum,omitempty"`
}

// AIDeploymentStatus defines the observed state of AIDeployment
type AIDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status,omitempty"`
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
