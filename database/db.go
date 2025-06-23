package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {

	var err error

	DB, err = sql.Open("sqlite3", "./game.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS characters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        class TEXT,
        hp INTEGER,
        mp INTEGER,
        alive BOOLEAN
    );
    `

	if _, err = DB.Exec(createTable); err != nil {
		log.Fatal(err)
	}
}
