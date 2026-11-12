package repository

import "github.com/flight/internal/src/repository/provider"

func Init(provider map[string]provider.Provider) *repository {
	return &repository{
		flightProvider: provider,
	}
}
