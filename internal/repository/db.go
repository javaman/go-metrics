package repository

import "github.com/javaman/go-metrics/internal/db"

type databaseStorage struct {
	db db.Database
}

func (d *databaseStorage) SaveGauge(name string, v float64) {
	d.db.Upsert(name, "gauge", 0, v)
}

func (d *databaseStorage) GetGauge(name string) (float64, bool) {
	_, value, err := d.db.Select(name, "gauge")
	if err == nil {
		return value, true
	}
	return 0, false
}

func (d *databaseStorage) AllGauges(f func(string, float64)) {
	d.db.SelectAll("gauge", func(id string, delta int64, value float64) {
		f(id, value)
	})
}

func (d *databaseStorage) SaveCounter(name string, v int64) {
	d.db.Upsert(name, "counter", v, 0)
}

func (d *databaseStorage) GetCounter(name string) (int64, bool) {
	delta, _, err := d.db.Select(name, "counter")
	if err == nil {
		return delta, true
	}
	return 0, false
}

func (d *databaseStorage) AllCounters(f func(string, int64)) {
	d.db.SelectAll("counter", func(id string, delta int64, value float64) {
		f(id, delta)
	})
}

func (d *databaseStorage) WriteToFile(file string) {
	//Ok to do nothing
}

func NewDatabaseStorage(db db.Database) *databaseStorage {
	return &databaseStorage{db}
}
