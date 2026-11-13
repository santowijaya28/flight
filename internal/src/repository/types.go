package repository

import (
	"context"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository/provider"
)

type providerResult struct {
	name string
	res  model.SearchResult
	err  error
}

type repository struct {
	flightProvider map[string]provider.Provider
}

//go:generate go run go.uber.org/mock/mockgen -source=types.go -destination=mocks/mock_provider_repo.go -package=mocks

type ProviderRepo interface {
	Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error)
}
