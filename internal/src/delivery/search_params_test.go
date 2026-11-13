package delivery

import (
	"errors"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/flight/internal/src/model"
)

func ptr[T any](value T) *T {
	return &value
}

func TestParseOptionalFloat(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    *float64
		wantErr bool
	}{
		{name: "empty", value: "", want: nil},
		{name: "integer", value: "500000", want: ptr(500000.0)},
		{name: "decimal", value: " 500000.5 ", want: ptr(500000.5)},
		{name: "negative", value: "-1", wantErr: true},
		{name: "invalid", value: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptionalFloat(tt.value, "test")
			if tt.wantErr {
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Fatalf("error = %v, want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOptionalInt(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    *int
		wantErr bool
	}{
		{name: "empty", value: "", want: nil},
		{name: "valid", value: "2", want: ptr(2)},
		{name: "with whitespace", value: " 2 ", want: ptr(2)},
		{name: "negative", value: "-1", wantErr: true},
		{name: "invalid", value: "2.5", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptionalInt(tt.value, "test")
			if tt.wantErr {
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Fatalf("error = %v, want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOptionalInt64(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    *int64
		wantErr bool
	}{
		{name: "empty", value: "", want: nil},
		{name: "valid", value: "120", want: ptr(int64(120))},
		{name: "negative", value: "-10", wantErr: true},
		{name: "invalid", value: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptionalInt64(tt.value, "test")
			if tt.wantErr {
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Fatalf("error = %v, want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseClock(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "empty", value: "", want: ""},
		{name: "valid", value: "06:00", want: "06:00"},
		{name: "with whitespace", value: " 18:30 ", want: "18:30"},
		{name: "single digit hour", value: "6:00", want: "6:00"},
		{name: "invalid hour", value: "24:00", wantErr: true},
		{name: "invalid", value: "morning", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseClock(tt.value, "test")
			if tt.wantErr {
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Fatalf("error = %v, want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseFilters(t *testing.T) {
	query := url.Values{
		"min_price":      []string{"500000"},
		"max_price":      []string{"1500000"},
		"max_stops":      []string{"1"},
		"min_duration":   []string{"90"},
		"max_duration":   []string{"240"},
		"departure_from": []string{"06:00"},
		"departure_to":   []string{"18:00"},
		"arrival_from":   []string{"08:00"},
		"arrival_to":     []string{"23:00"},
		"airlines":       []string{"QZ, JT", "GA"},
	}

	var got model.SearchInput
	if err := parseFilters(query, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := model.SearchInput{
		MinPrice:           ptr(500000.0),
		MaxPrice:           ptr(1500000.0),
		MaxStops:           ptr(1),
		MinDurationMinutes: ptr(int64(90)),
		MaxDurationMinutes: ptr(int64(240)),
		DepartureFrom:      "06:00",
		DepartureTo:        "18:00",
		ArrivalFrom:        "08:00",
		ArrivalTo:          "23:00",
		Airlines:           []string{"QZ", "JT", "GA"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseFiltersInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		query url.Values
	}{
		{name: "invalid min price", query: url.Values{"min_price": []string{"abc"}}},
		{name: "invalid max stops", query: url.Values{"max_stops": []string{"-1"}}},
		{name: "invalid duration", query: url.Values{"min_duration": []string{"abc"}}},
		{name: "invalid depart time", query: url.Values{"departure_from": []string{"25:00"}}},
		{name: "invalid arrival time", query: url.Values{"arrival_to": []string{"noon"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got model.SearchInput
			err := parseFilters(tt.query, &got)
			if !errors.Is(err, model.ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestValidateSearchInput(t *testing.T) {
	valid := func() model.SearchInput {
		return model.SearchInput{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
			Passengers:    1,
		}
	}

	tests := []struct {
		name    string
		modify  func(*model.SearchInput)
		wantErr bool
	}{
		{name: "valid"},
		{name: "missing departure date", modify: func(in *model.SearchInput) { in.DepartureDate = time.Time{} }, wantErr: true},
		{name: "missing origin", modify: func(in *model.SearchInput) { in.Origin = "" }, wantErr: true},
		{name: "missing destination", modify: func(in *model.SearchInput) { in.Destination = "" }, wantErr: true},
		{name: "invalid passengers", modify: func(in *model.SearchInput) { in.Passengers = 0 }, wantErr: true},
		{
			name: "min price exceeds max price",
			modify: func(in *model.SearchInput) {
				in.MinPrice = ptr(1500000.0)
				in.MaxPrice = ptr(500000.0)
			},
			wantErr: true,
		},
		{
			name: "min duration exceeds max duration",
			modify: func(in *model.SearchInput) {
				in.MinDurationMinutes = ptr(int64(240))
				in.MaxDurationMinutes = ptr(int64(90))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			if tt.modify != nil {
				tt.modify(&in)
			}

			err := validateSearchInput(in)
			if tt.wantErr {
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Fatalf("error = %v, want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
