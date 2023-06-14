package handlers

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/javaman/go-metrics/internal/model"
	"github.com/javaman/go-metrics/internal/services"
	"github.com/labstack/echo/v4"
)

func BadRequest(c echo.Context) error {
	return c.NoContent(http.StatusBadRequest)
}

func NotFound(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func ValueGauge(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		if value, found := s.GetGauge(measureName); found {
			return c.String(http.StatusOK, strconv.FormatFloat(value, 'f', -1, 64))
		} else {
			return NotFound(c)
		}
	}
}

func UpdateGauge(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		if strings.TrimSpace(measureName) == "" {
			return NotFound(c)
		}
		if measureValue, err := strconv.ParseFloat(c.Param("measureValue"), 64); err == nil {
			s.SaveGauge(measureName, measureValue)
			return c.NoContent(http.StatusOK)
		} else {
			return BadRequest(c)
		}
	}
}

func ValueCounter(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		fmt.Println(c.ParamNames())
		fmt.Println(c.ParamValues())
		if value, found := s.GetCounter(measureName); found {
			return c.String(http.StatusOK, fmt.Sprintf("%d", value))
		} else {
			return NotFound(c)
		}
	}
}
func UpdateCounter(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		metricName := c.Param("measureName")
		if strings.TrimSpace(metricName) == "" {
			return NotFound(c)
		}
		if metricValue, err := strconv.ParseInt(c.Param("measureValue"), 10, 64); err == nil {
			s.SaveCounter(metricName, metricValue)
			return c.NoContent(http.StatusOK)
		} else {
			return BadRequest(c)
		}
	}
}

func ListAll(s services.MetricsService) func(echo.Context) error {
	return func(e echo.Context) error {
		var b strings.Builder
		b.WriteString("<html><head><title>AllMetrics</title></head><body><table>")
		s.AllGauges(func(n string, v float64) {
			b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%f</td></tr>", n, v))
		})
		s.AllCounters(func(n string, v int64) {
			b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", n, v))
		})
		b.WriteString("</table></body></html>")
		return e.HTML(http.StatusOK, b.String())
	}
}

func Update(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		var m model.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&m)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		res, err := s.Save(&m)
		if err != nil {
			switch err {
			case services.ErrIDRequired:
				return NotFound(c)
			default:
				return BadRequest(c)
			}
		}
		return c.JSON(http.StatusOK, res)
	}
}

func Value(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		var m model.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&m)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		res, err := s.Value(&m)
		if err != nil {
			switch err {
			case services.ErrIDNotFound:
				return c.NoContent(http.StatusNotFound)
			default:
				return BadRequest(c)
			}
		}
		return c.JSON(http.StatusOK, res)
	}
}

func Updates(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		var metrics []model.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&metrics)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		s.Updates(metrics)
		return c.NoContent(http.StatusOK)
	}
}

func Ping(pinger driver.Pinger) func(echo.Context) error {
	return func(c echo.Context) error {
		if err := pinger.Ping(context.TODO()); err == nil {
			return c.NoContent(http.StatusOK)
		}
		return c.NoContent(http.StatusInternalServerError)
	}
}
