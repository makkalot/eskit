package camconfig_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
	"os"
)

func TestCamConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CamConfig Suite")
}

var (
	camConfigEndpoint string
)

var _ = BeforeSuite(func() {
	camConfigEndpoint = os.Getenv("CAMCONFIG_ENDPOINT")
	if camConfigEndpoint == "" {
		camConfigEndpoint = "localhost:8081"
	}
})
