package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// Storage структура для объекта Storage
type Storage struct {
	db *sql.DB
}

// New конструктор объекта Storage
func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite3", storagePath) // Подключаемся к БД
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаем таблицу если её ещё нет
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS entry(
	    id INTEGER PRIMARY KEY,
	    url TEXT NOT NULL UNIQUE,	    
	    title TEXT NOT NULL,
	    summary TEXT NOT NULL,
	    content TEXT NOT NULL,
	    updated TIMESTAMP ,
		published TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}
