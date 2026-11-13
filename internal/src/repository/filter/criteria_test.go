package filter

import (
	"testing"
	"time"

	"github.com/flight/internal/src/model"
)

func ptr[T any](value T) *T {
	return &value
}

func baseFlight() model.Flight {
	return model.Flight{
		ID: "QZ520_AirAsia",
		Airline: model.AirLine{
			Name: "AirAsia",
			Code: "QZ",
		},
		Departure: model.Departure{
			Airport:   "CGK",
			City:      "Jakarta",
			DateTime:  "2025-12-15T04:45:00+07:00",
			Timestamp: 1765748700,
		},
		Arrival: model.Arrival{
			Airport:   "DPS",
			City:      "Denpasar",
			DateTime:  "2025-12-15T07:25:00+08:00",
			Timestamp: 1765754700,
		},
		Duration: model.Duration{
			TotalMinutes: 100,
		},
		Stops:          0,
		Price:          model.Price{Amount: 650000},
		AvailableSeats: 67,
		CabinClass:     "economy",
	}
}

func TestMatchesFlight(t *testing.T) {
	departureDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	otherDate := time.Date(2025, 12, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		criteria     model.SearchInput
		modifyFlight func(*model.Flight)
		want         bool
	}{
		{name: "no filters", want: true},
		{
			name:     "origin case and whitespace",
			criteria: model.SearchInput{Origin: " cgk "},
			want:     true,
		},
		{
			name:     "origin mismatch",
			criteria: model.SearchInput{Origin: "SUB"},
			want:     false,
		},
		{
			name:     "destination match",
			criteria: model.SearchInput{Destination: "dps"},
			want:     true,
		},
		{
			name:     "destination mismatch",
			criteria: model.SearchInput{Destination: "SUB"},
			want:     false,
		},
		{
			name:     "cabin match",
			criteria: model.SearchInput{CabinClass: " ECONOMY "},
			want:     true,
		},
		{
			name:     "cabin mismatch",
			criteria: model.SearchInput{CabinClass: "business"},
			want:     false,
		},
		{
			name:     "passengers equal seats",
			criteria: model.SearchInput{Passengers: 67},
			want:     true,
		},
		{
			name:     "passengers exceed seats",
			criteria: model.SearchInput{Passengers: 68},
			want:     false,
		},
		{
			name:     "zero passengers means no filter",
			criteria: model.SearchInput{Passengers: 0},
			want:     true,
		},
		{
			name:     "date match",
			criteria: model.SearchInput{DepartureDate: departureDate},
			want:     true,
		},
		{
			name:     "date mismatch",
			criteria: model.SearchInput{DepartureDate: otherDate},
			want:     false,
		},
		{
			name:     "minimum price equal",
			criteria: model.SearchInput{MinPrice: ptr(650000.0)},
			want:     true,
		},
		{
			name:     "minimum price too high",
			criteria: model.SearchInput{MinPrice: ptr(650001.0)},
			want:     false,
		},
		{
			name:     "maximum price equal",
			criteria: model.SearchInput{MaxPrice: ptr(650000.0)},
			want:     true,
		},
		{
			name:     "maximum price too low",
			criteria: model.SearchInput{MaxPrice: ptr(649999.0)},
			want:     false,
		},
		{
			name:     "price within range",
			criteria: model.SearchInput{MinPrice: ptr(500000.0), MaxPrice: ptr(700000.0)},
			want:     true,
		},
		{
			name:     "max stops zero",
			criteria: model.SearchInput{MaxStops: ptr(0)},
			want:     true,
		},
		{
			name:         "stops exceed maximum",
			criteria:     model.SearchInput{MaxStops: ptr(0)},
			modifyFlight: func(f *model.Flight) { f.Stops = 1 },
			want:         false,
		},
		{
			name:     "airline code match",
			criteria: model.SearchInput{Airlines: []string{"qz"}},
			want:     true,
		},
		{
			name:     "airline name match",
			criteria: model.SearchInput{Airlines: []string{"airasia"}},
			want:     true,
		},
		{
			name:     "airline mismatch",
			criteria: model.SearchInput{Airlines: []string{"GA"}},
			want:     false,
		},
		{
			name:     "minimum duration equal",
			criteria: model.SearchInput{MinDurationMinutes: ptr(int64(100))},
			want:     true,
		},
		{
			name:     "minimum duration too high",
			criteria: model.SearchInput{MinDurationMinutes: ptr(int64(101))},
			want:     false,
		},
		{
			name:     "maximum duration equal",
			criteria: model.SearchInput{MaxDurationMinutes: ptr(int64(100))},
			want:     true,
		},
		{
			name:     "maximum duration too low",
			criteria: model.SearchInput{MaxDurationMinutes: ptr(int64(99))},
			want:     false,
		},
		{
			name:     "departure window match",
			criteria: model.SearchInput{DepartureFrom: "04:00", DepartureTo: "05:00"},
			want:     true,
		},
		{
			name:     "departure window mismatch",
			criteria: model.SearchInput{DepartureFrom: "05:00", DepartureTo: "06:00"},
			want:     false,
		},
		{
			name:     "departure overnight window",
			criteria: model.SearchInput{DepartureFrom: "22:00", DepartureTo: "06:00"},
			want:     true,
		},
		{
			name:     "arrival window match",
			criteria: model.SearchInput{ArrivalFrom: "07:00", ArrivalTo: "08:00"},
			want:     true,
		},
		{
			name:     "arrival window mismatch",
			criteria: model.SearchInput{ArrivalFrom: "08:00", ArrivalTo: "09:00"},
			want:     false,
		},
		{
			name:         "invalid datetime with window",
			criteria:     model.SearchInput{DepartureFrom: "00:00"},
			modifyFlight: func(f *model.Flight) { f.Departure.DateTime = "invalid" },
			want:         false,
		},
		{
			name: "all filters match",
			criteria: model.SearchInput{
				Origin:             "CGK",
				Destination:        "DPS",
				CabinClass:         "economy",
				Passengers:         1,
				DepartureDate:      departureDate,
				MinPrice:           ptr(500000.0),
				MaxPrice:           ptr(700000.0),
				MaxStops:           ptr(0),
				Airlines:           []string{"QZ"},
				MinDurationMinutes: ptr(int64(90)),
				MaxDurationMinutes: ptr(int64(110)),
				DepartureFrom:      "04:00",
				DepartureTo:        "05:00",
				ArrivalFrom:        "07:00",
				ArrivalTo:          "08:00",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flight := baseFlight()
			if tt.modifyFlight != nil {
				tt.modifyFlight(&flight)
			}

			if got := MatchesFlight(flight, tt.criteria); got != tt.want {
				t.Fatalf("MatchesFlight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesTimeWindow(t *testing.T) {
	datetime := "2025-12-15T23:30:00+07:00"

	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		{name: "no window", want: true},
		{name: "from only match", from: "23:00", want: true},
		{name: "from only mismatch", from: "23:45", want: false},
		{name: "to only match", to: "23:45", want: true},
		{name: "to only mismatch", to: "23:00", want: false},
		{name: "same day match", from: "23:00", to: "23:59", want: true},
		{name: "same day mismatch", from: "22:00", to: "23:00", want: false},
		{name: "overnight match", from: "23:00", to: "06:00", want: true},
		{name: "overnight mismatch", from: "22:00", to: "23:00", want: false},
		{name: "invalid datetime", from: "00:00", to: "01:00", want: false},
		{name: "invalid from", from: "bad", want: false},
		{name: "invalid to", to: "bad", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := datetime
			if tt.name == "invalid datetime" {
				value = "invalid"
			}

			if got := matchesTimeWindow(value, tt.from, tt.to); got != tt.want {
				t.Fatalf("matchesTimeWindow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseClock(t *testing.T) {
	tests := []struct {
		value      string
		wantMinute int
		wantOK     bool
	}{
		{value: "", wantMinute: 0, wantOK: false},
		{value: "06:00", wantMinute: 360, wantOK: true},
		{value: "23:30", wantMinute: 1410, wantOK: true},
		{value: "6:00", wantMinute: 360, wantOK: true},
		{value: "24:00", wantMinute: 0, wantOK: false},
		{value: "bad", wantMinute: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			gotMinute, gotOK := parseClock(tt.value)
			if gotMinute != tt.wantMinute || gotOK != tt.wantOK {
				t.Fatalf("parseClock(%q) = (%d, %v), want (%d, %v)",
					tt.value, gotMinute, gotOK, tt.wantMinute, tt.wantOK)
			}
		})
	}
}
