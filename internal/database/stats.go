package database

import (
	"database/sql"
	"fmt"
	"time"

	"solo-leveling/internal/models"
)

// ============================================================
// Character
// ============================================================

func (db *DB) GetOrCreateCharacter(name string) (*models.Character, error) {
	var char models.Character
	var activeTitle sql.NullString
	err := db.conn.QueryRow("SELECT id, name, attempts, COALESCE(active_title,'') FROM character LIMIT 1").Scan(&char.ID, &char.Name, &char.Attempts, &activeTitle)
	if err == sql.ErrNoRows {
		res, err := db.conn.Exec("INSERT INTO character (name, attempts) VALUES (?, 0)", name)
		if err != nil {
			return nil, err
		}
		char.ID, _ = res.LastInsertId()
		char.Name = name
		char.Attempts = 0

		for _, stat := range models.AllStats {
			_, err := db.conn.Exec(
				"INSERT INTO stat_levels (char_id, stat_type, level, current_exp, total_exp) VALUES (?, ?, 1, 0, 0)",
				char.ID, string(stat),
			)
			if err != nil {
				return nil, err
			}
		}
		return &char, nil
	}
	if err != nil {
		return nil, err
	}
	if activeTitle.Valid {
		char.ActiveTitle = activeTitle.String
	}
	return &char, nil
}

func (db *DB) AddAttempts(charID int64, amount int) (int, error) {
	_, err := db.conn.Exec(
		"UPDATE character SET attempts = MIN(attempts + ?, ?) WHERE id = ?",
		amount, models.MaxAttempts, charID,
	)
	if err != nil {
		return 0, err
	}
	var current int
	err = db.conn.QueryRow("SELECT attempts FROM character WHERE id = ?", charID).Scan(&current)
	return current, err
}

func (db *DB) SpendAttempt(charID int64) error {
	res, err := db.conn.Exec("UPDATE character SET attempts = attempts - 1 WHERE id = ? AND attempts > 0", charID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("нет попыток для боя")
	}
	return nil
}

func (db *DB) GetAttempts(charID int64) (int, error) {
	var attempts int
	err := db.conn.QueryRow("SELECT attempts FROM character WHERE id = ?", charID).Scan(&attempts)
	return attempts, err
}

// Streak titles
func (db *DB) InsertStreakTitle(charID int64, title string, streakDays int) error {
	_, err := db.conn.Exec(
		"INSERT OR IGNORE INTO streak_titles (char_id, title, streak_days) VALUES (?, ?, ?)",
		charID, title, streakDays,
	)
	return err
}

func (db *DB) GetStreakTitles(charID int64) ([]string, error) {
	rows, err := db.conn.Query(
		"SELECT title FROM streak_titles WHERE char_id = ? ORDER BY streak_days",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var titles []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	return titles, nil
}

func (db *DB) UpdateCharacterName(id int64, name string) error {
	_, err := db.conn.Exec("UPDATE character SET name = ? WHERE id = ?", name, id)
	return err
}

func (db *DB) SetActiveTitle(charID int64, title string) error {
	_, err := db.conn.Exec("UPDATE character SET active_title = ? WHERE id = ?", title, charID)
	return err
}

// GetAllTitles collects every title the character has earned (battle rewards and streak titles).
func (db *DB) GetAllTitles(charID int64) ([]string, error) {
	var titles []string

	// Battle reward titles
	rows, err := db.conn.Query("SELECT title FROM battle_rewards WHERE char_id = ? ORDER BY awarded_at", charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		if t != "" {
			titles = append(titles, t)
		}
	}

	// Streak titles
	rows2, err := db.conn.Query("SELECT title FROM streak_titles WHERE char_id = ? ORDER BY streak_days", charID)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var t string
		if err := rows2.Scan(&t); err != nil {
			return nil, err
		}
		if t != "" {
			titles = append(titles, t)
		}
	}

	return titles, nil
}

// ============================================================
// Stat Levels
// ============================================================

func (db *DB) GetStatLevels(charID int64) ([]models.StatLevel, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, stat_type, level, current_exp, total_exp FROM stat_levels WHERE char_id = ? ORDER BY stat_type",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.StatLevel
	for rows.Next() {
		var s models.StatLevel
		if err := rows.Scan(&s.ID, &s.CharID, &s.StatType, &s.Level, &s.CurrentEXP, &s.TotalEXP); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (db *DB) UpdateStatLevel(stat *models.StatLevel) error {
	_, err := db.conn.Exec(
		"UPDATE stat_levels SET level = ?, current_exp = ?, total_exp = ? WHERE id = ?",
		stat.Level, stat.CurrentEXP, stat.TotalEXP, stat.ID,
	)
	return err
}

// ============================================================
// Skills
// ============================================================

func (db *DB) CreateSkill(s *models.Skill) error {
	res, err := db.conn.Exec(
		"INSERT INTO skills (char_id, name, description, stat_type, multiplier, unlocked_at, active) VALUES (?, ?, ?, ?, ?, ?, 1)",
		s.CharID, s.Name, s.Description, string(s.StatType), s.Multiplier, s.UnlockedAt,
	)
	if err != nil {
		return err
	}
	s.ID, _ = res.LastInsertId()
	s.Active = true
	return nil
}

func (db *DB) GetSkills(charID int64) ([]models.Skill, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, name, description, stat_type, multiplier, unlocked_at, active FROM skills WHERE char_id = ? ORDER BY stat_type, unlocked_at",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var active int
		if err := rows.Scan(&s.ID, &s.CharID, &s.Name, &s.Description, &s.StatType, &s.Multiplier, &s.UnlockedAt, &active); err != nil {
			return nil, err
		}
		s.Active = active == 1
		skills = append(skills, s)
	}
	return skills, nil
}

func (db *DB) ToggleSkill(skillID int64, active bool) error {
	val := 0
	if active {
		val = 1
	}
	_, err := db.conn.Exec("UPDATE skills SET active = ? WHERE id = ?", val, skillID)
	return err
}

// ============================================================
// Daily Activity
// ============================================================

func (db *DB) RecordDailyActivity(charID int64, questsCompleted, questsFailed, expEarned int) error {
	today := time.Now().Format("2006-01-02")
	_, err := db.conn.Exec(`
		INSERT INTO daily_activity (char_id, date, quests_completed, quests_failed, exp_earned)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(char_id, date) DO UPDATE SET
			quests_completed = quests_completed + excluded.quests_completed,
			quests_failed = quests_failed + excluded.quests_failed,
			exp_earned = exp_earned + excluded.exp_earned
	`, charID, today, questsCompleted, questsFailed, expEarned)
	return err
}

func (db *DB) GetDailyActivityLast30(charID int64) ([]models.DailyActivity, error) {
	since := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	rows, err := db.conn.Query(
		"SELECT id, char_id, date, quests_completed, quests_failed, exp_earned FROM daily_activity WHERE char_id = ? AND date >= ? ORDER BY date",
		charID, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.DailyActivity
	for rows.Next() {
		var a models.DailyActivity
		if err := rows.Scan(&a.ID, &a.CharID, &a.Date, &a.QuestsComplete, &a.QuestsFailed, &a.EXPEarned); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

// GetStreak calculates the current streak of consecutive days with completed quests
func (db *DB) GetStreak(charID int64) (int, error) {
	rows, err := db.conn.Query(
		"SELECT date FROM daily_activity WHERE char_id = ? AND quests_completed > 0 ORDER BY date DESC",
		charID,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	streak := 0
	expectedDate := time.Now()

	// If no activity today yet, start from yesterday
	todayStr := expectedDate.Format("2006-01-02")
	yesterdayStr := expectedDate.AddDate(0, 0, -1).Format("2006-01-02")

	firstIteration := true
	for rows.Next() {
		var dateStr string
		if err := rows.Scan(&dateStr); err != nil {
			return 0, err
		}

		if firstIteration {
			firstIteration = false
			if dateStr != todayStr && dateStr != yesterdayStr {
				return 0, nil
			}
			if dateStr == todayStr {
				expectedDate = time.Now()
			} else {
				expectedDate = time.Now().AddDate(0, 0, -1)
			}
			if dateStr == expectedDate.Format("2006-01-02") {
				streak = 1
				expectedDate = expectedDate.AddDate(0, 0, -1)
				continue
			}
			return 0, nil
		}

		if dateStr == expectedDate.Format("2006-01-02") {
			streak++
			expectedDate = expectedDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}
	return streak, nil
}

// ============================================================
// Statistics queries
// ============================================================

func (db *DB) GetTotalCompletedCount(charID int64) (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM quests WHERE char_id = ? AND status = ?", charID, string(models.QuestCompleted)).Scan(&count)
	return count, err
}

func (db *DB) GetTotalFailedCount(charID int64) (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM quests WHERE char_id = ? AND status = ?", charID, string(models.QuestFailed)).Scan(&count)
	return count, err
}

func (db *DB) GetCompletedCountByRank(charID int64) (map[models.QuestRank]int, error) {
	rows, err := db.conn.Query(
		`SELECT CASE
			WHEN exp <= 10 THEN 'E'
			WHEN exp <= 18 THEN 'D'
			WHEN exp <= 28 THEN 'C'
			WHEN exp <= 40 THEN 'B'
			WHEN exp <= 55 THEN 'A'
			ELSE 'S'
		END AS rank_group, COUNT(*)
		FROM quests
		WHERE char_id = ? AND status = ?
		GROUP BY rank_group`,
		charID, string(models.QuestCompleted),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[models.QuestRank]int)
	for rows.Next() {
		var rank models.QuestRank
		var count int
		if err := rows.Scan(&rank, &count); err != nil {
			return nil, err
		}
		result[rank] = count
	}
	return result, nil
}

func (db *DB) GetTotalEXPEarned(charID int64) (int, error) {
	var total int
	err := db.conn.QueryRow("SELECT COALESCE(SUM(total_exp), 0) FROM stat_levels WHERE char_id = ?", charID).Scan(&total)
	return total, err
}
