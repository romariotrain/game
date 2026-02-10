package database

import (
	"database/sql"
	"time"
)

func (db *DB) GetAIProfileText() (string, error) {
	var text string
	err := db.conn.QueryRow("SELECT profile_text FROM ai_profile WHERE id = 1").Scan(&text)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return text, nil
}

func (db *DB) SaveAIProfileText(profileText string) error {
	_, err := db.conn.Exec(`
		INSERT INTO ai_profile (id, profile_text, updated_at)
		VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			profile_text = excluded.profile_text,
			updated_at = excluded.updated_at
	`, profileText, time.Now())
	return err
}

func (db *DB) LogAISuggestions(rawJSON, model, errText string) error {
	_, err := db.conn.Exec(
		"INSERT INTO ai_suggestions (created_at, raw_json, model, error) VALUES (?, ?, ?, ?)",
		time.Now(), rawJSON, model, nullableString(errText),
	)
	return err
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
