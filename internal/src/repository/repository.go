package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/provider"
)

func (r *repository) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	results := make(chan providerResult, len(r.flightProvider))

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

	res.TotalResult = len(res.Flights)
	res.IsSuccess = successfulProviders > 0
	res.CacheHit = false

	if len(errs) > 0 {
		err = errors.Join(errs...)
	}

	return res, err
}
