package model

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var connection *sql.DB

func init() {
	var err error
	connection, err = sql.Open("postgres", "postgres://postgres:root@localhost/postgres?sslmode=disable")

	if err != nil {
		panic(err.Error())
	}
}
