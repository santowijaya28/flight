package model

import (
	"errors"
	"fmt"
)

var ErrArrivalBeforeDeparture = errors.New("arrival time must be after departure time")

func ValidateArrivalAfterDeparture(departureTS, arrivalTS int64) error {
	if arrivalTS <= departureTS {
		return fmt.Errorf(
			"%w: departure timestamp %d, arrival timestamp %d",
			ErrArrivalBeforeDeparture,
			departureTS,
			arrivalTS,
		)
	}

	return nil
}
