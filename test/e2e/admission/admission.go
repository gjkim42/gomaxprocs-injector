package admission

import (
	"context"
	"fmt"

	"github.com/gjkim42/gomaxprocs-injector/test/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = ginkgo.Describe("Admission controller", func() {
	f := framework.NewDefaultFramework("admission")

	var ns string
	var client kubernetes.Interface

	ginkgo.BeforeEach(func() {
		ns = f.Namespace.Name
		client = f.ClientSet
	})

	ginkgo.AfterEach(func() {
	})

	ginkgo.It("should apply appropriate GOMAXPROCS env to pods", func() {
		ginkgo.By("creating a pod")
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{
						Name:  "init-0",
						Image: "busybox",
						Command: []string{
							"sh",
							"-c",
							"exit 0",
						},
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("100m"),
							},
						},
					},
				},
				Containers: []v1.Container{
					{
						Name:  "container-0",
						Image: "nginx",
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("2"),
							},
						},
					},
					{
						Name:  "container-1",
						Image: "nginx",
						Env: []v1.EnvVar{
							{
								Name:  "GOMAXPROCS",
								Value: "1",
							},
						},
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("2"),
							},
						},
					},
					{
						Name:  "container-2",
						Image: "nginx",
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("2100m"),
							},
						},
					},
					{
						Name:  "container-3",
						Image: "nginx",
					},
					{
						Name:  "container-4",
						Image: "nginx",
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("2100m"),
							},
						},
					},
				},
			},
		}
		_, err := client.CoreV1().Pods(ns).Create(context.TODO(), pod, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("getting the pod")
		pod, err = client.CoreV1().Pods(ns).Get(context.TODO(), pod.Name, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("checking if an init container has appropriate GOMAXPROCS env")
		err = checkGOMAXPROCSEnvSetForContainer(&pod.Spec.InitContainers[0], "1")

		ginkgo.By("checking if the container with integer cpu limit has appropriate GOMAXPROCS env")
		err = checkGOMAXPROCSEnvSetForContainer(&pod.Spec.Containers[0], "2")

		ginkgo.By("checking if the container that already has GOMAXPROCS env has not been modified")
		err = checkGOMAXPROCSEnvSetForContainer(&pod.Spec.Containers[1], "1")

		ginkgo.By("checking if the container with fractional cpu limit has appropriate GOMAXPROCS env")
		err = checkGOMAXPROCSEnvSetForContainer(&pod.Spec.Containers[2], "2")

		ginkgo.By("checking if a container without cpu limit does not have the GOMAXPROCS env")
		err = checkGOMAXPROCSEnvNotSetForContainer(&pod.Spec.Containers[3])
	})

	ginkgo.It("should allow pod that does not need any patch", func() {
		ginkgo.By("creating a pod successfully")
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{
						Name:  "init-0",
						Image: "busybox",
						Command: []string{
							"sh",
							"-c",
							"exit 0",
						},
					},
				},
				Containers: []v1.Container{
					{
						Name:  "container-0",
						Image: "nginx",
						Env: []v1.EnvVar{
							{
								Name:  "GOMAXPROCS",
								Value: "2",
							},
						},
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("2"),
							},
						},
					},
					{
						Name:  "container-1",
						Image: "nginx",
					},
				},
			},
		}
		_, err := client.CoreV1().Pods(ns).Create(context.TODO(), pod, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("should allow pod that does not need any patch", func() {
		ginkgo.By("creating a pod successfully")
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{
						Name:  "init-0",
						Image: "busybox",
						Command: []string{
							"sh",
							"-c",
							"exit 0",
						},
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("100m"),
							},
						},
					},
				},
				Containers: []v1.Container{
					{
						Name:  "container-0",
						Image: "nginx",
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("2"),
							},
						},
					},
				},
			},
		}
		_, err := client.CoreV1().Pods(ns).Create(context.TODO(), pod, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("getting the pod")
		pod, err = client.CoreV1().Pods(ns).Get(context.TODO(), pod.Name, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("ensuring that the pod has not been modified")
		for _, container := range pod.Spec.InitContainers {
			err = checkGOMAXPROCSEnvNotSetForContainer(&container)
		}
		for _, container := range pod.Spec.Containers {
			err = checkGOMAXPROCSEnvNotSetForContainer(&container)
		}
	})
})

func checkGOMAXPROCSEnvNotSetForContainer(container *v1.Container) error {
	for _, env := range container.Env {
		if env.Name == "GOMAXPROCS" {
			return fmt.Errorf("GOMAXPROCS env is set, the value is %s", env.Value)
		}
	}

	return nil
}

func checkGOMAXPROCSEnvSetForContainer(container *v1.Container, gomaxprocs string) error {
	for _, env := range container.Env {
		if env.Name == "GOMAXPROCS" {
			if env.Value != gomaxprocs {
				return fmt.Errorf("GOMAXPROCS env is set, but the value is not %s, got %s", gomaxprocs, env.Value)
			}
			return nil
		}
	}

	return fmt.Errorf("GOMAXPROCS env is not set, expected %s", gomaxprocs)
}
