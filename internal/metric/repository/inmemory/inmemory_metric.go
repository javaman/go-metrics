package inmemory

import (
	"encoding/json"
	"os"
	"time"

	"github.com/javaman/go-metrics/internal/domain"
)

func MakeFlushedOnEachSave(r domain.MetricRepository, fname string) domain.MetricRepository {
	return &wrappingSaveToFile{
		MetricRepository: r,
		fileName:         fname,
	}
}

func NewFromFile(file string) domain.MetricRepository {
	var result memStorage
	if data, err := os.ReadFile(file); err == nil {
		json.Unmarshal(data, &result)
	}
	if result.counters == nil {
		result.counters = make(map[string]int64)
	}
	if result.gauges == nil {
		result.gauges = make(map[string]float64)
	}
	return &result
}

func New() *memStorage {
	return &memStorage{make(map[string]int64), make(map[string]float64)}
}

type memStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

func (m *memStorage) WriteToFile(fname string) error {
	data, err := json.MarshalIndent(m, "", "   ")
	if err == nil {
		os.WriteFile(fname, data, 0666)
		return nil
	} else {
		return err
	}
}

func (m *memStorage) Save(metric *domain.Metric) error {
	switch metric.MType {
	case domain.Gauge:
		m.gauges[metric.ID] = *metric.Value
		return nil
	case domain.Counter:
		m.counters[metric.ID] = *metric.Delta
		return nil
	default:
		return domain.ErrorInvalidType
	}
}

func (m *memStorage) Get(metric *domain.Metric) (*domain.Metric, error) {
	switch metric.MType {
	case domain.Gauge:
		value, present := m.gauges[metric.ID]
		if present {
			return &domain.Metric{ID: metric.ID, MType: domain.Gauge, Delta: nil, Value: &value}, nil
		}
	case domain.Counter:
		delta, present := m.counters[metric.ID]
		if present {
			return &domain.Metric{ID: metric.ID, MType: domain.Counter, Delta: &delta, Value: nil}, nil
		}
	default:
		return nil, domain.ErrorInvalidType
	}
	return nil, domain.ErrorNotFound
}

func (m *memStorage) List() ([]*domain.Metric, error) {
	var result []*domain.Metric
	for id, value := range m.gauges {
		x := value
		result = append(result, &domain.Metric{ID: id, MType: domain.Gauge, Value: &x})
	}
	for id, delta := range m.counters {
		x := delta
		result = append(result, &domain.Metric{ID: id, MType: domain.Counter, Delta: &x})
	}
	return result, nil
}

func (m *memStorage) Ping() bool {
	return false
}

type wrappingSaveToFile struct {
	domain.MetricRepository
	fileName string
}

func (m *wrappingSaveToFile) Save(metric *domain.Metric) error {
	err := m.MetricRepository.Save(metric)
	if err == nil {
		return m.MetricRepository.WriteToFile(m.fileName)
	}
	return err
}

func (m *memStorage) UnmarshalJSON(b []byte) error {
	var tmp struct {
		Counters map[string]int64   `json:"counters"`
		Gauges   map[string]float64 `json:"gauges"`
	}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	m.counters = tmp.Counters
	m.gauges = tmp.Gauges
	return nil
}

func (m *memStorage) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Counters map[string]int64   `json:"counters"`
		Gauges   map[string]float64 `json:"gauges"`
	}{
		Counters: m.counters,
		Gauges:   m.gauges,
	})
}

func FlushStorageInBackground(r domain.MetricRepository, fname string, interval int) {
	go func() {
		for {
			time.Sleep(time.Duration(interval) * time.Second)
			r.WriteToFile(fname)
		}
	}()
}
