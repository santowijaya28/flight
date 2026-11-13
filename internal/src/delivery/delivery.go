package delivery

import (
	"errors"
	"fmt"
	"net/http"
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
		passenger, err := parseOptionalInt(pass, "passengers")
		if err != nil {
			response.WriteErrorResponse(w, response.Response{
				Data:           nil,
				Error:          err.Error(),
				HttpStatusCode: http.StatusBadRequest,
			})
			return
		}
		param.Passengers = *passenger
	}
	param.CabinClass = r.URL.Query().Get("cabin_class")
	param.Sort = r.URL.Query().Get("sort")

	if err := parseFilters(r.URL.Query(), &param); err != nil {
		response.WriteErrorResponse(w, response.Response{
			Data:           nil,
			Error:          err.Error(),
			HttpStatusCode: http.StatusBadRequest,
		})
		return
	}

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
		statusCode := http.StatusInternalServerError
		if errors.Is(err, model.ErrInvalidInput) {
			statusCode = http.StatusBadRequest
		}

		log.Errorf("failed to search flight %s", err.Error())
		response.WriteErrorResponse(w, response.Response{
			Data:           nil,
			Error:          err.Error(),
			HttpStatusCode: statusCode,
		})
		return
	}

	response.WriteSuccessResponse(w, response.Response{
		Data: resp,
	})
}
