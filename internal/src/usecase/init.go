package usecase

import "github.com/flight/internal/src/repository"

type usecase struct {
	repo repository.ProviderRepo
}

func Init(repo repository.ProviderRepo) *usecase {
	return &usecase{
		repo: repo,
	}
}
