package db

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database interface {
	Ping() error
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
		return errors.New("Wrong value pingig database")
	}
	return nil
}

func New(connstr string) Database {
	db, _ := sql.Open("pgx", connstr)
	return &defaultDatabase{db: db}
}
