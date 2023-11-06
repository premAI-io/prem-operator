package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/premAI-io/saas-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("simple test", func() {
	var artifactName string
	var sds, pods dynamic.ResourceInterface
	var scheme *runtime.Scheme
	var artifactLabelSelector labels.Selector

	BeforeEach(func() {
		k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
		scheme = runtime.NewScheme()
		err := api.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())

		sds = k8s.Resource(schema.GroupVersionResource{Group: api.GroupVersion.Group, Version: api.GroupVersion.Version, Resource: "simpledeployments"}).Namespace("default")
		pods = k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("default")

		artifact := &api.SimpleDeployments{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SimpleDeployments",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "simple-",
			},
			Spec: api.SimpleDeploymentsSpec{
				MLEngine: "localai",
			},
		}

		uArtifact := unstructured.Unstructured{}
		uArtifact.Object, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(artifact)
		resp, err := sds.Create(context.TODO(), &uArtifact, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred())
		artifactName = resp.GetName()

		artifactLabelSelectorReq, err := labels.NewRequirement("ai-workload", selection.Equals, []string{artifactName})
		Expect(err).ToNot(HaveOccurred())
		artifactLabelSelector = labels.NewSelector().Add(*artifactLabelSelectorReq)
	})

	AfterEach(func() {
		err := sds.Delete(context.Background(), artifactName, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("works", func() {
		By("starting the deployment")
		Eventually(func(g Gomega) bool {
			w, err := pods.Watch(context.TODO(), metav1.ListOptions{LabelSelector: artifactLabelSelector.String()})
			g.Expect(err).ToNot(HaveOccurred())
			var stopped bool
			for !stopped {
				event, _ := <-w.ResultChan()
				fmt.Println(event)
				pod := event.Object.DeepCopyObject().(*unstructured.Unstructured)
				fmt.Println(pod)
				dat, err := json.Marshal(pod)
				g.Expect(err).ToNot(HaveOccurred())
				fmt.Println(string(dat))
				podDefini := &corev1.Pod{}
				err = json.Unmarshal(dat, podDefini)
				g.Expect(err).ToNot(HaveOccurred())
				labels := podDefini.GetLabels()
				g.Expect(labels).ToNot(BeNil())
				g.Expect(labels["ai-workload"]).ToNot(BeNil())
				return labels["ai-workload"] == artifactName
			}
			return false
		}).WithTimeout(time.Hour).Should(BeTrue())

		// By("checking the deployment status")
		// Eventually(func(g Gomega) bool {
		// 	w, err := sds.Watch(context.TODO(), metav1.ListOptions{})
		// 	Expect(err).ToNot(HaveOccurred())

		// 	var artifact api.SimpleDeployments
		// 	var stopped bool
		// 	for !stopped {
		// 		event, ok := <-w.ResultChan()
		// 		stopped = !ok
		// 		if event.Type == watch.Modified && event.Object.(*unstructured.Unstructured).GetName() == artifactName {
		// 			err := scheme.Convert(event.Object, &artifact, nil)
		// 			Expect(err).ToNot(HaveOccurred())
		// 			stopped = artifact.Status.Status != "Ready"
		// 			return stopped
		// 		}
		// 	}
		// 	return false
		// }).WithTimeout(time.Hour).Should(Succeed())
	})
})
