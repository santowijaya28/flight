package usecase

import (
	"context"

	"github.com/flight/internal/src/model"
)

//go:generate go run go.uber.org/mock/mockgen -source=types.go -destination=mocks/mock_usecase.go -package=mocks

type Usecase interface {
	Search(ctx context.Context, in model.SearchInput) (res model.SearchResponse, err error)
}
