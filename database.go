package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB - инициализирует базу данных и возвращает соединение
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./notes.db")
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Создаем таблицу
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			text TEXT NOT NULL,
			content TEXT,
			original_url TEXT,
			tags TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
