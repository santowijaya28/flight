package lionair

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

func sampleResponse() LionAirSearchResp {
	return LionAirSearchResp{
		Success: true,
		Data: DataResp{
			AvailableFlights: []Flight{
				{
					ID: "JT740",
					Carrier: Carrier{
						Name: "Lion Air",
						IATA: "JT",
					},
					Route: Route{
						From: RoutePoint{
							Code: "CGK",
							Name: "Soekarno-Hatta International",
							City: "Jakarta",
						},
						To: RoutePoint{
							Code: "DPS",
							Name: "Ngurah Rai International",
							City: "Denpasar",
						},
					},
					Schedule: Schedule{
						Departure:         "2025-12-15T05:30:00",
						DepartureTimezone: "Asia/Jakarta",
						Arrival:           "2025-12-15T08:15:00",
						ArrivalTimezone:   "Asia/Makassar",
					},
					FlightTime: 105,
					IsDirect:   true,
					Pricing: Pricing{
						Total:    950000,
						Currency: "IDR",
						FareType: "ECONOMY",
					},
					SeatsLeft: 45,
					PlaneType: "Boeing 737-900ER",
					Services: Services{
						BaggageAllowance: BaggageAllowance{
							Cabin: "7 kg",
							Hold:  "20 kg",
						},
					},
				},
				{
					ID: "JT650",
					Carrier: Carrier{
						Name: "Lion Air",
						IATA: "JT",
					},
					Route: Route{
						From: RoutePoint{
							Code: "CGK",
							Name: "Soekarno-Hatta International",
							City: "Jakarta",
						},
						To: RoutePoint{
							Code: "DPS",
							Name: "Ngurah Rai International",
							City: "Denpasar",
						},
					},
					Schedule: Schedule{
						Departure:         "2025-12-15T16:20:00",
						DepartureTimezone: "Asia/Jakarta",
						Arrival:           "2025-12-15T21:10:00",
						ArrivalTimezone:   "Asia/Makassar",
					},
					FlightTime: 230,
					IsDirect:   false,
					StopCount:  1,
					Layovers: []Layover{
						{Airport: "SUB", DurationMinutes: 75},
					},
					Pricing: Pricing{
						Total:    780000,
						Currency: "IDR",
						FareType: "ECONOMY",
					},
					SeatsLeft: 52,
					PlaneType: "Boeing 737-800",
					Services: Services{
						WifiAvailable: true,
						MealsIncluded: true,
						BaggageAllowance: BaggageAllowance{
							Cabin: "7 kg",
							Hold:  "20 kg",
						},
					},
				},
			},
		},
	}
}

func TestParseLionAirTime(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		timezone     string
		wantDatetime string
		wantTS       int64
		wantErr      bool
	}{
		{
			name:         "jakarta",
			value:        "2025-12-15T05:30:00",
			timezone:     "Asia/Jakarta",
			wantDatetime: "2025-12-15T05:30:00+07:00",
			wantTS:       1765751400,
		},
		{
			name:         "makassar",
			value:        "2025-12-15T08:15:00",
			timezone:     "Asia/Makassar",
			wantDatetime: "2025-12-15T08:15:00+08:00",
			wantTS:       1765757700,
		},
		{name: "empty timezone", value: "2025-12-15T05:30:00", wantErr: true},
		{name: "invalid timezone", value: "2025-12-15T05:30:00", timezone: "Invalid/Zone", wantErr: true},
		{name: "invalid time", value: "not-a-time", timezone: "Asia/Jakarta", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDatetime, gotTS, err := parseLionAirTime(tt.value, tt.timezone)
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

func TestNormalizeCabinClass(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "ECONOMY", want: "economy"},
		{input: " economy ", want: "economy"},
		{input: "PREMIUM_ECONOMY", want: "premium economy"},
		{input: "", want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizeCabinClass(tt.input); got != tt.want {
				t.Fatalf("normalizeCabinClass(%q) = %q, want %q", tt.input, got, tt.want)
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
		{minutes: 230, want: "3h 50m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatDuration(tt.minutes); got != tt.want {
				t.Fatalf("formatDuration(%d) = %q, want %q", tt.minutes, got, tt.want)
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
	if direct.ID != "JT740_LionAir" {
		t.Fatalf("direct ID = %q", direct.ID)
	}
	if direct.Airline.Name != "Lion Air" || direct.Airline.Code != "JT" {
		t.Fatalf("airline = %+v", direct.Airline)
	}
	if direct.Departure.Airport != "CGK" || direct.Departure.City != "Jakarta" {
		t.Fatalf("departure = %+v", direct.Departure)
	}
	if direct.Departure.Timezone != model.TimezoneWIB || direct.Arrival.Timezone != model.TimezoneWITA {
		t.Fatalf("timezones = %q -> %q", direct.Departure.Timezone, direct.Arrival.Timezone)
	}
	if direct.Duration.TotalMinutes != 105 || direct.Duration.Formatted != "1h 45m" {
		t.Fatalf("duration = %+v", direct.Duration)
	}
	if direct.Price.Formatted != "Rp 950.000" {
		t.Fatalf("formatted price = %q", direct.Price.Formatted)
	}
	if direct.CabinClass != "economy" {
		t.Fatalf("cabin class = %q", direct.CabinClass)
	}
	if direct.Aircraft == nil || direct.Aircraft.Model != "Boeing 737-900ER" {
		t.Fatalf("aircraft = %+v", direct.Aircraft)
	}
	if direct.Amenities == nil || len(direct.Amenities) != 0 {
		t.Fatalf("amenities = %#v", direct.Amenities)
	}
	if direct.Baggage.CarryOn != "7 kg" || direct.Baggage.Checked != "20 kg" {
		t.Fatalf("baggage = %+v", direct.Baggage)
	}

	withStop := res.Flights[1]
	if withStop.ID != "JT650_LionAir" {
		t.Fatalf("stop ID = %q", withStop.ID)
	}
	if withStop.Stops != 1 {
		t.Fatalf("stops = %d, want 1", withStop.Stops)
	}
	if withStop.Duration.TotalMinutes != 230 || withStop.Duration.Formatted != "3h 50m" {
		t.Fatalf("duration = %+v", withStop.Duration)
	}
	if !reflect.DeepEqual(withStop.Amenities, []string{"meal", "wifi"}) {
		t.Fatalf("amenities = %#v", withStop.Amenities)
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
			wantIDs: []string{"JT740_LionAir", "JT650_LionAir"},
		},
		{
			name:     "wrong destination",
			criteria: model.SearchInput{Destination: "SUB"},
			wantIDs:  []string{},
		},
		{
			name:     "direct only",
			criteria: model.SearchInput{MaxStops: ptr(0)},
			wantIDs:  []string{"JT740_LionAir"},
		},
		{
			name:     "minimum price",
			criteria: model.SearchInput{MinPrice: ptr(800000.0)},
			wantIDs:  []string{"JT740_LionAir"},
		},
		{
			name:     "minimum seats",
			criteria: model.SearchInput{Passengers: 50},
			wantIDs:  []string{"JT650_LionAir"},
		},
		{
			name:     "departure time window",
			criteria: model.SearchInput{DepartureFrom: "10:00"},
			wantIDs:  []string{"JT650_LionAir"},
		},
		{
			name:     "airline code",
			criteria: model.SearchInput{Airlines: []string{"JT"}},
			wantIDs:  []string{"JT740_LionAir", "JT650_LionAir"},
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
			modify: func(f *Flight) { f.Schedule.Departure = "invalid" },
		},
		{
			name:   "invalid timezone",
			modify: func(f *Flight) { f.Schedule.DepartureTimezone = "Invalid/Zone" },
		},
		{
			name: "arrival before departure",
			modify: func(f *Flight) {
				f.Schedule.Departure = "2025-12-15T10:00:00"
				f.Schedule.Arrival = "2025-12-15T09:00:00"
			},
			wantErr: model.ErrArrivalBeforeDeparture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := sampleResponse()
			resp.Data.AvailableFlights = resp.Data.AvailableFlights[:1]
			tt.modify(&resp.Data.AvailableFlights[0])

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
			name: "unsuccessful response",
			path: func(t *testing.T) string {
				return writeTempFile(t, []byte(`{"success":false,"data":{"available_flights":[]}}`))
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

	path := filepath.Join(t.TempDir(), "lion-air.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
