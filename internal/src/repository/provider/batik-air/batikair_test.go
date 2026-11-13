package batikair

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/pkg/simulation"
)

func ptr[T any](value T) *T {
	return &value
}

func sampleResponse() BatikAirSearchResp {
	return BatikAirSearchResp{
		Code:    200,
		Message: "OK",
		Results: []Flight{
			{
				FlightNumber:      "ID6514",
				AirlineName:       "Batik Air",
				AirlineIATA:       "ID",
				Origin:            "CGK",
				Destination:       "DPS",
				DepartureDateTime: "2025-12-15T07:15:00+0700",
				ArrivalDateTime:   "2025-12-15T10:00:00+0800",
				Fare: Fare{
					TotalPrice:   1100000,
					CurrencyCode: "IDR",
					Class:        "Y",
				},
				SeatsAvailable:  32,
				AircraftModel:   "Airbus A320",
				BaggageInfo:     "7kg cabin, 20kg checked",
				OnboardServices: []string{"Snack", "Beverage"},
			},
			{
				FlightNumber:      "ID7042",
				AirlineName:       "Batik Air",
				AirlineIATA:       "ID",
				Origin:            "CGK",
				Destination:       "DPS",
				DepartureDateTime: "2025-12-15T18:45:00+0700",
				ArrivalDateTime:   "2025-12-15T23:50:00+0800",
				Fare: Fare{
					TotalPrice:   950000,
					CurrencyCode: "IDR",
					Class:        "Y",
				},
				Connections: []Connection{
					{StopAirPort: "UPG", StopDuration: "55m"},
				},
				SeatsAvailable:  41,
				AircraftModel:   "Airbus A320",
				BaggageInfo:     "7kg cabin, 20kg checked",
				OnboardServices: []string{"Snack"},
			},
		},
	}
}

func TestParseBatikAirTime(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDatetime string
		wantTS       int64
		wantErr      bool
	}{
		{
			name:         "offset without colon",
			input:        "2025-12-15T07:15:00+0700",
			wantDatetime: "2025-12-15T07:15:00+07:00",
			wantTS:       1765757700,
		},
		{
			name:         "rfc3339 offset",
			input:        "2025-12-15T10:00:00+08:00",
			wantDatetime: "2025-12-15T10:00:00+08:00",
			wantTS:       1765764000,
		},
		{
			name:         "whitespace",
			input:        " 2025-12-15T07:15:00+0700 ",
			wantDatetime: "2025-12-15T07:15:00+07:00",
			wantTS:       1765757700,
		},
		{name: "invalid", input: "not-a-time", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDatetime, gotTS, err := parseBatikAirTime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotDatetime != tt.wantDatetime {
				t.Fatalf("datetime = %q, want %q", gotDatetime, tt.wantDatetime)
			}
			if gotTS != tt.wantTS {
				t.Fatalf("timestamp = %d, want %d", gotTS, tt.wantTS)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		minutes int64
		want    string
	}{
		{minutes: 0, want: "0h 0m"},
		{minutes: 105, want: "1h 45m"},
		{minutes: 110, want: "1h 50m"},
		{minutes: 245, want: "4h 5m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatDuration(tt.minutes); got != tt.want {
				t.Fatalf("formatDuration(%d) = %q, want %q", tt.minutes, got, tt.want)
			}
		})
	}
}

func TestParseBatikAirBaggage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  model.Baggage
	}{
		{
			name:  "cabin and checked",
			input: "7kg cabin, 20kg checked",
			want: model.Baggage{
				CarryOn: "7kg",
				Checked: "20kg",
			},
		},
		{
			name:  "spaces",
			input: " 7 kg cabin , 20 kg checked ",
			want: model.Baggage{
				CarryOn: "7 kg",
				Checked: "20 kg",
			},
		},
		{
			name:  "cabin only",
			input: "7kg cabin",
			want:  model.Baggage{CarryOn: "7kg"},
		},
		{name: "empty", input: "", want: model.Baggage{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBatikAirBaggage(tt.input); got != tt.want {
				t.Fatalf("parseBatikAirBaggage(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseResponseData(t *testing.T) {
	res, err := parseResponseData(sampleResponse(), model.SearchInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Flights) != 2 {
		t.Fatalf("flights = %d, want 2", len(res.Flights))
	}

	first := res.Flights[0]
	if first.ID != "ID6514_BatikAir" {
		t.Fatalf("first ID = %q", first.ID)
	}
	if first.Airline.Name != "Batik Air" || first.Airline.Code != "ID" {
		t.Fatalf("airline = %+v", first.Airline)
	}
	if first.Departure.Airport != "CGK" || first.Departure.City != "Jakarta" {
		t.Fatalf("departure = %+v", first.Departure)
	}
	if first.Departure.Timezone != model.TimezoneWIB {
		t.Fatalf("departure timezone = %q", first.Departure.Timezone)
	}
	if first.Arrival.Timezone != model.TimezoneWITA {
		t.Fatalf("arrival timezone = %q", first.Arrival.Timezone)
	}
	if first.Duration.TotalMinutes != 105 || first.Duration.Formatted != "1h 45m" {
		t.Fatalf("duration = %+v", first.Duration)
	}
	if first.Price.Formatted != "Rp 1.100.000" {
		t.Fatalf("formatted price = %q", first.Price.Formatted)
	}
	if first.CabinClass != "economy" {
		t.Fatalf("cabin class = %q", first.CabinClass)
	}
	if first.Aircraft == nil || first.Aircraft.Model != "Airbus A320" {
		t.Fatalf("aircraft = %+v", first.Aircraft)
	}
	if !reflect.DeepEqual(first.Amenities, []string{"beverage", "snack"}) {
		t.Fatalf("amenities = %#v", first.Amenities)
	}
	if first.Baggage.CarryOn != "7 kg" || first.Baggage.Checked != "20 kg" {
		t.Fatalf("baggage = %+v", first.Baggage)
	}

	second := res.Flights[1]
	if second.ID != "ID7042_BatikAir" {
		t.Fatalf("second ID = %q", second.ID)
	}
	if second.Stops != 1 {
		t.Fatalf("stops = %d, want 1", second.Stops)
	}
	if second.Duration.TotalMinutes != 245 || second.Duration.Formatted != "4h 5m" {
		t.Fatalf("duration = %+v", second.Duration)
	}
}

func TestParseResponseDataFilters(t *testing.T) {
	tests := []struct {
		name     string
		criteria model.SearchInput
		wantIDs  []string
	}{
		{
			name:    "no filter",
			wantIDs: []string{"ID6514_BatikAir", "ID7042_BatikAir"},
		},
		{
			name:     "wrong destination",
			criteria: model.SearchInput{Destination: "SUB"},
			wantIDs:  []string{},
		},
		{
			name:     "direct only",
			criteria: model.SearchInput{MaxStops: ptr(0)},
			wantIDs:  []string{"ID6514_BatikAir"},
		},
		{
			name:     "minimum price",
			criteria: model.SearchInput{MinPrice: ptr(1000000.0)},
			wantIDs:  []string{"ID6514_BatikAir"},
		},
		{
			name:     "minimum seats",
			criteria: model.SearchInput{Passengers: 40},
			wantIDs:  []string{"ID7042_BatikAir"},
		},
		{
			name:     "departure time window",
			criteria: model.SearchInput{DepartureFrom: "10:00"},
			wantIDs:  []string{"ID7042_BatikAir"},
		},
		{
			name:     "airline code",
			criteria: model.SearchInput{Airlines: []string{"ID"}},
			wantIDs:  []string{"ID6514_BatikAir", "ID7042_BatikAir"},
		},
		{
			name:     "cabin class",
			criteria: model.SearchInput{CabinClass: "economy"},
			wantIDs:  []string{"ID6514_BatikAir", "ID7042_BatikAir"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseResponseData(sampleResponse(), tt.criteria)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotIDs := make([]string, 0, len(res.Flights))
			for _, flight := range res.Flights {
				gotIDs = append(gotIDs, flight.ID)
			}
			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Fatalf("IDs = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestParseResponseDataValidation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Flight)
		wantErr error
	}{
		{
			name:   "invalid departure time",
			modify: func(f *Flight) { f.DepartureDateTime = "invalid" },
		},
		{
			name:   "invalid arrival time",
			modify: func(f *Flight) { f.ArrivalDateTime = "invalid" },
		},
		{
			name: "arrival before departure",
			modify: func(f *Flight) {
				f.DepartureDateTime = "2025-12-15T10:00:00+0700"
				f.ArrivalDateTime = "2025-12-15T09:00:00+0700"
			},
			wantErr: model.ErrArrivalBeforeDeparture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := sampleResponse()
			resp.Results = resp.Results[:1]
			tt.modify(&resp.Results[0])

			_, err := parseResponseData(resp, model.SearchInput{})
			if err == nil {
				t.Fatal("expected error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	originalSimulate := simulation.Simulate
	simulation.Simulate = func(
		context.Context,
		string,
		time.Duration,
		time.Duration,
		int,
	) error {
		return nil
	}
	t.Cleanup(func() {
		simulation.Simulate = originalSimulate
	})

	successPayload, err := json.Marshal(sampleResponse())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		path        func(t *testing.T) string
		wantErr     bool
		wantFlights int
	}{
		{
			name:    "empty path",
			path:    func(t *testing.T) string { return "" },
			wantErr: true,
		},
		{
			name: "missing file",
			path: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.json")
			},
			wantErr: true,
		},
		{
			name: "invalid json",
			path: func(t *testing.T) string {
				return writeTempFile(t, []byte("{"))
			},
			wantErr: true,
		},
		{
			name: "non-200 response code",
			path: func(t *testing.T) string {
				return writeTempFile(t, []byte(`{"code":500,"message":"error","results":[]}`))
			},
			wantErr: true,
		},
		{
			name: "success",
			path: func(t *testing.T) string {
				return writeTempFile(t, successPayload)
			},
			wantFlights: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Init(tt.path(t)).Search(context.Background(), model.SearchInput{})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(res.Flights) != tt.wantFlights {
				t.Fatalf("flights = %d, want %d", len(res.Flights), tt.wantFlights)
			}
		})
	}
}

func writeTempFile(t *testing.T, data []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "batik-air.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
