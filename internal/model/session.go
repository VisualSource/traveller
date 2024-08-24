package model

import (
	"database/sql"
	"encoding/json"
)

type Session struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Admin   string   `json:"admin"`
	Players []string `json:"players"`
	// TODO other session config stuff here
}

func (s *Session) GetSession(db *sql.DB, id string) error {
	stmt, err := db.Prepare("SELECT * FROM session WHERE id = ?")
	if err != nil {
		return err
	}
	row := stmt.QueryRow(id)
	err = s.ScanRow(row)
	if err != nil {

		return err
	}
	return nil
}

func (s *Session) ScanRow(row *sql.Row) error {
	var players string
	err := row.Scan(&s.Id, &s.Name, &s.Admin, &players)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(players), &s.Players)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) Scan(row *sql.Rows) error {
	var players string
	err := row.Scan(&s.Id, &s.Name, &s.Admin, &players)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(players), &s.Players)
	if err != nil {
		return err
	}

	return nil
}
