package provider

import (
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	mockProvider := make(map[string]Provider)
	tests := []struct {
		name      string
		providers map[string]Provider
		want      *provider
	}{
		{
			name:      "success",
			providers: mockProvider,
			want: &provider{
				providers: mockProvider,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Init(tt.providers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}
