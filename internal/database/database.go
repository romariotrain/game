package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func New() (*DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".solo-leveling")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "game.db")
	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
