package main

import (
	"ciranda/internal/cli"
	"ciranda/internal/database"
	"ciranda/internal/http"
	"database/sql"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cli.Menu()

	var err error
	var path string
	var db *sql.DB

	path, err = database.DefaultPath()
	if err != nil {
		return err
	}

	db, err = database.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()
	return http.Init()
}
