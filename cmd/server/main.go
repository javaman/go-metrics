package main

import (
	"database/sql"
	"database/sql/driver"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/handlers"
	"github.com/javaman/go-metrics/internal/repository"
	"github.com/javaman/go-metrics/internal/services"
)

func configureWithDatabase(cfg *config.ServerConfiguration) (services.MetricsService, driver.Pinger) {
	db, _ := sql.Open("pgx", cfg.DBDsn)
	storage := repository.NewDatabaseStorage(db)
	return services.NewMetricsService(storage), storage
}

func configureInMemtory(cfg *config.ServerConfiguration) (services.MetricsService, driver.Pinger) {
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

	return services.NewMetricsService(storage), storage
}

func main() {
	cfg := config.ConfigureServer()
	var service services.MetricsService
	var ping driver.Pinger

	if cfg.DBDsn != "" {
		service, ping = configureWithDatabase(cfg)
	} else {
		service, ping = configureInMemtory(cfg)
	}

	e := handlers.New(service, ping)

	e.Logger.Fatal(e.Start(cfg.Address))
}
