package users_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"os"
)

func TestUsers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Users Suite")
}

var (
	userEndpoint string
)

var _ = BeforeSuite(func() {

	userEndpoint = os.Getenv("USERS_ENDPOINT")
	if userEndpoint == "" {
		Fail("USERS_ENDPOINT is required")
	}

})
