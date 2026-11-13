package repository

import (
	"context"
	"errors"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/provider"
	"github.com/flight/internal/src/repository/provider/mocks"
	"github.com/flight/pkg/log"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	log.Init()
	os.Exit(m.Run())
}

type providerCase struct {
	result model.SearchResult
	err    error
}

func TestSearch(t *testing.T) {
	tests := []struct {
		name          string
		providers     map[string]providerCase
		wantIDs       []string
		wantQueried   int
		wantSucceeded int
		wantFailed    int
		wantErr       bool
		wantAllDown   bool
	}{
		{
			name: "all providers succeed",
			providers: map[string]providerCase{
				"provider-a": {
					result: model.SearchResult{
						Flights: []model.Flight{{ID: "A1"}, {ID: "A2"}},
					},
				},
				"provider-b": {
					result: model.SearchResult{
						Flights: []model.Flight{{ID: "B1"}},
					},
				},
			},
			wantIDs:       []string{"A1", "A2", "B1"},
			wantQueried:   2,
			wantSucceeded: 2,
		},
		{
			name: "one provider fails",
			providers: map[string]providerCase{
				"provider-a": {
					result: model.SearchResult{
						Flights: []model.Flight{{ID: "A1"}},
					},
				},
				"provider-b": {
					err: errors.New("provider b failed"),
				},
			},
			wantIDs:       []string{"A1"},
			wantQueried:   2,
			wantSucceeded: 1,
			wantFailed:    1,
			wantErr:       true,
		},
		{
			name: "all providers fail",
			providers: map[string]providerCase{
				"provider-a": {err: errors.New("provider a failed")},
				"provider-b": {err: errors.New("provider b failed")},
			},
			wantIDs:       []string{},
			wantQueried:   2,
			wantSucceeded: 0,
			wantFailed:    2,
			wantErr:       true,
			wantAllDown:   true,
		},
		{
			name: "provider error still keeps partial result",
			providers: map[string]providerCase{
				"provider-a": {
					result: model.SearchResult{
						Flights: []model.Flight{{ID: "A1"}},
					},
					err: errors.New("provider a failed after partial result"),
				},
				"provider-b": {
					result: model.SearchResult{
						Flights: []model.Flight{{ID: "B1"}},
					},
				},
			},
			wantIDs:       []string{"A1", "B1"},
			wantQueried:   2,
			wantSucceeded: 1,
			wantFailed:    1,
			wantErr:       true,
		},
		{
			name:          "empty provider map",
			providers:     map[string]providerCase{},
			wantIDs:       []string{},
			wantQueried:   0,
			wantSucceeded: 0,
			wantFailed:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			providers := make(map[string]provider.Provider, len(tt.providers))

			for name, testProvider := range tt.providers {
				mockProvider := mocks.NewMockProvider(ctrl)
				mockProvider.EXPECT().
					Search(gomock.Any(), gomock.Any()).
					Return(testProvider.result, testProvider.err).
					Times(1)
				providers[name] = mockProvider
			}

			repo := Init(providers)
			result, err := repo.Search(context.Background(), model.SearchInput{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			isAllDown := errors.Is(err, model.ErrAllProviderDown)
			if isAllDown != tt.wantAllDown {
				t.Fatalf("all provider down = %v, want %v", isAllDown, tt.wantAllDown)
			}

			gotIDs := make([]string, 0, len(result.Flights))
			for _, flight := range result.Flights {
				gotIDs = append(gotIDs, flight.ID)
			}
			sort.Strings(gotIDs)
			sort.Strings(tt.wantIDs)

			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Fatalf("flight IDs = %v, want %v", gotIDs, tt.wantIDs)
			}
			if result.TotalResults != len(tt.wantIDs) {
				t.Fatalf("total results = %d, want %d", result.TotalResults, len(tt.wantIDs))
			}
			if result.ProvidersQueried != tt.wantQueried {
				t.Fatalf("providers queried = %d, want %d", result.ProvidersQueried, tt.wantQueried)
			}
			if result.ProvidersSucceeded != tt.wantSucceeded {
				t.Fatalf("providers succeeded = %d, want %d", result.ProvidersSucceeded, tt.wantSucceeded)
			}
			if result.ProvidersFailed != tt.wantFailed {
				t.Fatalf("providers failed = %d, want %d", result.ProvidersFailed, tt.wantFailed)
			}
		})
	}
}
