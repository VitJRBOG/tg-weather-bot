package db

import (
	"database/sql"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"github.com/VitJRBOG/TelegramWeatherBot/internal/tools"
	_ "github.com/go-sql-driver/mysql"
)

func Connect(dbConn tools.DBConn) (*sql.DB, error) {
	c := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		dbConn.Login, dbConn.Password, dbConn.Address, dbConn.DBName)
	db, err := sql.Open("mysql", c)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type User struct {
	ID           int
	Name         string
	Username     string
	UserID       int
	RequestCount int
}

func (u *User) InsertInto(db *sql.DB) (int, int, error) {
	query := `INSERT INTO user(name, username, user_id, request_count) values(?, ?, ?, ?)`

	result, err := db.Exec(query, u.Name, u.Username, u.UserID, u.RequestCount)
	if err != nil {
		return 0, 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return int(id), int(count), nil
}

func (u *User) SelectFrom(db *sql.DB) ([]User, error) {
	query := `SELECT * FROM user`

	var f []interface{}

	if u.ID > 0 {
		query += ` WHERE id = ?`
		f = append(f, u.ID)
	}

	if u.Name != "" {
		if strings.Contains(query, "WHERE") {
			query += ` AND name = ?`
		} else {
			query += ` WHERE name = ?`
		}
		f = append(f, u.Name)
	}

	if u.Username != "" {
		if strings.Contains(query, "WHERE") {
			query += ` AND username = ?`
		} else {
			query += ` WHERE username = ?`
		}
		f = append(f, u.Username)
	}

	if u.UserID > 0 {
		if strings.Contains(query, "WHERE") {
			query += ` AND user_id = ?`
		} else {
			query += ` WHERE user_id = ?`
		}
		f = append(f, u.UserID)
	}

	rows, err := db.Query(query, f...)
	if err != nil {
		return []User{}, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("%s\n\n%s\n", err, debug.Stack())
		}
	}()

	var users []User

	for rows.Next() {
		var user User

		if err := rows.Scan(&user.ID, &user.Name, &user.Username,
			&user.UserID, &user.RequestCount); err != nil {
			panic(err.Error())
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return []User{}, err
	}

	return users, nil
}

func (u *User) Update(db *sql.DB) (int, int, error) {
	query := `UPDATE user SET name = ?, username = ?, user_id = ?, request_count = ? WHERE id = ?`

	result, err := db.Exec(query, u.Name, u.Username, u.UserID, u.RequestCount, u.ID)
	if err != nil {
		return 0, 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return int(id), int(count), nil
}
