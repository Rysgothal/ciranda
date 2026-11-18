package main

import (
	"ciranda/internal/cli"
	"ciranda/internal/database"
	"ciranda/internal/http"
	"context"
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

	ctx := context.Background()
	path, err = database.DefaultPath()
	if err != nil {
		return err
	}

	db, err = database.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = database.Migrate(ctx, db); err != nil {
		return err
	}

	server := http.NewServer(db)
	return server.Run()
}
