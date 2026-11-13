package garudaindonesia

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/filter"
)

func (g *gi) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	if g.datapath == "" {
		return res, fmt.Errorf("empty path for garuda-indonesia provider")
	}

	data, err := os.ReadFile(g.datapath)
	if err != nil {
		return res, fmt.Errorf("error oppening files %+w", err)
	}

	var resp GarudaIndonesiaSearchResp
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return res, fmt.Errorf("error unmarshal data %+w", err)
	}

	if resp.Status != "success" {
		return res, fmt.Errorf("failed when getting data from garida-indonesia")
	}

	return parseResponseData(resp, in)
}

func parseResponseData(in GarudaIndonesiaSearchResp, criteria model.SearchInput) (res model.SearchResult, err error) {

	for _, f := range in.Flights {
		departureAirport := strings.ToUpper(f.Departure.Airport)
		arrivalAirport := strings.ToUpper(f.Arrival.Airport)

		departureCity := f.Departure.City
		arrivalCity := f.Arrival.City

		departureTime := f.Departure.Time
		arrivalTime := f.Arrival.Time

		stops := f.Stops

		if len(f.Segments) > 0 {
			firstSegment := f.Segments[0]
			lastSegment := f.Segments[len(f.Segments)-1]

			departureAirport = strings.ToUpper(firstSegment.Departure.Airport)
			arrivalAirport = strings.ToUpper(lastSegment.Arrival.Airport)

			departureTime = firstSegment.Departure.Time
			arrivalTime = lastSegment.Arrival.Time

			stops = len(f.Segments) - 1

			if city := model.CityByAirportCode[departureAirport]; city != "" {
				departureCity = city
			}
			if city := model.CityByAirportCode[arrivalAirport]; city != "" {
				arrivalCity = city
			}
		}

		departureDatetime, departureTS, err := parseGarudaTime(departureTime)
		if err != nil {
			return res, err
		}

		arrivalDatetime, arrivalTS, err := parseGarudaTime(arrivalTime)
		if err != nil {
			return res, err
		}

		if err := model.ValidateArrivalAfterDeparture(departureTS, arrivalTS); err != nil {
			return res, err
		}

		departureDatetime, departureTimezone := model.FormatAirportDateTime(departureDatetime, departureAirport)
		arrivalDatetime, arrivalTimezone := model.FormatAirportDateTime(arrivalDatetime, arrivalAirport)

		totalMinutes := (arrivalTS - departureTS) / 60
		baggage := model.NormalizeBaggage(parseGarudaBaggage(f.Baggage))

		fl := model.Flight{
			ID:       fmt.Sprintf("%s_%s", f.FlightID, model.ProviderGarudaIndonesia),
			Provider: model.ProviderGarudaIndonesia,
			Airline: model.AirLine{
				Name: f.Airline,
				Code: f.AirlineCode,
			},
			FlightNumber: f.FlightID,
			Departure: model.Departure{
				Airport:   departureAirport,
				City:      departureCity,
				Timezone:  departureTimezone,
				DateTime:  departureDatetime,
				Timestamp: departureTS,
			},
			Arrival: model.Arrival{
				Airport:   arrivalAirport,
				City:      arrivalCity,
				Timezone:  arrivalTimezone,
				DateTime:  arrivalDatetime,
				Timestamp: arrivalTS,
			},
			Duration: model.Duration{
				TotalMinutes: totalMinutes,
				Formatted:    formatDuration(totalMinutes),
			},
			Stops:     stops,
			Amenities: model.NormalizeAmenities(f.Amenities),
			Price: model.Price{
				Amount:    f.Price.Amount,
				Currency:  f.Price.Currency,
				Formatted: model.FormatIDR(f.Price.Amount),
			},
			AvailableSeats: int64(f.AvailableSeats),
			CabinClass:     f.FareClass,
			Aircraft:       model.OptionalAircraft(f.Aircraft),
			Baggage:        baggage,
		}

		if filter.MatchesFlight(fl, criteria) {
			res.Flights = append(res.Flights, fl)
		}
	}
	return
}

func parseGarudaBaggage(baggage Baggage) model.Baggage {
	return model.Baggage{
		CarryOn: formatBaggageCount(baggage.CarryOn, "carry-on bag", "carry-on bags"),
		Checked: formatBaggageCount(baggage.Checked, "checked bag", "checked bags"),
	}
}

func formatBaggageCount(count int, singular, plural string) string {
	switch {
	case count <= 0:
		return ""
	case count == 1:
		return "1 " + singular
	default:
		return fmt.Sprintf("%d %s", count, plural)
	}
}

func parseGarudaTime(value string) (string, int64, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", 0, fmt.Errorf("parse time %q: %w", value, err)
	}

	return t.Format(time.RFC3339), t.Unix(), nil
}

func formatDuration(minutes int64) string {
	return fmt.Sprintf("%dh %dm", minutes/60, minutes%60)
}
