package delivery

import (
	"net/http"

	"github.com/flight/internal/src/repository"
)

type delivery struct {
}

func Init(repo repository.ProviderRepo) *delivery {
	return &delivery{}
}

func (d *delivery) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
