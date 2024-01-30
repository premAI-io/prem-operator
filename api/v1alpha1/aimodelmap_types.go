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

// +enum
type AIModelFormat string

const (
	AIModelFormatGguf        AIModelFormat = "gguf"
	AIModelFormatGgml        AIModelFormat = "ggml"
	AIModelFormatOnnx        AIModelFormat = "onnx"
	AIModelFormatCtranslate  AIModelFormat = "ctranslate"
	AIModelFormatPickle      AIModelFormat = "pickle"
	AIModelFormatSafeTensor  AIModelFormat = "safetensor"
	AIModelFormatProtobuf    AIModelFormat = "protobuf"
	AIModelFormatMessagePack AIModelFormat = "msgpack"
)

// +enum
type AIModelFramework string

const (
	AIModelFrameworkPytorch    AIModelFramework = "pytorch"
	AIModelFrameworkTensorflow AIModelFramework = "tensorflow"
	AIModelFrameworkSklearn    AIModelFramework = "sklearn"
	AIModelFrameworkKeras      AIModelFramework = "keras"
	AIModelFrameworkXgboost    AIModelFramework = "xgboost"
	AIModelFrameworkLightgbm   AIModelFramework = "lightgbm"
	AIModelFrameworkPaddle     AIModelFramework = "paddlepaddle"
)

// +enum
type AIModelQuantization string

const (
	AIModelQuantizationAWQ  AIModelQuantization = "awq"
	AIModelQuantizationGPTQ AIModelQuantization = "gptq"
)

// +enum
type AIModelDataType string

const (
	AIModelDataTypeInt4     AIModelDataType = "int4"
	AIModelDataTypeInt8     AIModelDataType = "int8"
	AIModelDataTypeInt16    AIModelDataType = "int16"
	AIModelDataTypeInt32    AIModelDataType = "int32"
	AIModelDataTypeBFloat16 AIModelDataType = "bfloat16"
	AIModelDataTypeFloat16  AIModelDataType = "float16"
	AIModelDataTypeFloat32  AIModelDataType = "float32"
)

// Parameters that match Prem's UI
// TODO: Add to AIModelSpec or remove
type AIModelRuntimeParams struct {
	// +optional
	SystemPrompt string `json:"systemprompt,omitempty"`
	MaxTokens    int    `json:"maxtokens,omitempty"`
	// +kubebuilder:validation:Pattern=`^-?\d+(\.\d+)?$`
	Temperature string `json:"temperature,omitempty"`
	// +kubebuilder:validation:Pattern=`^-?\d+(\.\d+)?$`
	TopP string `json:"topp,omitempty"`
	// +kubebuilder:validation:Pattern=`^-?\d+(\.\d+)?$`
	FreqPenalty string `json:"freqpenalty,omitempty"`
	// +kubebuilder:validation:Pattern=`^-?\d+(\.\d+)?$`
	PresPenalty string `json:"prespenalty,omitempty"`
}

// Information needed to use a model
// TODO: Uncomment or remove fields
type AIModelSpec struct {
	Uri string `json:"uri,omitempty"`
	//CacheUri          string			  `json:"cache,omitempty"`
	//Format           AIModelFormat       `json:"format,omitempty"`
	//Framework        AIModelFramework    `json:"framework,omitempty"`
	// +optional
	Quantization AIModelQuantization `json:"quantization,omitempty"`
	// +optional
	DataType AIModelDataType `json:"dataType,omitempty"`
	//ParameterCount   string              `json:"parametercount,omitempty"`
	//BitsPerParameter uint8               `json:"bitsperparameter,omitempty"`

	//DefaultRuntimeParams *AIModelRuntimeParams `json:"defaultruntimeparams,omitempty"`
}

type AIModelMapReference struct {
	// +optional
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	Variant   string `json:"variant"`
}

type AIModelVariant struct {
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Name        string `json:"name"`
	AIModelSpec `json:",inline"`
}

// AIModelMapSpec defines the desired state of AIModelMap
type AIModelMapSpec struct {
	Localai      []AIModelVariant `json:"localai,omitempty"`
	Vllm         []AIModelVariant `json:"vllm,omitempty"`
	DeepSpeedMii []AIModelVariant `json:"deepspeed-mii,omitempty"`
}

// AIModelMapStatus defines the observed state of AIModelMap
type AIModelMapStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AIModelMap is the Schema for the aimodelmaps API
type AIModelMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AIModelMapSpec   `json:"spec,omitempty"`
	Status AIModelMapStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AIModelMapList contains a list of AIModelMap
type AIModelMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AIModelMap `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AIModelMap{}, &AIModelMapList{})
}