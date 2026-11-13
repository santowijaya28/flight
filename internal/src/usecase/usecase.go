package usecase

import (
	"context"
	"errors"

	"github.com/flight/internal/src/model"
)

func (u *usecase) Search(ctx context.Context, in model.SearchInput) (response model.SearchResponse, err error) {
	sortOptions, err := parseSort(in.Sort)
	if err != nil {
		return response, err
	}

	res, err := u.repo.Search(ctx, in)
	if errors.Is(err, model.ErrAllProviderDown) {
		return response, err
	}

	sortFlights(res.Flights, sortOptions)

	response = model.SearchResponse{
		SearchCriteria: buildSearchCriteria(in),
		Metadata: model.SearchMetadata{
			TotalResults:       res.TotalResults,
			ProvidersQueried:   res.ProvidersQueried,
			ProvidersSucceeded: res.ProvidersSucceeded,
			ProvidersFailed:    res.ProvidersFailed,
			SearchTimeMs:       res.SearchTimeMs,
		},
		Flights: res.Flights,
	}

	return
}
