package framework

import (
	"flag"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

type TestContextType struct {
	KubeConfig string
}

// TestContext should be used by all tests to get access to common context data
var TestContext TestContextType

func RegisterFlags(flags *flag.FlagSet) {
	flags.StringVar(&TestContext.KubeConfig, clientcmd.RecommendedConfigPathFlag, os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "kubeconfig path to use")
}
