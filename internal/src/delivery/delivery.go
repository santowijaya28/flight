package delivery

import (
	"net/http"

	"github.com/flight/internal/src/usecase"
)

type delivery struct {
	useCase usecase.Usecase
}

func Init(useCase usecase.Usecase) *delivery {
	return &delivery{
		useCase: useCase,
	}
}

func (d *delivery) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
