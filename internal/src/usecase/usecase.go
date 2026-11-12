package usecase

import (
	"context"

	"github.com/flight/internal/src/model"
)

func (u *usecase) Search(ctx context.Context, in model.SearchInput) (res model.SearchResult, err error) {
	res, err = u.repo.Search(ctx, model.SearchInput{})

	return
}
