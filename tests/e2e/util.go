package e2e_test

import (
	"context"

	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

func getObjectWithLabel(resource dynamic.ResourceInterface, obj metav1.Object, labelName string, labelValue string) bool {
	list, err := resource.List(context.TODO(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	found := false

	for _, item := range list.Items {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, obj)

		Expect(err).ToNot(HaveOccurred())

		if v, ok := obj.GetLabels()[labelName]; ok && v == labelValue {
			found = ok
			break
		}
	}

	if !found {
		obj = nil
	}

	return found
}

func getObjectWithAnnotation(resource dynamic.ResourceInterface, obj metav1.Object, labelName string, labelValue string) bool {
	list, err := resource.List(context.TODO(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	found := false

	for _, item := range list.Items {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, obj)

		Expect(err).ToNot(HaveOccurred())

		if v, ok := obj.GetAnnotations()[labelName]; ok && v == labelValue {
			found = ok
			break
		}
	}

	if !found {
		obj = nil
	}

	return found
}

func getObjectWithName(resource dynamic.ResourceInterface, obj metav1.Object, name string) bool {
	list, err := resource.List(context.TODO(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	found := false

	for _, item := range list.Items {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, obj)

		Expect(err).ToNot(HaveOccurred())

		if obj.GetName() == name {
			found = true
			break
		}
	}

	if !found {
		obj = nil
	}

	return found
}
