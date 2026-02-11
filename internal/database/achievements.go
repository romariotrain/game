package database

import (
	"database/sql"
	"time"

	"solo-leveling/internal/models"
)

func (db *DB) SeedAchievements(list []models.Achievement) error {
	for _, a := range list {
		_, err := db.conn.Exec(
			`INSERT OR IGNORE INTO achievements (key, title, description, category, is_unlocked)
			 VALUES (?, ?, ?, ?, 0)`,
			a.Key, a.Title, a.Description, a.Category,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// UnlockAchievement marks the achievement as unlocked once.
// Returns true when the row has been changed for the first time.
func (db *DB) UnlockAchievement(key string) (bool, error) {
	res, err := db.conn.Exec(
		`UPDATE achievements
		 SET is_unlocked = 1,
		     obtained_at = COALESCE(obtained_at, ?)
		 WHERE key = ? AND is_unlocked = 0`,
		time.Now(), key,
	)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (db *DB) GetAchievements() ([]models.Achievement, error) {
	rows, err := db.conn.Query(
		`SELECT id, key, title, description, category, obtained_at, is_unlocked
		 FROM achievements
		 ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Achievement
	for rows.Next() {
		var a models.Achievement
		var obtainedAt sql.NullTime
		var unlocked int
		if err := rows.Scan(&a.ID, &a.Key, &a.Title, &a.Description, &a.Category, &obtainedAt, &unlocked); err != nil {
			return nil, err
		}
		if obtainedAt.Valid {
			t := obtainedAt.Time
			a.ObtainedAt = &t
		}
		a.IsUnlocked = unlocked == 1
		out = append(out, a)
	}
	return out, nil
}
