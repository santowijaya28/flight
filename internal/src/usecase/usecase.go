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

	criteria := model.SearchCriteria{
		Origin:      in.Origin,
		Destination: in.Destination,
		Passengers:  in.Passengers,
		CabinClass:  in.CabinClass,
	}
	if !in.DepartureDate.IsZero() {
		criteria.DepartureDate = in.DepartureDate.Format("2006-01-02")
	}
	if in.ReturnDate != nil {
		returnDate := in.ReturnDate.Format("2006-01-02")
		criteria.ReturnDate = &returnDate
	}

	response = model.SearchResponse{
		SearchCriteria: criteria,
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
