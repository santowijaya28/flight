package airasia

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/filter"
	"github.com/flight/pkg/simulation"
)

func (a *airasia) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	if a.dataPath == "" {
		return res, fmt.Errorf("empty path for airasia provider")
	}

	if err := simulation.Simulate(ctx, "airasia", 50*time.Millisecond, 150*time.Millisecond, 10); err != nil {
		return res, err
	}

	data, err := os.ReadFile(a.dataPath)
	if err != nil {
		return res, fmt.Errorf("error oppening files %+w", err)
	}

	var resp AirAsiaSearchResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return res, fmt.Errorf("error unmarshal data %+w", err)
	}

	if resp.Status != "ok" {
		return res, fmt.Errorf("Status is not ok")
	}

	return parseResponseData(resp, in)
}

func parseResponseData(in AirAsiaSearchResponse, criteria model.SearchInput) (res model.SearchResult, err error) {

	for _, f := range in.Flights {
		departureDatetime, departureTS, err := parseAirAsiaTime(f.DepartTime)
		if err != nil {
			return res, err
		}

		arrivallDatetime, arrivalTS, err := parseAirAsiaTime(f.ArriveTime)
		if err != nil {
			return res, err
		}

		if err := model.ValidateArrivalAfterDeparture(departureTS, arrivalTS); err != nil {
			return res, err
		}

		departureDatetime, departureTimezone := model.FormatAirportDateTime(departureDatetime, f.FromAirport)
		arrivallDatetime, arrivalTimezone := model.FormatAirportDateTime(arrivallDatetime, f.ToAirport)

		totalMinutes := (arrivalTS - departureTS) / 60
		baggage := parseAirAsiaBaggage(f.BaggageNotes)

		fligt := model.Flight{
			ID:       fmt.Sprintf("%s_%s", f.FlightCode, f.Airline),
			Provider: model.ProviderAirAsia,
			Airline: model.AirLine{
				Name: f.Airline,
				Code: f.FlightCode[0:2],
			},
			FlightNumber: f.FlightCode,
			Departure: model.Departure{
				Airport:   f.FromAirport,
				City:      model.CityByAirportCode[f.FromAirport],
				Timezone:  departureTimezone,
				DateTime:  departureDatetime,
				Timestamp: departureTS,
			},
			Arrival: model.Arrival{
				Airport:   f.ToAirport,
				City:      model.CityByAirportCode[f.ToAirport],
				Timezone:  arrivalTimezone,
				DateTime:  arrivallDatetime,
				Timestamp: arrivalTS,
			},
			Duration: model.Duration{
				TotalMinutes: totalMinutes,
				Formatted:    formatDuration(totalMinutes),
			},
			Stops: len(f.Stops),
			Price: model.Price{
				Amount:    f.PriceIDR,
				Currency:  "IDR",
				Formatted: model.FormatIDR(f.PriceIDR),
			},
			AvailableSeats: f.Seats,
			CabinClass:     f.CabinClass,
			Amenities:      model.NormalizeAmenities(nil),
			Baggage: model.NormalizeBaggage(model.Baggage{
				CarryOn: baggage.CarryOn,
				Checked: baggage.Checked,
			}),
		}

		if filter.MatchesFlight(fligt, criteria) {
			res.Flights = append(res.Flights, fligt)
		}
	}

	return
}

func parseAirAsiaTime(value string) (string, int64, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", 0, fmt.Errorf("parse time %q: %w", value, err)
	}

	return t.Format(time.RFC3339), t.Unix(), nil
}

func formatDuration(minutes int64) string {
	return fmt.Sprintf("%dh %dm", minutes/60, minutes%60)
}

func parseAirAsiaBaggage(note string) Baggage {
	var b Baggage

	parts := strings.SplitN(note, ",", 2)
	if len(parts) > 0 {
		b.CarryOn = strings.TrimSpace(parts[0])
	}

	if len(parts) == 2 {
		checked := strings.TrimSpace(parts[1])
		checked = strings.TrimPrefix(checked, "checked bags ")
		b.Checked = capitalizeFirst(checked)
	}

	return b
}

func capitalizeFirst(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
