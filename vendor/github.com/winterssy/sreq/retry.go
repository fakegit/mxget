package sreq

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultMaxAttempts = 1
	defaultWaitTime    = 1 * time.Second
	defaultMaxWaitTime = 30 * time.Second
)

var (
	defaultRetry = &Retry{
		MaxAttempts: defaultMaxAttempts,
		WaitTime:    defaultWaitTime,
		MaxWaitTime: defaultMaxWaitTime,
		Backoff:     DefaultBackoff,
		Trigger:     DefaultTrigger,
	}
)

type (
	// Backoff specifies a policy for how long to wait between retries.
	// It is called after a failing request to determine the amount of time
	// that should pass before trying again.
	Backoff func(min time.Duration, max time.Duration, attemptNum int, resp *Response) time.Duration

	// Retry specifies the retry policy for handling retries.
	Retry struct {
		// MaxAttempts specifies the max attempts of the retry policy, default is 1, it means no retries.
		MaxAttempts int

		// WaitTime specifies the wait time to sleep before retrying request, default is 1s.
		WaitTime time.Duration

		// MaxWaitTime specifies the max wait time to sleep before retrying request, default is 30s.
		MaxWaitTime time.Duration

		// Backoff specifies the backoff of the retry policy. It is called
		// after a failing request to determine the amount of time
		// that should pass before trying again.
		Backoff Backoff

		// Trigger specifies the trigger for handling retries. It is called
		// following each request with the response values returned by Client.
		// If the trigger returns true, the Client will retry the request.
		Trigger func(resp *Response) bool
	}
)

// Merge merges r2 into r and returns the merged result. It keeps the non-zero values of r.
func (r *Retry) Merge(r2 *Retry) *Retry {
	if r == nil {
		return r2
	}

	if r.MaxAttempts == 0 {
		r.MaxAttempts = r2.MaxAttempts
	}
	if r.WaitTime == 0 {
		r.WaitTime = r2.WaitTime
	}
	if r.MaxWaitTime == 0 {
		r.MaxWaitTime = r2.MaxWaitTime
	}
	if r.Backoff == nil {
		r.Backoff = r2.Backoff
	}
	if r.Trigger == nil {
		r.Trigger = r2.Trigger
	}
	return r
}

// DefaultBackoff provides a callback for the default retry policy which
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

// DefaultTrigger provides a trigger for the default retry policy, it returns true if the error of resp isn't nil.
func DefaultTrigger(resp *Response) bool {
	return resp.err != nil
}
