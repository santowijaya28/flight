package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	errTransient := errors.New("transient")
	errPermanent := errors.New("permanent")

	tests := []struct {
		name         string
		failures     int
		failureErr   error
		retryable    bool
		nilRetryable bool
		maxAttempts  int
		wantAttempts int
		wantErr      bool
	}{
		{
			name:         "success first attempt",
			maxAttempts:  3,
			retryable:    true,
			wantAttempts: 1,
		},
		{
			name:         "succeeds after retry",
			failures:     1,
			failureErr:   errTransient,
			retryable:    true,
			maxAttempts:  3,
			wantAttempts: 2,
		},
		{
			name:         "non retryable error",
			failures:     1,
			failureErr:   errPermanent,
			retryable:    false,
			maxAttempts:  3,
			wantAttempts: 1,
			wantErr:      true,
		},
		{
			name:         "nil retryable stops after first",
			failures:     1,
			failureErr:   errTransient,
			nilRetryable: true,
			maxAttempts:  3,
			wantAttempts: 1,
			wantErr:      true,
		},
		{
			name:         "exhausts attempts",
			failures:     99,
			failureErr:   errTransient,
			retryable:    true,
			maxAttempts:  3,
			wantAttempts: 3,
			wantErr:      true,
		},
		{
			name:         "zero max attempts is clamped to one",
			failures:     1,
			failureErr:   errTransient,
			retryable:    true,
			maxAttempts:  0,
			wantAttempts: 1,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0

			policy := Policy{
				MaxAttempts:    tt.maxAttempts,
				InitialBackoff: time.Millisecond,
				MaxBackoff:     time.Millisecond,
			}
			if !tt.nilRetryable {
				policy.Retryable = func(error) bool {
					return tt.retryable
				}
			}

			got, err := Execute(context.Background(), policy, func(context.Context) (string, error) {
				attempts++
				if attempts <= tt.failures {
					return "", tt.failureErr
				}
				return "ok", nil
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != "ok" {
					t.Fatalf("value = %q, want %q", got, "ok")
				}
			}

			if attempts != tt.wantAttempts {
				t.Fatalf("attempts = %d, want %d", attempts, tt.wantAttempts)
			}
		})
	}
}

func TestExecuteTimeout(t *testing.T) {
	_, err := Execute(context.Background(), Policy{
		Timeout:     10 * time.Millisecond,
		MaxAttempts: 1,
	}, func(ctx context.Context) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestExecuteCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0
	_, err := Execute(ctx, Policy{MaxAttempts: 3}, func(context.Context) (string, error) {
		attempts++
		return "ok", nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if attempts != 0 {
		t.Fatalf("attempts = %d, want 0", attempts)
	}
}

func TestExecuteBackoff(t *testing.T) {
	tests := []struct {
		name           string
		initialBackoff time.Duration
		maxBackoff     time.Duration
		failures       int
		maxAttempts    int
		wantMinElapsed time.Duration
	}{
		{
			name:           "negative backoff is clamped",
			initialBackoff: -time.Millisecond,
			maxBackoff:     0,
			failures:       1,
			maxAttempts:    2,
		},
		{
			name:           "max backoff below initial uses initial",
			initialBackoff: 5 * time.Millisecond,
			maxBackoff:     time.Millisecond,
			failures:       1,
			maxAttempts:    2,
			wantMinElapsed: 5 * time.Millisecond,
		},
		{
			name:           "backoff is capped",
			initialBackoff: 3 * time.Millisecond,
			maxBackoff:     3 * time.Millisecond,
			failures:       3,
			maxAttempts:    4,
			wantMinElapsed: 9 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0
			start := time.Now()

			got, err := Execute(context.Background(), Policy{
				MaxAttempts:    tt.maxAttempts,
				InitialBackoff: tt.initialBackoff,
				MaxBackoff:     tt.maxBackoff,
				Retryable:      func(error) bool { return true },
			}, func(context.Context) (string, error) {
				attempts++
				if attempts <= tt.failures {
					return "", errors.New("transient")
				}
				return "ok", nil
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != "ok" {
				t.Fatalf("value = %q, want %q", got, "ok")
			}
			if elapsed := time.Since(start); elapsed < tt.wantMinElapsed {
				t.Fatalf("elapsed = %v, want at least %v", elapsed, tt.wantMinElapsed)
			}
		})
	}
}

func TestExecuteCanceledDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	_, err := Execute(ctx, Policy{
		MaxAttempts:    3,
		InitialBackoff: time.Second,
		MaxBackoff:     time.Second,
		Retryable:      func(error) bool { return true },
	}, func(context.Context) (string, error) {
		attempts++
		cancel()
		return "", errors.New("transient")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestExecuteWithTimeoutPreCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	_, err := executeWithTimeout(ctx, time.Second, func(context.Context) (string, error) {
		called = true
		return "ok", nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if called {
		t.Fatal("operation should not be called")
	}
}
