package app

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type App struct {
	http.ServeMux
	db *sqlx.DB
}

func New(dsn string) (*App, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	m := http.NewServeMux()
	a := App{
		ServeMux: *m,
		db:       db,
	}
	return &a, nil
}
