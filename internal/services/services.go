package services

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/javaman/go-metrics/internal/model"
	"github.com/javaman/go-metrics/internal/repository"
)

var (
	ErrIDRequired    error = errors.New("ID Required")
	ErrInvalidMType  error = errors.New("MType must be counter or delta")
	ErrDeltaRequired error = errors.New("delta is required")
	ErrValueRequired error = errors.New("dalue is required")
	ErrIDNotFound    error = errors.New("ID not found")
)

type MetricsService interface {
	SaveGauge(name string, v float64)
	GetGauge(name string) (float64, bool)
	AllGauges(func(string, float64))
	SaveCounter(name string, v int64)
	GetCounter(name string) (int64, bool)
	AllCounters(func(string, int64))
	Save(m *model.Metrics) (*model.Metrics, error)
	Value(m *model.Metrics) (*model.Metrics, error)
}

type defaultMetricsService struct {
	storage   repository.Storage
	validator *validator.Validate
}

func (dm *defaultMetricsService) SaveGauge(name string, v float64) {
	dm.storage.SaveGauge(name, v)
}

func (dm *defaultMetricsService) GetGauge(name string) (float64, bool) {
	return dm.storage.GetGauge(name)
}

func (dm *defaultMetricsService) AllGauges(f func(string, float64)) {
	dm.storage.AllGauges(f)
}

func (dm *defaultMetricsService) saveCounter(name string, v int64) int64 {
	if value, ok := dm.storage.GetCounter(name); ok {
		value += v
		dm.storage.SaveCounter(name, value)
		return value
	} else {
		dm.storage.SaveCounter(name, v)
		return v
	}
}

func (dm *defaultMetricsService) SaveCounter(name string, v int64) {
	dm.saveCounter(name, v)
}

func (dm *defaultMetricsService) GetCounter(name string) (int64, bool) {
	return dm.storage.GetCounter(name)
}

func (dm *defaultMetricsService) AllCounters(f func(string, int64)) {
	dm.storage.AllCounters(f)
}

func (dm *defaultMetricsService) saveCounterStruct(m *model.Metrics) (*model.Metrics, error) {
	result := &model.Metrics{ID: m.ID, MType: m.MType}
	if strings.TrimSpace(m.ID) == "" {
		return nil, ErrIDRequired
	}
	if m.Delta == nil {
		return nil, ErrDeltaRequired
	}
	newDelta := dm.saveCounter(m.ID, *m.Delta)
	result.Delta = &newDelta
	return result, nil
}

func (dm *defaultMetricsService) saveGaugeStruct(m *model.Metrics) (*model.Metrics, error) {
	result := &model.Metrics{ID: m.ID, MType: m.MType}
	if strings.TrimSpace(m.ID) == "" {
		return nil, ErrIDRequired
	}
	if m.Value == nil {
		return nil, ErrValueRequired
	}
	dm.SaveGauge(m.ID, *m.Value)
	var newValue float64 = *m.Value
	result.Value = &newValue
	return result, nil
}

func (dm *defaultMetricsService) Save(m *model.Metrics) (*model.Metrics, error) {
	switch metricType := m.MType; metricType {
	case "counter":
		return dm.saveCounterStruct(m)
	case "gauge":
		return dm.saveGaugeStruct(m)
	default:
		return nil, ErrInvalidMType
	}
}

func (dm *defaultMetricsService) valueCounterStruct(m *model.Metrics) (*model.Metrics, error) {
	if delta, ok := dm.GetCounter(m.ID); ok {
		m.Delta = &delta
		return m, nil
	}
	return nil, ErrIDNotFound
}

func (dm *defaultMetricsService) valueGaugeStruct(m *model.Metrics) (*model.Metrics, error) {
	if value, ok := dm.GetGauge(m.ID); ok {
		m.Value = &value
		return m, nil
	}
	return nil, ErrIDNotFound
}

func (dm *defaultMetricsService) Value(m *model.Metrics) (*model.Metrics, error) {
	result := &model.Metrics{ID: m.ID, MType: m.MType}
	switch m.MType {
	case "counter":
		return dm.valueCounterStruct(result)
	case "gauge":
		return dm.valueGaugeStruct(result)
	default:
		return nil, ErrInvalidMType
	}
}

func NewMetricsService(repository repository.Storage) *defaultMetricsService {
	return &defaultMetricsService{repository, validator.New()}
}

func FlushStorageInBackground(storage repository.Storage, fname string, interval int) {
	go func() {
		for {
			time.Sleep(time.Duration(interval) * time.Second)
			storage.Save(fname)
		}
	}()
}
