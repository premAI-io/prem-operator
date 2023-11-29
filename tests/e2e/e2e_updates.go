package e2e_test

import (
	"context"
	"fmt"
	"math/rand"
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

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

var _ = Describe("update test", func() {
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

		artifact := &api.AIDeployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AIDeployment",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "update-",
			},
			Spec: api.AIDeploymentSpec{
				Engine: api.AIEngine{Name: "localai"},
			},
		}

		uArtifact := unstructured.Unstructured{}
		uArtifact.Object, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(artifact)
		resp, err := sds.Create(context.TODO(), &uArtifact, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred(), fmt.Sprint(artifact))
		artifactName = resp.GetName()
	})

	AfterEach(func() {
		err := sds.Delete(context.Background(), artifactName, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("starts a deployment and updates it", func() {
		By("starting the workload with the associated label")
		Eventually(findWorkloadPod(pods, artifactName)).WithPolling(30 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
		u, err := sds.Get(context.Background(), artifactName, metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		sd := &api.AIDeployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sd)
		Expect(err).ToNot(HaveOccurred())

		testString := randomString(5)

		sd.Spec.Env = append(sd.Spec.Env, corev1.EnvVar{
			Name:  "TEST",
			Value: testString,
		})

		un, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sd)
		Expect(err).ToNot(HaveOccurred())

		d := &unstructured.Unstructured{}
		d.SetUnstructuredContent(un)

		_, err = sds.Update(context.Background(), d, metav1.UpdateOptions{})
		Expect(err).ToNot(HaveOccurred())

		Eventually(
			func(g Gomega) bool {
				list, err := pods.List(context.TODO(), metav1.ListOptions{})
				g.Expect(err).ToNot(HaveOccurred())
				found := false

				for _, pod := range list.Items {
					p := &corev1.Pod{}
					err := runtime.DefaultUnstructuredConverter.FromUnstructured(pod.Object, p)
					g.Expect(err).ToNot(HaveOccurred())
					envs := p.Spec.Containers[0].Env
					for _, env := range envs {
						if env.Name == "TEST" && env.Value == testString {
							found = true
						}
					}
				}
				return found
			}).WithPolling(30 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
	})
})
