package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javaman/go-metrics/internal/model"
)

type databaseStorage struct {
	db *sql.DB
	tx *sql.Tx
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
		fmt.Println(err)
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

func (d *databaseStorage) get(id string, mtype string) (int64, float64, bool, error) {
	queryFunc := d.db.QueryRow
	if d.tx != nil {
		queryFunc = d.tx.QueryRow
	}

	var delta int64
	var value float64

	err := retry(func() error {
		return queryFunc("SELECT delta, value FROM metrics WHERE id=$1 and mtype=$2", id, mtype).Scan(&delta, &value)
	})

	switch {
	case err == sql.ErrNoRows:
		return delta, value, false, nil
	case err != nil:
		return 0, 0, false, err
	default:
		return delta, value, true, nil
	}
}

func (d *databaseStorage) getAll(mtype string, consumer func(string, int64, float64) bool) error {
	queryFunc := d.db.Query
	if d.tx != nil {
		queryFunc = d.tx.Query
	}

	err := retry(func() error {
		rows, err := queryFunc("SELECT id, mtype, delta, value FROM metrics WHERE mtype=$1", mtype)
		if err != nil {
			return err
		}
		defer func() {
			_ = rows.Close()
			_ = rows.Err()
		}()
		var id string
		var delta int64
		var value float64
		for rows.Next() {
			if err := rows.Scan(&id, &delta, &value); err != nil {
				return err
			}
			consumer(id, delta, value)
		}
		return nil
	})
	return err
}

func (d *databaseStorage) save(id string, mtype string, delta int64, value float64) error {
	queryFunc := d.db.Exec
	if d.tx != nil {
		queryFunc = d.tx.Exec
	}
	err := retry(func() error {
		_, err := queryFunc(`
			insert into metrics(id, mtype, delta, value) values($1, $2, $3, $4) 
			on conflict on constraint metrics_pk do update set delta = excluded.delta, value = excluded.value`, id, mtype, delta, value)
		return err
	})
	return err
}

func (d *databaseStorage) SaveGauge(name string, v float64) {
	d.save(name, model.Gauge, 0, v)
}

func (d *databaseStorage) GetGauge(name string) (float64, bool) {
	_, value, found, err := d.get(name, model.Gauge)
	if err != nil {
		return 0, false
	}
	return value, found
}

func (d *databaseStorage) AllGauges(f func(string, float64)) {
	d.getAll(model.Gauge, func(id string, delta int64, value float64) bool {
		f(id, value)
		return true
	})
}

func (d *databaseStorage) SaveCounter(name string, v int64) {
	d.save(name, model.Counter, v, 0)
}

func (d *databaseStorage) GetCounter(name string) (int64, bool) {
	delta, _, found, err := d.get(name, model.Counter)
	if err != nil {
		return 0, false
	}
	return delta, found
}

func (d *databaseStorage) AllCounters(f func(string, int64)) {
	d.getAll(model.Counter, func(id string, delta int64, value float64) bool {
		f(id, delta)
		return true
	})
}

func (d *databaseStorage) Lock() (LockedStorage, error) {
	var tx *sql.Tx
	err := retry(func() error {
		var err error
		tx, err = d.db.BeginTx(context.TODO(), &sql.TxOptions{Isolation: sql.LevelSerializable})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &databaseStorage{d.db, tx}, nil
}

func (d *databaseStorage) Unlock() error {
	return d.tx.Commit()
}

func (d *databaseStorage) WriteToFile(file string) {
	//Ok to do nothing
}

func (d *databaseStorage) Ping(ctx context.Context) error {
	return d.db.Ping()
}

func NewDatabaseStorage(db *sql.DB) *databaseStorage {
	createTable(db)
	return &databaseStorage{db: db}
}
