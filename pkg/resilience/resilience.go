package resilience

import (
	"context"
	"time"
)

type Operation[T any] func(context.Context) (T, error)

type Policy struct {
	Timeout        time.Duration
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Retryable      func(error) bool
}

func Execute[T any](ctx context.Context, policy Policy, operation Operation[T]) (T, error) {
	attempts := policy.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}

	backoff := policy.InitialBackoff
	if backoff < 0 {
		backoff = 0
	}

	maxBackoff := policy.MaxBackoff
	if maxBackoff < backoff {
		maxBackoff = backoff
	}

	var result T
	var err error

	for attempt := 1; attempt <= attempts; attempt++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return result, ctxErr
		}

		result, err = executeWithTimeout(ctx, policy.Timeout, operation)
		if err == nil {
			return result, nil
		}

		if attempt == attempts || policy.Retryable == nil || !policy.Retryable(err) {
			return result, err
		}

		if waitErr := wait(ctx, backoff); waitErr != nil {
			return result, waitErr
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	return result, err
}

func executeWithTimeout[T any](
	ctx context.Context,
	timeout time.Duration,
	operation Operation[T],
) (T, error) {
	if timeout <= 0 {
		return operation(ctx)
	}

	if err := ctx.Err(); err != nil {
		var zero T
		return zero, err
	}

	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type outcome struct {
		value T
		err   error
	}

	results := make(chan outcome, 1)
	go func() {
		value, err := operation(attemptCtx)
		results <- outcome{value: value, err: err}
	}()

	select {
	case <-attemptCtx.Done():
		var zero T
		return zero, attemptCtx.Err()
	case result := <-results:
		return result.value, result.err
	}
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
