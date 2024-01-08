package e2e_test

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/constants"
	"github.com/premAI-io/saas-controller/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("generic test", func() {
	var artifactName string
	var deps, sds, svc, pods dynamic.ResourceInterface
	var scheme *runtime.Scheme
	var artifact *api.AIDeployment
	var startTime time.Time

	model := []api.AIModel{{
		ModelName: "buns",
	}}

	JustBeforeEach(func() {
		startTime = time.Now()
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		sds = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "aideployments"}).Namespace("default")
		svc = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "services"}).Namespace("default")
		pods = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("default")
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

		checkLogs(startTime)
	})

	When("the config is good", func() {
		BeforeEach(func() {
			artifact = &api.AIDeployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AIDeployment",
					APIVersion: api.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "generic-",
				},
				Spec: api.AIDeploymentSpec{
					Engine: api.AIEngine{
						Name: "generic",
					},
					Endpoint: []api.Endpoint{{
						Domain: "foo.127.0.0.1.nip.io",
						Port:   8081,
					}},
					Deployment: api.Deployment{
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("3Gi"),
							},
						},
						PodTemplate: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ImagePullSecrets: []corev1.LocalObjectReference{
									{
										Name: "regcred-dockerhub",
									},
								},
								Containers: []corev1.Container{
									{
										Name:  "bun",
										Image: "oven/bun",
									},
								},
							},
						},
					},
					Models: model,
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
				g.Expect(c.Name).To(HavePrefix("bun"))
				g.Expect(c.Image).To(Equal("oven/bun"))

				g.Expect(c.Resources.Requests["memory"]).To(Equal(resource.MustParse("3Gi")))
				g.Expect(c.Resources.Requests["cpu"]).To(Equal(resource.MustParse("2")))
				mems := c.Resources.Limits["memory"]
				g.Expect(mems.Cmp(c.Resources.Requests["memory"])).To(BeNumerically(">", 0))
				cpus := c.Resources.Limits["cpu"]
				g.Expect(cpus.Cmp(c.Resources.Requests["cpu"])).To(BeNumerically(">=", 0))

				return true
			}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())

			Eventually(func(g Gomega) bool {
				deploymentService := &corev1.Service{}
				if !getObjectWithAnnotation(svc, deploymentService, resources.DefaultAnnotation, artifactName) {
					return false
				}

				g.Expect(deploymentService.Spec.Ports[0].Port).To(Equal(int32(8081)))

				return true

			}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
		})

		When("We specify a GPU", func() {
			BeforeEach(func() {
				artifact = &api.AIDeployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "AIDeployment",
						APIVersion: api.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "generic-",
					},
					Spec: api.AIDeploymentSpec{
						Engine: api.AIEngine{
							Name: "generic",
						},
						Endpoint: []api.Endpoint{{
							Domain: "foo.127.0.0.1.nip.io",
							Port:   8000,
						},
						},
						Models: model,
						Deployment: api.Deployment{
							Accelerator: &api.Accelerator{
								Interface: api.AcceleratorInterfaceCUDA,
							},
							Resources: corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									"memory":                 resource.MustParse("200Gi"),
									constants.NvidiaGPULabel: resource.MustParse("3"),
								},
							},
							PodTemplate: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									ImagePullSecrets: []corev1.LocalObjectReference{
										{
											Name: "regcred-dockerhub",
										},
									},
									Containers: []corev1.Container{
										{
											Name:  "bun",
											Image: "oven/bun",
										},
									},
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
					g.Expect(c.Name).To(HavePrefix(constants.ContainerEngineName))
					g.Expect(c.Resources.Requests["memory"]).To(Equal(resource.MustParse("200Gi")))
					g.Expect(c.Resources.Requests["cpu"]).To(Equal(resource.MustParse("2")))
					g.Expect(c.Resources.Requests[constants.NvidiaGPULabel]).To(Equal(resource.MustParse("3")))
					g.Expect(c.Resources.Limits[constants.NvidiaGPULabel]).To(Equal(resource.MustParse("3")))

					return true
				}).WithPolling(5 * time.Second).WithTimeout(time.Minute).Should(BeTrue())
			})
		})
	})
})
