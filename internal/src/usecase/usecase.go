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

	res, err := u.repo.Search(ctx, in)
	if errors.Is(err, model.ErrAllProviderDown) {
		return response, err
	}

	sortFlights(res.Flights, sortOptions)
	response = buildSearchResponse(in, res, false, res.SearchTimeMs)

	if cacheKeyErr == nil && err == nil {
		u.cache.Set(cacheKey, res)
	}

	return
}
