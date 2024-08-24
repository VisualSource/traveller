package handler

import (
	"database/sql"
	"visualsource/traveller/internal/socket"

	_ "github.com/mattn/go-sqlite3"
)

type Handler struct {
	// DB connection here
	db  *sql.DB
	hub *socket.Hub
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
	CREATE TABLE IF NOT EXISTS user (id TEXT PRIMARY KEY, username TEXT UNIQUE NOT NULL, password TEXT NOT NULL, sessions jsonb NOT NULL DEFAULT '[]');
	CREATE TABLE IF NOT EXISTS session (id TEXT PRIMARY KEY, name TEXT NOT NULL, admin TEXT NOT NULL, players jsonb NOT NULL DEFAULT '[]');
	CREATE TABLE IF NOT EXISTS character (id TEXT PRIMARY key, owner TEXT NOT NULL, session_id TEXT NOT NULL);
	CREATE TABLE IF NOT EXISTS ship (id TEXT PRIMARY KEY, name TEXT NOT NULL);
	`)
	if err != nil {
		return nil, err
	}

	hub := socket.NewHub()

	go hub.Run()

	return &Handler{
		db:  db,
		hub: hub,
	}, nil
}
