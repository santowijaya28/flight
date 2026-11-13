package provider

import (
	"context"

	"github.com/flight/internal/src/model"
)

//go:generate go run go.uber.org/mock/mockgen -source=types.go -destination=mocks/mock_provider.go -package=mocks

type Provider interface {
	Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error)
}
