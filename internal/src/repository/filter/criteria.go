package filter

import (
	"strings"

	"github.com/flight/internal/src/model"
)

func MatchesFlight(flight model.Flight, in model.SearchInput) bool {
	return matchesText(flight.Departure.Airport, in.Origin) &&
		matchesText(flight.Arrival.Airport, in.Destination) &&
		matchesText(flight.CabinClass, in.CabinClass) &&
		(in.Passengers <= 0 || flight.AvailableSeats >= int64(in.Passengers)) &&
		matchesDate(flight.Departure.DateTime, in)
}

func matchesText(value, filter string) bool {
	if filter == "" {
		return true
	}

	return strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(filter))
}

func matchesDate(datetime string, in model.SearchInput) bool {
	if in.DepartureDate.IsZero() {
		return true
	}

	return strings.HasPrefix(datetime, in.DepartureDate.Format("2006-01-02"))
}
