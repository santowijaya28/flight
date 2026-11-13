package model

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestOptionalAircraft(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNil   bool
		wantModel string
	}{
		{name: "empty", input: "", wantNil: true},
		{name: "spaces", input: "   ", wantNil: true},
		{name: "value", input: "Boeing 737-800", wantModel: "Boeing 737-800"},
		{name: "trimmed value", input: "  Airbus A320  ", wantModel: "Airbus A320"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptionalAircraft(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("got %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("got nil, want aircraft")
			}
			if got.Model != tt.wantModel {
				t.Fatalf("model = %q, want %q", got.Model, tt.wantModel)
			}
		})
	}
}

func TestNormalizeAmenities(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{name: "nil", input: nil, want: []string{}},
		{name: "empty", input: []string{}, want: []string{}},
		{
			name:  "trim lower deduplicate sort",
			input: []string{" Wifi ", "power_outlet", "WIFI", "", "Meal", "meal"},
			want:  []string{"meal", "power outlet", "wifi"},
		},
		{
			name:  "already normalized",
			input: []string{"beverage", "snack"},
			want:  []string{"beverage", "snack"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAmenities(tt.input)
			if got == nil {
				t.Fatal("got nil, want non-nil slice")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNormalizeBaggage(t *testing.T) {
	tests := []struct {
		name  string
		input Baggage
		want  Baggage
	}{
		{
			name:  "weight spacing",
			input: Baggage{CarryOn: " 7kg ", Checked: "20KG"},
			want:  Baggage{CarryOn: "7 kg", Checked: "20 kg"},
		},
		{
			name:  "decimal weight",
			input: Baggage{CarryOn: "7.5kg", Checked: "20 kg"},
			want:  Baggage{CarryOn: "7.5 kg", Checked: "20 kg"},
		},
		{
			name:  "descriptive values unchanged",
			input: Baggage{CarryOn: "Cabin baggage only", Checked: "Additional fee"},
			want:  Baggage{CarryOn: "Cabin baggage only", Checked: "Additional fee"},
		},
		{
			name:  "empty",
			input: Baggage{},
			want:  Baggage{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeBaggage(tt.input); got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFormatThousands(t *testing.T) {
	tests := []struct {
		value int64
		want  string
	}{
		{value: 0, want: "0"},
		{value: 999, want: "999"},
		{value: 1000, want: "1.000"},
		{value: 1250000, want: "1.250.000"},
		{value: -1500, want: "-1.500"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatThousands(tt.value); got != tt.want {
				t.Fatalf("formatThousands(%d) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestFormatIDR(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{amount: 0, want: "Rp 0"},
		{amount: 485000, want: "Rp 485.000"},
		{amount: 1250000, want: "Rp 1.250.000"},
		{amount: 1850000.4, want: "Rp 1.850.000"},
		{amount: 1850000.6, want: "Rp 1.850.001"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatIDR(tt.amount); got != tt.want {
				t.Fatalf("FormatIDR(%v) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}

func TestCityByAirportCode(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: "CGK", want: "Jakarta"},
		{code: "DPS", want: "Denpasar"},
		{code: "SUB", want: "Surabaya"},
		{code: "SOC", want: "Surakarta (Solo)"},
		{code: "UPG", want: "Makassar"},
		{code: "XXX", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := CityByAirportCode[tt.code]; got != tt.want {
				t.Fatalf("city for %s = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestGetCabinClass(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: "Y", want: "economy"},
		{code: "y", want: "economy"},
		{code: " Y ", want: "economy"},
		{code: "C", want: "business"},
		{code: "F", want: "first"},
		{code: "W", want: "premium economy"},
		{code: "?", want: "unknown"},
		{code: "", want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := GetCabinClass(tt.code); got != tt.want {
				t.Fatalf("GetCabinClass(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestAirportTimezone(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{code: "CGK", want: TimezoneWIB},
		{code: " cgk ", want: TimezoneWIB},
		{code: "DPS", want: TimezoneWITA},
		{code: "DJJ", want: TimezoneWIT},
		{code: "XXX", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := AirportTimezone(tt.code); got != tt.want {
				t.Fatalf("AirportTimezone(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestFormatAirportDateTime(t *testing.T) {
	tests := []struct {
		name         string
		datetime     string
		airport      string
		wantDatetime string
		wantTimezone string
	}{
		{
			name:         "convert utc to wib",
			datetime:     "2025-12-15T08:15:00Z",
			airport:      "CGK",
			wantDatetime: "2025-12-15T15:15:00+07:00",
			wantTimezone: TimezoneWIB,
		},
		{
			name:         "convert utc to wita",
			datetime:     "2025-12-15T08:15:00Z",
			airport:      "DPS",
			wantDatetime: "2025-12-15T16:15:00+08:00",
			wantTimezone: TimezoneWITA,
		},
		{
			name:         "convert utc to wit",
			datetime:     "2025-12-15T08:15:00Z",
			airport:      "DJJ",
			wantDatetime: "2025-12-15T17:15:00+09:00",
			wantTimezone: TimezoneWIT,
		},
		{
			name:         "invalid datetime keeps original",
			datetime:     "invalid",
			airport:      "CGK",
			wantDatetime: "invalid",
			wantTimezone: TimezoneWIB,
		},
		{
			name:         "unknown airport",
			datetime:     "2025-12-15T08:15:00Z",
			airport:      "XXX",
			wantDatetime: "2025-12-15T08:15:00Z",
			wantTimezone: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDatetime, gotTimezone := FormatAirportDateTime(tt.datetime, tt.airport)
			if gotDatetime != tt.wantDatetime {
				t.Fatalf("datetime = %q, want %q", gotDatetime, tt.wantDatetime)
			}
			if gotTimezone != tt.wantTimezone {
				t.Fatalf("timezone = %q, want %q", gotTimezone, tt.wantTimezone)
			}
		})
	}
}

func TestTimezoneLocation(t *testing.T) {
	tests := []struct {
		name       string
		timezone   string
		wantOffset int
	}{
		{name: "wib", timezone: TimezoneWIB, wantOffset: 7 * 60 * 60},
		{name: "wita", timezone: TimezoneWITA, wantOffset: 8 * 60 * 60},
		{name: "wit", timezone: TimezoneWIT, wantOffset: 9 * 60 * 60},
		{name: "default", timezone: "UNKNOWN", wantOffset: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, offset := time.Now().In(timezoneLocation(tt.timezone)).Zone()
			if offset != tt.wantOffset {
				t.Fatalf("offset = %d, want %d", offset, tt.wantOffset)
			}
		})
	}
}

func TestValidateArrivalAfterDeparture(t *testing.T) {
	tests := []struct {
		name        string
		departureTS int64
		arrivalTS   int64
		wantErr     bool
	}{
		{name: "valid", departureTS: 100, arrivalTS: 200},
		{name: "same time", departureTS: 100, arrivalTS: 100, wantErr: true},
		{name: "arrival before departure", departureTS: 200, arrivalTS: 100, wantErr: true},
		{name: "zero values", departureTS: 0, arrivalTS: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArrivalAfterDeparture(tt.departureTS, tt.arrivalTS)
			if tt.wantErr {
				if !errors.Is(err, ErrArrivalBeforeDeparture) {
					t.Fatalf("error = %v, want ErrArrivalBeforeDeparture", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
