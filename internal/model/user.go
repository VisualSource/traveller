package model

import (
	"database/sql"
	"encoding/json"
)

type User struct {
	Id       string   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Sessions []string `json:"sessions"`
	//Characters []string
}

func (u *User) GetUser(db *sql.DB, id string) error {
	stmt, err := db.Prepare("SELECT * FROM user WHERE id = ?;")
	if err != nil {
		return err
	}
	row := stmt.QueryRow(id)

	err = u.ScanRow(row)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) ScanRow(row *sql.Row) error {
	var sessions string
	err := row.Scan(&u.Id, &u.Username, &u.Password, &sessions)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(sessions), &u.Sessions)
	if err != nil {
		return err
	}

	return nil
}
func (u *User) Scan(row *sql.Rows) error {
	var sessions string
	err := row.Scan(&u.Id, &u.Username, &u.Password, &u.Sessions)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(sessions), &u.Sessions)
	if err != nil {
		return err
	}

	return nil
}
