package sreq

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultMaxAttempts = 4
	defaultWaitTime    = 1 * time.Second
	defaultMaxWaitTime = 30 * time.Second
)

type (
	// Backoff specifies a policy for how long to wait between retries.
	// It is called after a failing request to determine the amount of time
	// that should pass before trying again.
	Backoff func(min time.Duration, max time.Duration, attemptNum int, resp *Response) time.Duration

	retry struct {
		maxAttempts int
		waitTime    time.Duration
		maxWaitTime time.Duration
		backoff     Backoff
		triggers    []func(resp *Response) bool
	}

	// RetryOption provides a convenient way to configure the retry policy.
	RetryOption func(r *retry)
)

// WithMaxAttempts specifies the max attempts of the retry policy, default is 4.
func WithMaxAttempts(maxAttempts int) RetryOption {
	return func(r *retry) {
		if maxAttempts > 1 {
			r.maxAttempts = maxAttempts
		}
	}
}

// WithWaitTime specifies the wait time to sleep before retrying request, default is 1s.
func WithWaitTime(waitTime time.Duration) RetryOption {
	return func(r *retry) {
		r.waitTime = waitTime
	}
}

// WithMaxWaitTime specifies the max wait time to sleep before retrying request, default is 30s.
func WithMaxWaitTime(maxWaitTime time.Duration) RetryOption {
	return func(r *retry) {
		r.maxWaitTime = maxWaitTime
	}
}

// WithBackoff specifies the backoff of the retry policy. It is called
// after a failing request to determine the amount of time
// that should pass before trying again.
func WithBackoff(backoff Backoff) RetryOption {
	return func(r *retry) {
		r.backoff = backoff
	}
}

// WithTriggers specifies the triggers for handling retries. It is called
// following each request with the response values returned by Client.
// If one of the triggers returns true, the Client will retry the request.
func WithTriggers(triggers ...func(resp *Response) bool) RetryOption {
	return func(r *retry) {
		r.triggers = triggers
	}
}

// DefaultBackoff provides a default callback for the default retry policy which
// will perform exponential backoff with jitter based on the attempt number and limited
// by the provided minimum and maximum durations.
// See: https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
func DefaultBackoff(min time.Duration, max time.Duration, attempt int, _ *Response) time.Duration {
	temp := math.Min(float64(max), float64(min)*math.Exp2(float64(attempt)))
	n := int(temp / 2)
	if n <= 0 {
		n = math.MaxInt32 // max int for arch 386
	}
	sleep := time.Duration(math.Abs(float64(n + rand.Intn(n))))
	if sleep < min {
		sleep = min
	}
	return sleep
}
