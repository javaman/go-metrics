package main

import (
	"net/http"

	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/handlers"
	"github.com/javaman/go-metrics/internal/repository"
	"github.com/javaman/go-metrics/internal/services"

	"github.com/labstack/echo/v4"
)

func main() {

	storage := repository.NewInMemoryStorage()
	service := services.NewMetricsService(storage)

	e := echo.New()

	e.GET("/", handlers.ListAll(service))

	e.GET("/value/counter/:measureName", handlers.ValueCounter(service))
	e.GET("/value/gauge/:measureName", handlers.ValueGauge(service))

	e.GET("/update/*", func(c echo.Context) error { return c.NoContent(http.StatusMethodNotAllowed) })
	e.POST("/update/:measureType/*", handlers.BadRequest)

	e.POST("/update/counter/:measureName/:measureValue", handlers.UpdateCounter(storage))
	e.POST("/update/counter/", handlers.NotFound)
	e.POST("/update/gauge/:measureName/:measureValue", handlers.UpdateGauge(storage))
	e.POST("/update/gauge/", handlers.NotFound)

	cfg := config.ConfigureServer()
	e.Logger.Fatal(e.Start(cfg.Address))
}
