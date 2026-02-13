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
		"INSERT INTO quests (char_id, title, description, congratulations, exp, target_stat, status, created_at, is_daily, template_id, expedition_id, expedition_task_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		q.CharID,
		q.Title,
		q.Description,
		q.Congratulations,
		q.Exp,
		string(q.TargetStat),
		string(models.QuestActive),
		time.Now(),
		isDaily,
		q.TemplateID,
		q.ExpeditionID,
		q.ExpeditionTaskID,
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
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, expedition_id, expedition_task_id FROM quests WHERE char_id = ? AND status = ? ORDER BY created_at DESC",
		charID,
		string(models.QuestActive),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
}

func (db *DB) GetCompletedQuests(charID int64, limit int) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, expedition_id, expedition_task_id FROM quests WHERE char_id = ? AND status = ? ORDER BY completed_at DESC LIMIT ?",
		charID,
		string(models.QuestCompleted),
		limit,
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
		var expeditionID sql.NullInt64
		var expeditionTaskID sql.NullInt64
		if err := rows.Scan(
			&q.ID,
			&q.CharID,
			&q.Title,
			&q.Description,
			&q.Congratulations,
			&q.Exp,
			&q.TargetStat,
			&q.Status,
			&q.CreatedAt,
			&completedAt,
			&isDaily,
			&templateID,
			&expeditionID,
			&expeditionTaskID,
		); err != nil {
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
		if expeditionID.Valid {
			v := expeditionID.Int64
			q.ExpeditionID = &v
		}
		if expeditionTaskID.Valid {
			v := expeditionTaskID.Int64
			q.ExpeditionTaskID = &v
		}
		quests = append(quests, q)
	}
	return quests, nil
}

func (db *DB) CompleteQuest(questID int64) error {
	now := time.Now()
	_, err := db.conn.Exec(
		"UPDATE quests SET status = ?, completed_at = ? WHERE id = ?",
		string(models.QuestCompleted),
		now,
		questID,
	)
	return err
}

func (db *DB) FailQuest(questID int64) error {
	_, err := db.conn.Exec(
		"UPDATE quests SET status = ? WHERE id = ?",
		string(models.QuestFailed),
		questID,
	)
	return err
}

func (db *DB) SetQuestCreatedAt(questID int64, createdAt time.Time) error {
	_, err := db.conn.Exec(
		"UPDATE quests SET created_at = ? WHERE id = ?",
		createdAt,
		questID,
	)
	return err
}

func (db *DB) DeleteQuest(questID int64) error {
	_, err := db.conn.Exec("DELETE FROM quests WHERE id = ?", questID)
	return err
}

// GetQuestByID returns a single quest by its ID.
func (db *DB) GetQuestByID(questID int64) (*models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, expedition_id, expedition_task_id FROM quests WHERE id = ?",
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

// GetExpeditionActiveQuests returns active quests for a given expedition.
func (db *DB) GetExpeditionActiveQuests(charID int64, expeditionID int64) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, expedition_id, expedition_task_id FROM quests WHERE char_id = ? AND expedition_id = ? AND status = ? ORDER BY id",
		charID,
		expeditionID,
		string(models.QuestActive),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
}

// GetExpeditionAllQuests returns all quests (any status) for a given expedition.
func (db *DB) GetExpeditionAllQuests(charID int64, expeditionID int64) ([]models.Quest, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, title, description, congratulations, exp, target_stat, status, created_at, completed_at, is_daily, template_id, expedition_id, expedition_task_id FROM quests WHERE char_id = ? AND expedition_id = ? ORDER BY id",
		charID,
		expeditionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanQuestsExt(rows)
}

func (db *DB) HasActiveQuestForExpeditionTask(charID int64, taskID int64) (bool, error) {
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM quests WHERE char_id = ? AND expedition_task_id = ? AND status = ?",
		charID,
		taskID,
		string(models.QuestActive),
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *DB) FailActiveQuestsByExpedition(charID int64, expeditionID int64) error {
	_, err := db.conn.Exec(
		"UPDATE quests SET status = ? WHERE char_id = ? AND expedition_id = ? AND status = ?",
		string(models.QuestFailed),
		charID,
		expeditionID,
		string(models.QuestActive),
	)
	return err
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
		t.CharID,
		t.Title,
		t.Description,
		t.Congratulations,
		t.Exp,
		string(t.TargetStat),
		time.Now(),
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

// HasDailyQuestForToday checks if a quest from this template already exists today.
func (db *DB) HasDailyQuestForToday(charID int64, templateID int64) (bool, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM quests WHERE char_id = ? AND template_id = ? AND substr(created_at, 1, 10) = ?",
		charID,
		templateID,
		today,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ============================================================
// Expeditions
// ============================================================

func (db *DB) GetExpeditionCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM expeditions").Scan(&count)
	return count, err
}

func (db *DB) InsertExpedition(e *models.Expedition) error {
	rewardStats, err := marshalRewardStats(e.RewardStats)
	if err != nil {
		return err
	}
	if e.Status == "" {
		e.Status = models.ExpeditionActive
	}

	res, err := db.conn.Exec(
		"INSERT INTO expeditions (name, description, deadline, reward_exp, reward_stats, is_repeatable, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		e.Name,
		e.Description,
		e.Deadline,
		e.RewardEXP,
		rewardStats,
		boolToSQLiteInt(e.IsRepeatable),
		string(e.Status),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}
	e.ID, _ = res.LastInsertId()

	for i := range e.Tasks {
		t := &e.Tasks[i]
		t.ExpeditionID = e.ID
		if t.ProgressTarget <= 0 {
			t.ProgressTarget = 1
		}
		if t.ProgressCurrent < 0 {
			t.ProgressCurrent = 0
		}
		if t.TargetStat == "" {
			t.TargetStat = models.StatStrength
		}
		if t.RewardEXP <= 0 {
			t.RewardEXP = 20
		}
		if t.ProgressCurrent >= t.ProgressTarget {
			t.IsCompleted = true
			t.ProgressCurrent = t.ProgressTarget
		}

		resTask, err := db.conn.Exec(
			"INSERT INTO expedition_tasks (expedition_id, title, description, is_completed, progress_current, progress_target, reward_exp, target_stat, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			t.ExpeditionID,
			t.Title,
			t.Description,
			boolToSQLiteInt(t.IsCompleted),
			t.ProgressCurrent,
			t.ProgressTarget,
			t.RewardEXP,
			string(t.TargetStat),
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return err
		}
		t.ID, _ = resTask.LastInsertId()
	}

	return nil
}

func (db *DB) GetAllExpeditions() ([]models.Expedition, error) {
	rows, err := db.conn.Query(
		"SELECT id, name, description, deadline, reward_exp, reward_stats, is_repeatable, status, created_at, updated_at FROM expeditions ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expeditions []models.Expedition
	for rows.Next() {
		e, err := scanExpeditionRow(rows)
		if err != nil {
			return nil, err
		}
		tasks, err := db.GetExpeditionTasks(e.ID)
		if err != nil {
			return nil, err
		}
		e.Tasks = tasks
		expeditions = append(expeditions, *e)
	}
	return expeditions, nil
}

func (db *DB) GetExpeditionByID(expeditionID int64) (*models.Expedition, error) {
	rows, err := db.conn.Query(
		"SELECT id, name, description, deadline, reward_exp, reward_stats, is_repeatable, status, created_at, updated_at FROM expeditions WHERE id = ?",
		expeditionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("expedition not found: %d", expeditionID)
	}
	e, err := scanExpeditionRow(rows)
	if err != nil {
		return nil, err
	}
	tasks, err := db.GetExpeditionTasks(e.ID)
	if err != nil {
		return nil, err
	}
	e.Tasks = tasks
	return e, nil
}

func scanExpeditionRow(rows *sql.Rows) (*models.Expedition, error) {
	var e models.Expedition
	var deadline sql.NullTime
	var rewardStatsRaw string
	var isRepeatable int
	if err := rows.Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&deadline,
		&e.RewardEXP,
		&rewardStatsRaw,
		&isRepeatable,
		&e.Status,
		&e.CreatedAt,
		&e.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if deadline.Valid {
		t := deadline.Time
		e.Deadline = &t
	}
	e.IsRepeatable = isRepeatable == 1
	rewardStats, err := unmarshalRewardStats(rewardStatsRaw)
	if err != nil {
		return nil, err
	}
	e.RewardStats = rewardStats
	return &e, nil
}

func (db *DB) GetExpeditionTasks(expeditionID int64) ([]models.ExpeditionTask, error) {
	rows, err := db.conn.Query(
		"SELECT id, expedition_id, title, description, is_completed, progress_current, progress_target, reward_exp, target_stat, created_at, updated_at FROM expedition_tasks WHERE expedition_id = ? ORDER BY id",
		expeditionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.ExpeditionTask
	for rows.Next() {
		t, err := scanExpeditionTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, nil
}

func (db *DB) GetExpeditionTaskByID(taskID int64) (*models.ExpeditionTask, error) {
	rows, err := db.conn.Query(
		"SELECT id, expedition_id, title, description, is_completed, progress_current, progress_target, reward_exp, target_stat, created_at, updated_at FROM expedition_tasks WHERE id = ?",
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("expedition task not found: %d", taskID)
	}
	return scanExpeditionTaskRow(rows)
}

func scanExpeditionTaskRow(rows *sql.Rows) (*models.ExpeditionTask, error) {
	var t models.ExpeditionTask
	var completed int
	if err := rows.Scan(
		&t.ID,
		&t.ExpeditionID,
		&t.Title,
		&t.Description,
		&completed,
		&t.ProgressCurrent,
		&t.ProgressTarget,
		&t.RewardEXP,
		&t.TargetStat,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if t.ProgressTarget <= 0 {
		t.ProgressTarget = 1
	}
	if t.ProgressCurrent < 0 {
		t.ProgressCurrent = 0
	}
	t.IsCompleted = completed == 1 || t.ProgressCurrent >= t.ProgressTarget
	if t.IsCompleted && t.ProgressCurrent < t.ProgressTarget {
		t.ProgressCurrent = t.ProgressTarget
	}
	if t.RewardEXP <= 0 {
		t.RewardEXP = 20
	}
	if t.TargetStat == "" {
		t.TargetStat = models.StatStrength
	}
	return &t, nil
}

func (db *DB) IncrementExpeditionTaskProgress(taskID int64, delta int) (*models.ExpeditionTask, error) {
	task, err := db.GetExpeditionTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if delta == 0 {
		return task, nil
	}

	next := task.ProgressCurrent + delta
	if next < 0 {
		next = 0
	}
	completed := next >= task.ProgressTarget
	if completed {
		next = task.ProgressTarget
	}

	_, err = db.conn.Exec(
		"UPDATE expedition_tasks SET progress_current = ?, is_completed = ?, updated_at = ? WHERE id = ?",
		next,
		boolToSQLiteInt(completed),
		time.Now(),
		taskID,
	)
	if err != nil {
		return nil, err
	}
	return db.GetExpeditionTaskByID(taskID)
}

func (db *DB) FindNextIncompleteExpeditionTaskByTitle(expeditionID int64, title string) (*models.ExpeditionTask, error) {
	rows, err := db.conn.Query(
		"SELECT id, expedition_id, title, description, is_completed, progress_current, progress_target, reward_exp, target_stat, created_at, updated_at FROM expedition_tasks WHERE expedition_id = ? AND title = ? AND is_completed = 0 ORDER BY id LIMIT 1",
		expeditionID,
		title,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	return scanExpeditionTaskRow(rows)
}

func (db *DB) ResetExpeditionTasks(expeditionID int64) error {
	_, err := db.conn.Exec(
		"UPDATE expedition_tasks SET is_completed = 0, progress_current = 0, updated_at = ? WHERE expedition_id = ?",
		time.Now(),
		expeditionID,
	)
	return err
}

func (db *DB) UpdateExpeditionStatus(expeditionID int64, status models.ExpeditionStatus) error {
	_, err := db.conn.Exec(
		"UPDATE expeditions SET status = ?, updated_at = ? WHERE id = ?",
		string(status),
		time.Now(),
		expeditionID,
	)
	return err
}

func (db *DB) CompleteExpedition(charID int64, expeditionID int64) error {
	_, err := db.conn.Exec(
		"INSERT INTO completed_expeditions (char_id, expedition_id, completed_at) VALUES (?, ?, ?)",
		charID,
		expeditionID,
		time.Now(),
	)
	return err
}

func (db *DB) GetCompletedExpeditions(charID int64) ([]models.CompletedExpedition, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, expedition_id, completed_at FROM completed_expeditions WHERE char_id = ? ORDER BY completed_at DESC",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.CompletedExpedition
	for rows.Next() {
		var c models.CompletedExpedition
		if err := rows.Scan(&c.ID, &c.CharID, &c.ExpeditionID, &c.CompletedAt); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, nil
}

func (db *DB) IsExpeditionCompleted(charID int64, expeditionID int64) (bool, error) {
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM completed_expeditions WHERE char_id = ? AND expedition_id = ?",
		charID,
		expeditionID,
	).Scan(&count)
	return count > 0, err
}

func marshalRewardStats(stats map[models.StatType]int) (string, error) {
	if len(stats) == 0 {
		return "{}", nil
	}
	raw := make(map[string]int, len(stats))
	for stat, value := range stats {
		raw[string(stat)] = value
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalRewardStats(raw string) (map[models.StatType]int, error) {
	if raw == "" {
		return map[models.StatType]int{}, nil
	}
	var parsed map[string]int
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	result := make(map[models.StatType]int, len(parsed))
	for k, v := range parsed {
		result[models.StatType(k)] = v
	}
	return result, nil
}

func boolToSQLiteInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
