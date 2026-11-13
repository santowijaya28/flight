package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/flight/internal/src/cache"
	"github.com/flight/internal/src/model"
	repomocks "github.com/flight/internal/src/repository/mocks"
	"go.uber.org/mock/gomock"
)

func searchInput() model.SearchInput {
	return model.SearchInput{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		Passengers:    1,
		CabinClass:    "economy",
		Sort:          "price_asc",
	}
}

func searchFlight(id string) model.Flight {
	return model.Flight{
		ID:       id,
		Provider: model.ProviderAirAsia,
		Airline: model.AirLine{
			Name: "AirAsia",
			Code: "QZ",
		},
		FlightNumber: id,
		Departure: model.Departure{
			Airport:   "CGK",
			City:      "Jakarta",
			Timezone:  model.TimezoneWIB,
			DateTime:  "2025-12-15T04:45:00+07:00",
			Timestamp: 1765748700,
		},
		Arrival: model.Arrival{
			Airport:   "DPS",
			City:      "Denpasar",
			Timezone:  model.TimezoneWITA,
			DateTime:  "2025-12-15T07:25:00+08:00",
			Timestamp: 1765754700,
		},
		Duration: model.Duration{
			TotalMinutes: 100,
			Formatted:    "1h 40m",
		},
		Stops: 0,
		Price: model.Price{
			Amount:    650000,
			Currency:  "IDR",
			Formatted: "Rp 650.000",
		},
		AvailableSeats: 67,
		CabinClass:     "economy",
		Amenities:      []string{},
		Baggage: model.Baggage{
			CarryOn: "Cabin baggage only",
			Checked: "Additional fee",
		},
	}
}

func searchResult(ids ...string) model.SearchResult {
	flights := make([]model.Flight, 0, len(ids))
	for _, id := range ids {
		flights = append(flights, searchFlight(id))
	}

	return model.SearchResult{
		TotalResults:       len(ids),
		ProvidersQueried:   2,
		ProvidersSucceeded: 2,
		Flights:            flights,
	}
}

func withoutFlightTimes(flights []model.Flight) []model.Flight {
	result := make([]model.Flight, len(flights))
	copy(result, flights)

	for i := range result {
		result[i].Departure.DateTime = ""
		result[i].Departure.Timestamp = 0
		result[i].Arrival.DateTime = ""
		result[i].Arrival.Timestamp = 0
	}

	return result
}

func partialSearchResult() model.SearchResult {
	result := searchResult("A")
	result.ProvidersSucceeded = 1
	result.ProvidersFailed = 1
	return result
}

func TestSearch(t *testing.T) {
	tests := []struct {
		name          string
		input         model.SearchInput
		repoResult    model.SearchResult
		repoErr       error
		setupCache    bool
		cachedResult  model.SearchResult
		wantErr       error
		wantMetadata  model.SearchMetadata
		wantCacheHit  bool
		wantRepoCalls int
	}{
		{
			name:       "successful search",
			input:      searchInput(),
			repoResult: searchResult("A", "B", "C"),
			wantMetadata: model.SearchMetadata{
				TotalResults:       3,
				ProvidersQueried:   2,
				ProvidersSucceeded: 2,
			},
			wantRepoCalls: 1,
		},
		{
			name:         "cache hit",
			input:        searchInput(),
			setupCache:   true,
			cachedResult: searchResult("cached"),
			wantMetadata: model.SearchMetadata{
				TotalResults:       1,
				ProvidersQueried:   2,
				ProvidersSucceeded: 2,
				CacheHit:           true,
			},
			wantCacheHit:  true,
			wantRepoCalls: 0,
		},
		{
			name:          "invalid sort",
			input:         model.SearchInput{Sort: "invalid"},
			wantErr:       model.ErrInvalidInput,
			wantRepoCalls: 0,
		},
		{
			name:          "all providers down",
			input:         searchInput(),
			repoErr:       model.ErrAllProviderDown,
			wantErr:       model.ErrAllProviderDown,
			wantRepoCalls: 1,
		},
		{
			name:       "partial provider failure",
			input:      searchInput(),
			repoResult: partialSearchResult(),
			repoErr:    errors.New("provider failed"),
			wantMetadata: model.SearchMetadata{
				TotalResults:       1,
				ProvidersQueried:   2,
				ProvidersSucceeded: 1,
				ProvidersFailed:    1,
			},
			wantRepoCalls: 1,
		},
		{
			name:       "successful empty result",
			input:      searchInput(),
			repoResult: searchResult(),
			wantMetadata: model.SearchMetadata{
				ProvidersQueried:   2,
				ProvidersSucceeded: 2,
			},
			wantRepoCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repomocks.NewMockProviderRepo(ctrl)

			searchCache, err := cache.New[model.SearchResult](10, time.Minute)
			if err != nil {
				t.Fatal(err)
			}

			if tt.setupCache {
				key, err := buildCacheKey(tt.input)
				if err != nil {
					t.Fatal(err)
				}
				searchCache.Set(key, tt.cachedResult)
			}

			if tt.wantRepoCalls > 0 {
				repo.EXPECT().
					Search(gomock.Any(), tt.input).
					Return(tt.repoResult, tt.repoErr).
					Times(tt.wantRepoCalls)
			}

			u := Init(repo, searchCache)
			response, err := u.Search(context.Background(), tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wantFlights := tt.repoResult.Flights
			if tt.setupCache {
				wantFlights = tt.cachedResult.Flights
			}
			if !reflect.DeepEqual(withoutFlightTimes(response.Flights), withoutFlightTimes(wantFlights)) {
				t.Fatalf("flights = %+v, want %+v", response.Flights, wantFlights)
			}
			if response.Metadata.CacheHit != tt.wantCacheHit {
				t.Fatalf("cache hit = %v, want %v", response.Metadata.CacheHit, tt.wantCacheHit)
			}
			gotMetadata := response.Metadata
			gotMetadata.SearchTimeMs = 0
			if !reflect.DeepEqual(gotMetadata, tt.wantMetadata) {
				t.Fatalf("metadata = %+v, want %+v", gotMetadata, tt.wantMetadata)
			}
			wantCriteria := buildSearchCriteria(tt.input)
			if !reflect.DeepEqual(response.SearchCriteria, wantCriteria) {
				t.Fatalf("criteria = %+v, want %+v", response.SearchCriteria, wantCriteria)
			}
		})
	}
}

func TestSearchDoesNotCacheUncacheableResults(t *testing.T) {
	tests := []struct {
		name         string
		repoResult   model.SearchResult
		repoErr      error
		wantMetadata model.SearchMetadata
	}{
		{
			name:       "partial provider failure",
			repoResult: partialSearchResult(),
			repoErr:    errors.New("provider failed"),
			wantMetadata: model.SearchMetadata{
				TotalResults:       1,
				ProvidersQueried:   2,
				ProvidersSucceeded: 1,
				ProvidersFailed:    1,
			},
		},
		{
			name:       "successful empty result",
			repoResult: searchResult(),
			wantMetadata: model.SearchMetadata{
				ProvidersQueried:   2,
				ProvidersSucceeded: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repomocks.NewMockProviderRepo(ctrl)
			input := searchInput()

			repo.EXPECT().
				Search(gomock.Any(), input).
				Return(tt.repoResult, tt.repoErr).
				Times(2)

			searchCache, err := cache.New[model.SearchResult](10, time.Minute)
			if err != nil {
				t.Fatal(err)
			}

			u := Init(repo, searchCache)
			for i := 0; i < 2; i++ {
				response, err := u.Search(context.Background(), input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if response.Metadata.CacheHit {
					t.Fatal("cache hit = true, want false")
				}

				if !reflect.DeepEqual(
					withoutFlightTimes(response.Flights),
					withoutFlightTimes(tt.repoResult.Flights),
				) {
					t.Fatalf("flights = %+v, want %+v", response.Flights, tt.repoResult.Flights)
				}

				gotMetadata := response.Metadata
				gotMetadata.SearchTimeMs = 0
				if !reflect.DeepEqual(gotMetadata, tt.wantMetadata) {
					t.Fatalf("metadata = %+v, want %+v", gotMetadata, tt.wantMetadata)
				}
			}
		})
	}
}
