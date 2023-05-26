package main

import (
	"net/http"

	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/handlers"
	"github.com/javaman/go-metrics/internal/repository"
	"github.com/javaman/go-metrics/internal/services"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	storage := repository.NewInMemoryStorage()
	service := services.NewMetricsService(storage)

	e := echo.New()

	e.GET("/", handlers.ListAll(service))

	e.GET("/value/counter/:measureName", handlers.ValueCounter(service))
	e.GET("/value/gauge/:measureName", handlers.ValueGauge(service))
	e.POST("/value/", handlers.Value(service))

	e.GET("/update/*", func(c echo.Context) error { return c.NoContent(http.StatusMethodNotAllowed) })
	e.POST("/update/:measureType/*", handlers.BadRequest)

	e.POST("/update/counter/:measureName/:measureValue", handlers.UpdateCounter(service))
	e.POST("/update/counter/", handlers.NotFound)
	e.POST("/update/gauge/:measureName/:measureValue", handlers.UpdateGauge(service))
	e.POST("/update/gauge/", handlers.NotFound)
	e.POST("/update/", handlers.Update(service))

	cfg := config.ConfigureServer()

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogMethod:       true,
		LogLatency:      true,
		LogStatus:       true,
		LogResponseSize: true,
		LogValuesFunc: func(e echo.Context, v middleware.RequestLoggerValues) error {
			sugar.Infow("request",
				zap.String("URI", v.URI),
				zap.String("method", v.Method),
				zap.Int64("latency", v.Latency.Nanoseconds()),
			)
			sugar.Infow("response",
				zap.Int("status", v.Status),
				zap.Int64("size", v.ResponseSize),
			)
			return nil
		},
	}))

	e.Logger.Fatal(e.Start(cfg.Address))
}
