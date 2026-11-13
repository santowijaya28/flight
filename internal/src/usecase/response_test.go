package usecase

import (
	"reflect"
	"testing"
	"time"

	"github.com/flight/internal/src/model"
)

func Test_buildSearchResponse(t *testing.T) {
	tests := []struct {
		name         string
		in           model.SearchInput
		res          model.SearchResult
		cacheHit     bool
		searchTimeMs int64
		want         model.SearchResponse
	}{
		{
			name: "success",
			in: model.SearchInput{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
				Passengers:    1,
				CabinClass:    "economy",
			},
			res: model.SearchResult{
				TotalResults:       1,
				ProvidersQueried:   4,
				ProvidersSucceeded: 4,
				ProvidersFailed:    0,
				SearchTimeMs:       1,
				Flights: []model.Flight{
					{
						ID:           "TESTFLIGHT",
						Provider:     model.ProviderAirAsia,
						FlightNumber: "TESTNUMBER",
					},
				},
			},
			want: model.SearchResponse{
				SearchCriteria: buildSearchCriteria(model.SearchInput{
					Origin:        "CGK",
					Destination:   "DPS",
					DepartureDate: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
					Passengers:    1,
					CabinClass:    "economy",
				}),
				Metadata: model.SearchMetadata{
					TotalResults:       1,
					ProvidersQueried:   4,
					ProvidersSucceeded: 4,
					ProvidersFailed:    0,
				},
				Flights: []model.Flight{
					{
						ID:           "TESTFLIGHT",
						Provider:     model.ProviderAirAsia,
						FlightNumber: "TESTNUMBER",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSearchResponse(tt.in, tt.res, tt.cacheHit, tt.searchTimeMs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildSearchResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
