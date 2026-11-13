package provider

import (
	"context"
	"errors"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/pkg/resilience"
	"github.com/flight/pkg/simulation"
)

type resilientProvider struct {
	next   Provider
	policy resilience.Policy
}

func WithResilience(
	next Provider,
	timeout time.Duration,
	maxAttempts int,
	initialBackoff time.Duration,
	maxBackoff time.Duration,
) Provider {
	return &resilientProvider{
		next: next,
		policy: resilience.Policy{
			Timeout:        timeout,
			MaxAttempts:    maxAttempts,
			InitialBackoff: initialBackoff,
			MaxBackoff:     maxBackoff,
			Retryable:      isRetryable,
		},
	}
}

func (p *resilientProvider) Search(
	ctx context.Context,
	in model.SearchInput,
) (model.SearchResult, error) {
	return resilience.Execute(ctx, p.policy, func(ctx context.Context) (model.SearchResult, error) {
		return p.next.Search(ctx, in)
	})
}

func isRetryable(err error) bool {
	return errors.Is(err, simulation.ErrSimulatedUpstreamFailure) ||
		errors.Is(err, context.DeadlineExceeded)
}
