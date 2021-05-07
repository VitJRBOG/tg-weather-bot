package db

import (
	"database/sql"
	"fmt"

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
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	UserID       int    `json:"user_id"`
	RequestCount int    `json:"request_count"`
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

func (u *User) SelectByUserID(db *sql.DB) []User {
	query := `SELECT * FROM user WHERE user_id = ?`

	rows, err := db.Query(query, u.UserID)
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
