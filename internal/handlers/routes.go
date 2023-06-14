package handlers

import (
	"database/sql/driver"
	"net/http"

	mymiddleware "github.com/javaman/go-metrics/internal/middleware"
	"github.com/javaman/go-metrics/internal/services"
	"github.com/labstack/echo/v4"
)

func New(service services.MetricsService, pinger driver.Pinger) *echo.Echo {
	e := echo.New()

	e.GET("/", ListAll(service))

	e.GET("/value/counter/:measureName", ValueCounter(service))
	e.GET("/value/gauge/:measureName", ValueGauge(service))
	e.POST("/value/", Value(service))

	e.GET("/update/*", func(c echo.Context) error { return c.NoContent(http.StatusMethodNotAllowed) })
	e.POST("/update/:measureType/*", BadRequest)

	e.POST("/update/counter/:measureName/:measureValue", UpdateCounter(service))
	e.POST("/update/counter/", NotFound)
	e.POST("/update/gauge/:measureName/:measureValue", UpdateGauge(service))
	e.POST("/update/gauge/", NotFound)
	e.POST("/update/", Update(service))

	e.GET("/ping", Ping(pinger))

	e.POST("/updates/", Updates(service))

	e.Use(mymiddleware.Logger())
	e.Use(mymiddleware.CompressDecompress)
	return e
}
