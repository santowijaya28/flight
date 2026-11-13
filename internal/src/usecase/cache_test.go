package usecase

import (
	"strings"
	"testing"
	"time"

	"github.com/flight/internal/src/model"
)

func ptr[T any](value T) *T {
	return &value
}

func baseCacheInput() model.SearchInput {
	return model.SearchInput{
		Origin:             "CGK",
		Destination:        "DPS",
		DepartureDate:      time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		ReturnDate:         ptr(time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)),
		Passengers:         1,
		CabinClass:         "economy",
		Sort:               "price_asc",
		MinPrice:           ptr(500000.0),
		MaxPrice:           ptr(1500000.0),
		MaxStops:           ptr(1),
		Airlines:           []string{"QZ", "JT"},
		MinDurationMinutes: ptr(int64(90)),
		MaxDurationMinutes: ptr(int64(240)),
		DepartureFrom:      "06:00",
		DepartureTo:        "18:00",
		ArrivalFrom:        "08:00",
		ArrivalTo:          "23:00",
	}
}

func TestBuildCacheKey(t *testing.T) {
	baseKey, err := buildCacheKey(baseCacheInput())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		modify   func(*model.SearchInput)
		wantSame bool
	}{
		{name: "same input", wantSame: true},
		{
			name: "origin case and whitespace",
			modify: func(in *model.SearchInput) {
				in.Origin = " cgk "
			},
			wantSame: true,
		},
		{
			name: "destination case and whitespace",
			modify: func(in *model.SearchInput) {
				in.Destination = " dps "
			},
			wantSame: true,
		},
		{
			name: "cabin class case and whitespace",
			modify: func(in *model.SearchInput) {
				in.CabinClass = " Economy "
			},
			wantSame: true,
		},
		{
			name: "sort case and whitespace",
			modify: func(in *model.SearchInput) {
				in.Sort = " PRICE_ASC "
			},
			wantSame: true,
		},
		{
			name: "airlines order case and whitespace",
			modify: func(in *model.SearchInput) {
				in.Airlines = []string{" jt ", "qz"}
			},
			wantSame: true,
		},
		{
			name: "empty airlines",
			modify: func(in *model.SearchInput) {
				in.Airlines = []string{"", "  "}
			},
			wantSame: false,
		},
		{
			name: "nil airlines",
			modify: func(in *model.SearchInput) {
				in.Airlines = nil
			},
			wantSame: false,
		},
		{
			name: "departure window whitespace",
			modify: func(in *model.SearchInput) {
				in.DepartureFrom = " 06:00 "
				in.DepartureTo = " 18:00 "
			},
			wantSame: true,
		},
		{
			name: "arrival window whitespace",
			modify: func(in *model.SearchInput) {
				in.ArrivalFrom = " 08:00 "
				in.ArrivalTo = " 23:00 "
			},
			wantSame: true,
		},
		{
			name: "different origin",
			modify: func(in *model.SearchInput) {
				in.Origin = "SUB"
			},
		},
		{
			name: "different destination",
			modify: func(in *model.SearchInput) {
				in.Destination = "SUB"
			},
		},
		{
			name: "different departure date",
			modify: func(in *model.SearchInput) {
				in.DepartureDate = time.Date(2025, 12, 16, 0, 0, 0, 0, time.UTC)
			},
		},
		{
			name: "different return date",
			modify: func(in *model.SearchInput) {
				in.ReturnDate = ptr(time.Date(2025, 12, 21, 0, 0, 0, 0, time.UTC))
			},
		},
		{
			name: "different passengers",
			modify: func(in *model.SearchInput) {
				in.Passengers = 2
			},
		},
		{
			name: "different cabin class",
			modify: func(in *model.SearchInput) {
				in.CabinClass = "business"
			},
		},
		{
			name: "different sort",
			modify: func(in *model.SearchInput) {
				in.Sort = "duration_asc"
			},
		},
		{
			name: "different min price",
			modify: func(in *model.SearchInput) {
				in.MinPrice = ptr(600000.0)
			},
		},
		{
			name: "different max price",
			modify: func(in *model.SearchInput) {
				in.MaxPrice = ptr(1600000.0)
			},
		},
		{
			name: "different max stops",
			modify: func(in *model.SearchInput) {
				in.MaxStops = ptr(0)
			},
		},
		{
			name: "different airlines",
			modify: func(in *model.SearchInput) {
				in.Airlines = []string{"GA"}
			},
		},
		{
			name: "different min duration",
			modify: func(in *model.SearchInput) {
				in.MinDurationMinutes = ptr(int64(100))
			},
		},
		{
			name: "different max duration",
			modify: func(in *model.SearchInput) {
				in.MaxDurationMinutes = ptr(int64(300))
			},
		},
		{
			name: "different departure from",
			modify: func(in *model.SearchInput) {
				in.DepartureFrom = "07:00"
			},
		},
		{
			name: "different departure to",
			modify: func(in *model.SearchInput) {
				in.DepartureTo = "19:00"
			},
		},
		{
			name: "different arrival from",
			modify: func(in *model.SearchInput) {
				in.ArrivalFrom = "09:00"
			},
		},
		{
			name: "different arrival to",
			modify: func(in *model.SearchInput) {
				in.ArrivalTo = "22:00"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := baseCacheInput()
			if tt.modify != nil {
				tt.modify(&in)
			}

			key, err := buildCacheKey(in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantSame && key != baseKey {
				t.Fatalf("key = %q, want same as %q", key, baseKey)
			}
			if !tt.wantSame && key == baseKey {
				t.Fatalf("key = %q, want different from base key", key)
			}
		})
	}
}

func TestBuildCacheKeyPrefix(t *testing.T) {
	key, err := buildCacheKey(baseCacheInput())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "search:v1:") {
		t.Fatalf("key = %q, want search:v1: prefix", key)
	}
}
