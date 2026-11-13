package garudaindonesia

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

func sampleResponse() GarudaIndonesiaSearchResp {
	return GarudaIndonesiaSearchResp{
		Status: "success",
		Flights: []Flight{
			{
				FlightID:    "GA400",
				Airline:     "Garuda Indonesia",
				AirlineCode: "GA",
				Departure: Departure{
					Airport: "CGK",
					City:    "Jakarta",
					Time:    "2025-12-15T06:00:00+07:00",
				},
				Arrival: Arrival{
					Airport: "DPS",
					City:    "Denpasar",
					Time:    "2025-12-15T08:50:00+08:00",
				},
				DurationMinutes: 110,
				Aircraft:        "Boeing 737-800",
				Price: Price{
					Amount:   1250000,
					Currency: "IDR",
				},
				AvailableSeats: 28,
				FareClass:      "economy",
				Baggage:        Baggage{CarryOn: 1, Checked: 2},
				Amenities:      []string{"wifi", "meal", "entertainment"},
			},
			{
				FlightID:    "GA315",
				Airline:     "Garuda Indonesia",
				AirlineCode: "GA",
				Departure: Departure{
					Airport: "CGK",
					City:    "Jakarta",
					Time:    "2025-12-15T14:00:00+07:00",
				},
				Arrival: Arrival{
					Airport: "SUB",
					City:    "Surabaya",
					Time:    "2025-12-15T15:30:00+07:00",
				},
				DurationMinutes: 90,
				Aircraft:        "Boeing 737",
				Price: Price{
					Amount:   1850000,
					Currency: "IDR",
				},
				AvailableSeats: 22,
				FareClass:      "economy",
				Baggage:        Baggage{CarryOn: 1, Checked: 2},
				Segments: []Segmennt{
					{
						FlightNumber: "GA315",
						Departure: SegmentDeparture{
							Airport: "CGK",
							Time:    "2025-12-15T14:00:00+07:00",
						},
						Arrival: SegmentArrival{
							Airport: "SUB",
							Time:    "2025-12-15T15:30:00+07:00",
						},
						DurationMinutes: 90,
					},
					{
						FlightNumber: "GA332",
						Departure: SegmentDeparture{
							Airport: "SUB",
							Time:    "2025-12-15T17:15:00+07:00",
						},
						Arrival: SegmentArrival{
							Airport: "DPS",
							Time:    "2025-12-15T18:45:00+08:00",
						},
						DurationMinutes: 90,
						LayOverMinutes:  105,
					},
				},
			},
		},
	}
}

func TestParseGarudaTime(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDatetime string
		wantTS       int64
		wantErr      bool
	}{
		{
			name:         "valid",
			input:        "2025-12-15T06:00:00+07:00",
			wantDatetime: "2025-12-15T06:00:00+07:00",
			wantTS:       1765753200,
		},
		{
			name:         "different offset",
			input:        "2025-12-15T08:50:00+08:00",
			wantDatetime: "2025-12-15T08:50:00+08:00",
			wantTS:       1765759800,
		},
		{name: "invalid", input: "not-a-time", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDatetime, gotTS, err := parseGarudaTime(tt.input)
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
		{minutes: 110, want: "1h 50m"},
		{minutes: 115, want: "1h 55m"},
		{minutes: 225, want: "3h 45m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatDuration(tt.minutes); got != tt.want {
				t.Fatalf("formatDuration(%d) = %q, want %q", tt.minutes, got, tt.want)
			}
		})
	}
}

func TestParseGarudaBaggage(t *testing.T) {
	tests := []struct {
		name    string
		baggage Baggage
		want    model.Baggage
	}{
		{
			name:    "one and two",
			baggage: Baggage{CarryOn: 1, Checked: 2},
			want: model.Baggage{
				CarryOn: "1 carry-on bag",
				Checked: "2 checked bags",
			},
		},
		{
			name:    "zero values",
			baggage: Baggage{},
			want:    model.Baggage{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseGarudaBaggage(tt.baggage); got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
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

	direct := res.Flights[0]
	if direct.ID != "GA400_GarudaIndonesia" {
		t.Fatalf("direct ID = %q", direct.ID)
	}
	if direct.Departure.Timezone != model.TimezoneWIB || direct.Arrival.Timezone != model.TimezoneWITA {
		t.Fatalf("timezones = %q -> %q", direct.Departure.Timezone, direct.Arrival.Timezone)
	}
	if direct.Duration.TotalMinutes != 110 || direct.Duration.Formatted != "1h 50m" {
		t.Fatalf("duration = %+v", direct.Duration)
	}
	if direct.Price.Formatted != "Rp 1.250.000" {
		t.Fatalf("formatted price = %q", direct.Price.Formatted)
	}
	if !reflect.DeepEqual(direct.Amenities, []string{"entertainment", "meal", "wifi"}) {
		t.Fatalf("amenities = %#v", direct.Amenities)
	}
	if direct.Baggage.CarryOn != "1 carry-on bag" || direct.Baggage.Checked != "2 checked bags" {
		t.Fatalf("baggage = %+v", direct.Baggage)
	}
	if direct.Aircraft == nil || direct.Aircraft.Model != "Boeing 737-800" {
		t.Fatalf("aircraft = %+v", direct.Aircraft)
	}

	segmented := res.Flights[1]
	if segmented.ID != "GA315_GarudaIndonesia" {
		t.Fatalf("segmented ID = %q", segmented.ID)
	}
	if segmented.Arrival.Airport != "DPS" || segmented.Arrival.City != "Denpasar" {
		t.Fatalf("segmented arrival = %+v", segmented.Arrival)
	}
	if segmented.Duration.TotalMinutes != 225 || segmented.Duration.Formatted != "3h 45m" {
		t.Fatalf("segmented duration = %+v", segmented.Duration)
	}
	if segmented.Stops != 1 {
		t.Fatalf("segmented stops = %d, want 1", segmented.Stops)
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
			wantIDs: []string{"GA400_GarudaIndonesia", "GA315_GarudaIndonesia"},
		},
		{
			name:     "destination uses final segment",
			criteria: model.SearchInput{Destination: "DPS"},
			wantIDs:  []string{"GA400_GarudaIndonesia", "GA315_GarudaIndonesia"},
		},
		{
			name:     "first segment destination does not match",
			criteria: model.SearchInput{Destination: "SUB"},
			wantIDs:  []string{},
		},
		{
			name:     "direct only",
			criteria: model.SearchInput{MaxStops: ptr(0)},
			wantIDs:  []string{"GA400_GarudaIndonesia"},
		},
		{
			name:     "minimum price",
			criteria: model.SearchInput{MinPrice: ptr(1500000.0)},
			wantIDs:  []string{"GA315_GarudaIndonesia"},
		},
		{
			name:     "minimum seats",
			criteria: model.SearchInput{Passengers: 25},
			wantIDs:  []string{"GA400_GarudaIndonesia"},
		},
		{
			name:     "departure time window",
			criteria: model.SearchInput{DepartureFrom: "10:00"},
			wantIDs:  []string{"GA315_GarudaIndonesia"},
		},
		{
			name:     "airline code",
			criteria: model.SearchInput{Airlines: []string{"GA"}},
			wantIDs:  []string{"GA400_GarudaIndonesia", "GA315_GarudaIndonesia"},
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
			modify: func(f *Flight) { f.Departure.Time = "invalid" },
		},
		{
			name:   "invalid arrival time",
			modify: func(f *Flight) { f.Arrival.Time = "invalid" },
		},
		{
			name: "arrival before departure",
			modify: func(f *Flight) {
				f.Departure.Time = "2025-12-15T10:00:00+07:00"
				f.Arrival.Time = "2025-12-15T09:00:00+07:00"
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
			name: "status not success",
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

	path := filepath.Join(t.TempDir(), "garuda-indonesia.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
