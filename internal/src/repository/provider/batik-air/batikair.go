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
)

func (b *batikAir) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	if b.dataPath == "" {
		return res, fmt.Errorf("empty path for airasia provider")
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

	return parseResponseData(resp)
}

func parseResponseData(in BatikAirSearchResp) (res model.SearchResult, err error) {
	for _, f := range in.Results {
		departureDatetime, departureTS, err := parseBatikAirTime(f.DepartureDateTime)
		if err != nil {
			return res, err
		}

		arrivallDatetime, arrivalTS, err := parseBatikAirTime(f.ArrivalDateTime)
		if err != nil {
			return res, err
		}

		totalMinutes := (arrivalTS - departureTS) / 60

		fl := model.Flight{
			ID:       fmt.Sprintf("%s_%s", f.FlightNumber, model.ProviderBatikAir),
			Provider: model.ProviderBatikAir,
			Airline: model.AirLine{
				Name: f.AirlineName,
				Code: f.FlightNumber[0:2],
			},
			FlightNumber: f.FlightNumber,
			Departure: model.Departure{
				Airport:   f.Origin,
				City:      model.CityByAirportCode[f.Origin],
				DateTime:  departureDatetime,
				Timestamp: departureTS,
			},
			Arrival: model.Arrival{
				Airport:   f.Destination,
				City:      model.CityByAirportCode[f.Destination],
				DateTime:  arrivallDatetime,
				Timestamp: arrivalTS,
			},
			Duration: model.Duration{
				TotalMinutes: totalMinutes,
				Formatted:    formatDuration(totalMinutes),
			},
			Stops: len(f.Connections),
			Price: model.Price{
				Amount:   f.Fare.TotalPrice,
				Currency: f.Fare.CurrencyCode,
			},
			AvailableSeats: int64(f.SeatsAvailable),
			CabinClass:     model.GetCabinClass(f.Fare.Class),
			Amenities:      f.OnboardServices,
			Baggage:        parseBatikAirBaggage(f.BaggageInfo),
		}
		res.Flights = append(res.Flights, fl)
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
