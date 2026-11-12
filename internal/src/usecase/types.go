package usecase

import (
	"context"

	"github.com/flight/internal/src/model"
)

type Usecase interface {
	Search(ctx context.Context, in model.SearchInput) (res model.SearchResponse, err error)
}
