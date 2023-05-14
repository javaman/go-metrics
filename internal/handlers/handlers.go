package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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