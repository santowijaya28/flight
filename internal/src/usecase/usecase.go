package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/flight/internal/src/model"
)

func (u *usecase) Search(ctx context.Context, in model.SearchInput) (response model.SearchResponse, err error) {
	start := time.Now()

	sortOptions, err := parseSort(in.Sort)
	if err != nil {
		return response, err
	}

	cacheKey, cacheKeyErr := buildCacheKey(in)
	if cacheKeyErr == nil {
		if cached, exists := u.cache.Get(cacheKey); exists {
			return buildSearchResponse(in, cached, true, time.Since(start).Milliseconds()), nil
		}
	}

	res, providerErr := u.repo.Search(ctx, in)
	if errors.Is(providerErr, model.ErrAllProviderDown) {
		return response, providerErr
	}

	sortFlights(res.Flights, sortOptions)
	// set cache hit as false
	response = buildSearchResponse(in, res, false, res.SearchTimeMs)

	// Partial provider failures are logged by the repository and returned as
	// successful partial results. Only cache fully successful searches.
	if cacheKeyErr == nil && providerErr == nil && len(res.Flights) > 0 {
		u.cache.Set(cacheKey, res)
	}

	return response, nil
}
