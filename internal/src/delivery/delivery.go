package delivery

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/usecase"
	"github.com/flight/pkg/log"
	"github.com/flight/pkg/response"
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
			response.WriteErrorResponse(w, response.Response{
				Data:           nil,
				Error:          err.Error(),
				HttpStatusCode: http.StatusBadRequest,
			})
			return
		}
		param.DepartureDate = dd
	}

	pass := r.URL.Query().Get("passengers")
	if pass != "" {
		passenger, err := strconv.Atoi(pass)
		if err != nil {
			response.WriteErrorResponse(w, response.Response{
				Data:           nil,
				Error:          err.Error(),
				HttpStatusCode: http.StatusBadRequest,
			})
			return
		}
		param.Passengers = passenger
	}
	param.CabinClass = r.URL.Query().Get("cabin_class")

	err := validateSearchInput(param)
	if err != nil {
		response.WriteErrorResponse(w, response.Response{
			Data:           nil,
			Error:          err.Error(),
			HttpStatusCode: http.StatusBadRequest,
		})
		return
	}

	resp, err := d.useCase.Search(r.Context(), param)
	if err != nil {
		log.Errorf("failed to search flight %s", err.Error())
		response.WriteErrorResponse(w, response.Response{
			Data:           nil,
			Error:          err.Error(),
			HttpStatusCode: http.StatusInternalServerError,
		})
		return
	}

	response.WriteSuccessResponse(w, response.Response{
		Data: resp,
	})
}

func validateSearchInput(param model.SearchInput) error {
	if param.DepartureDate.IsZero() {
		return fmt.Errorf("empty departure_date: %+w", model.ErrInvalidInput)
	}

	if param.Origin == "" {
		return fmt.Errorf("empty origin: %+w", model.ErrInvalidInput)
	}

	if param.Destination == "" {
		return fmt.Errorf("empty destination: %+w", model.ErrInvalidInput)
	}

	if param.Passengers <= 0 {
		return fmt.Errorf("invalid passenger: %+w", model.ErrInvalidInput)
	}

	return nil
}
