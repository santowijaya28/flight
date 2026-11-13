package usecase

import (
	"github.com/flight/internal/src/cache"
	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository"
)

type usecase struct {
	repo  repository.ProviderRepo
	cache *cache.Cache[model.SearchResult]
}

func Init(repo repository.ProviderRepo) *usecase {
	return &usecase{
		repo:  repo,
		cache: cache.New[model.SearchResult](searchCacheTTL),
	}
}
