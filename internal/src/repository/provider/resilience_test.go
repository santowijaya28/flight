package provider

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/provider/mocks"
	"github.com/flight/pkg/simulation"
	"go.uber.org/mock/gomock"
)

func Test_isRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "error is upstream failure simulation",
			err:  simulation.ErrSimulatedUpstreamFailure,
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("random error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryable(tt.err); got != tt.want {
				t.Errorf("isRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithResilience(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProvider := mocks.NewMockProvider(ctrl)

	want := model.SearchResult{
		TotalResults: 1,
		Flights:      []model.Flight{{ID: "test-flight"}},
	}

	mockProvider.EXPECT().
		Search(gomock.Any(), gomock.Any()).
		Return(want, nil).
		Times(1)

	wrapped := WithResilience(
		mockProvider,
		time.Second,
		3,
		time.Millisecond,
		10*time.Millisecond,
	)

	got, err := wrapped.Search(context.Background(), model.SearchInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %+v, want %+v", got, want)
	}
}

func TestWithResilienceNilProvider(t *testing.T) {
	got := WithResilience(
		nil,
		time.Second,
		3,
		time.Millisecond,
		10*time.Millisecond,
	)
	if got != nil {
		t.Fatalf("result = %v, want nil", got)
	}
}
