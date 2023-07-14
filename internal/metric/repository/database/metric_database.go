package database

import (
	"database/sql"
	"errors"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javaman/go-metrics/internal/domain"
)

type databaseStorage struct {
	db *sql.DB
}

func retry(f func() error) error {
	sleep := [...]int{1, 3, 5}
	var err error
	for _, s := range sleep {
		err = f()
		var netError net.Error
		if err == nil || !errors.As(err, &netError) || pgconn.SafeToRetry(err) {
			break
		}
		time.Sleep(time.Duration(s) * time.Second)
	}
	return err
}

func createTable(db *sql.DB) error {
	return retry(func() error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id text,
			mtype text,
			delta bigint,
			value double precision,
			CONSTRAINT metrics_pk PRIMARY KEY(id, mtype)
		)
	        `)
		return err
	})
}

func mapRow(src interface{ Scan(dest ...any) error }, dst *domain.Metric) error {
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := src.Scan(&dst.ID, &dst.MType, &delta, &value)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.ErrorNotFound
	case err != nil:
		return err
	}

	switch dst.MType {
	case domain.Gauge:
		dst.Value = &value.Float64
	case domain.Counter:
		dst.Delta = &delta.Int64
	}

	return nil
}

func (d *databaseStorage) Save(metric *domain.Metric) error {
	return retry(func() error {
		_, err := d.db.Exec(`insert into metrics(id, mtype, delta, value) values($1, $2, $3, $4) 
			   on conflict on constraint metrics_pk do update set delta = excluded.delta, value = excluded.value`, metric.ID, metric.MType, metric.Delta, metric.Value)
		return err
	})
}

func (d *databaseStorage) Get(metric *domain.Metric) (*domain.Metric, error) {
	result := &domain.Metric{}
	err := retry(func() error {
		return mapRow(d.db.QueryRow("SELECT id, mtype, delta, value FROM metrics WHERE id=$1 and mtype=$2", metric.ID, metric.MType), result)
	})
	if err == nil {
		return result, nil
	}
	return nil, err
}

func (d *databaseStorage) List() ([]*domain.Metric, error) {
	var result []*domain.Metric
	err := retry(func() error {
		rows, err := d.db.Query("SELECT id, mtype, delta, value FROM metrics")
		if err != nil {
			return err
		}
		defer func() {
			_ = rows.Close()
			_ = rows.Err()
		}()
		for rows.Next() {
			metric := &domain.Metric{}
			if err := mapRow(rows, metric); err != nil {
				return err
			}
			result = append(result, metric)
		}
		return nil

	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *databaseStorage) WriteToFile(file string) error {
	//Ok to do nothing
	return nil
}

func (d *databaseStorage) Ping() bool {
	return d.db.Ping() == nil
}

func New(db *sql.DB) *databaseStorage {
	createTable(db)
	return &databaseStorage{db: db}
}
