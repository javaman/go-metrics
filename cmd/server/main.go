package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/db"
	"github.com/javaman/go-metrics/internal/handlers"
	"github.com/javaman/go-metrics/internal/repository"
	"github.com/javaman/go-metrics/internal/services"
)

func configureWithDatabase(cfg *config.ServerConfiguration) services.MetricsService {
	database := db.New(cfg.DBDsn)
	storage := repository.NewDatabaseStorage(database)
	return services.NewMetricsService(storage, database)
}

func configureInMemtory(cfg *config.ServerConfiguration) services.MetricsService {
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

	return services.NewMetricsService(storage, db.NewStub())
}

func main() {

	cfg := config.ConfigureServer()
	var service services.MetricsService

	if cfg.DBDsn != "" {
		service = configureWithDatabase(cfg)
	} else {
		service = configureInMemtory(cfg)
	}

	e := handlers.New(service)

	e.Logger.Fatal(e.Start(cfg.Address))
}
