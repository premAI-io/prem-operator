package e2e_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("localai test", func() {
	var artifactName string
	var oper, sds, pods, svc, ingr dynamic.ResourceInterface
	var scheme *runtime.Scheme
	var artifact *api.AIDeployment
	var startTime time.Time

	JustBeforeEach(func() {
		startTime = time.Now()
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		sds = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "aideployments"}).Namespace("default")
		svc = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "services"}).Namespace("default")
		ingr = k8s.Resource(schema.GroupVersionResource{Group: networkv1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "ingresses"}).Namespace("default")

		pods = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("default")
		oper = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("saas-operator-system")

		uArtifact := unstructured.Unstructured{}
		uArtifact.Object, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(artifact)
		resp, err := sds.Create(context.TODO(), &uArtifact, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred())
		artifactName = resp.GetName()
		GinkgoWriter.Printf("artifactName: %s\n", artifactName)
	})

	AfterEach(func() {
		err := sds.Delete(context.Background(), artifactName, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())

		controllerPod := &corev1.Pod{}
		getObjectWithLabel(oper, controllerPod, "control-plane", "controller-manager")
		Expect(controllerPod).ToNot(BeNil())

		clientset := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
		req := clientset.CoreV1().Pods(controllerPod.Namespace).GetLogs(controllerPod.Name, &corev1.PodLogOptions{SinceTime: &metav1.Time{Time: startTime}, Container: "manager"})
		logs, err := req.Stream(context.Background())
		Expect(err).ToNot(HaveOccurred())
		defer logs.Close()

		logReader := bufio.NewReader(logs)
		lines := 0
		for {
			line, err := logReader.ReadString('\n')
			if err == io.EOF {
				break
			}
			GinkgoWriter.Printf("log: %s", line)
			Expect(err).ToNot(HaveOccurred())
			Expect(line).ToNot(ContainSubstring("ERROR"))
			lines++
		}
		Expect(lines).To(BeNumerically(">", 0))
	})

	When("the config is good", func() {
		BeforeEach(func() {
			artifact = &api.AIDeployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AIDeployment",
					APIVersion: api.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "localai-",
				},
				Spec: api.AIDeploymentSpec{
					Engine: api.AIEngine{
						Name: "localai",
					},
					Endpoint: []api.Endpoint{{
						Domain: "foo.127.0.0.1.nip.io",
					}},
					Deployment: api.Deployment{
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("70Mi"),
							},
						},
					},
					Models: []api.AIModel{
						{
							Custom: &api.AIModelCustom{
								Format: "ggml",
								Name:   "gpt-4",
								Url:    "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en-q5_1.bin?download=true",
							},
						},
					},
				},
			}
		})

		It("starts the API", func() {
			By("starting the workload with the associated label")
			Eventually(func(g Gomega) bool {
				deploymentPod := &corev1.Pod{}
				if !getObjectWithLabel(pods, deploymentPod, resources.DefaultAnnotation, artifactName) {
					return false
				}

				c := deploymentPod.Spec.Containers[0]
				g.Expect(c.Name).To(HavePrefix("localai"))
				g.Expect(c.StartupProbe).ToNot(BeNil())
				g.Expect(c.StartupProbe.InitialDelaySeconds).To(Equal(int32(60)))
				g.Expect(c.StartupProbe.PeriodSeconds).To(Equal(int32(30)))
				g.Expect(c.StartupProbe.FailureThreshold).To(Equal(int32(120)))

				g.Expect(c.ReadinessProbe).ToNot(BeNil())
				g.Expect(c.ReadinessProbe.FailureThreshold).To(Equal(int32(3)))

				g.Expect(c.LivenessProbe).ToNot(BeNil())
				g.Expect(c.LivenessProbe.PeriodSeconds).To(Equal(int32(30)))
				g.Expect(c.LivenessProbe.TimeoutSeconds).To(Equal(int32(15)))
				g.Expect(c.LivenessProbe.FailureThreshold).To(Equal(int32(10)))

				g.Expect(c.Resources.Requests["memory"]).To(Equal(resource.MustParse("70Mi")))
				g.Expect(c.Resources.Requests["cpu"]).To(Equal(resource.MustParse("2")))
				mems := c.Resources.Limits["memory"]
				g.Expect(mems.Cmp(c.Resources.Requests["memory"])).To(BeNumerically(">", 0))
				cpus := c.Resources.Limits["cpu"]
				g.Expect(cpus.Cmp(c.Resources.Requests["cpu"])).To(BeNumerically(">=", 0))

				g.Expect(c.Resources.Requests["nvidia.com/gpu"]).To(Equal(resource.Quantity{}))

				return true
			}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())

			Eventually(func(g Gomega) bool {
				deploymentPod := &corev1.Pod{}
				if getObjectWithLabel(pods, deploymentPod, resources.DefaultAnnotation, artifactName) {
					return deploymentPod.Status.Phase == corev1.PodRunning
				}

				return false
			}).WithPolling(5 * time.Second).WithTimeout(time.Hour).Should(BeTrue())

			Eventually(func(g Gomega) bool {
				deploymentPod := &corev1.Pod{}
				if !getObjectWithLabel(pods, deploymentPod, resources.DefaultAnnotation, artifactName) {
					return false
				}

				for _, cond := range deploymentPod.Status.Conditions {
					switch ctype := cond.Type; ctype {
					case "Ready":
						return cond.Status == "True"
					default:
					}
				}

				return false
			}).WithPolling(5 * time.Second).WithTimeout(time.Hour).Should(BeTrue())

			Eventually(func(g Gomega) bool {
				p := &corev1.Service{}
				return getObjectWithAnnotation(svc, p, resources.DefaultAnnotation, artifactName)
			}).WithTimeout(time.Minute).Should(BeTrue())

			Eventually(func(g Gomega) bool {
				p := &networkv1.Ingress{}
				return getObjectWithAnnotation(ingr, p, resources.DefaultAnnotation, artifactName)
			}).WithTimeout(time.Minute).Should(BeTrue())

			Eventually(func(g Gomega) string {
				url := "http://foo.127.0.0.1.nip.io:8080/v1/models"
				req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
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
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Error reading request body:", err)
					return ""
				}
				return string(body)
			}).WithPolling(5 * time.Second).WithTimeout(time.Hour).Should(ContainSubstring("gpt-4"))
		})
	})

	When("we override the probe values", func() {
		initialDelay := int32(66)

		BeforeEach(func() {
			artifact = &api.AIDeployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AIDeployment",
					APIVersion: api.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "localai-",
				},
				Spec: api.AIDeploymentSpec{
					Engine: api.AIEngine{
						Name: "localai",
					},
					Endpoint: []api.Endpoint{{
						Domain: "foo.127.0.0.1.nip.io",
					},
					},
					Models: []api.AIModel{
						{
							Custom: &api.AIModelCustom{
								Format: "ggml",
								Name:   "gpt-4",
								Url:    "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en-q5_1.bin?download=true",
							},
						},
					},
					Deployment: api.Deployment{
						StartupProbe: &api.Probe{
							InitialDelaySeconds: &initialDelay,
							PeriodSeconds:       33,
							TimeoutSeconds:      12,
							FailureThreshold:    13,
						},
						ReadinessProbe: &api.Probe{
							SuccessThreshold: 14,
						},
						LivenessProbe: &api.Probe{
							PeriodSeconds: 21,
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("70Mi"),
							},
						},
					},
				},
			}
		})

		It("starts the API with the merged probe values", func() {
			By("starting the workload with the associated label")
			Eventually(func(g Gomega) bool {
				deploymentPod := &corev1.Pod{}
				if !getObjectWithLabel(pods, deploymentPod, resources.DefaultAnnotation, artifactName) {
					return false
				}

				c := deploymentPod.Spec.Containers[0]
				g.Expect(c.Name).To(HavePrefix("localai"))
				g.Expect(c.StartupProbe).ToNot(BeNil())
				g.Expect(c.StartupProbe.InitialDelaySeconds).To(Equal(int32(66)))
				g.Expect(c.StartupProbe.PeriodSeconds).To(Equal(int32(33)))
				g.Expect(c.StartupProbe.TimeoutSeconds).To(Equal(int32(12)))
				g.Expect(c.StartupProbe.FailureThreshold).To(Equal(int32(13)))

				g.Expect(c.ReadinessProbe).ToNot(BeNil())
				g.Expect(c.ReadinessProbe.FailureThreshold).To(Equal(int32(3)))
				g.Expect(c.ReadinessProbe.SuccessThreshold).To(Equal(int32(14)))

				g.Expect(c.LivenessProbe).ToNot(BeNil())
				g.Expect(c.LivenessProbe.PeriodSeconds).To(Equal(int32(21)))
				g.Expect(c.LivenessProbe.TimeoutSeconds).To(Equal(int32(15)))
				g.Expect(c.LivenessProbe.FailureThreshold).To(Equal(int32(10)))

				return true
			}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
		})
	})

	When("We specify a GPU", func() {
		BeforeEach(func() {
			artifact = &api.AIDeployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AIDeployment",
					APIVersion: api.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "localai-",
				},
				Spec: api.AIDeploymentSpec{
					Engine: api.AIEngine{
						Name: "localai",
					},
					Endpoint: []api.Endpoint{{
						Domain: "foo.127.0.0.1.nip.io",
					},
					},
					Models: []api.AIModel{
						{
							Custom: &api.AIModelCustom{
								Format: "ggml",
								Name:   "gpt-4",
								Url:    "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en-q5_1.bin?download=true",
							},
						},
					},
					Deployment: api.Deployment{
						NodeSelector: map[string]string{
							"nvidia.com/gpu.memory": "81920",
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("200Gi"),
							},
						},
					},
				},
			}
		})

		It("Creates a deployment with the correct GPU count", func() {
			By("creating the workload with the associated label")
			Eventually(func(g Gomega) bool {
				deploymentPod := &corev1.Pod{}
				if !getObjectWithLabel(pods, deploymentPod, resources.DefaultAnnotation, artifactName) {
					return false
				}

				c := deploymentPod.Spec.Containers[0]
				g.Expect(c.Name).To(HavePrefix("localai"))
				g.Expect(c.Resources.Requests["memory"]).To(Equal(resource.MustParse("200Gi")))
				g.Expect(c.Resources.Requests["cpu"]).To(Equal(resource.MustParse("2")))
				g.Expect(c.Resources.Requests["nvidia.com/gpu"]).To(Equal(resource.MustParse("3")))
				g.Expect(c.Resources.Limits["nvidia.com/gpu"]).To(Equal(resource.MustParse("3")))

				return true
			}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
		})
	})
})
