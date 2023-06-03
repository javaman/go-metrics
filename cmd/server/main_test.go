package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javaman/go-metrics/internal/db"
	"github.com/javaman/go-metrics/internal/handlers"
	"github.com/javaman/go-metrics/internal/repository"
	"github.com/javaman/go-metrics/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {
	storage := repository.NewInMemoryStorage()
	storage.SaveGauge("g1", 3.14)
	storage.SaveCounter("counter1", 42)
	service := services.NewMetricsService(storage, db.NewStub())
	testData := []struct {
		path           string
		method         string
		expectedStatus int
		paramNames     []string
		paramValues    []string
		handler        func(echo.Context) error
	}{
		{
			"/",
			http.MethodGet,
			http.StatusOK,
			nil,
			nil,
			handlers.ListAll(service),
		},
		{
			"/value/:metricType/:measureName",
			http.MethodGet,
			http.StatusOK,
			[]string{"metricType", "measureName"},
			[]string{"counter", "counter1"},
			handlers.ValueCounter(service),
		},
		{
			"/value/:metricType/:measureName",
			http.MethodGet,
			http.StatusNotFound,
			[]string{"metricType", "measureName"},
			[]string{"counter", "counter2"},
			handlers.ValueCounter(service),
		},
		{
			"/value/:metricType/:measureName",
			http.MethodGet,
			http.StatusOK,
			[]string{"metricType", "measureName"},
			[]string{"gauge", "g1"},
			handlers.ValueGauge(service),
		},
		{
			"/value/:metricType/:measureName",
			http.MethodGet,
			http.StatusNotFound,
			[]string{"metricType", "measureName"},
			[]string{"gauge", "guage2"},
			handlers.ValueGauge(service),
		},

		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusOK,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"counter", "cx1", "42"},
			handlers.UpdateCounter(service),
		},
		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusNotFound,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"counter", "", "42"},
			handlers.UpdateCounter(service),
		},
		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusBadRequest,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"counter", "cx3", ""},
			handlers.UpdateCounter(service),
		},
		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusBadRequest,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"counter", "cx3", "notanumber"},
			handlers.UpdateCounter(service),
		},

		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusOK,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"gauge", "gx1", "3.14"},
			handlers.UpdateGauge(service),
		},
		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusNotFound,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"gauge", "", "3.14"},
			handlers.UpdateGauge(service),
		},
		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusBadRequest,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"gauge", "gx3", ""},
			handlers.UpdateGauge(service),
		},
		{
			"/update/:metricType/:measureName/:measureValue",
			http.MethodPost,
			http.StatusBadRequest,
			[]string{"metricType", "measureName", "measureValue"},
			[]string{"gauge", "gx4", "notanumber"},
			handlers.UpdateGauge(service),
		},
	}

	e := echo.New()

	for _, test := range testData {
		req := httptest.NewRequest(test.method, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(test.path)
		if test.paramNames != nil {
			c.SetParamNames(test.paramNames...)
			c.SetParamValues(test.paramValues...)
		}
		test.handler(c)
		assert.Equal(t, test.expectedStatus, rec.Code)
	}
}
