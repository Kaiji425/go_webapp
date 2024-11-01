// models/user.go
package models

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

func InitializeDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./data.db")
	if err != nil {
		return nil, err
	}

	query := `
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func FindUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, created_at FROM users WHERE username = ?`
	row := db.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
