package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flight/internal/src/model"
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

func (d *delivery) Search(w http.ResponseWriter, r *http.Request) {
	param := model.SearchInput{}

	param.Origin = r.URL.Query().Get("origin")
	param.Destination = r.URL.Query().Get("destination")
	dt := r.URL.Query().Get("departure_date")
	if dt != "" {
		dd, err := time.Parse("2006-01-02", dt)
		if err != nil {
			err = fmt.Errorf("invalid departure_time format: format should be 'YYYY-MM-DD' : %+w", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		param.DepartureDate = dd
	}

	pass := r.URL.Query().Get("passengers")
	if pass != "" {
		passenger, err := strconv.Atoi(pass)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		param.Passengers = passenger
	}
	param.CabinClass = r.URL.Query().Get("cabin_class")

	resp, err := d.useCase.Search(r.Context(), param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
