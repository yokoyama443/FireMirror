package models

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./attack_db.sql")
	defer DB.Close()
	cmd := `CREATE TABLE IF NOT EXISTS attack_list(
				ID INTEGER PRIMARY KEY AUTOINCREMENT,
				CreateAt TIMESTAMP,
				UpdateAt TIMESTAMP,
				IP STRING,
				Status STRING
			)`
	_, err = DB.Exec(cmd)
	if err != nil {
		log.Fatalln(err)
	}
}
