package models

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./waf_db.sql")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// テーブルの作成
	cmd := `CREATE TABLE IF NOT EXISTS block_list(
				ID INTEGER PRIMARY KEY AUTOINCREMENT,
				CreateAt TIMESTAMP,
				UpdateAt TIMESTAMP,
				IP STRING,
				Count INTEGER,
				LastEvent TIMESTAMP
			)`
	_, err = DB.Exec(cmd)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
