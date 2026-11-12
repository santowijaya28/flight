package main

import (
	"net/http"

	"github.com/flight/internal/src/delivery"
	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository"
	"github.com/flight/internal/src/repository/provider"
	"github.com/flight/internal/src/repository/provider/airasia"
	batikair "github.com/flight/internal/src/repository/provider/batik-air"
	"github.com/flight/internal/src/usecase"
	"github.com/flight/pkg/log"
	"github.com/gorilla/mux"
)

func main() {
	log.Init()

	// init provider
	airAsia := airasia.Init("external/airasia_search_response.json")
	batikAir := batikair.Init("external/batik_air_search_response.json")

	// initRepo
	repo := repository.Init(
		map[string]provider.Provider{
			model.ProviderAirAsia:  airAsia,
			model.ProviderBatikAir: batikAir,
		},
	)

	// init usecase
	usecase := usecase.Init(repo)

	// init delivery
	dlv := delivery.Init(usecase)

	r := mux.NewRouter()

	r.Handle("/ping", http.HandlerFunc(dlv.Ping)).Methods("GET")
	r.Handle("/search", http.HandlerFunc(dlv.Search)).Methods("GET")

	// start server
	log.Info("Starting the server...")
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("can't start server", err)
	}
}
