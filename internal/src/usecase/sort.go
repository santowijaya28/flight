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
	sortBestValueAsc
	sortBestValueDesc
)

const (
	bestValuePriceWeight    = 0.60
	bestValueDurationWeight = 0.30
	bestValueStopsWeight    = 0.10
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
	case "best_value", "best_value_asc":
		return sortBestValueAsc, nil
	case "best_value_desc":
		return sortBestValueDesc, nil
	default:
		return 0, fmt.Errorf("%w: invalid sort %q", model.ErrInvalidInput, value)
	}
}

func sortFlights(flights []model.Flight, options []sortOption) {
	var scores map[string]float64
	if requiresBestValue(options) {
		scores = bestValueScores(flights)
	}

	sort.SliceStable(flights, func(i, j int) bool {
		a := flights[i]
		b := flights[j]

		for _, option := range options {
			comparison := compareFlight(a, b, option, scores)
			if comparison == 0 {
				continue
			}

			return comparison < 0
		}

		return a.ID < b.ID
	})
}

func compareFlight(a, b model.Flight, option sortOption, scores map[string]float64) int {
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
	case sortBestValueAsc:
		return compareFloat(scores[a.ID], scores[b.ID])
	case sortBestValueDesc:
		return compareFloat(scores[b.ID], scores[a.ID])
	default:
		return 0
	}
}

func requiresBestValue(options []sortOption) bool {
	for _, option := range options {
		if option == sortBestValueAsc || option == sortBestValueDesc {
			return true
		}
	}

	return false
}

func bestValueScores(flights []model.Flight) map[string]float64 {
	if len(flights) == 0 {
		return nil
	}

	minPrice := flights[0].Price.Amount
	maxPrice := flights[0].Price.Amount

	minDuration := float64(flights[0].Duration.TotalMinutes)
	maxDuration := float64(flights[0].Duration.TotalMinutes)

	minStops := float64(flights[0].Stops)
	maxStops := float64(flights[0].Stops)

	for _, flight := range flights[1:] {
		price := flight.Price.Amount
		if price < minPrice {
			minPrice = price
		}
		if price > maxPrice {
			maxPrice = price
		}

		duration := float64(flight.Duration.TotalMinutes)
		if duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}

		stops := float64(flight.Stops)
		if stops < minStops {
			minStops = stops
		}
		if stops > maxStops {
			maxStops = stops
		}
	}

	scores := make(map[string]float64, len(flights))
	for _, flight := range flights {
		priceScore := normalizeBestValue(flight.Price.Amount, minPrice, maxPrice)
		durationScore := normalizeBestValue(float64(flight.Duration.TotalMinutes), minDuration, maxDuration)
		stopsScore := normalizeBestValue(float64(flight.Stops), minStops, maxStops)

		scores[flight.ID] = bestValuePriceWeight*priceScore +
			bestValueDurationWeight*durationScore +
			bestValueStopsWeight*stopsScore
	}

	return scores
}

func normalizeBestValue(value, minValue, maxValue float64) float64 {
	if maxValue <= minValue {
		return 0
	}

	return (value - minValue) / (maxValue - minValue)
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
