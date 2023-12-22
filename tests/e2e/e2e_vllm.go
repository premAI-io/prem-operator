package e2e_test

import (
	"bufio"
	"context"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("vllm test", func() {
	var artifactName string
	var deps, oper, sds, pods dynamic.ResourceInterface
	var scheme *runtime.Scheme
	var artifact *api.AIDeployment
	var startTime time.Time

	custModel := []api.AIModel{{
		Custom: &api.AIModelCustom{
			Format: "pickle",
			Name:   "phi-1-5",
			Url:    "https://huggingface.co/microsoft/phi-1_5/resolve/main/pytorch_model.bin?download=true",
		},
	}}

	JustBeforeEach(func() {
		startTime = time.Now()
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		sds = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "aideployments"}).Namespace("default")
		pods = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("default")
		oper = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("saas-operator-system")
		deps = k8s.Resource(schema.GroupVersionResource{Group: appsv1.GroupName, Version: appsv1.SchemeGroupVersion.Version, Resource: "deployments"}).Namespace("default")

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
					GenerateName: "vllm-",
				},
				Spec: api.AIDeploymentSpec{
					Engine: api.AIEngine{
						Name: "vllm",
					},
					Endpoint: []api.Endpoint{{
						Domain: "foo.127.0.0.1.nip.io",
					}},
					Deployment: api.Deployment{
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("3Gi"),
							},
						},
					},
					Models: custModel,
				},
			}
		})

		It("creates the deployment", func() {
			By("starting the workload with the associated label")
			Eventually(func(g Gomega) bool {
				deploymentPod := &corev1.Pod{}
				if !getObjectWithLabel(pods, deploymentPod, resources.DefaultAnnotation, artifactName) {
					return false
				}

				c := deploymentPod.Spec.Containers[0]
				g.Expect(c.Name).To(HavePrefix("phi-1-5"))
				g.Expect(c.StartupProbe).ToNot(BeNil())
				g.Expect(c.StartupProbe.InitialDelaySeconds).To(Equal(int32(3)))
				g.Expect(c.StartupProbe.PeriodSeconds).To(Equal(int32(1)))
				g.Expect(c.StartupProbe.FailureThreshold).To(Equal(int32(120)))

				g.Expect(c.ReadinessProbe).ToNot(BeNil())
				g.Expect(c.ReadinessProbe.FailureThreshold).To(Equal(int32(3)))

				g.Expect(c.LivenessProbe).ToNot(BeNil())
				g.Expect(c.LivenessProbe.PeriodSeconds).To(Equal(int32(30)))
				g.Expect(c.LivenessProbe.TimeoutSeconds).To(Equal(int32(15)))
				g.Expect(c.LivenessProbe.FailureThreshold).To(Equal(int32(10)))

				g.Expect(c.Resources.Requests["memory"]).To(Equal(resource.MustParse("3Gi")))
				g.Expect(c.Resources.Requests["cpu"]).To(Equal(resource.MustParse("2")))
				mems := c.Resources.Limits["memory"]
				g.Expect(mems.Cmp(c.Resources.Requests["memory"])).To(BeNumerically(">", 0))
				cpus := c.Resources.Limits["cpu"]
				g.Expect(cpus.Cmp(c.Resources.Requests["cpu"])).To(BeNumerically(">=", 0))

				return true
			}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
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
						GenerateName: "vllm-",
					},
					Spec: api.AIDeploymentSpec{
						Engine: api.AIEngine{
							Name: "vllm",
						},
						Endpoint: []api.Endpoint{{
							Domain: "foo.127.0.0.1.nip.io",
						},
						},
						Models: custModel,
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
					g.Expect(c.Name).To(HavePrefix("phi-1-5"))
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
						GenerateName: "vllm-",
					},
					Spec: api.AIDeploymentSpec{
						Engine: api.AIEngine{
							Name: "vllm",
						},
						Endpoint: []api.Endpoint{{
							Domain: "foo.127.0.0.1.nip.io",
						},
						},
						Models: custModel,
						Deployment: api.Deployment{
							Accelerator: &api.Accelerator{
								Interface: api.AcceleratorInterfaceCUDA,
								MinVersion: &api.Version{
									Major: 7,
								},
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
					deployment := &appsv1.Deployment{}
					if !getObjectWithName(deps, deployment, artifactName) {
						return false
					}

					nvidia := "nvidia"
					g.Expect(deployment.Spec.Template.Spec.RuntimeClassName).To(Equal(&nvidia))

					c := deployment.Spec.Template.Spec.Containers[0]
					g.Expect(c.Name).To(HavePrefix("phi-1-5"))
					g.Expect(c.Resources.Requests["memory"]).To(Equal(resource.MustParse("200Gi")))
					g.Expect(c.Resources.Requests["cpu"]).To(Equal(resource.MustParse("2")))
					g.Expect(c.Resources.Requests["nvidia.com/gpu"]).To(Equal(resource.MustParse("1")))
					g.Expect(c.Resources.Limits["nvidia.com/gpu"]).To(Equal(resource.MustParse("1")))

					return true
				}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
			})
		})
	})
})
