package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/VitJRBOG/TelegramWeatherBot/tools"
	_ "github.com/go-sql-driver/mysql"
)

func Connect(dbConnection tools.DBConnection) *sql.DB {
	c := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		dbConnection.Login, dbConnection.Password, dbConnection.Address, dbConnection.DBName)
	db, err := sql.Open("mysql", c)
	if err != nil {
		panic(err.Error())
	}

	return db
}

type User struct {
	ID           int
	Name         string
	Username     string
	UserID       int
	RequestCount int
}

func (u *User) InsertInto(db *sql.DB) (int, int) {
	query := `INSERT INTO user(name, username, user_id, request_count) values(?, ?, ?, ?)`

	result, err := db.Exec(query, u.Name, u.Username, u.UserID, u.RequestCount)
	if err != nil {
		panic(err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}

	count, err := result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

	return int(id), int(count)
}

func (u *User) SelectFrom(db *sql.DB) []User {
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
		panic(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			panic(err.Error())
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
		panic(err.Error())
	}

	return users
}

func (u *User) Update(db *sql.DB) (int, int) {
	query := `UPDATE user SET name = ?, username = ?, user_id = ?, request_count = ? WHERE id = ?`

	result, err := db.Exec(query, u.Name, u.Username, u.UserID, u.RequestCount, u.ID)
	if err != nil {
		panic(err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}

	count, err := result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

	return int(id), int(count)
}
