package handler

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Handler struct {
	// DB connection here
	db *sql.DB
}

func (h *Handler) Close() {
	h.db.Close()
}

func New() (*Handler, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS user (id TEXT PRIMARY KEY, username TEXT UNIQUE NOT NULL, password TEXT NOT NULL);
	CREATE TABLE IF NOT EXISTS session (id TEXT PRIMARY KEY, name TEXT NOT NULL, admin TEXT NOT NULL);
	CREATE TABLE IF NOT EXISTS character (id TEXT PRIMARY key, owner TEXT NOT NULL);
	CREATE TABLE IF NOT EXISTS ship (id TEXT PRIMARY KEY, name TEXT NOT NULL);
	`)
	if err != nil {
		return nil, err
	}

	return &Handler{
		db: db,
	}, nil
}
