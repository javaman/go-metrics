package repository

import (
	"encoding/json"
	"fmt"
	"os"
)

type Storage interface {
	SaveGauge(name string, v float64)
	GetGauge(name string) (float64, bool)
	AllGauges(func(string, float64))
	SaveCounter(name string, v int64)
	GetCounter(name string) (int64, bool)
	AllCounters(func(string, int64))
	WriteToFile(file string)
}

func MakeStorageFlushedOnEachCall(s Storage, fname string) Storage {
	return &wrappingSaveToFile{
		Storage:  s,
		fileName: fname,
	}
}

func NewInMemoryStorageFromFile(file string) Storage {
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

func NewInMemoryStorage() *memStorage {
	return &memStorage{make(map[string]int64), make(map[string]float64)}
}

type memStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

func (m *memStorage) WriteToFile(fname string) {
	data, err := json.MarshalIndent(m, "", "   ")
	if err == nil {
		os.WriteFile(fname, data, 0666)
	}
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

type wrappingSaveToFile struct {
	Storage
	fileName string
}

func (m *wrappingSaveToFile) SaveCounter(name string, v int64) {
	m.Storage.SaveCounter(name, v)
	m.Storage.WriteToFile(m.fileName)
}

func (m *wrappingSaveToFile) SaveGauge(name string, v float64) {
	m.Storage.SaveGauge(name, v)
	m.Storage.WriteToFile(m.fileName)
}

func (m *memStorage) UnmarshalJSON(b []byte) error {
	fmt.Println("Here")
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
