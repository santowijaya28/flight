package usecase

import (
	"errors"
	"reflect"
	"testing"

	"github.com/flight/internal/src/model"
)

func sortableFlights() []model.Flight {
	return []model.Flight{
		{
			ID:        "A",
			Price:     model.Price{Amount: 300},
			Duration:  model.Duration{TotalMinutes: 120},
			Stops:     0,
			Departure: model.Departure{Timestamp: 30},
			Arrival:   model.Arrival{Timestamp: 100},
		},
		{
			ID:        "B",
			Price:     model.Price{Amount: 100},
			Duration:  model.Duration{TotalMinutes: 180},
			Stops:     1,
			Departure: model.Departure{Timestamp: 10},
			Arrival:   model.Arrival{Timestamp: 200},
		},
		{
			ID:        "C",
			Price:     model.Price{Amount: 200},
			Duration:  model.Duration{TotalMinutes: 90},
			Stops:     0,
			Departure: model.Departure{Timestamp: 20},
			Arrival:   model.Arrival{Timestamp: 150},
		},
		{
			ID:        "D",
			Price:     model.Price{Amount: 200},
			Duration:  model.Duration{TotalMinutes: 120},
			Stops:     2,
			Departure: model.Departure{Timestamp: 40},
			Arrival:   model.Arrival{Timestamp: 50},
		},
	}
}

func TestParseSort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []sortOption
		wantErr bool
	}{
		{name: "empty defaults to price asc", input: "", want: []sortOption{sortPriceAsc}},
		{name: "whitespace defaults to price asc", input: "   ", want: []sortOption{sortPriceAsc}},
		{name: "price asc", input: "price_asc", want: []sortOption{sortPriceAsc}},
		{name: "uppercase", input: " PRICE_ASC ", want: []sortOption{sortPriceAsc}},
		{name: "price desc", input: "price_desc", want: []sortOption{sortPriceDesc}},
		{name: "duration asc", input: "duration_asc", want: []sortOption{sortDurationAsc}},
		{name: "duration desc", input: "duration_desc", want: []sortOption{sortDurationDesc}},
		{name: "departure asc", input: "departure_asc", want: []sortOption{sortDepartureAsc}},
		{name: "departure desc", input: "departure_desc", want: []sortOption{sortDepartureDesc}},
		{name: "arrival asc", input: "arrival_asc", want: []sortOption{sortArrivalAsc}},
		{name: "arrival desc", input: "arrival_desc", want: []sortOption{sortArrivalDesc}},
		{name: "best value", input: "best_value", want: []sortOption{sortBestValueAsc}},
		{name: "best value asc", input: "best_value_asc", want: []sortOption{sortBestValueAsc}},
		{name: "best value desc", input: "best_value_desc", want: []sortOption{sortBestValueDesc}},
		{
			name:  "multiple keys",
			input: "price_asc, duration_asc",
			want:  []sortOption{sortPriceAsc, sortDurationAsc},
		},
		{
			name:  "skips empty keys",
			input: "price_asc,,duration_asc",
			want:  []sortOption{sortPriceAsc, sortDurationAsc},
		},
		{name: "only commas defaults", input: ",,", want: []sortOption{sortPriceAsc}},
		{name: "invalid", input: "unknown", wantErr: true},
		{name: "invalid in list", input: "price_asc,unknown", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSort(tt.input)
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

func TestSortFlights(t *testing.T) {
	tests := []struct {
		name    string
		options []sortOption
		wantIDs []string
	}{
		{
			name:    "no options uses id",
			options: nil,
			wantIDs: []string{"A", "B", "C", "D"},
		},
		{
			name:    "price asc",
			options: []sortOption{sortPriceAsc},
			wantIDs: []string{"B", "C", "D", "A"},
		},
		{
			name:    "price desc",
			options: []sortOption{sortPriceDesc},
			wantIDs: []string{"A", "C", "D", "B"},
		},
		{
			name:    "duration asc",
			options: []sortOption{sortDurationAsc},
			wantIDs: []string{"C", "A", "D", "B"},
		},
		{
			name:    "duration desc",
			options: []sortOption{sortDurationDesc},
			wantIDs: []string{"B", "A", "D", "C"},
		},
		{
			name:    "departure asc",
			options: []sortOption{sortDepartureAsc},
			wantIDs: []string{"B", "C", "A", "D"},
		},
		{
			name:    "departure desc",
			options: []sortOption{sortDepartureDesc},
			wantIDs: []string{"D", "A", "C", "B"},
		},
		{
			name:    "arrival asc",
			options: []sortOption{sortArrivalAsc},
			wantIDs: []string{"D", "A", "C", "B"},
		},
		{
			name:    "arrival desc",
			options: []sortOption{sortArrivalDesc},
			wantIDs: []string{"B", "C", "A", "D"},
		},
		{
			name:    "best value",
			options: []sortOption{sortBestValueAsc},
			wantIDs: []string{"C", "B", "D", "A"},
		},
		{
			name:    "best value desc",
			options: []sortOption{sortBestValueDesc},
			wantIDs: []string{"A", "D", "B", "C"},
		},
		{
			name:    "multi key price and duration asc",
			options: []sortOption{sortPriceAsc, sortDurationAsc},
			wantIDs: []string{"B", "C", "D", "A"},
		},
		{
			name:    "multi key price asc duration desc",
			options: []sortOption{sortPriceAsc, sortDurationDesc},
			wantIDs: []string{"B", "D", "C", "A"},
		},
		{
			name:    "multi key best value then price",
			options: []sortOption{sortBestValueAsc, sortPriceAsc},
			wantIDs: []string{"C", "B", "D", "A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flights := sortableFlights()
			sortFlights(flights, tt.options)

			gotIDs := make([]string, 0, len(flights))
			for _, flight := range flights {
				gotIDs = append(gotIDs, flight.ID)
			}
			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Fatalf("IDs = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestSortFlightsEmpty(t *testing.T) {
	tests := []struct {
		name    string
		options []sortOption
	}{
		{name: "price", options: []sortOption{sortPriceAsc}},
		{name: "best value", options: []sortOption{sortBestValueAsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var flights []model.Flight
			sortFlights(flights, tt.options)
			if len(flights) != 0 {
				t.Fatalf("flights = %d, want 0", len(flights))
			}
		})
	}
}

func TestBestValueScores(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := bestValueScores(nil); got != nil {
			t.Fatalf("scores = %v, want nil", got)
		}
	})

	t.Run("all values equal", func(t *testing.T) {
		flights := []model.Flight{
			{ID: "A", Price: model.Price{Amount: 100}, Duration: model.Duration{TotalMinutes: 100}, Stops: 0},
			{ID: "B", Price: model.Price{Amount: 100}, Duration: model.Duration{TotalMinutes: 100}, Stops: 0},
		}

		scores := bestValueScores(flights)
		for _, flight := range flights {
			if scores[flight.ID] != 0 {
				t.Fatalf("score for %s = %v, want 0", flight.ID, scores[flight.ID])
			}
		}
	})
}

func TestNormalizeBestValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		minValue float64
		maxValue float64
		want     float64
	}{
		{name: "minimum", value: 10, minValue: 10, maxValue: 20, want: 0},
		{name: "middle", value: 15, minValue: 10, maxValue: 20, want: 0.5},
		{name: "maximum", value: 20, minValue: 10, maxValue: 20, want: 1},
		{name: "equal range", value: 10, minValue: 10, maxValue: 10, want: 0},
		{name: "invalid range", value: 10, minValue: 20, maxValue: 10, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeBestValue(tt.value, tt.minValue, tt.maxValue); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequiresBestValue(t *testing.T) {
	tests := []struct {
		name    string
		options []sortOption
		want    bool
	}{
		{name: "nil", options: nil, want: false},
		{name: "price", options: []sortOption{sortPriceAsc}, want: false},
		{name: "best value asc", options: []sortOption{sortBestValueAsc}, want: true},
		{name: "best value desc", options: []sortOption{sortBestValueDesc}, want: true},
		{name: "mixed", options: []sortOption{sortPriceAsc, sortBestValueAsc}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requiresBestValue(tt.options); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareInt64(t *testing.T) {
	tests := []struct {
		a    int64
		b    int64
		want int
	}{
		{a: 1, b: 2, want: -1},
		{a: 2, b: 1, want: 1},
		{a: 1, b: 1, want: 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := compareInt64(tt.a, tt.b); got != tt.want {
				t.Fatalf("compareInt64(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCompareFloat(t *testing.T) {
	tests := []struct {
		a    float64
		b    float64
		want int
	}{
		{a: 1, b: 2, want: -1},
		{a: 2, b: 1, want: 1},
		{a: 1, b: 1, want: 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := compareFloat(tt.a, tt.b); got != tt.want {
				t.Fatalf("compareFloat(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
