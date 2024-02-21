package e2e_test

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/aimodelmap"
	"github.com/premAI-io/saas-controller/controllers/constants"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AIModelMap", func() {
	var artifactName string
	var conf, aim dynamic.ResourceInterface
	var scheme *runtime.Scheme
	var artifact *api.AIModelMap
	var startTime time.Time

	typedClient := getTypedClient()

	JustBeforeEach(func() {
		startTime = time.Now()
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		aim = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "aimodelmaps"}).Namespace("default")
		conf = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "configmaps"}).Namespace("default")

		uArtifact := unstructured.Unstructured{}
		uArtifact.Object, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(artifact)
		resp, err := aim.Create(context.TODO(), &uArtifact, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred())
		artifactName = resp.GetName()
		GinkgoWriter.Printf("artifactName: %s\n", artifactName)
	})

	AfterEach(func() {
		err := aim.Delete(context.Background(), artifactName, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())

		checkLogs(startTime)
	})

	When("We have an AIModelMap with variants that have config", func() {
		BeforeEach(func() {
			artifact = &api.AIModelMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AIModelMap",
					APIVersion: api.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-aimodelmap-",
				},
				Spec: api.AIModelMapSpec{
					Localai: []api.AIModelVariant{
						{
							Variant: "variant1",
							AIModelSpec: api.AIModelSpec{
								Uri:              "s3://prem-ai/aimodels/variant1",
								EngineConfigFile: "engine config file content",
							},
						},
						{
							Variant: "variant2",
							AIModelSpec: api.AIModelSpec{
								Uri:              "s3://prem-ai/aimodels/variant2",
								EngineConfigFile: "engine config file content",
							},
						},
					},
					Vllm: []api.AIModelVariant{
						{
							Variant: "variant1",
							AIModelSpec: api.AIModelSpec{
								Uri:              "s3://prem-ai/aimodels/variant1",
								EngineConfigFile: "engine config file content",
							},
						},
					},
				},
			}
		})

		It("Creates the configmaps", func() {
			By("creating the AIModelMap")
			Eventually(func(g Gomega) bool {
				modelMap := &api.AIModelMap{}

				if err := typedClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: artifactName}, modelMap); err != nil {
					g.Expect(errors.IsNotFound(err)).To(BeTrue())

					return false
				}

				return true
			}, time.Minute, time.Second).Should(BeTrue())

			By("checking the configmaps")
			Eventually(func(g Gomega) bool {
				configMap := &corev1.ConfigMap{}
				if !getObjectWithLabel(conf, configMap, constants.AIModelMapDefaultLabelKey, artifactName) {
					return false
				}

				g.Expect(aimodelmap.GetEngineConfigFileData(configMap, api.AIEngineNameLocalai, "variant1")).To(Equal("engine config file content"))
				g.Expect(aimodelmap.GetEngineConfigFileData(configMap, api.AIEngineNameLocalai, "variant2")).To(Equal("engine config file content"))
				g.Expect(aimodelmap.GetEngineConfigFileData(configMap, api.AIEngineNameVLLM, "variant1")).To(Equal("engine config file content"))

				return true
			}, time.Minute, time.Second).Should(BeTrue())
		})
	})
})
