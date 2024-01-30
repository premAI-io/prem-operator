package aimodelmap

import (
	"context"
	"fmt"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/pkg/utils"
)

type ResolvedModel struct {
	Name string
	Spec a1.AIModelSpec
}

func Resolve(d *a1.AIDeployment, ctx context.Context, c ctrlClient.Client) ([]ResolvedModel, error) {
	ms := make([]ResolvedModel, 0, len(d.Spec.Models))

	for _, m := range d.Spec.Models {
		rm, err := resolveOne(&m, d, ctx, c)
		if err != nil {
			return nil, err
		}

		if rm != nil {
			ms = append(ms, *rm)
		}
	}

	return ms, nil
}

func findVariant(mv []a1.AIModelVariant, name string) *a1.AIModelSpec {
	for _, v := range mv {
		if v.Name == name {
			return &v.AIModelSpec
		}
	}

	return nil
}

func mergeModelSpecs(primary *a1.AIModelSpec, secondary *a1.AIModelSpec) *a1.AIModelSpec {
	result := primary.DeepCopy()

	if secondary == nil {
		return result
	}

	if result.Uri == "" {
		result.Uri = secondary.Uri
	}

	if result.DataType == "" {
		result.DataType = secondary.DataType
	}

	if result.Quantization == "" {
		result.Quantization = secondary.Quantization
	}

	return result
}

func resolveOne(m *a1.AIModel, d *a1.AIDeployment, ctx context.Context, c ctrlClient.Client) (*ResolvedModel, error) {
	if m.ModelMapRef == nil {
		return &ResolvedModel{
			Name: utils.ToHostName(d.Name + "-model"),
			Spec: m.AIModelSpec,
		}, nil
	}

	if m.ModelMapRef.Name == "" {
		return nil, fmt.Errorf("deployment %s/%s has modelMapRef with no name", d.Namespace, d.Name)
	}

	name := m.ModelMapRef.Name
	namespace := m.ModelMapRef.Namespace
	if namespace == "" {
		namespace = d.Namespace
	}

	if m.ModelMapRef.Variant == "" {
		return nil, fmt.Errorf("deployment %s/%s has modelMapRef with no variant", d.Namespace, d.Name)
	}

	mm := &a1.AIModelMap{}
	if err := c.Get(ctx, ctrlClient.ObjectKey{Namespace: namespace, Name: name}, mm); err != nil {
		return nil, err
	}

	var variants []a1.AIModelVariant
	switch d.Spec.Engine.Name {
	case a1.AIEngineNameLocalai:
		variants = mm.Spec.Localai
	case a1.AIEngineNameVLLM:
		variants = mm.Spec.Vllm
	case a1.AIEngineNameDeepSpeedMii:
		variants = mm.Spec.DeepSpeedMii
	case a1.AIEngineNameGeneric:
		return nil, fmt.Errorf("deployment %s/%s: Can't specify a model with generic engine", d.Namespace, d.Name)
	default:
		return nil, fmt.Errorf("deployment %s/%s has unknown engine %s", d.Namespace, d.Name, d.Spec.Engine.Name)
	}

	variant := findVariant(variants, m.ModelMapRef.Variant)
	if variant == nil {
		return nil, fmt.Errorf("deployment %s/%s has no model variant %s for %s", d.Namespace, d.Name, m.ModelMapRef.Variant, a1.AIEngineNameLocalai)
	}

	merged := mergeModelSpecs(&m.AIModelSpec, variant)

	return &ResolvedModel{
		Name: utils.ToHostName(m.ModelMapRef.Name + "-" + m.ModelMapRef.Variant),
		Spec: *merged,
	}, nil
}
