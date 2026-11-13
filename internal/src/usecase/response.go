package usecase

import "github.com/flight/internal/src/model"

func buildSearchResponse(
	in model.SearchInput,
	res model.SearchResult,
	cacheHit bool,
	searchTimeMs int64,
) model.SearchResponse {
	return model.SearchResponse{
		SearchCriteria: buildSearchCriteria(in),
		Metadata: model.SearchMetadata{
			TotalResults:       res.TotalResults,
			ProvidersQueried:   res.ProvidersQueried,
			ProvidersSucceeded: res.ProvidersSucceeded,
			ProvidersFailed:    res.ProvidersFailed,
			SearchTimeMs:       searchTimeMs,
			CacheHit:           cacheHit,
		},
		Flights: res.Flights,
	}
}
