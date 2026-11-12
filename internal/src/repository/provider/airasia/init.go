package airasia

import "github.com/flight/internal/src/repository/provider"

type airasia struct {
	dataPath string
}

func Init(path string) provider.Provider {
	return &airasia{
		dataPath: path,
	}
}
