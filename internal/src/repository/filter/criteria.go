package filter

import (
	"strings"
	"time"

	"github.com/flight/internal/src/model"
)

func MatchesFlight(flight model.Flight, in model.SearchInput) bool {
	return matchesText(flight.Departure.Airport, in.Origin) &&
		matchesText(flight.Arrival.Airport, in.Destination) &&
		matchesText(flight.CabinClass, in.CabinClass) &&
		(in.Passengers <= 0 || flight.AvailableSeats >= int64(in.Passengers)) &&
		matchesDate(flight.Departure.DateTime, in) &&
		matchesPrice(flight.Price.Amount, in) &&
		matchesStops(flight.Stops, in.MaxStops) &&
		matchesAirline(flight.Airline, in.Airlines) &&
		matchesDuration(flight.Duration.TotalMinutes, in) &&
		matchesTimeWindow(flight.Departure.DateTime, in.DepartureFrom, in.DepartureTo) &&
		matchesTimeWindow(flight.Arrival.DateTime, in.ArrivalFrom, in.ArrivalTo)
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

func matchesPrice(price float64, in model.SearchInput) bool {
	if in.MinPrice != nil && price < *in.MinPrice {
		return false
	}
	if in.MaxPrice != nil && price > *in.MaxPrice {
		return false
	}
	return true
}

func matchesStops(stops int, maxStops *int) bool {
	return maxStops == nil || stops <= *maxStops
}

func matchesAirline(airline model.AirLine, airlines []string) bool {
	if len(airlines) == 0 {
		return true
	}

	for _, value := range airlines {
		value = strings.TrimSpace(value)
		if strings.EqualFold(value, airline.Code) || strings.EqualFold(value, airline.Name) {
			return true
		}
	}

	return false
}

func matchesDuration(minutes int64, in model.SearchInput) bool {
	if in.MinDurationMinutes != nil && minutes < *in.MinDurationMinutes {
		return false
	}
	if in.MaxDurationMinutes != nil && minutes > *in.MaxDurationMinutes {
		return false
	}
	return true
}

func matchesTimeWindow(datetime, from, to string) bool {
	if from == "" && to == "" {
		return true
	}

	value, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return false
	}

	current := value.Hour()*60 + value.Minute()

	fromMinutes, fromSet := parseClock(from)
	toMinutes, toSet := parseClock(to)

	switch {
	case fromSet && toSet && fromMinutes <= toMinutes:
		return current >= fromMinutes && current <= toMinutes
	case fromSet && toSet:
		return current >= fromMinutes || current <= toMinutes
	case fromSet:
		return current >= fromMinutes
	case toSet:
		return current <= toMinutes
	default:
		return false
	}
}

func parseClock(value string) (int, bool) {
	if value == "" {
		return 0, false
	}

	t, err := time.Parse("15:04", value)
	if err != nil {
		return 0, false
	}

	return t.Hour()*60 + t.Minute(), true
}
