package delivery

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flight/internal/src/model"
)

func parseFilters(query url.Values, param *model.SearchInput) error {
	var err error

	param.MinPrice, err = parseOptionalFloat(query.Get("min_price"), "min_price")
	if err != nil {
		return err
	}

	param.MaxPrice, err = parseOptionalFloat(query.Get("max_price"), "max_price")
	if err != nil {
		return err
	}

	param.MaxStops, err = parseOptionalInt(query.Get("max_stops"), "max_stops")
	if err != nil {
		return err
	}

	param.MinDurationMinutes, err = parseOptionalInt64(query.Get("min_duration"), "min_duration")
	if err != nil {
		return err
	}

	param.MaxDurationMinutes, err = parseOptionalInt64(query.Get("max_duration"), "max_duration")
	if err != nil {
		return err
	}

	param.DepartureFrom, err = parseClock(query.Get("departure_from"), "departure_from")
	if err != nil {
		return err
	}

	param.DepartureTo, err = parseClock(query.Get("departure_to"), "departure_to")
	if err != nil {
		return err
	}

	param.ArrivalFrom, err = parseClock(query.Get("arrival_from"), "arrival_from")
	if err != nil {
		return err
	}

	param.ArrivalTo, err = parseClock(query.Get("arrival_to"), "arrival_to")
	if err != nil {
		return err
	}

	for _, value := range query["airlines"] {
		for _, airline := range strings.Split(value, ",") {
			airline = strings.TrimSpace(airline)
			if airline != "" {
				param.Airlines = append(param.Airlines, airline)
			}
		}
	}

	return nil
}

func parseOptionalFloat(value, name string) (*float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("%w: invalid %s", model.ErrInvalidInput, name)
	}

	return &parsed, nil
}

func parseOptionalInt(value, name string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("%w: invalid %s", model.ErrInvalidInput, name)
	}

	return &parsed, nil
}

func parseOptionalInt64(value, name string) (*int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("%w: invalid %s", model.ErrInvalidInput, name)
	}

	return &parsed, nil
}

func parseClock(value, name string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	if _, err := time.Parse("15:04", value); err != nil {
		return "", fmt.Errorf("%w: invalid %s, expected HH:MM", model.ErrInvalidInput, name)
	}

	return value, nil
}

func validateSearchInput(param model.SearchInput) error {
	if param.DepartureDate.IsZero() {
		return fmt.Errorf("empty departure_date: %+w", model.ErrInvalidInput)
	}

	if param.Origin == "" {
		return fmt.Errorf("empty origin: %+w", model.ErrInvalidInput)
	}

	if param.Destination == "" {
		return fmt.Errorf("empty destination: %+w", model.ErrInvalidInput)
	}

	if param.Passengers <= 0 {
		return fmt.Errorf("invalid passenger: %+w", model.ErrInvalidInput)
	}

	if param.MinPrice != nil && param.MaxPrice != nil && *param.MinPrice > *param.MaxPrice {
		return fmt.Errorf("%w: min_price cannot exceed max_price", model.ErrInvalidInput)
	}

	if param.MinDurationMinutes != nil && param.MaxDurationMinutes != nil &&
		*param.MinDurationMinutes > *param.MaxDurationMinutes {
		return fmt.Errorf("%w: min_duration cannot exceed max_duration", model.ErrInvalidInput)
	}

	return nil
}
