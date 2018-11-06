package util

import (
	"time"
	"github.com/cenkalti/backoff"
	"github.com/onsi/gomega"
)

func AssertRetry(cb func() error, maxTime time.Duration) {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime

	err := backoff.Retry(cb, b)
	gomega.Expect(err).To(gomega.BeNil())
}
