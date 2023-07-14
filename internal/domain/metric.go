package domain

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrorInvalidType = errors.New("invalid metric type")
	ErrorNotFound    = errors.New("metric not found")
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m Metric) String() string {
	switch m.MType {
	case Gauge:
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	case Counter:
		return fmt.Sprintf("%d", *m.Delta)
	default:
		return ""
	}
}

type MetricUsecase interface {
	Save(m *Metric) (*Metric, error)
	SaveAll(ms []Metric) error
	Get(m *Metric) (*Metric, error)
	List() ([]*Metric, error)
	Ping() bool
}

type MetricRepository interface {
	Save(m *Metric) error
	Get(m *Metric) (*Metric, error)
	List() ([]*Metric, error)
	Ping() bool
	WriteToFile(fname string) error
}
