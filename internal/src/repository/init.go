package repository

import "github.com/flight/internal/src/repository/provider"

type repository struct {
	flightProvider map[string]provider.Provider
}

func Init(provider map[string]provider.Provider) *repository {
	return &repository{
		flightProvider: provider,
	}
}
