package airasia

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

func sampleResponse() AirAsiaSearchResponse {
	return AirAsiaSearchResponse{
		Status: "ok",
		Flights: []Flight{
			{
				FlightCode:    "QZ520",
				Airline:       "AirAsia",
				FromAirport:   "CGK",
				ToAirport:     "DPS",
				DepartTime:    "2025-12-15T04:45:00+07:00",
				ArriveTime:    "2025-12-15T07:25:00+08:00",
				DurationHours: 1.67,
				DirectFlight:  true,
				PriceIDR:      650000,
				Seats:         67,
				CabinClass:    "economy",
				BaggageNotes:  "Cabin baggage only, checked bags additional fee",
			},
			{
				FlightCode:    "QZ7250",
				Airline:       "AirAsia",
				FromAirport:   "CGK",
				ToAirport:     "DPS",
				DepartTime:    "2025-12-15T15:15:00+07:00",
				ArriveTime:    "2025-12-15T20:35:00+08:00",
				DurationHours: 4.33,
				DirectFlight:  false,
				PriceIDR:      485000,
				Seats:         88,
				CabinClass:    "economy",
				BaggageNotes:  "Cabin baggage only, checked bags additional fee",
				Stops: []Stops{
					{Airport: "SOC", WaitTimeMinutes: 95},
				},
			},
		},
	}
}

func TestParseAirAsiaTime(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDatetime string
		wantTS       int64
		wantErr      bool
	}{
		{
			name:         "valid offset",
			input:        "2025-12-15T10:00:00+07:00",
			wantDatetime: "2025-12-15T10:00:00+07:00",
			wantTS:       1765767600,
		},
		{
			name:         "valid different offset",
			input:        "2025-12-15T12:45:00+08:00",
			wantDatetime: "2025-12-15T12:45:00+08:00",
			wantTS:       1765773900,
		},
		{name: "invalid", input: "not-a-time", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDatetime, gotTS, err := parseAirAsiaTime(tt.input)
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
		{minutes: 65, want: "1h 5m"},
		{minutes: 100, want: "1h 40m"},
		{minutes: 260, want: "4h 20m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatDuration(tt.minutes); got != tt.want {
				t.Fatalf("formatDuration(%d) = %q, want %q", tt.minutes, got, tt.want)
			}
		})
	}
}

func TestParseAirAsiaBaggage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Baggage
	}{
		{
			name:  "cabin and checked",
			input: "Cabin baggage only, checked bags additional fee",
			want: Baggage{
				CarryOn: "Cabin baggage only",
				Checked: "Additional fee",
			},
		},
		{
			name:  "cabin only",
			input: "Cabin baggage only",
			want:  Baggage{CarryOn: "Cabin baggage only"},
		},
		{
			name:  "extra whitespace",
			input: " Cabin baggage only , checked bags additional fee ",
			want: Baggage{
				CarryOn: "Cabin baggage only",
				Checked: "Additional fee",
			},
		},
		{name: "empty", input: "", want: Baggage{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseAirAsiaBaggage(tt.input); got != tt.want {
				t.Fatalf("parseAirAsiaBaggage(%q) = %+v, want %+v", tt.input, got, tt.want)
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
	if first.ID != "QZ520_AirAsia" {
		t.Fatalf("first ID = %q", first.ID)
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
	if first.Duration.TotalMinutes != 100 || first.Duration.Formatted != "1h 40m" {
		t.Fatalf("duration = %+v", first.Duration)
	}
	if first.Price.Formatted != "Rp 650.000" {
		t.Fatalf("formatted price = %q", first.Price.Formatted)
	}
	if first.Baggage.CarryOn != "Cabin baggage only" || first.Baggage.Checked != "Additional fee" {
		t.Fatalf("baggage = %+v", first.Baggage)
	}
	if first.Amenities == nil || len(first.Amenities) != 0 {
		t.Fatalf("amenities = %#v, want empty non-nil slice", first.Amenities)
	}

	second := res.Flights[1]
	if second.ID != "QZ7250_AirAsia" {
		t.Fatalf("second ID = %q", second.ID)
	}
	if second.Stops != 1 {
		t.Fatalf("stops = %d, want 1", second.Stops)
	}
	if second.Duration.TotalMinutes != 260 || second.Duration.Formatted != "4h 20m" {
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
			wantIDs: []string{"QZ520_AirAsia", "QZ7250_AirAsia"},
		},
		{
			name:     "wrong destination",
			criteria: model.SearchInput{Destination: "SUB"},
			wantIDs:  []string{},
		},
		{
			name:     "direct only",
			criteria: model.SearchInput{MaxStops: ptr(0)},
			wantIDs:  []string{"QZ520_AirAsia"},
		},
		{
			name:     "minimum price",
			criteria: model.SearchInput{MinPrice: ptr(600000.0)},
			wantIDs:  []string{"QZ520_AirAsia"},
		},
		{
			name:     "minimum seats",
			criteria: model.SearchInput{Passengers: 80},
			wantIDs:  []string{"QZ7250_AirAsia"},
		},
		{
			name:     "departure time window",
			criteria: model.SearchInput{DepartureFrom: "10:00"},
			wantIDs:  []string{"QZ7250_AirAsia"},
		},
		{
			name:     "airline code",
			criteria: model.SearchInput{Airlines: []string{"QZ"}},
			wantIDs:  []string{"QZ520_AirAsia", "QZ7250_AirAsia"},
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
			modify: func(f *Flight) { f.DepartTime = "invalid" },
		},
		{
			name:   "invalid arrival time",
			modify: func(f *Flight) { f.ArriveTime = "invalid" },
		},
		{
			name: "arrival before departure",
			modify: func(f *Flight) {
				f.DepartTime = "2025-12-15T10:00:00+07:00"
				f.ArriveTime = "2025-12-15T09:00:00+07:00"
			},
			wantErr: model.ErrArrivalBeforeDeparture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := sampleResponse()
			resp.Flights = resp.Flights[:1]
			tt.modify(&resp.Flights[0])

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
			name: "status not ok",
			path: func(t *testing.T) string {
				return writeTempFile(t, []byte(`{"status":"error","flights":[]}`))
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

	path := filepath.Join(t.TempDir(), "airasia.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func Test_capitalizeFirst(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "success",
			value: "test",
			want:  "Test",
		},
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capitalizeFirst(tt.value)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("capitalizeFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}
