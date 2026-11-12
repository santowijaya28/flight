package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/provider"
)

func (r *repository) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	results := make(chan providerResult, len(r.flightProvider))

	t := time.Now()
	var wg sync.WaitGroup
	for name, p := range r.flightProvider {
		wg.Add(1)
		go func(name string, p provider.Provider) {
			defer wg.Done()

			providerRes, providerErr := p.Search(ctx, in)
			results <- providerResult{
				name: name,
				res:  providerRes,
				err:  providerErr,
			}
		}(name, p)
	}

	wg.Wait()
	close(results)

	res.Flights = make([]model.Flight, 0)
	successfulProviders := 0
	var errs []error

	for result := range results {
		res.Flights = append(res.Flights, result.res.Flights...)

		if result.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", result.name, result.err))
			continue
		}

		successfulProviders++
	}

	if len(errs) > 0 {
		err = errors.Join(errs...)
	}

	res.ProvidersFailed = len(errs)
	res.ProvidersQueried = len(r.flightProvider)
	res.ProvidersSucceeded = len(r.flightProvider) - len(errs)
	res.TotalResults = len(res.Flights)
	res.SearchTimeMs = int64(time.Since(t).Milliseconds())

	if len(r.flightProvider) > 0 && len(errs) == len(r.flightProvider) {
		err = errors.Join(model.ErrAllProviderDown, err)
		return res, err
	}

	return res, err
}
