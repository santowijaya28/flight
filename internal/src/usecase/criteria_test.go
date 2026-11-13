package usecase

import (
	"reflect"
	"testing"
	"time"

	"github.com/flight/internal/src/model"
)

func TestBuildSearchCriteria(t *testing.T) {
	departureDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		input model.SearchInput
		want  model.SearchCriteria
	}{
		{
			name: "full input",
			input: model.SearchInput{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: departureDate,
				ReturnDate:    &returnDate,
				Passengers:    2,
				CabinClass:    "economy",
			},
			want: model.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				ReturnDate:    ptr("2025-12-20"),
				Passengers:    2,
				CabinClass:    "economy",
			},
		},
		{
			name: "one way",
			input: model.SearchInput{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: departureDate,
				Passengers:    1,
			},
			want: model.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
			},
		},
		{
			name:  "zero departure date",
			input: model.SearchInput{},
			want:  model.SearchCriteria{},
		},
		{
			name: "date with non utc location",
			input: model.SearchInput{
				DepartureDate: time.Date(
					2025, 12, 16, 0, 0, 0, 0,
					time.FixedZone("WIB", 7*60*60),
				),
			},
			want: model.SearchCriteria{
				DepartureDate: "2025-12-16",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSearchCriteria(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
