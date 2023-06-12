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

/*
 *  утильные функции выполнить запрос вернуть ошибки, закрыть ресуры
 */
func update(exec func(string, ...any) (sql.Result, error), queryString string, args ...any) error {
	_, err := exec(queryString, args...)
	return err
}

func query(queryFunc func(string, ...any) (*sql.Rows, error), queryString string, consumer func(string, string, int64, float64) bool, args ...any) error {
	rows, err := queryFunc(queryString, args...)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	var id string
	var mtype string
	var delta int64
	var value float64

	for rows.Next() {
		if err := rows.Scan(&id, &mtype, &delta, &value); err != nil {
			return err
		}
		if !consumer(id, mtype, delta, value) {
			break
		}
	}

	return nil
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

/*
 * некий промежуточный слой - связать все вместе - контекст транзакционный или нет, переповтор, сам запрос, параметры, куда будут результаты возвращены
 */
func (d *databaseStorage) get(id string, mtype string) (int64, float64, bool, error) {
	queryFunc := d.db.Query
	if d.tx != nil {
		queryFunc = d.tx.Query
	}
	found := false
	var delta int64
	var value float64

	err := retry(func() error {
		return query(queryFunc, "SELECT id, mtype, delta, value FROM metrics WHERE id=$1 and mtype=$2", func(i string, m string, d int64, v float64) bool {
			found = true
			delta = d
			value = v
			return false

		}, id, mtype)
	})
	if err == nil {
		return delta, value, found, nil
	}
	return 0, 0, false, err
}

func (d *databaseStorage) getAll(mtype string, consumer func(string, int64, float64) bool) error {
	queryFunc := d.db.Query
	if d.tx != nil {
		queryFunc = d.tx.Query
	}
	return retry(func() error {
		return query(queryFunc, "SELECT id, mtype, delta, value FROM metrics WHERE mtype=$1", func(i string, m string, d int64, v float64) bool {
			consumer(i, d, v)
			return true
		}, mtype)
	})
}

func (d *databaseStorage) save(id string, mtype string, delta int64, value float64) error {
	queryFunc := d.db.Exec
	if d.tx != nil {
		queryFunc = d.tx.Exec
	}
	err := retry(func() error {
		return update(queryFunc, `
			insert into metrics(id, mtype, delta, value) values($1, $2, $3, $4) 
			on conflict on constraint metrics_pk do update set delta = excluded.delta, value = excluded.value
		`, id, mtype, delta, value)
	})
	return err
}

/*
 *  сама реализация интерфейса
 */
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

func NewDatabaseStorage(db *sql.DB) *databaseStorage {
	createTable(db)
	return &databaseStorage{db: db}
}

func PingDB(db *sql.DB) func() error {
	return func() error {
		return db.Ping()
	}
}
