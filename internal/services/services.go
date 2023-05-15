package services

import (
	"github.com/javaman/go-metrics/internal/repository"
)

type MetricsService interface {
	SaveGauge(name string, v float64)
	GetGauge(name string) (float64, bool)
	AllGauges(func(string, float64))
	SaveCounter(name string, v int64)
	GetCounter(name string) (int64, bool)
	AllCounters(func(string, int64))
}

type defaultMetricsService struct {
	storage repository.Storage
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

func (dm *defaultMetricsService) SaveCounter(name string, v int64) {
	if value, ok := dm.storage.GetCounter(name); ok {
		value += v
		dm.storage.SaveCounter(name, value)
	} else {
		dm.storage.SaveCounter(name, v)
	}
}

func (dm *defaultMetricsService) GetCounter(name string) (int64, bool) {
	return dm.storage.GetCounter(name)
}

func (dm *defaultMetricsService) AllCounters(f func(string, int64)) {
	dm.storage.AllCounters(f)
}

func NewMetricsService(repository repository.Storage) *defaultMetricsService {
	return &defaultMetricsService{repository}
}
