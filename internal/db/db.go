package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrNotFound error = errors.New("Not found")

type Database interface {
	Ping() error
	Select(id string, mtype string) (int64, float64, error)
	SelectAll(mtype string, f func(string, int64, float64)) error
	Upsert(id string, mtype string, delta int64, value float64) error
	CreateTable() error
}

type defaultDatabase struct {
	db *sql.DB
}

func (d *defaultDatabase) Ping() error {
	var i int
	row := d.db.QueryRow("SELECT 42")
	if err := row.Scan(&i); err != nil {
		return err
	}
	if i != 42 {
		return errors.New("wrong value pingig database")
	}
	return nil
}

func (d *defaultDatabase) Select(id string, mtype string) (int64, float64, error) {
	var delta int64
	var value float64
	if err := d.db.QueryRow("SELECT delta, value FROM metrics WHERE id = $1 and mtype = $2", id, mtype).Scan(&delta, &value); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, ErrNotFound
		}
		return 0, 0, err
	}
	return delta, value, nil
}

func (d *defaultDatabase) SelectAll(mtype string, f func(string, int64, float64)) error {
	rows, err := d.db.Query("SELECT id, delta, value FROM metrics WHERE mtype=$1 ORDER BY ID", mtype)
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

func (d *defaultDatabase) Upsert(id string, mtype string, delta int64, value float64) error {
	_, err := d.db.Exec(`
		INSERT INTO metrics(id, mtype, delta, value) VALUES($1, $2, $3, $4) 
		ON CONFLICT ON CONSTRAINT metrics_pk DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value
	`, id, mtype, delta, value)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (d *defaultDatabase) CreateTable() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id text,
			mtype text,
			delta bigint,
			value double precision,
			CONSTRAINT metrics_pk PRIMARY KEY(id, mtype)
		)
	`)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

var ErrWithoutDatabase = errors.New("without database")

type withoutDatabase struct {
}

func (d *withoutDatabase) Ping() error {
	return ErrWithoutDatabase
}

func (d *withoutDatabase) Select(id string, mtype string) (int64, float64, error) {
	return 0, 0, ErrWithoutDatabase
}

func (d *withoutDatabase) SelectAll(mtype string, f func(string, int64, float64)) error {
	return ErrWithoutDatabase
}

func (d *withoutDatabase) Upsert(id string, mtype string, delta int64, value float64) error {
	return ErrWithoutDatabase
}

func (d *withoutDatabase) CreateTable() error {
	return ErrWithoutDatabase
}

func New(connstr string) Database {
	db, _ := sql.Open("pgx", connstr)
	return &defaultDatabase{db: db}
}

func NewStub() Database {
	return &withoutDatabase{}
}
