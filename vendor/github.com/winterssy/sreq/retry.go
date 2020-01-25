package sreq

import (
	"math"
	"math/rand"
	"time"
)

var (
	// DefaultBackoff specifies an exponential backoff with jitter
	// whose minWaitTime is 1s and maxWaitTime is 30s.
	DefaultBackoff = NewExponentialBackoff(1*time.Second, 30*time.Second, true)

	noRetry = &Retry{
		MaxAttempts: 1,
	}
)

type (
	// Backoff is the interface that specifies a backoff strategy for handling retries.
	Backoff interface {
		// WaitTime returns the wait time to sleep before retrying request.
		WaitTime(attemptNum int, resp *Response) time.Duration
	}

	// Retry specifies the retry policy for handling retries.
	Retry struct {
		// MaxAttempts specifies the max attempts of the retry policy, 1 means no retries.
		MaxAttempts int

		// Backoff specifies the backoff of the retry policy. It is called
		// after a failing request to determine the amount of time
		// that should pass before trying again.
		Backoff Backoff

		// Triggers specifies a group of triggers for handling retries. It is called
		// following each request with the response values returned by Client.
		// If the triggers not specified, default is the error of resp isn't nil.
		// Otherwise, the Client will only retry the request when the response meets one of the triggers.
		Triggers []func(resp *Response) bool
	}

	exponentialBackoff struct {
		minWaitTime time.Duration
		maxWaitTime time.Duration
		jitter      bool
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewExponentialBackoff provides a callback for the retry policy which
// will perform exponential backoff with jitter based on the attempt number and limited
// by the provided minimum and maximum durations.
// See: https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
func NewExponentialBackoff(minWaitTime, maxWaitTime time.Duration, jitter bool) Backoff {
	return &exponentialBackoff{
		minWaitTime: minWaitTime,
		maxWaitTime: maxWaitTime,
		jitter:      jitter,
	}
}

// WaitTime implements Backoff interface.
func (eb *exponentialBackoff) WaitTime(attemptNum int, _ *Response) time.Duration {
	sleep := math.Min(float64(eb.maxWaitTime), float64(eb.minWaitTime)*math.Exp2(float64(attemptNum)))
	if eb.jitter {
		n := int64(sleep / 2)
		sleep = float64(n + rand.Int63n(n))
	}
	return time.Duration(sleep)
}
