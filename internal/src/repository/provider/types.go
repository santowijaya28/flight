package provider

import (
	"context"

	"github.com/flight/internal/src/model"
)

type Provider interface {
	Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error)
}
