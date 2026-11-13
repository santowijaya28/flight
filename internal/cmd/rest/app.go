package main

import (
	"net/http"
	"time"

	"github.com/flight/internal/src/cache"
	"github.com/flight/internal/src/config"
	"github.com/flight/internal/src/delivery"
	"github.com/flight/internal/src/model"
	"github.com/flight/internal/src/repository"
	"github.com/flight/internal/src/repository/provider"
	"github.com/flight/internal/src/repository/provider/airasia"
	batikair "github.com/flight/internal/src/repository/provider/batik-air"
	garudaindonesia "github.com/flight/internal/src/repository/provider/garuda-indonesia"
	lionair "github.com/flight/internal/src/repository/provider/lion-air"
	"github.com/flight/internal/src/usecase"
	"github.com/flight/pkg/log"
	"github.com/gorilla/mux"
)

func main() {
	log.Init()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal("failed to init config : ", err)
		return
	}

	// init provider
	airAsiaProvider := withProviderResilience(airasia.Init(cfg.MockDataPath.AirAsia), cfg.Provider)
	batikAirProvider := withProviderResilience(batikair.Init(cfg.MockDataPath.BatikAir), cfg.Provider)
	garudaIndonesiaProvider := withProviderResilience(garudaindonesia.Init(cfg.MockDataPath.GarudaIndonesia), cfg.Provider)
	lionAirProvider := withProviderResilience(lionair.Init(cfg.MockDataPath.LionAir), cfg.Provider)

	// initRepo
	repo := repository.Init(
		map[string]provider.Provider{
			model.ProviderAirAsia:         airAsiaProvider,
			model.ProviderBatikAir:        batikAirProvider,
			model.ProviderGarudaIndonesia: garudaIndonesiaProvider,
			model.ProviderLionAir:         lionAirProvider,
		},
	)

	// init cache
	searchCache, err := cache.New[model.SearchResult](
		cfg.Cache.MaxEntries,
		time.Duration(cfg.Cache.TTLSeconds)*time.Second,
	)
	if err != nil {
		log.Fatal("failed to init search cache: ", err)
		return
	}

	// init usecase
	usecase := usecase.Init(repo, searchCache)

	// init delivery
	dlv := delivery.Init(usecase)

	r := mux.NewRouter()

	r.Handle("/ping", http.HandlerFunc(dlv.Ping)).Methods("GET")
	r.Handle("/search", http.HandlerFunc(dlv.Search)).Methods("GET")

	// start server
	log.Info("Starting the server...")
	err = http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("can't start server", err)
	}
}

func withProviderResilience(next provider.Provider, cfg config.ProviderConfig) provider.Provider {
	return provider.WithResilience(
		next,
		time.Duration(cfg.TimeoutMs)*time.Millisecond,
		cfg.MaxAttempts,
		time.Duration(cfg.InitialBackoffMs)*time.Millisecond,
		time.Duration(cfg.MaxBackoffMs)*time.Millisecond,
	)
}
