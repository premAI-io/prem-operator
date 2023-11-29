package e2e_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("replicas test", func() {
	var artifactName string
	var sds, pods dynamic.ResourceInterface
	var scheme *runtime.Scheme

	BeforeEach(func() {
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		sds = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "aideployments"}).Namespace("default")
		pods = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("default")
		replicas := int32(2)
		artifact := &api.AIDeployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AIDeployment",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "simple-",
			},
			Spec: api.AIDeploymentSpec{
				Deployment: api.Deployment{
					Replicas: &replicas,
				},
				Engine: api.AIEngine{
					Name: "localai",
				},
			},
		}

		uArtifact := unstructured.Unstructured{}
		uArtifact.Object, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(artifact)
		resp, err := sds.Create(context.TODO(), &uArtifact, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred())
		artifactName = resp.GetName()
	})

	AfterEach(func() {
		err := sds.Delete(context.Background(), artifactName, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("starts a deployment", func() {
		By("starting the workload with the associated label")

		Eventually(func(g Gomega) bool {
			list, err := pods.List(context.TODO(), metav1.ListOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			found := 0

			for _, pod := range list.Items {
				p := &corev1.Pod{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(pod.Object, p)
				g.Expect(err).ToNot(HaveOccurred())
				if v, ok := p.Labels[resources.DefaultAnnotation]; ok && v == artifactName {
					found++
				}
			}

			return found == 2
		}).WithTimeout(time.Minute).Should(BeTrue())
	})
})
