package common

import (
	"github.com/cenkalti/backoff"
	"time"
)

// RetryNormal retries a function with a normal backoff (15 sec)
func RetryNormal(cb func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Second * 15

	return backoff.Retry(cb, b)
}

// RetryShort retries a function with a short backoff (1 sec)
func RetryShort(cb func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Second * 1

	return backoff.Retry(cb, b)
}
