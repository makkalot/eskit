package store

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"os"
	"github.com/makkalot/eskit/services/clients"
	"golang.org/x/net/context"
	"fmt"
)

func TestUsers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Store Suite")
}

var (
	storeEndpoint     string
	crudStoreEndpoint string
	consumerEndpoint  string
)

var _ = BeforeSuite(func() {

	storeEndpoint = os.Getenv("EVENTSTORE_ENDPOINT")
	if storeEndpoint == "" {
		Fail("EVENTSTORE_ENDPOINT is required")
	}

	consumerEndpoint = os.Getenv("CONSUMERSTORE_ENDPOINT")
	if consumerEndpoint == "" {
		Fail("CONSUMERSTORE_ENDPOINT is required")
	}

	crudStoreEndpoint = os.Getenv("CRUDSTORE_ENDPOINT")
	if crudStoreEndpoint == "" {
		Fail("CRUDSTORE_ENDPOINT is required")
	}

	waitForPreReqServices()
})

func waitForPreReqServices() {
	_, err := clients.NewStoreClientWithWait(context.Background(), storeEndpoint)
	if err != nil {
		Fail(fmt.Sprintf("couldn't connect to %s because of %v", storeEndpoint, err))
	}
}
