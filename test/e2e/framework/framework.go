package framework

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Framework supports common operations used by e2e tests
type Framework struct {
	BaseName string

	Namespace *v1.Namespace
	ClientSet kubernetes.Interface
}

func NewDefaultFramework(baseName string) *Framework {
	f := &Framework{
		BaseName: baseName,
	}

	ginkgo.BeforeEach(func() {
		config, err := clientcmd.BuildConfigFromFlags("", TestContext.KubeConfig)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		clientSet, err := kubernetes.NewForConfig(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		f.ClientSet = clientSet

		ns, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: f.BaseName,
			},
		}, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		f.Namespace = ns
	})

	ginkgo.AfterEach(func() {
		err := f.ClientSet.CoreV1().Namespaces().Delete(context.TODO(), f.BaseName, metav1.DeleteOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		f.Namespace = nil
	})

	return f
}
