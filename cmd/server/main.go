package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/handlers"
	"github.com/javaman/go-metrics/internal/repository"
	"github.com/javaman/go-metrics/internal/services"
)

func configureWithDatabase(cfg *config.ServerConfiguration) (services.MetricsService, func() error) {
	db, _ := sql.Open("pgx", cfg.DBDsn)
	storage := repository.NewDatabaseStorage(db)
	return services.NewMetricsService(storage), repository.PingDB(db)
}

func configureInMemtory(cfg *config.ServerConfiguration) (services.MetricsService, func() error) {
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

	return services.NewMetricsService(storage), func() error { return nil }
}

func main() {
	cfg := config.ConfigureServer()
	var service services.MetricsService
	var ping func() error

	if cfg.DBDsn != "" {
		service, ping = configureWithDatabase(cfg)
	} else {
		service, ping = configureInMemtory(cfg)
	}

	e := handlers.New(service, ping)

	e.Logger.Fatal(e.Start(cfg.Address))
}
