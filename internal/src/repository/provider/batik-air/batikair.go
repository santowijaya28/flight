package batikair

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/filter"
	"github.com/flight/internal/src/repository/provider"
)

func (b *batikAir) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	if b.dataPath == "" {
		return res, fmt.Errorf("empty path for batik-air provider")
	}

	if err := provider.Simulate(ctx, "batik-air", 200*time.Millisecond, 400*time.Millisecond, 0); err != nil {
		return res, err
	}

	data, err := os.ReadFile(b.dataPath)
	if err != nil {
		return res, fmt.Errorf("error oppening files %+w", err)
	}

	var resp BatikAirSearchResp
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return res, fmt.Errorf("error unmarshal data %+w", err)
	}

	if resp.Code != http.StatusOK {
		return res, fmt.Errorf("error when getting data from batikair: %s", resp.Message)
	}

	return parseResponseData(resp, in)
}

func parseResponseData(in BatikAirSearchResp, criteria model.SearchInput) (res model.SearchResult, err error) {
	for _, f := range in.Results {
		cabinClass := model.GetCabinClass(f.Fare.Class)

		departureDatetime, departureTS, err := parseBatikAirTime(f.DepartureDateTime)
		if err != nil {
			return res, err
		}

		arrivallDatetime, arrivalTS, err := parseBatikAirTime(f.ArrivalDateTime)
		if err != nil {
			return res, err
		}

		if err := model.ValidateArrivalAfterDeparture(departureTS, arrivalTS); err != nil {
			return res, err
		}

		departureDatetime, departureTimezone := model.FormatAirportDateTime(departureDatetime, f.Origin)
		arrivallDatetime, arrivalTimezone := model.FormatAirportDateTime(arrivallDatetime, f.Destination)

		totalMinutes := (arrivalTS - departureTS) / 60

		fl := model.Flight{
			ID:       fmt.Sprintf("%s_%s", f.FlightNumber, model.ProviderBatikAir),
			Provider: model.ProviderBatikAir,
			Airline: model.AirLine{
				Name: f.AirlineName,
				Code: f.AirlineIATA,
			},
			FlightNumber: f.FlightNumber,
			Departure: model.Departure{
				Airport:   f.Origin,
				City:      model.CityByAirportCode[f.Origin],
				Timezone:  departureTimezone,
				DateTime:  departureDatetime,
				Timestamp: departureTS,
			},
			Arrival: model.Arrival{
				Airport:   f.Destination,
				City:      model.CityByAirportCode[f.Destination],
				Timezone:  arrivalTimezone,
				DateTime:  arrivallDatetime,
				Timestamp: arrivalTS,
			},
			Duration: model.Duration{
				TotalMinutes: totalMinutes,
				Formatted:    formatDuration(totalMinutes),
			},
			Stops: len(f.Connections),
			Price: model.Price{
				Amount:    f.Fare.TotalPrice,
				Currency:  f.Fare.CurrencyCode,
				Formatted: model.FormatIDR(f.Fare.TotalPrice),
			},
			AvailableSeats: int64(f.SeatsAvailable),
			CabinClass:     cabinClass,
			Aircraft:       model.OptionalAircraft(f.AircraftModel),
			Amenities:      model.NormalizeAmenities(f.OnboardServices),
			Baggage:        model.NormalizeBaggage(parseBatikAirBaggage(f.BaggageInfo)),
		}
		if filter.MatchesFlight(fl, criteria) {
			res.Flights = append(res.Flights, fl)
		}
	}

	return
}

func parseBatikAirTime(value string) (string, int64, error) {
	value = strings.TrimSpace(value)

	for _, layout := range batikAirTimeLayouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t.Format(time.RFC3339), t.Unix(), nil
		}
	}

	return "", 0, fmt.Errorf("parse time %q: unsupported format", value)
}

func formatDuration(minutes int64) string {
	return fmt.Sprintf("%dh %dm", minutes/60, minutes%60)
}

func parseBatikAirBaggage(info string) model.Baggage {
	var baggage model.Baggage

	for _, part := range strings.Split(info, ",") {
		part = strings.TrimSpace(part)
		lower := strings.ToLower(part)

		switch {
		case strings.HasSuffix(lower, "cabin"):
			baggage.CarryOn = strings.TrimSpace(part[:len(part)-len("cabin")])
		case strings.HasSuffix(lower, "checked"):
			baggage.Checked = strings.TrimSpace(part[:len(part)-len("checked")])
		}
	}

	return baggage
}
