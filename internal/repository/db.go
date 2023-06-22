package repository

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound error = errors.New("not found")

func upsert(tx *sql.Tx, id string, mtype string, delta int64, value float64) error {
	_, err := tx.Exec(`
		INSERT INTO metrics(id, mtype, delta, value) VALUES($1, $2, $3, $4) 
		ON CONFLICT ON CONSTRAINT metrics_pk DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value
	`, id, mtype, delta, value)
	return err
}

func query(tx *sql.Tx, id string, mtype string) (int64, float64, error) {
	var delta int64
	var value float64
	if err := tx.QueryRow("SELECT delta, value FROM metrics WHERE id = $1 and mtype = $2", id, mtype).Scan(&delta, &value); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, ErrNotFound
		}
		return 0, 0, err
	}
	return delta, value, nil
}

func queryAll(tx *sql.Tx, mtype string, f func(string, int64, float64)) error {
	rows, err := tx.Query("SELECT id, delta, value FROM metrics WHERE mtype=$1 ORDER BY ID", mtype)
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
		f(id, delta, value)
	}

	return nil
}

func createTable(db *sql.DB) {
	db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id text,
			mtype text,
			delta bigint,
			value double precision,
			CONSTRAINT metrics_pk PRIMARY KEY(id, mtype)
		)
	`)
}

type databaseStorage struct {
	db *sql.DB
	tx *sql.Tx
}

func (d *databaseStorage) useTransactionOrNew() (*sql.Tx, bool) {
	if d.tx == nil {
		result, _ := d.db.BeginTx(context.TODO(), &sql.TxOptions{Isolation: sql.LevelSerializable})
		return result, true
	}
	return d.tx, false
}

func (d *databaseStorage) SaveGauge(name string, v float64) {
	tx, created := d.useTransactionOrNew()
	if created {
		defer tx.Commit()
	}
	upsert(tx, name, "gauge", 0, v)
}

func (d *databaseStorage) GetGauge(name string) (float64, bool) {
	tx, created := d.useTransactionOrNew()
	if created {
		defer tx.Commit()
	}
	_, value, err := query(tx, name, "gauge")
	if err == nil {
		return value, true
	}
	return 0, false
}

func (d *databaseStorage) AllGauges(f func(string, float64)) {
	tx, created := d.useTransactionOrNew()
	if created {
		defer tx.Commit()
	}
	queryAll(tx, "gauge", func(id string, delta int64, value float64) {
		f(id, value)
	})
}

func (d *databaseStorage) SaveCounter(name string, v int64) {
	tx, created := d.useTransactionOrNew()
	if created {
		defer tx.Commit()
	}
	upsert(tx, name, "counter", v, 0)
}

func (d *databaseStorage) GetCounter(name string) (int64, bool) {
	tx, created := d.useTransactionOrNew()
	if created {
		defer tx.Commit()
	}
	delta, _, err := query(tx, name, "counter")
	if err == nil {
		return delta, true
	}
	return 0, false
}

func (d *databaseStorage) AllCounters(f func(string, int64)) {
	tx, created := d.useTransactionOrNew()
	if created {
		defer tx.Commit()
	}
	queryAll(tx, "counter", func(id string, delta int64, value float64) {
		f(id, delta)
	})
}

func (d *databaseStorage) Lock() LockedStorage {
	tx, _ := d.db.BeginTx(context.TODO(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	return &databaseStorage{d.db, tx}
}

func (d *databaseStorage) Unlock() {
	d.tx.Commit()
}

func (d *databaseStorage) WriteToFile(file string) {
	//Ok to do nothing
}

func NewDatabaseStorage(constr string) (*databaseStorage, func() error) {
	db, _ := sql.Open("pgx", constr)
	createTable(db)
	return &databaseStorage{db: db}, func() error { return db.Ping() }
}
