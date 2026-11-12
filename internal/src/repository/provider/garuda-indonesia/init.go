package garudaindonesia

import "github.com/flight/internal/src/repository/provider"

type gi struct {
	datapath string
}

func Init(path string) provider.Provider {
	return &gi{
		datapath: path,
	}
}
