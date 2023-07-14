package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/javaman/go-metrics/internal/domain"
	"github.com/labstack/echo/v4"
)

type MetricHandler struct {
	usecase domain.MetricUsecase
}

func New(e *echo.Echo, uc domain.MetricUsecase) {
	handler := &MetricHandler{
		usecase: uc,
	}

	e.GET("/", handler.ListAll)

	e.GET("/value/counter/:measureName", handler.ValueCounter)
	e.GET("/value/gauge/:measureName", handler.ValueGauge)
	e.POST("/value/", handler.Value)

	e.GET("/update/*", handler.NotAllowed)
	e.POST("/update/:measureType/*", handler.BadRequest)

	e.POST("/update/counter/:measureName/:measureValue", handler.UpdateCounter)
	e.POST("/update/counter/", handler.NotFound)
	e.POST("/update/gauge/:measureName/:measureValue", handler.UpdateGauge)
	e.POST("/update/gauge/", handler.NotFound)
	e.POST("/update/", handler.Update)

	e.GET("/ping", handler.Ping)

	e.POST("/updates/", handler.Updates)
}

func (h *MetricHandler) ListAll(c echo.Context) error {
	list, err := h.usecase.List()
	if err != nil {
		return h.InternalServerError(c)
	}
	var b strings.Builder
	b.WriteString("<html><head><title>AllMetrics</title></head><body><table>")
	for _, m := range list {
		b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", m.ID, m))
	}
	b.WriteString("</table></body></html>")
	return c.HTML(http.StatusOK, b.String())
}

func (h *MetricHandler) ValueCounter(c echo.Context) error {
	measureName := c.Param("measureName")
	m := &domain.Metric{ID: measureName, MType: domain.Counter}

	counter, err := h.usecase.Get(m)

	switch err {
	case nil:
		return c.String(http.StatusOK, counter.String())
	case domain.ErrorNotFound:
		return h.NotFound(c)
	default:
		return h.InternalServerError(c)
	}
}

func (h *MetricHandler) ValueGauge(c echo.Context) error {
	measureName := c.Param("measureName")
	m := &domain.Metric{ID: measureName, MType: domain.Gauge}

	counter, err := h.usecase.Get(m)

	switch err {
	case nil:
		return c.String(http.StatusOK, counter.String())
	case domain.ErrorNotFound:
		return h.NotFound(c)
	default:
		return h.InternalServerError(c)
	}
}

func (h *MetricHandler) Value(c echo.Context) error {
	var m domain.Metric

	err := json.NewDecoder(c.Request().Body).Decode(&m)

	if err != nil {
		return h.BadRequest(c)
	}

	res, err := h.usecase.Get(&m)

	switch err {
	case nil:
		return c.JSON(http.StatusOK, res)
	case domain.ErrorNotFound:
		return h.NotFound(c)
	default:
		return h.InternalServerError(c)
	}
}

func (h *MetricHandler) NotAllowed(c echo.Context) error {
	return c.NoContent(http.StatusMethodNotAllowed)
}

func (h *MetricHandler) BadRequest(c echo.Context) error {
	return c.NoContent(http.StatusBadRequest)
}

func (h *MetricHandler) UpdateCounter(c echo.Context) error {
	metricName := c.Param("measureName")
	if strings.TrimSpace(metricName) == "" {
		return h.NotFound(c)
	}
	if delta, err := strconv.ParseInt(c.Param("measureValue"), 10, 64); err == nil {
		counter := &domain.Metric{ID: metricName, MType: domain.Counter, Delta: &delta}
		h.usecase.Save(counter)
		return h.OK(c)
	} else {
		return h.BadRequest(c)
	}
}

func (h *MetricHandler) NotFound(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func (h *MetricHandler) UpdateGauge(c echo.Context) error {
	measureName := c.Param("measureName")
	if strings.TrimSpace(measureName) == "" {
		return h.NotFound(c)
	}
	if value, err := strconv.ParseFloat(c.Param("measureValue"), 64); err == nil {
		gauge := &domain.Metric{ID: measureName, MType: domain.Gauge, Value: &value}
		h.usecase.Save(gauge)
		return h.OK(c)
	} else {
		return h.BadRequest(c)
	}
}

func (h *MetricHandler) Update(c echo.Context) error {
	var m domain.Metric
	err := json.NewDecoder(c.Request().Body).Decode(&m)
	if err != nil {
		return h.BadRequest(c)
	}
	res, err := h.usecase.Save(&m)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, res)
	default:
		return h.BadRequest(c)
	}
}

func (h *MetricHandler) Ping(c echo.Context) error {
	if h.usecase.Ping() {
		return h.OK(c)
	}
	return h.InternalServerError(c)
}

func (h *MetricHandler) Updates(c echo.Context) error {
	var metrics []domain.Metric
	err := json.NewDecoder(c.Request().Body).Decode(&metrics)
	if err != nil {
		return h.BadRequest(c)
	}
	h.usecase.SaveAll(metrics)
	return h.OK(c)
}

func (h *MetricHandler) InternalServerError(c echo.Context) error {
	return c.NoContent(http.StatusInternalServerError)
}

func (h *MetricHandler) OK(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
