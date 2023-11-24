package e2e_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/premAI-io/saas-controller/controllers/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "saas-operator e2e test Suite")
}

func findWorkloadPod(d dynamic.ResourceInterface, artifactName string) func(g Gomega) bool {
	return func(g Gomega) bool {
		list, err := d.List(context.TODO(), metav1.ListOptions{})
		g.Expect(err).ToNot(HaveOccurred())
		found := false

		for _, pod := range list.Items {
			p := &corev1.Pod{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(pod.Object, p)
			g.Expect(err).ToNot(HaveOccurred())
			if v, ok := p.Labels[resources.DefaultAnnotation]; ok && v == artifactName {
				found = ok
			}
		}

		return found
	}
}
