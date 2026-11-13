package usecase

import "github.com/flight/internal/src/model"

func buildSearchCriteria(in model.SearchInput) model.SearchCriteria {
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

	return criteria
}
