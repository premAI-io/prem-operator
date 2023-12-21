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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutoNodeLabelerSpec defines the desired state of AutoNodeLabeler
type AutoNodeLabelerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	MatchExpression LabelSelectorRequirementApplyConfiguration `json:"matchExpression,omitempty"`
	Labels          map[string]string                          `json:"labels,omitempty"`
}

type LabelSelectorRequirementApplyConfiguration struct {
	Key      *string                       `json:"key,omitempty"`
	Operator *metav1.LabelSelectorOperator `json:"operator,omitempty"`
	Values   []string                      `json:"values,omitempty"`
}

// AutoNodeLabelerStatus defines the observed state of AutoNodeLabeler
type AutoNodeLabelerStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AutoNodeLabeler is the Schema for the autonodelabelers API
type AutoNodeLabeler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoNodeLabelerSpec   `json:"spec,omitempty"`
	Status AutoNodeLabelerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutoNodeLabelerList contains a list of AutoNodeLabeler
type AutoNodeLabelerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoNodeLabeler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoNodeLabeler{}, &AutoNodeLabelerList{})
}
