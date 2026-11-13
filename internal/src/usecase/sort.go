package usecase

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flight/internal/src/model"
)

type sortOption int

const (
	sortPriceAsc sortOption = iota
	sortPriceDesc
	sortDurationAsc
	sortDurationDesc
	sortDepartureAsc
	sortDepartureDesc
	sortArrivalAsc
	sortArrivalDesc
)

func parseSort(value string) ([]sortOption, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []sortOption{sortPriceAsc}, nil
	}

	parts := strings.Split(value, ",")
	options := make([]sortOption, 0, len(parts))

	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}

		option, err := parseSortOption(part)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	if len(options) == 0 {
		return []sortOption{sortPriceAsc}, nil
	}

	return options, nil
}

func parseSortOption(value string) (sortOption, error) {
	switch value {
	case "price_asc":
		return sortPriceAsc, nil
	case "price_desc":
		return sortPriceDesc, nil
	case "duration_asc":
		return sortDurationAsc, nil
	case "duration_desc":
		return sortDurationDesc, nil
	case "departure_asc":
		return sortDepartureAsc, nil
	case "departure_desc":
		return sortDepartureDesc, nil
	case "arrival_asc":
		return sortArrivalAsc, nil
	case "arrival_desc":
		return sortArrivalDesc, nil
	default:
		return 0, fmt.Errorf("%w: invalid sort %q", model.ErrInvalidInput, value)
	}
}

func sortFlights(flights []model.Flight, options []sortOption) {
	sort.SliceStable(flights, func(i, j int) bool {
		a := flights[i]
		b := flights[j]

		for _, option := range options {
			comparison := compareFlight(a, b, option)
			if comparison == 0 {
				continue
			}

			return comparison < 0
		}

		return a.ID < b.ID
	})
}

func compareFlight(a, b model.Flight, option sortOption) int {
	switch option {
	case sortPriceAsc:
		return compareFloat(a.Price.Amount, b.Price.Amount)
	case sortPriceDesc:
		return compareFloat(b.Price.Amount, a.Price.Amount)
	case sortDurationAsc:
		return compareInt64(a.Duration.TotalMinutes, b.Duration.TotalMinutes)
	case sortDurationDesc:
		return compareInt64(b.Duration.TotalMinutes, a.Duration.TotalMinutes)
	case sortDepartureAsc:
		return compareInt64(a.Departure.Timestamp, b.Departure.Timestamp)
	case sortDepartureDesc:
		return compareInt64(b.Departure.Timestamp, a.Departure.Timestamp)
	case sortArrivalAsc:
		return compareInt64(a.Arrival.Timestamp, b.Arrival.Timestamp)
	case sortArrivalDesc:
		return compareInt64(b.Arrival.Timestamp, a.Arrival.Timestamp)
	default:
		return 0
	}
}

func compareInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
