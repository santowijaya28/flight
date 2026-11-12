package repository

import (
	"context"

	"github.com/flight/internal/src/model"
)

type ProviderRepo interface {
	Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error)
}
