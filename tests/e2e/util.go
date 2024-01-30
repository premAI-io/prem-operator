package e2e_test

import (
	"bufio"
	"context"
	"io"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func getObjectWithLabel(resource dynamic.ResourceInterface, obj metav1.Object, labelName string, labelValue string) bool {
	list, err := resource.List(context.TODO(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	found := false

	for _, item := range list.Items {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, obj)

		Expect(err).ToNot(HaveOccurred())

		if v, ok := obj.GetLabels()[labelName]; ok && v == labelValue {
			found = ok
			break
		}
	}

	if !found {
		obj = nil
	}

	return found
}

func getObjectWithAnnotation(resource dynamic.ResourceInterface, obj metav1.Object, labelName string, labelValue string) bool {
	list, err := resource.List(context.TODO(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	found := false

	for _, item := range list.Items {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, obj)

		Expect(err).ToNot(HaveOccurred())

		if v, ok := obj.GetAnnotations()[labelName]; ok && v == labelValue {
			found = ok
			break
		}
	}

	if !found {
		obj = nil
	}

	return found
}

func getObjectWithName(resource dynamic.ResourceInterface, obj metav1.Object, name string) bool {
	list, err := resource.List(context.TODO(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	found := false

	for _, item := range list.Items {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, obj)

		Expect(err).ToNot(HaveOccurred())

		if obj.GetName() == name {
			found = true
			break
		}
	}

	if !found {
		obj = nil
	}

	return found
}

func checkLogs(startTime time.Time) {
	k8s := dynamic.NewForConfigOrDie(ctrl.GetConfigOrDie())
	oper := k8s.Resource(schema.GroupVersionResource{Group: corev1.GroupName, Version: corev1.SchemeGroupVersion.Version, Resource: "pods"}).Namespace("saas-operator-system")

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
		//Expect(line).ToNot(ContainSubstring("ERROR"))
		lines++
	}
	Expect(lines).To(BeNumerically(">", 0))
}
