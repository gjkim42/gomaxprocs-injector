package e2e

import (
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"

	_ "github.com/gjkim42/gomaxprocs-injector/test/e2e/admission"
	"github.com/gjkim42/gomaxprocs-injector/test/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	framework.RegisterFlags(flag.CommandLine)
	flag.Parse()

	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func TestE2e(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "E2E Suite")
}
