package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/usecase/mocks"
	"github.com/flight/pkg/log"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	log.Init()
	os.Exit(m.Run())
}

func TestPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	handler := Init(mocks.NewMockUsecase(ctrl))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)

	handler.Ping(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "pong" {
		t.Fatalf("body = %q, want %q", recorder.Body.String(), "pong")
	}
}

func TestSearch(t *testing.T) {
	departureDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	successResponse := model.SearchResponse{
		SearchCriteria: model.SearchCriteria{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
			Passengers:    1,
			CabinClass:    "economy",
		},
		Metadata: model.SearchMetadata{
			TotalResults: 1,
			CacheHit:     true,
		},
		Flights: []model.Flight{{ID: "test-flight"}},
	}

	tests := []struct {
		name       string
		query      url.Values
		response   model.SearchResponse
		searchErr  error
		wantStatus int
		wantCalled bool
		wantInput  *model.SearchInput
	}{
		{
			name: "success with filters",
			query: url.Values{
				"origin":         []string{"CGK"},
				"destination":    []string{"DPS"},
				"departure_date": []string{"2025-12-15"},
				"passengers":     []string{"1"},
				"cabin_class":    []string{"economy"},
				"sort":           []string{"price_asc"},
				"min_price":      []string{"500000"},
				"max_stops":      []string{"1"},
				"airlines":       []string{"QZ,JT"},
			},
			response:   successResponse,
			wantStatus: http.StatusOK,
			wantCalled: true,
			wantInput: &model.SearchInput{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: departureDate,
				Passengers:    1,
				CabinClass:    "economy",
				Sort:          "price_asc",
				MinPrice:      ptr(500000.0),
				MaxStops:      ptr(1),
				Airlines:      []string{"QZ", "JT"},
			},
		},
		{
			name: "missing departure date",
			query: url.Values{
				"origin":      []string{"CGK"},
				"destination": []string{"DPS"},
				"passengers":  []string{"1"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid departure date",
			query: url.Values{
				"origin":         []string{"CGK"},
				"destination":    []string{"DPS"},
				"departure_date": []string{"15-12-2025"},
				"passengers":     []string{"1"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "zero passengers",
			query: url.Values{
				"origin":         []string{"CGK"},
				"destination":    []string{"DPS"},
				"departure_date": []string{"2025-12-15"},
				"passengers":     []string{"0"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid passengers format",
			query: url.Values{
				"origin":         []string{"CGK"},
				"destination":    []string{"DPS"},
				"departure_date": []string{"2025-12-15"},
				"passengers":     []string{"abc"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid filter",
			query: url.Values{
				"origin":         []string{"CGK"},
				"destination":    []string{"DPS"},
				"departure_date": []string{"2025-12-15"},
				"passengers":     []string{"1"},
				"min_price":      []string{"not-a-number"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "usecase invalid input",
			query: url.Values{
				"origin":         []string{"CGK"},
				"destination":    []string{"DPS"},
				"departure_date": []string{"2025-12-15"},
				"passengers":     []string{"1"},
			},
			searchErr:  model.ErrInvalidInput,
			wantStatus: http.StatusBadRequest,
			wantCalled: true,
		},
		{
			name: "usecase internal error",
			query: url.Values{
				"origin":         []string{"CGK"},
				"destination":    []string{"DPS"},
				"departure_date": []string{"2025-12-15"},
				"passengers":     []string{"1"},
			},
			searchErr:  errors.New("unexpected failure"),
			wantStatus: http.StatusInternalServerError,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := mocks.NewMockUsecase(ctrl)

			var gotInput model.SearchInput
			if tt.wantCalled {
				uc.EXPECT().
					Search(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, in model.SearchInput) (model.SearchResponse, error) {
						gotInput = in
						return tt.response, tt.searchErr
					}).
					Times(1)
			}

			handler := Init(uc)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/search?"+tt.query.Encode(), nil)

			handler.Search(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if tt.wantInput != nil && !reflect.DeepEqual(gotInput, *tt.wantInput) {
				t.Fatalf("input = %+v, want %+v", gotInput, *tt.wantInput)
			}

			if tt.wantStatus >= http.StatusBadRequest {
				var body struct {
					Header struct {
						Error string `json:"error"`
					} `json:"header"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatalf("invalid error JSON: %v", err)
				}
				if body.Header.Error == "" {
					t.Fatal("error message is empty")
				}
				return
			}

			var body model.SearchResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid success JSON: %v", err)
			}
			if !reflect.DeepEqual(body, successResponse) {
				t.Fatalf("body = %+v, want %+v", body, successResponse)
			}
		})
	}
}
