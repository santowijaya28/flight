package simulation

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSimulate(t *testing.T) {
	tests := []struct {
		name           string
		minLatency     time.Duration
		maxLatency     time.Duration
		failureRate    int
		wantErr        error
		wantMinElapsed time.Duration
	}{
		{
			name: "no latency no failure",
		},
		{
			name:        "guaranteed failure",
			failureRate: 100,
			wantErr:     ErrSimulatedUpstreamFailure,
		},
		{
			name:        "failure rate above 100",
			failureRate: 150,
			wantErr:     ErrSimulatedUpstreamFailure,
		},
		{
			name:        "negative failure rate",
			failureRate: -10,
		},
		{
			name:           "minimum latency",
			minLatency:     2 * time.Millisecond,
			maxLatency:     2 * time.Millisecond,
			wantMinElapsed: 2 * time.Millisecond,
		},
		{
			name:           "max below min uses min",
			minLatency:     2 * time.Millisecond,
			maxLatency:     time.Millisecond,
			wantMinElapsed: 2 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := Simulate(
				context.Background(),
				"test-provider",
				tt.minLatency,
				tt.maxLatency,
				tt.failureRate,
			)
			elapsed := time.Since(start)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if elapsed < tt.wantMinElapsed {
				t.Fatalf("elapsed = %v, want at least %v", elapsed, tt.wantMinElapsed)
			}
		})
	}
}

func TestSimulateContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := Simulate(ctx, "test-provider", time.Second, time.Second, 0)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("cancellation took too long: %v", elapsed)
	}
}
