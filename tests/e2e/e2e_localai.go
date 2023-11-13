package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("localai test", func() {
	var artifactName string
	var sds, pods dynamic.ResourceInterface
	var scheme *runtime.Scheme

	BeforeEach(func() {
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		sds = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "simpledeployments"}).Namespace("default")

		artifact := &api.SimpleDeployments{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SimpleDeployments",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "localai-",
			},
			Spec: api.SimpleDeploymentsSpec{
				MLEngine: "localai",
				Domain:   "foo.127.0.0.1.nip.io",
				Models: []map[string]string{
					{
						"name": "gpt-4",
						"url":  "https://huggingface.co/TheBloke/WizardLM-7B-uncensored-GGUF/resolve/main/WizardLM-7B-uncensored.Q2_K.gguf",
					},
				},
			},
		}
		pods = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("default")

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

	It("starts the API", func() {
		By("starting the workload with the associated label")
		Eventually(func(g Gomega) bool {
			list, err := pods.List(context.TODO(), metav1.ListOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			found := false

			fmt.Println("PODS")
			for _, pod := range list.Items {
				p := &corev1.Pod{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(pod.Object, p)
				g.Expect(err).ToNot(HaveOccurred())
				if v, ok := p.Labels[resources.DefaultAnnotation]; ok && v == artifactName {
					found = ok
				}
				fmt.Println(pod)
			}

			return found
		}).WithPolling(30 * time.Second).WithTimeout(time.Minute).Should(BeTrue())

		Eventually(func(g Gomega) bool {
			list, err := pods.List(context.TODO(), metav1.ListOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			found := false

			fmt.Println("PODS")
			deploymentPod := &corev1.Pod{}
			for _, pod := range list.Items {
				p := &corev1.Pod{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(pod.Object, p)
				g.Expect(err).ToNot(HaveOccurred())
				if v, ok := p.Labels[resources.DefaultAnnotation]; ok && v == artifactName {
					found = ok
					deploymentPod = p
					fmt.Println(pod)
				}
			}

			if found {
				return deploymentPod.Status.Phase == corev1.PodRunning
			}

			return false
		}).WithPolling(30 * time.Second).WithTimeout(time.Hour).Should(BeTrue())

		Eventually(func(g Gomega) string {
			fmt.Println("Polling API for a response")
			url := "http://foo.127.0.0.1.nip.io:8080/v1/completions"
			payload := []byte(`{
				"model": "gpt-4",
				"prompt": "How are you?",
				"temperature": 0.1
			}`)

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
			if err != nil {
				fmt.Println("Error creating request:", err)
				return ""
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error making request:", err)
				return ""
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			return string(body)
		}).WithPolling(30 * time.Second).WithTimeout(time.Hour).Should(ContainSubstring("doing well"))
	})
})
