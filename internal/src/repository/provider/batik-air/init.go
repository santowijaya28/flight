package batikair

import "github.com/flight/internal/src/repository/provider"

type batikAir struct {
	dataPath string
}

func Init(path string) provider.Provider {
	return &batikAir{
		dataPath: path,
	}
}
