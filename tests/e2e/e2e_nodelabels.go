package e2e_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("automatic node labelling tests", func() {
	var artifactName string
	var lb, nodes dynamic.ResourceInterface
	var scheme *runtime.Scheme

	BeforeEach(func() {
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		lb = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "autonodelabelers"}).Namespace("default")
		nodes = k8s.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"})

		key := "kubernetes.io/arch"
		op := metav1.LabelSelectorOpExists

		ll := &api.AutoNodeLabeler{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AutoNodeLabeler",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "autolabel-",
			},
			Spec: api.AutoNodeLabelerSpec{
				MatchExpression: api.LabelSelectorRequirementApplyConfiguration{
					Key:      &key,
					Operator: &op,
				},
				Labels: map[string]string{
					"foo/bar.com": "baz",
				},
			},
		}

		uArtifact := unstructured.Unstructured{}
		uArtifact.Object, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(ll)
		resp, err := lb.Create(context.TODO(), &uArtifact, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred(), fmt.Sprint(ll))
		artifactName = resp.GetName()
	})

	AfterEach(func() {
		err := lb.Delete(context.Background(), artifactName, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("specify a rule", func() {
		By("relabeling the node")
		Eventually(
			func(g Gomega) bool {
				list, err := nodes.List(context.TODO(), metav1.ListOptions{})
				g.Expect(err).ToNot(HaveOccurred())
				found := false

				for _, node := range list.Items {
					p := &corev1.Node{}
					err := runtime.DefaultUnstructuredConverter.FromUnstructured(node.Object, p)
					g.Expect(err).ToNot(HaveOccurred())
					labels := p.Labels
					for l, v := range labels {
						if l == "foo/bar.com" && v == "baz" {
							found = true
						}
					}
				}
				return found
			}).WithPolling(4 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
	})
})
