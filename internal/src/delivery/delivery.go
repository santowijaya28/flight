package delivery

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func parseFilters(query url.Values, param *model.SearchInput) error {
	var err error

	param.MinPrice, err = parseOptionalFloat(query.Get("min_price"), "min_price")
	if err != nil {
		return err
	}

	param.MaxPrice, err = parseOptionalFloat(query.Get("max_price"), "max_price")
	if err != nil {
		return err
	}

	param.MaxStops, err = parseOptionalInt(query.Get("max_stops"), "max_stops")
	if err != nil {
		return err
	}

	param.MinDurationMinutes, err = parseOptionalInt64(query.Get("min_duration"), "min_duration")
	if err != nil {
		return err
	}

	param.MaxDurationMinutes, err = parseOptionalInt64(query.Get("max_duration"), "max_duration")
	if err != nil {
		return err
	}

	param.DepartureFrom, err = parseClock(query.Get("departure_from"), "departure_from")
	if err != nil {
		return err
	}

	param.DepartureTo, err = parseClock(query.Get("departure_to"), "departure_to")
	if err != nil {
		return err
	}

	param.ArrivalFrom, err = parseClock(query.Get("arrival_from"), "arrival_from")
	if err != nil {
		return err
	}

	param.ArrivalTo, err = parseClock(query.Get("arrival_to"), "arrival_to")
	if err != nil {
		return err
	}

	for _, value := range query["airlines"] {
		for _, airline := range strings.Split(value, ",") {
			airline = strings.TrimSpace(airline)
			if airline != "" {
				param.Airlines = append(param.Airlines, airline)
			}
		}
	}

	return nil
}

func parseOptionalFloat(value, name string) (*float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("%w: invalid %s", model.ErrInvalidInput, name)
	}

	return &parsed, nil
}

func parseOptionalInt(value, name string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("%w: invalid %s", model.ErrInvalidInput, name)
	}

	return &parsed, nil
}

func parseOptionalInt64(value, name string) (*int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("%w: invalid %s", model.ErrInvalidInput, name)
	}

	return &parsed, nil
}

func parseClock(value, name string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	if _, err := time.Parse("15:04", value); err != nil {
		return "", fmt.Errorf("%w: invalid %s, expected HH:MM", model.ErrInvalidInput, name)
	}

	return value, nil
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

	if param.MinPrice != nil && param.MaxPrice != nil && *param.MinPrice > *param.MaxPrice {
		return fmt.Errorf("%w: min_price cannot exceed max_price", model.ErrInvalidInput)
	}

	if param.MinDurationMinutes != nil && param.MaxDurationMinutes != nil &&
		*param.MinDurationMinutes > *param.MaxDurationMinutes {
		return fmt.Errorf("%w: min_duration cannot exceed max_duration", model.ErrInvalidInput)
	}

	return nil
}
