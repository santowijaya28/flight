package simulation

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

var ErrSimulatedUpstreamFailure = errors.New("simulated upstream failure")

var Simulate = func(
	ctx context.Context,
	name string,
	minLatency time.Duration,
	maxLatency time.Duration,
	failureRate int,
) error {
	return simulate(ctx, name, minLatency, maxLatency, failureRate)
}

func simulate(
	ctx context.Context,
	name string,
	minLatency time.Duration,
	maxLatency time.Duration,
	failureRate int,
) error {
	if maxLatency < minLatency {
		maxLatency = minLatency
	}

	delay := minLatency
	if maxLatency > minLatency {
		delay += time.Duration(rand.Int64N(int64(maxLatency-minLatency) + 1))
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}

	if failureRate <= 0 {
		return nil
	}
	if failureRate > 100 {
		failureRate = 100
	}

	if rand.IntN(100) < failureRate {
		return fmt.Errorf("%s: %w", name, ErrSimulatedUpstreamFailure)
	}

	return nil
}
