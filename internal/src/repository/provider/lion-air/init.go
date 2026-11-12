package lionair

import "github.com/flight/internal/src/repository/provider"

type lr struct {
	dataPath string
}

func Init(path string) provider.Provider {
	return &lr{
		dataPath: path,
	}
}
