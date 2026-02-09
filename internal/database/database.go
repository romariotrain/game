package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"solo-leveling/internal/models"
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

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS character (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL DEFAULT 'Hunter'
	);

	CREATE TABLE IF NOT EXISTS stat_levels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		stat_type TEXT NOT NULL,
		level INTEGER NOT NULL DEFAULT 1,
		current_exp INTEGER NOT NULL DEFAULT 0,
		total_exp INTEGER NOT NULL DEFAULT 0,
		UNIQUE(char_id, stat_type)
	);

	CREATE TABLE IF NOT EXISTS quests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		rank TEXT NOT NULL DEFAULT 'E',
		target_stat TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		is_daily INTEGER NOT NULL DEFAULT 0,
		template_id INTEGER,
		dungeon_id INTEGER
	);

	CREATE TABLE IF NOT EXISTS skills (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		stat_type TEXT NOT NULL,
		multiplier REAL NOT NULL DEFAULT 1.1,
		unlocked_at INTEGER NOT NULL DEFAULT 1,
		active INTEGER NOT NULL DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS daily_quest_templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		rank TEXT NOT NULL DEFAULT 'E',
		target_stat TEXT NOT NULL,
		active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS daily_activity (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		date TEXT NOT NULL,
		quests_completed INTEGER NOT NULL DEFAULT 0,
		quests_failed INTEGER NOT NULL DEFAULT 0,
		exp_earned INTEGER NOT NULL DEFAULT 0,
		UNIQUE(char_id, date)
	);

	CREATE TABLE IF NOT EXISTS dungeons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		requirements_json TEXT NOT NULL DEFAULT '[]',
		status TEXT NOT NULL DEFAULT 'locked',
		reward_title TEXT NOT NULL DEFAULT '',
		reward_exp INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS dungeon_quests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		dungeon_id INTEGER NOT NULL REFERENCES dungeons(id),
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		rank TEXT NOT NULL DEFAULT 'E',
		target_stat TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS completed_dungeons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		dungeon_id INTEGER NOT NULL REFERENCES dungeons(id),
		completed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		earned_title TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS enemies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		rank TEXT NOT NULL DEFAULT 'E',
		type TEXT NOT NULL DEFAULT 'regular',
		hp INTEGER NOT NULL DEFAULT 100,
		attack INTEGER NOT NULL DEFAULT 10,
		pattern_size INTEGER NOT NULL DEFAULT 4,
		show_time REAL NOT NULL DEFAULT 3.0,
		reward_exp INTEGER NOT NULL DEFAULT 20,
		reward_crystals INTEGER NOT NULL DEFAULT 10,
		drop_material TEXT NOT NULL DEFAULT '',
		drop_chance REAL NOT NULL DEFAULT 0.0
	);

	CREATE TABLE IF NOT EXISTS player_resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		crystals INTEGER NOT NULL DEFAULT 0,
		material_common INTEGER NOT NULL DEFAULT 0,
		material_rare INTEGER NOT NULL DEFAULT 0,
		material_epic INTEGER NOT NULL DEFAULT 0,
		UNIQUE(char_id)
	);

	CREATE TABLE IF NOT EXISTS equipment (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		name TEXT NOT NULL,
		slot TEXT NOT NULL DEFAULT 'weapon',
		rarity TEXT NOT NULL DEFAULT 'common',
		level INTEGER NOT NULL DEFAULT 1,
		current_exp INTEGER NOT NULL DEFAULT 0,
		bonus_attack INTEGER NOT NULL DEFAULT 0,
		bonus_hp INTEGER NOT NULL DEFAULT 0,
		bonus_time REAL NOT NULL DEFAULT 0.0,
		equipped INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS gacha_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		banner TEXT NOT NULL DEFAULT 'normal',
		equipment_id INTEGER NOT NULL,
		rarity TEXT NOT NULL DEFAULT 'common',
		pulled_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS battles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		enemy_id INTEGER NOT NULL,
		enemy_name TEXT NOT NULL DEFAULT '',
		result TEXT NOT NULL DEFAULT 'lose',
		damage_dealt INTEGER NOT NULL DEFAULT 0,
		damage_taken INTEGER NOT NULL DEFAULT 0,
		accuracy REAL NOT NULL DEFAULT 0.0,
		critical_hits INTEGER NOT NULL DEFAULT 0,
		dodges INTEGER NOT NULL DEFAULT 0,
		reward_exp INTEGER NOT NULL DEFAULT 0,
		reward_crystals INTEGER NOT NULL DEFAULT 0,
		material_drop TEXT NOT NULL DEFAULT '',
		fought_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS craft_recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		result_slot TEXT NOT NULL DEFAULT 'weapon',
		result_rarity TEXT NOT NULL DEFAULT 'common',
		cost_crystals INTEGER NOT NULL DEFAULT 0,
		cost_common INTEGER NOT NULL DEFAULT 0,
		cost_rare INTEGER NOT NULL DEFAULT 0,
		cost_epic INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS daily_rewards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		claimed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		day INTEGER NOT NULL DEFAULT 1,
		crystals INTEGER NOT NULL DEFAULT 50
	);

	CREATE TABLE IF NOT EXISTS gacha_pity (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		normal_pity INTEGER NOT NULL DEFAULT 0,
		advanced_pity INTEGER NOT NULL DEFAULT 0,
		UNIQUE(char_id)
	);
	`
	_, err := db.conn.Exec(schema)
	return err
}

// ============================================================
// Character
// ============================================================

func (db *DB) GetOrCreateCharacter(name string) (*models.Character, error) {
	var char models.Character
	err := db.conn.QueryRow("SELECT id, name FROM character LIMIT 1").Scan(&char.ID, &char.Name)
	if err == sql.ErrNoRows {
		res, err := db.conn.Exec("INSERT INTO character (name) VALUES (?)", name)
		if err != nil {
			return nil, err
		}
		char.ID, _ = res.LastInsertId()
		char.Name = name

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
	return &char, nil
}

func (db *DB) UpdateCharacterName(id int64, name string) error {
	_, err := db.conn.Exec("UPDATE character SET name = ? WHERE id = ?", name, id)
	return err
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
// Quests
// ============================================================

func (db *DB) CreateQuest(q *models.Quest) error {
	isDaily := 0
	if q.IsDaily {
		isDaily = 1
	}
	res, err := db.conn.Exec(
		"INSERT INTO quests (char_id, title, description, rank, target_stat, status, created_at, is_daily, template_id, dungeon_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		q.CharID, q.Title, q.Description, string(q.Rank), string(q.TargetStat), string(models.QuestActive), time.Now(), isDaily, q.TemplateID, q.DungeonID,
	)
	if err != nil {
		return err
	}
	q.ID, _ = res.LastInsertId()
	q.Status = models.QuestActive
	return nil
}

func (db *DB) GetActiveQuests(charID int64) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, rank, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND status = ? ORDER BY created_at DESC",
		charID, string(models.QuestActive),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
}

func (db *DB) GetCompletedQuests(charID int64, limit int) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, rank, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND status = ? ORDER BY completed_at DESC LIMIT ?",
		charID, string(models.QuestCompleted), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
}

func (db *DB) scanQuestsExt(rows *sql.Rows) ([]models.Quest, error) {
	var quests []models.Quest
	for rows.Next() {
		var q models.Quest
		var completedAt sql.NullTime
		var isDaily int
		var templateID sql.NullInt64
		var dungeonID sql.NullInt64
		if err := rows.Scan(&q.ID, &q.CharID, &q.Title, &q.Description, &q.Rank, &q.TargetStat, &q.Status, &q.CreatedAt, &completedAt, &isDaily, &templateID, &dungeonID); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			q.CompletedAt = &completedAt.Time
		}
		q.IsDaily = isDaily == 1
		if templateID.Valid {
			v := templateID.Int64
			q.TemplateID = &v
		}
		if dungeonID.Valid {
			v := dungeonID.Int64
			q.DungeonID = &v
		}
		quests = append(quests, q)
	}
	return quests, nil
}

func (db *DB) CompleteQuest(questID int64) error {
	now := time.Now()
	_, err := db.conn.Exec(
		"UPDATE quests SET status = ?, completed_at = ? WHERE id = ?",
		string(models.QuestCompleted), now, questID,
	)
	return err
}

func (db *DB) FailQuest(questID int64) error {
	_, err := db.conn.Exec(
		"UPDATE quests SET status = ? WHERE id = ?",
		string(models.QuestFailed), questID,
	)
	return err
}

func (db *DB) DeleteQuest(questID int64) error {
	_, err := db.conn.Exec("DELETE FROM quests WHERE id = ?", questID)
	return err
}

// GetQuestByID returns a single quest by its ID
func (db *DB) GetQuestByID(questID int64) (*models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, rank, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE id = ?",
		questID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	quests, err := db.scanQuestsExt(rows)
	if err != nil {
		return nil, err
	}
	if len(quests) == 0 {
		return nil, fmt.Errorf("quest not found: %d", questID)
	}
	return &quests[0], nil
}

// GetDungeonActiveQuests returns active quests for a given dungeon
func (db *DB) GetDungeonActiveQuests(charID int64, dungeonID int64) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, rank, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND dungeon_id = ? AND status = ? ORDER BY id",
		charID, dungeonID, string(models.QuestActive),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
}

// GetDungeonAllQuests returns all quests (any status) for a given dungeon
func (db *DB) GetDungeonAllQuests(charID int64, dungeonID int64) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, rank, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND dungeon_id = ? ORDER BY id",
		charID, dungeonID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
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
// Daily Quest Templates
// ============================================================

func (db *DB) CreateDailyTemplate(t *models.DailyQuestTemplate) error {
	res, err := db.conn.Exec(
		"INSERT INTO daily_quest_templates (char_id, title, description, rank, target_stat, active, created_at) VALUES (?, ?, ?, ?, ?, 1, ?)",
		t.CharID, t.Title, t.Description, string(t.Rank), string(t.TargetStat), time.Now(),
	)
	if err != nil {
		return err
	}
	t.ID, _ = res.LastInsertId()
	t.Active = true
	return nil
}

func (db *DB) GetActiveDailyTemplates(charID int64) ([]models.DailyQuestTemplate, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, rank, target_stat, active, created_at FROM daily_quest_templates WHERE char_id = ? AND active = 1 ORDER BY created_at",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.DailyQuestTemplate
	for rows.Next() {
		var t models.DailyQuestTemplate
		var active int
		if err := rows.Scan(&t.ID, &t.CharID, &t.Title, &t.Description, &t.Rank, &t.TargetStat, &active, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Active = active == 1
		templates = append(templates, t)
	}
	return templates, nil
}

func (db *DB) DisableDailyTemplate(templateID int64) error {
	_, err := db.conn.Exec("UPDATE daily_quest_templates SET active = 0 WHERE id = ?", templateID)
	return err
}

// HasDailyQuestForToday checks if a quest from this template already exists today
func (db *DB) HasDailyQuestForToday(charID int64, templateID int64) (bool, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM quests WHERE char_id = ? AND template_id = ? AND date(created_at) = ?",
		charID, templateID, today,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
		"SELECT rank, COUNT(*) FROM quests WHERE char_id = ? AND status = ? GROUP BY rank",
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

// ============================================================
// Dungeons
// ============================================================

func (db *DB) GetDungeonCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM dungeons").Scan(&count)
	return count, err
}

func (db *DB) InsertDungeon(d *models.Dungeon) error {
	reqJSON, err := json.Marshal(d.Requirements)
	if err != nil {
		return err
	}
	res, err := db.conn.Exec(
		"INSERT INTO dungeons (name, description, requirements_json, status, reward_title, reward_exp) VALUES (?, ?, ?, ?, ?, ?)",
		d.Name, d.Description, string(reqJSON), string(models.DungeonLocked), d.RewardTitle, d.RewardEXP,
	)
	if err != nil {
		return err
	}
	d.ID, _ = res.LastInsertId()

	for i := range d.QuestDefinitions {
		qd := &d.QuestDefinitions[i]
		qd.DungeonID = d.ID
		res2, err := db.conn.Exec(
			"INSERT INTO dungeon_quests (dungeon_id, title, description, rank, target_stat) VALUES (?, ?, ?, ?, ?)",
			qd.DungeonID, qd.Title, qd.Description, string(qd.Rank), string(qd.TargetStat),
		)
		if err != nil {
			return err
		}
		qd.ID, _ = res2.LastInsertId()
	}
	return nil
}

func (db *DB) GetAllDungeons() ([]models.Dungeon, error) {
	rows, err := db.conn.Query(
		"SELECT id, name, description, requirements_json, status, reward_title, reward_exp FROM dungeons ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dungeons []models.Dungeon
	for rows.Next() {
		var d models.Dungeon
		var reqJSON string
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &reqJSON, &d.Status, &d.RewardTitle, &d.RewardEXP); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(reqJSON), &d.Requirements); err != nil {
			d.Requirements = nil
		}
		// Load quest definitions
		defs, err := db.GetDungeonQuestDefs(d.ID)
		if err != nil {
			return nil, err
		}
		d.QuestDefinitions = defs
		dungeons = append(dungeons, d)
	}
	return dungeons, nil
}

func (db *DB) GetDungeonQuestDefs(dungeonID int64) ([]models.DungeonQuestDef, error) {
	rows, err := db.conn.Query(
		"SELECT id, dungeon_id, title, description, rank, target_stat FROM dungeon_quests WHERE dungeon_id = ? ORDER BY id",
		dungeonID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []models.DungeonQuestDef
	for rows.Next() {
		var d models.DungeonQuestDef
		if err := rows.Scan(&d.ID, &d.DungeonID, &d.Title, &d.Description, &d.Rank, &d.TargetStat); err != nil {
			return nil, err
		}
		defs = append(defs, d)
	}
	return defs, nil
}

func (db *DB) UpdateDungeonStatus(dungeonID int64, status models.DungeonStatus) error {
	_, err := db.conn.Exec("UPDATE dungeons SET status = ? WHERE id = ?", string(status), dungeonID)
	return err
}

func (db *DB) CompleteDungeon(charID, dungeonID int64, title string) error {
	_, err := db.conn.Exec(
		"INSERT INTO completed_dungeons (char_id, dungeon_id, completed_at, earned_title) VALUES (?, ?, ?, ?)",
		charID, dungeonID, time.Now(), title,
	)
	return err
}

func (db *DB) GetCompletedDungeons(charID int64) ([]models.CompletedDungeon, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, dungeon_id, completed_at, earned_title FROM completed_dungeons WHERE char_id = ? ORDER BY completed_at DESC",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.CompletedDungeon
	for rows.Next() {
		var c models.CompletedDungeon
		if err := rows.Scan(&c.ID, &c.CharID, &c.DungeonID, &c.CompletedAt, &c.EarnedTitle); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, nil
}

func (db *DB) IsDungeonCompleted(charID, dungeonID int64) (bool, error) {
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM completed_dungeons WHERE char_id = ? AND dungeon_id = ?",
		charID, dungeonID,
	).Scan(&count)
	return count > 0, err
}

// ============================================================
// Enemies
// ============================================================

func (db *DB) InsertEnemy(e *models.Enemy) error {
	res, err := db.conn.Exec(
		`INSERT INTO enemies (name, description, rank, type, hp, attack, pattern_size, show_time, reward_exp, reward_crystals, drop_material, drop_chance)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Name, e.Description, string(e.Rank), string(e.Type), e.HP, e.Attack,
		e.PatternSize, e.ShowTime, e.RewardEXP, e.RewardCrystals, string(e.DropMaterial), e.DropChance,
	)
	if err != nil {
		return err
	}
	e.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetAllEnemies() ([]models.Enemy, error) {
	rows, err := db.conn.Query(
		"SELECT id, name, description, rank, type, hp, attack, pattern_size, show_time, reward_exp, reward_crystals, drop_material, drop_chance FROM enemies ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enemies []models.Enemy
	for rows.Next() {
		var e models.Enemy
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Rank, &e.Type, &e.HP, &e.Attack,
			&e.PatternSize, &e.ShowTime, &e.RewardEXP, &e.RewardCrystals, &e.DropMaterial, &e.DropChance); err != nil {
			return nil, err
		}
		enemies = append(enemies, e)
	}
	return enemies, nil
}

func (db *DB) GetEnemyCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM enemies").Scan(&count)
	return count, err
}

func (db *DB) GetEnemyByID(id int64) (*models.Enemy, error) {
	var e models.Enemy
	err := db.conn.QueryRow(
		"SELECT id, name, description, rank, type, hp, attack, pattern_size, show_time, reward_exp, reward_crystals, drop_material, drop_chance FROM enemies WHERE id = ?",
		id,
	).Scan(&e.ID, &e.Name, &e.Description, &e.Rank, &e.Type, &e.HP, &e.Attack,
		&e.PatternSize, &e.ShowTime, &e.RewardEXP, &e.RewardCrystals, &e.DropMaterial, &e.DropChance)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// ============================================================
// Player Resources
// ============================================================

func (db *DB) GetOrCreateResources(charID int64) (*models.PlayerResources, error) {
	var r models.PlayerResources
	err := db.conn.QueryRow(
		"SELECT id, char_id, crystals, material_common, material_rare, material_epic FROM player_resources WHERE char_id = ?",
		charID,
	).Scan(&r.ID, &r.CharID, &r.Crystals, &r.MaterialCommon, &r.MaterialRare, &r.MaterialEpic)
	if err == sql.ErrNoRows {
		res, err := db.conn.Exec("INSERT INTO player_resources (char_id) VALUES (?)", charID)
		if err != nil {
			return nil, err
		}
		r.ID, _ = res.LastInsertId()
		r.CharID = charID
		return &r, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) UpdateResources(r *models.PlayerResources) error {
	_, err := db.conn.Exec(
		"UPDATE player_resources SET crystals = ?, material_common = ?, material_rare = ?, material_epic = ? WHERE id = ?",
		r.Crystals, r.MaterialCommon, r.MaterialRare, r.MaterialEpic, r.ID,
	)
	return err
}

func (db *DB) AddCrystals(charID int64, amount int) error {
	_, err := db.conn.Exec(
		"UPDATE player_resources SET crystals = crystals + ? WHERE char_id = ?",
		amount, charID,
	)
	return err
}

func (db *DB) AddMaterial(charID int64, tier models.MaterialTier, amount int) error {
	col := "material_common"
	switch tier {
	case models.MaterialRare:
		col = "material_rare"
	case models.MaterialEpic:
		col = "material_epic"
	}
	_, err := db.conn.Exec(
		fmt.Sprintf("UPDATE player_resources SET %s = %s + ? WHERE char_id = ?", col, col),
		amount, charID,
	)
	return err
}

// ============================================================
// Equipment
// ============================================================

func (db *DB) InsertEquipment(eq *models.Equipment) error {
	res, err := db.conn.Exec(
		`INSERT INTO equipment (char_id, name, slot, rarity, level, current_exp, bonus_attack, bonus_hp, bonus_time, equipped, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		eq.CharID, eq.Name, string(eq.Slot), string(eq.Rarity), eq.Level, eq.CurrentEXP,
		eq.BonusAttack, eq.BonusHP, eq.BonusTime, boolToInt(eq.Equipped), time.Now(),
	)
	if err != nil {
		return err
	}
	eq.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetAllEquipment(charID int64) ([]models.Equipment, error) {
	rows, err := db.conn.Query(
		`SELECT id, char_id, name, slot, rarity, level, current_exp, bonus_attack, bonus_hp, bonus_time, equipped, created_at
		FROM equipment WHERE char_id = ? ORDER BY equipped DESC, rarity DESC, level DESC`,
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanEquipment(rows)
}

func (db *DB) GetEquippedItems(charID int64) ([]models.Equipment, error) {
	rows, err := db.conn.Query(
		`SELECT id, char_id, name, slot, rarity, level, current_exp, bonus_attack, bonus_hp, bonus_time, equipped, created_at
		FROM equipment WHERE char_id = ? AND equipped = 1 ORDER BY slot`,
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanEquipment(rows)
}

func (db *DB) GetEquipmentByID(id int64) (*models.Equipment, error) {
	rows, err := db.conn.Query(
		`SELECT id, char_id, name, slot, rarity, level, current_exp, bonus_attack, bonus_hp, bonus_time, equipped, created_at
		FROM equipment WHERE id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := db.scanEquipment(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("equipment not found: %d", id)
	}
	return &items[0], nil
}

func (db *DB) UpdateEquipment(eq *models.Equipment) error {
	_, err := db.conn.Exec(
		`UPDATE equipment SET level = ?, current_exp = ?, bonus_attack = ?, bonus_hp = ?, bonus_time = ?, equipped = ? WHERE id = ?`,
		eq.Level, eq.CurrentEXP, eq.BonusAttack, eq.BonusHP, eq.BonusTime, boolToInt(eq.Equipped), eq.ID,
	)
	return err
}

func (db *DB) DeleteEquipment(id int64) error {
	_, err := db.conn.Exec("DELETE FROM equipment WHERE id = ?", id)
	return err
}

func (db *DB) UnequipSlot(charID int64, slot models.EquipmentSlot) error {
	_, err := db.conn.Exec(
		"UPDATE equipment SET equipped = 0 WHERE char_id = ? AND slot = ? AND equipped = 1",
		charID, string(slot),
	)
	return err
}

func (db *DB) GetEquipmentCount(charID int64) (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM equipment WHERE char_id = ?", charID).Scan(&count)
	return count, err
}

func (db *DB) scanEquipment(rows *sql.Rows) ([]models.Equipment, error) {
	var items []models.Equipment
	for rows.Next() {
		var eq models.Equipment
		var equipped int
		if err := rows.Scan(&eq.ID, &eq.CharID, &eq.Name, &eq.Slot, &eq.Rarity, &eq.Level, &eq.CurrentEXP,
			&eq.BonusAttack, &eq.BonusHP, &eq.BonusTime, &equipped, &eq.CreatedAt); err != nil {
			return nil, err
		}
		eq.Equipped = equipped == 1
		items = append(items, eq)
	}
	return items, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ============================================================
// Gacha
// ============================================================

func (db *DB) InsertGachaHistory(h *models.GachaHistory) error {
	res, err := db.conn.Exec(
		"INSERT INTO gacha_history (char_id, banner, equipment_id, rarity, pulled_at) VALUES (?, ?, ?, ?, ?)",
		h.CharID, string(h.Banner), h.EquipmentID, string(h.Rarity), time.Now(),
	)
	if err != nil {
		return err
	}
	h.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetGachaHistory(charID int64, limit int) ([]models.GachaHistory, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, banner, equipment_id, rarity, pulled_at FROM gacha_history WHERE char_id = ? ORDER BY pulled_at DESC LIMIT ?",
		charID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.GachaHistory
	for rows.Next() {
		var h models.GachaHistory
		if err := rows.Scan(&h.ID, &h.CharID, &h.Banner, &h.EquipmentID, &h.Rarity, &h.PulledAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (db *DB) GetOrCreateGachaPity(charID int64) (*models.GachaPity, error) {
	var pity models.GachaPity
	err := db.conn.QueryRow(
		"SELECT normal_pity, advanced_pity FROM gacha_pity WHERE char_id = ?", charID,
	).Scan(&pity.NormalPity, &pity.AdvancedPity)
	if err == sql.ErrNoRows {
		_, err = db.conn.Exec("INSERT INTO gacha_pity (char_id) VALUES (?)", charID)
		if err != nil {
			return nil, err
		}
		return &models.GachaPity{}, nil
	}
	if err != nil {
		return nil, err
	}
	return &pity, nil
}

func (db *DB) UpdateGachaPity(charID int64, pity *models.GachaPity) error {
	_, err := db.conn.Exec(
		"UPDATE gacha_pity SET normal_pity = ?, advanced_pity = ? WHERE char_id = ?",
		pity.NormalPity, pity.AdvancedPity, charID,
	)
	return err
}

// ============================================================
// Battles
// ============================================================

func (db *DB) InsertBattle(b *models.BattleRecord) error {
	res, err := db.conn.Exec(
		`INSERT INTO battles (char_id, enemy_id, enemy_name, result, damage_dealt, damage_taken, accuracy, critical_hits, dodges, reward_exp, reward_crystals, material_drop, fought_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.CharID, b.EnemyID, b.EnemyName, string(b.Result), b.DamageDealt, b.DamageTaken,
		b.Accuracy, b.CriticalHits, b.Dodges, b.RewardEXP, b.RewardCrystals, b.MaterialDrop, time.Now(),
	)
	if err != nil {
		return err
	}
	b.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetBattleHistory(charID int64, limit int) ([]models.BattleRecord, error) {
	rows, err := db.conn.Query(
		`SELECT id, char_id, enemy_id, enemy_name, result, damage_dealt, damage_taken, accuracy, critical_hits, dodges, reward_exp, reward_crystals, material_drop, fought_at
		FROM battles WHERE char_id = ? ORDER BY fought_at DESC LIMIT ?`,
		charID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var battles []models.BattleRecord
	for rows.Next() {
		var b models.BattleRecord
		if err := rows.Scan(&b.ID, &b.CharID, &b.EnemyID, &b.EnemyName, &b.Result, &b.DamageDealt,
			&b.DamageTaken, &b.Accuracy, &b.CriticalHits, &b.Dodges, &b.RewardEXP, &b.RewardCrystals,
			&b.MaterialDrop, &b.FoughtAt); err != nil {
			return nil, err
		}
		battles = append(battles, b)
	}
	return battles, nil
}

func (db *DB) GetBattleStats(charID int64) (*models.BattleStatistics, error) {
	stats := &models.BattleStatistics{
		EnemiesDefeated: make(map[string]int),
	}

	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM battles WHERE char_id = ?", charID,
	).Scan(&stats.TotalBattles)
	if err != nil {
		return nil, err
	}

	err = db.conn.QueryRow(
		"SELECT COUNT(*) FROM battles WHERE char_id = ? AND result = 'win'", charID,
	).Scan(&stats.Wins)
	if err != nil {
		return nil, err
	}

	stats.Losses = stats.TotalBattles - stats.Wins
	if stats.TotalBattles > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.TotalBattles) * 100
	}

	err = db.conn.QueryRow(
		"SELECT COALESCE(SUM(damage_dealt), 0) FROM battles WHERE char_id = ?", charID,
	).Scan(&stats.TotalDamage)
	if err != nil {
		return nil, err
	}

	err = db.conn.QueryRow(
		"SELECT COALESCE(SUM(critical_hits), 0) FROM battles WHERE char_id = ?", charID,
	).Scan(&stats.TotalCrits)
	if err != nil {
		return nil, err
	}

	err = db.conn.QueryRow(
		"SELECT COALESCE(SUM(dodges), 0) FROM battles WHERE char_id = ?", charID,
	).Scan(&stats.TotalDodges)
	if err != nil {
		return nil, err
	}

	rows, err := db.conn.Query(
		"SELECT enemy_name, COUNT(*) FROM battles WHERE char_id = ? AND result = 'win' GROUP BY enemy_name",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, err
		}
		stats.EnemiesDefeated[name] = count
	}

	return stats, nil
}

// ============================================================
// Craft Recipes
// ============================================================

func (db *DB) InsertCraftRecipe(r *models.CraftRecipe) error {
	res, err := db.conn.Exec(
		`INSERT INTO craft_recipes (name, result_slot, result_rarity, cost_crystals, cost_common, cost_rare, cost_epic)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.Name, string(r.ResultSlot), string(r.ResultRarity), r.CostCrystals, r.CostCommon, r.CostRare, r.CostEpic,
	)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetAllCraftRecipes() ([]models.CraftRecipe, error) {
	rows, err := db.conn.Query(
		"SELECT id, name, result_slot, result_rarity, cost_crystals, cost_common, cost_rare, cost_epic FROM craft_recipes ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []models.CraftRecipe
	for rows.Next() {
		var r models.CraftRecipe
		if err := rows.Scan(&r.ID, &r.Name, &r.ResultSlot, &r.ResultRarity, &r.CostCrystals, &r.CostCommon, &r.CostRare, &r.CostEpic); err != nil {
			return nil, err
		}
		recipes = append(recipes, r)
	}
	return recipes, nil
}

func (db *DB) GetCraftRecipeCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM craft_recipes").Scan(&count)
	return count, err
}

// ============================================================
// Daily Rewards
// ============================================================

func (db *DB) GetLastDailyReward(charID int64) (*models.DailyReward, error) {
	var r models.DailyReward
	err := db.conn.QueryRow(
		"SELECT id, char_id, claimed_at, day, crystals FROM daily_rewards WHERE char_id = ? ORDER BY claimed_at DESC LIMIT 1",
		charID,
	).Scan(&r.ID, &r.CharID, &r.ClaimedAt, &r.Day, &r.Crystals)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) InsertDailyReward(r *models.DailyReward) error {
	res, err := db.conn.Exec(
		"INSERT INTO daily_rewards (char_id, claimed_at, day, crystals) VALUES (?, ?, ?, ?)",
		r.CharID, time.Now(), r.Day, r.Crystals,
	)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) CanClaimDailyReward(charID int64) (bool, error) {
	last, err := db.GetLastDailyReward(charID)
	if err != nil {
		return false, err
	}
	if last == nil {
		return true, nil
	}
	today := time.Now().Format("2006-01-02")
	lastDay := last.ClaimedAt.Format("2006-01-02")
	return today != lastDay, nil
}

func (db *DB) GetDailyRewardStreak(charID int64) (int, error) {
	last, err := db.GetLastDailyReward(charID)
	if err != nil {
		return 0, err
	}
	if last == nil {
		return 0, nil
	}
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	lastDay := last.ClaimedAt.Format("2006-01-02")
	if lastDay == today || lastDay == yesterday {
		return last.Day, nil
	}
	return 0, nil // streak broken
}

// ============================================================
// Gacha Statistics
// ============================================================

func (db *DB) GetGachaStats(charID int64) (*models.GachaStatistics, error) {
	stats := &models.GachaStatistics{
		PullsByBanner: make(map[models.GachaBanner]int),
		PullsByRarity: make(map[models.EquipmentRarity]int),
	}

	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM gacha_history WHERE char_id = ?", charID,
	).Scan(&stats.TotalPulls)
	if err != nil {
		return nil, err
	}

	rows, err := db.conn.Query(
		"SELECT banner, COUNT(*) FROM gacha_history WHERE char_id = ? GROUP BY banner", charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var banner models.GachaBanner
		var count int
		if err := rows.Scan(&banner, &count); err != nil {
			return nil, err
		}
		stats.PullsByBanner[banner] = count
	}

	rows2, err := db.conn.Query(
		"SELECT rarity, COUNT(*) FROM gacha_history WHERE char_id = ? GROUP BY rarity", charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var rarity models.EquipmentRarity
		var count int
		if err := rows2.Scan(&rarity, &count); err != nil {
			return nil, err
		}
		stats.PullsByRarity[rarity] = count
	}

	return stats, nil
}
