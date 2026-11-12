package usecase

import (
	"context"
	"errors"

	"github.com/flight/internal/src/model"
)

func (u *usecase) Search(ctx context.Context, in model.SearchInput) (response model.SearchResponse, err error) {
	res, err := u.repo.Search(ctx, in)
	if errors.Is(err, model.ErrAllProviderDown) {
		return response, err
	}

	response = model.SearchResponse{
		SearchCriteria: model.SearchCriteria{
			Origin:        in.Origin,
			Destination:   in.Destination,
			DepartureDate: in.DepartureDate,
			ReturnDate:    in.ReturnDate,
			Passengers:    in.Passengers,
			CabinClass:    in.CabinClass,
		},
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
