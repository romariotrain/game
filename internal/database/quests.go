package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"solo-leveling/internal/models"
)

// ============================================================
// Quests
// ============================================================

func (db *DB) CreateQuest(q *models.Quest) error {
	isDaily := 0
	if q.IsDaily {
		isDaily = 1
	}
	if q.Exp <= 0 {
		q.Exp = 20
	}
	res, err := db.conn.Exec(
		"INSERT INTO quests (char_id, title, description, congratulations, exp, target_stat, status, created_at, is_daily, template_id, dungeon_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		q.CharID, q.Title, q.Description, q.Congratulations, q.Exp, string(q.TargetStat), string(models.QuestActive), time.Now(), isDaily, q.TemplateID, q.DungeonID,
	)
	if err != nil {
		return err
	}
	q.ID, _ = res.LastInsertId()
	q.Status = models.QuestActive
	q.Rank = models.RankFromEXP(q.Exp)
	return nil
}

func (db *DB) GetActiveQuests(charID int64) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND status = ? ORDER BY created_at DESC",
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
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND status = ? ORDER BY completed_at DESC LIMIT ?",
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
		if err := rows.Scan(&q.ID, &q.CharID, &q.Title, &q.Description, &q.Congratulations, &q.Exp, &q.TargetStat, &q.Status, &q.CreatedAt, &completedAt, &isDaily, &templateID, &dungeonID); err != nil {
			return nil, err
		}
		if q.Exp <= 0 {
			q.Exp = 20
		}
		q.Rank = models.RankFromEXP(q.Exp)
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
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE id = ?",
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
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND dungeon_id = ? AND status = ? ORDER BY id",
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
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, dungeon_id FROM quests WHERE char_id = ? AND dungeon_id = ? ORDER BY id",
		charID, dungeonID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
}

// ============================================================
// Daily Quest Templates
// ============================================================

func (db *DB) CreateDailyTemplate(t *models.DailyQuestTemplate) error {
	if t.Exp <= 0 {
		t.Exp = 20
	}
	res, err := db.conn.Exec(
		"INSERT INTO daily_quest_templates (char_id, title, description, congratulations, exp, target_stat, active, created_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?)",
		t.CharID, t.Title, t.Description, t.Congratulations, t.Exp, string(t.TargetStat), time.Now(),
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
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, active, created_at FROM daily_quest_templates WHERE char_id = ? AND active = 1 ORDER BY created_at",
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
		if err := rows.Scan(&t.ID, &t.CharID, &t.Title, &t.Description, &t.Congratulations, &t.Exp, &t.TargetStat, &active, &t.CreatedAt); err != nil {
			return nil, err
		}
		if t.Exp <= 0 {
			t.Exp = 20
		}
		t.Rank = models.RankFromEXP(t.Exp)
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
		if qd.Exp <= 0 {
			qd.Exp = qd.Rank.BaseEXP()
		}
		res2, err := db.conn.Exec(
			"INSERT INTO dungeon_quests (dungeon_id, title, description, exp, rank, target_stat) VALUES (?, ?, ?, ?, ?, ?)",
			qd.DungeonID, qd.Title, qd.Description, qd.Exp, string(qd.Rank), string(qd.TargetStat),
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
		"SELECT id, dungeon_id, title, description, exp, rank, target_stat FROM dungeon_quests WHERE dungeon_id = ? ORDER BY id",
		dungeonID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []models.DungeonQuestDef
	for rows.Next() {
		var d models.DungeonQuestDef
		if err := rows.Scan(&d.ID, &d.DungeonID, &d.Title, &d.Description, &d.Exp, &d.Rank, &d.TargetStat); err != nil {
			return nil, err
		}
		if d.Exp <= 0 {
			d.Exp = d.Rank.BaseEXP()
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
