package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type Storage interface {
	SaveGauge(name string, v float64)
	GetGauge(name string) (float64, bool)
	AllGauges(func(string, float64))
	SaveCounter(name string, v int64)
	GetCounter(name string) (int64, bool)
	AllCounters(func(string, int64))
}

type memStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

func (m *memStorage) GetGauge(name string) (float64, bool) {
	v, found := m.gauges[name]
	return v, found
}

func (m *memStorage) SaveGauge(name string, v float64) {
	log.Printf("Received gauge %s = %f", name, v)
	m.gauges[name] = v
}

func (m *memStorage) AllGauges(f func(string, float64)) {
	for k, v := range m.gauges {
		f(k, v)
	}
}

func (m *memStorage) GetCounter(name string) (int64, bool) {
	v, found := m.counters[name]
	return v, found
}

func (m *memStorage) SaveCounter(name string, v int64) {
	log.Printf("Received counter %s = %d", name, v)
	if value, ok := m.counters[name]; ok {
		m.counters[name] = value + v
	} else {
		m.counters[name] = v
	}
}

func (m *memStorage) AllCounters(f func(string, int64)) {
	for k, v := range m.counters {
		f(k, v)
	}
}

func badRequest(c echo.Context) error {
	return c.NoContent(http.StatusBadRequest)
}

func notFound(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func valueGauge(s Storage) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		if value, found := s.GetGauge(measureName); found {
			return c.String(http.StatusOK, strconv.FormatFloat(value, 'f', -1, 64))
		} else {
			return notFound(c)
		}
	}
}

func updateGauge(s Storage) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		if strings.TrimSpace(measureName) == "" {
			return notFound(c)
		}
		if measureValue, err := strconv.ParseFloat(c.Param("measureValue"), 64); err == nil {
			s.SaveGauge(measureName, measureValue)
			return c.NoContent(http.StatusOK)
		} else {
			return badRequest(c)
		}
	}
}

func valueCounter(s Storage) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		if value, found := s.GetCounter(measureName); found {
			return c.String(http.StatusOK, fmt.Sprintf("%d", value))
		} else {
			return notFound(c)
		}
	}
}
func updateCounter(s Storage) func(echo.Context) error {
	return func(c echo.Context) error {
		metricName := c.Param("measureName")
		if strings.TrimSpace(metricName) == "" {
			return notFound(c)
		}
		if metricValue, err := strconv.ParseInt(c.Param("measureValue"), 10, 64); err == nil {
			s.SaveCounter(metricName, metricValue)
			return c.NoContent(http.StatusOK)
		} else {
			return badRequest(c)
		}
	}
}

func listAll(s Storage) func(echo.Context) error {
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

func main() {
	storage := &memStorage{make(map[string]int64), make(map[string]float64)}
	e := echo.New()

	e.GET("/", listAll(storage))

	e.GET("/value/counter/:measureName", valueCounter(storage))
	e.GET("/value/gauge/:measureName", valueGauge(storage))

	e.GET("/update/*", func(c echo.Context) error { return c.NoContent(http.StatusMethodNotAllowed) })
	e.POST("/update/:measureType/*", badRequest)

	e.POST("/update/counter/:measureName/:measureValue", updateCounter(storage))
	e.POST("/update/counter/", notFound)
	e.POST("/update/gauge/:measureName/:measureValue", updateGauge(storage))
	e.POST("/update/gauge/", notFound)

	e.Logger.Fatal(e.Start(":8080"))
}
