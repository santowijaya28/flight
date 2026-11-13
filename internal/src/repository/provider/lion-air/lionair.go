package lionair

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/filter"
)

func (l *lr) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	if l.dataPath == "" {
		return res, fmt.Errorf("empty path for lion-air provider")
	}

	data, err := os.ReadFile(l.dataPath)
	if err != nil {
		return res, fmt.Errorf("error opening file: %w", err)
	}

	var resp LionAirSearchResp
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return res, fmt.Errorf("error unmarshal data: %w", err)
	}

	if !resp.Success {
		return res, fmt.Errorf("failed when getting data from lion-air")
	}

	return parseResponseData(resp, in)
}

func parseResponseData(in LionAirSearchResp, criteria model.SearchInput) (res model.SearchResult, err error) {
	res = model.SearchResult{
		Flights: make([]model.Flight, 0, len(in.Data.AvailableFlights)),
	}

	for _, f := range in.Data.AvailableFlights {
		cabinClass := normalizeCabinClass(f.Pricing.FareType)

		departureDatetime, departureTS, err := parseLionAirTime(
			f.Schedule.Departure,
			f.Schedule.DepartureTimezone,
		)
		if err != nil {
			return res, err
		}

		arrivalDatetime, arrivalTS, err := parseLionAirTime(
			f.Schedule.Arrival,
			f.Schedule.ArrivalTimezone,
		)
		if err != nil {
			return res, err
		}

		if err := model.ValidateArrivalAfterDeparture(departureTS, arrivalTS); err != nil {
			return res, err
		}

		departureDatetime, departureTimezone := model.FormatAirportDateTime(departureDatetime, f.Route.From.Code)
		arrivalDatetime, arrivalTimezone := model.FormatAirportDateTime(arrivalDatetime, f.Route.To.Code)

		totalMinutes := (arrivalTS - departureTS) / 60
		stops := f.StopCount
		if stops == 0 {
			stops = len(f.Layovers)
		}

		amenities := make([]string, 0, 2)
		if f.Services.WifiAvailable {
			amenities = append(amenities, "wifi")
		}
		if f.Services.MealsIncluded {
			amenities = append(amenities, "meal")
		}

		flight := model.Flight{
			ID:       fmt.Sprintf("%s_%s", f.ID, model.ProviderLionAir),
			Provider: model.ProviderLionAir,
			Airline: model.AirLine{
				Name: f.Carrier.Name,
				Code: f.Carrier.IATA,
			},
			FlightNumber: f.ID,
			Departure: model.Departure{
				Airport:   strings.ToUpper(f.Route.From.Code),
				City:      f.Route.From.City,
				Timezone:  departureTimezone,
				DateTime:  departureDatetime,
				Timestamp: departureTS,
			},
			Arrival: model.Arrival{
				Airport:   strings.ToUpper(f.Route.To.Code),
				City:      f.Route.To.City,
				Timezone:  arrivalTimezone,
				DateTime:  arrivalDatetime,
				Timestamp: arrivalTS,
			},
			Duration: model.Duration{
				TotalMinutes: totalMinutes,
				Formatted:    formatDuration(totalMinutes),
			},
			Stops: stops,
			Price: model.Price{
				Amount:    f.Pricing.Total,
				Currency:  f.Pricing.Currency,
				Formatted: model.FormatIDR(f.Pricing.Total),
			},
			AvailableSeats: int64(f.SeatsLeft),
			CabinClass:     cabinClass,
			Aircraft:       model.OptionalAircraft(f.PlaneType),
			Amenities:      model.NormalizeAmenities(amenities),
			Baggage: model.NormalizeBaggage(model.Baggage{
				CarryOn: f.Services.BaggageAllowance.Cabin,
				Checked: f.Services.BaggageAllowance.Hold,
			}),
		}

		if filter.MatchesFlight(flight, criteria) {
			res.Flights = append(res.Flights, flight)
		}
	}

	return res, nil
}

func parseLionAirTime(value, timezone string) (string, int64, error) {
	if timezone == "" {
		return "", 0, fmt.Errorf("empty timezone for time %q", value)
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return "", 0, fmt.Errorf("load timezone %q: %w", timezone, err)
	}

	t, err := time.ParseInLocation(lionAirTimeLayout, value, location)
	if err != nil {
		return "", 0, fmt.Errorf("parse time %q: %w", value, err)
	}

	return t.Format(time.RFC3339), t.Unix(), nil
}

func normalizeCabinClass(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", " ")
	if value == "" {
		return "unknown"
	}
	return value
}

func formatDuration(minutes int64) string {
	return fmt.Sprintf("%dh %dm", minutes/60, minutes%60)
}
