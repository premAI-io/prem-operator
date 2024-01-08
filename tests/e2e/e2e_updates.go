package e2e_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

func randomString(length int) string {
	r := rand.New(rand.NewSource(1234567890))

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[r.Intn(len(charset))]
	}
	return string(result)
}

var _ = Describe("update test", func() {
	var artifactName string
	var svc, ingr, sds, pods dynamic.ResourceInterface
	var scheme *runtime.Scheme
	var startTime time.Time

	BeforeEach(func() {
		startTime = time.Now()
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		sds = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "aideployments"}).Namespace("default")
		pods = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("default")
		ingr = k8s.Resource(schema.GroupVersionResource{Group: networkv1.GroupName, Version: networkv1.SchemeGroupVersion.Version, Resource: "ingresses"}).Namespace("default")
		svc = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "services"}).Namespace("default")

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
				Deployment: api.Deployment{
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							"memory": resource.MustParse("70Mi"),
						},
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "DEBUG",
						Value: "true",
					},
				},
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

		checkLogs(startTime)
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

		sd.Spec.Endpoint = []api.Endpoint{
			{
				Domain: "test.127.0.0.1.nip.io",
			},
		}

		un, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sd)
		Expect(err).ToNot(HaveOccurred())

		d := &unstructured.Unstructured{}
		d.SetUnstructuredContent(un)

		u, err = sds.Update(context.Background(), d, metav1.UpdateOptions{})
		Expect(err).ToNot(HaveOccurred())
		artifactName = u.GetName()

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

				if !found {
					return false
				}

				GinkgoWriter.Printf("Found the pod ENV\n")

				i := &networkv1.Ingress{}
				if !getObjectWithAnnotation(ingr, i, resources.DefaultAnnotation, artifactName) {
					return false
				}

				g.Expect(i.Spec.Rules[0].Host).To(Equal("test.127.0.0.1.nip.io"))
				g.Expect(i.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number).To(Equal(int32(8080)))

				return true
			}).WithPolling(30 * time.Second).WithTimeout(time.Minute).Should(BeTrue())

		u, err = sds.Get(context.Background(), artifactName, metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sd)
		Expect(err).ToNot(HaveOccurred())

		sd.Spec.Service.Labels = map[string]string{"a-test-lable": "test"}
		sd.Spec.Ingress.Labels = map[string]string{"a-test-lable": "test"}

		un, err = runtime.DefaultUnstructuredConverter.ToUnstructured(sd)
		Expect(err).ToNot(HaveOccurred())

		d = &unstructured.Unstructured{}
		d.SetUnstructuredContent(un)

		u, err = sds.Update(context.Background(), d, metav1.UpdateOptions{})
		Expect(err).ToNot(HaveOccurred())
		artifactName = u.GetName()

		Eventually(func() bool {
			s := &corev1.Service{}
			if !getObjectWithLabel(svc, s, "a-test-lable", "test") {
				return false
			}

			i := &networkv1.Ingress{}
			return !getObjectWithLabel(ingr, i, "a-test-lable", "test")
		})
	})
})
