package repository

type Storage interface {
	SaveGauge(name string, v float64)
	GetGauge(name string) (float64, bool)
	AllGauges(func(string, float64))
	SaveCounter(name string, v int64)
	GetCounter(name string) (int64, bool)
	AllCounters(func(string, int64))
}

func NewInMemoryStorage() *memStorage {
	return &memStorage{make(map[string]int64), make(map[string]float64)}
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
	m.counters[name] = v
}

func (m *memStorage) AllCounters(f func(string, int64)) {
	for k, v := range m.counters {
		f(k, v)
	}
}
