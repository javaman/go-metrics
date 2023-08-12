package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/domain"
	"github.com/javaman/go-metrics/internal/metric/delivery/http"
	"github.com/javaman/go-metrics/internal/metric/delivery/http/middleware"
	"github.com/javaman/go-metrics/internal/metric/repository/database"
	"github.com/javaman/go-metrics/internal/metric/repository/inmemory"
	"github.com/javaman/go-metrics/internal/metric/usecase"
	"github.com/labstack/echo/v4"
)

func configureWithDatabase(cfg *config.ServerConfiguration) domain.MetricUsecase {
	db, err := sql.Open("pgx", cfg.DBDsn)
	if err != nil {
		panic(err)
	}
	repository := database.New(db)
	return usecase.New(repository)
}

func configureInMemory(cfg *config.ServerConfiguration) domain.MetricUsecase {
	var r domain.MetricRepository

	if cfg.Restore {
		r = inmemory.NewFromFile(cfg.FileStoragePath)
	} else {
		r = inmemory.New()
	}

	if cfg.StoreInterval > 0 {
		inmemory.FlushStorageInBackground(r, cfg.FileStoragePath, cfg.StoreInterval)
	} else {
		r = inmemory.MakeFlushedOnEachSave(r, cfg.FileStoragePath)
	}

	return usecase.New(r)
}

func main() {
	cfg := config.ConfigureServer()
	var metricUsecase domain.MetricUsecase

	if cfg.DBDsn != "" {
		metricUsecase = configureWithDatabase(cfg)
	} else {
		metricUsecase = configureInMemory(cfg)
	}

	e := echo.New()
	e.Use(middleware.CompressDecompress)
	if len(cfg.Key) > 0 {
		e.Use(middleware.VerifyHash(cfg.Key))
		e.Use(middleware.AppendHash(cfg.Key))
	}
	e.Use(middleware.Logger())
	http.New(e, metricUsecase)
	e.Logger.Fatal(e.Start(cfg.Address))
}
