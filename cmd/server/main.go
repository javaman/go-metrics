package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/db"
	"github.com/javaman/go-metrics/internal/handlers"
	"github.com/javaman/go-metrics/internal/repository"
	"github.com/javaman/go-metrics/internal/services"
)

func main() {

	cfg := config.ConfigureServer()

	var storage repository.Storage

	if cfg.Restore {
		storage = repository.NewInMemoryStorageFromFile(cfg.FileStoragePath)
	} else {
		storage = repository.NewInMemoryStorage()
	}

	if cfg.StoreInterval > 0 {
		services.FlushStorageInBackground(storage, cfg.FileStoragePath, cfg.StoreInterval)
	} else {
		storage = repository.MakeStorageFlushedOnEachCall(storage, cfg.FileStoragePath)
	}

	service := services.NewMetricsService(storage, db.New(cfg.DBDsn))

	e := handlers.New(service)

	e.Logger.Fatal(e.Start(cfg.Address))
}
