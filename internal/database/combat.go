package database

import (
	"database/sql"
	"time"

	"solo-leveling/internal/models"
)

// ============================================================
// Enemies
// ============================================================

func (db *DB) InsertEnemy(e *models.Enemy) error {
	res, err := db.conn.Exec(
		`INSERT INTO enemies (name, description, rank, type, hp, attack, floor)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.Name, e.Description, string(e.Rank), string(e.Type), e.HP, e.Attack, e.Floor,
	)
	if err != nil {
		return err
	}
	e.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetAllEnemies() ([]models.Enemy, error) {
	rows, err := db.conn.Query(
		"SELECT id, name, description, rank, type, hp, attack, floor FROM enemies ORDER BY floor, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enemies []models.Enemy
	for rows.Next() {
		var e models.Enemy
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Rank, &e.Type, &e.HP, &e.Attack, &e.Floor); err != nil {
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
		"SELECT id, name, description, rank, type, hp, attack, floor FROM enemies WHERE id = ?",
		id,
	).Scan(&e.ID, &e.Name, &e.Description, &e.Rank, &e.Type, &e.HP, &e.Attack, &e.Floor)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (db *DB) GetEnemiesByFloor(floor int) ([]models.Enemy, error) {
	rows, err := db.conn.Query(
		"SELECT id, name, description, rank, type, hp, attack, floor FROM enemies WHERE floor = ? ORDER BY id",
		floor,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enemies []models.Enemy
	for rows.Next() {
		var e models.Enemy
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Rank, &e.Type, &e.HP, &e.Attack, &e.Floor); err != nil {
			return nil, err
		}
		enemies = append(enemies, e)
	}
	return enemies, nil
}

func (db *DB) GetMaxFloor() (int, error) {
	var maxFloor int
	err := db.conn.QueryRow("SELECT COALESCE(MAX(floor), 0) FROM enemies").Scan(&maxFloor)
	return maxFloor, err
}

// ============================================================
// Enemy Unlocks
// ============================================================

func (db *DB) GetUnlockedEnemyIDs(charID int64) (map[int64]bool, error) {
	rows, err := db.conn.Query(
		"SELECT enemy_id FROM enemy_unlocks WHERE char_id = ?",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	unlocked := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		unlocked[id] = true
	}
	return unlocked, nil
}

func (db *DB) UnlockEnemy(charID, enemyID int64) error {
	_, err := db.conn.Exec(
		"INSERT OR IGNORE INTO enemy_unlocks (char_id, enemy_id) VALUES (?, ?)",
		charID, enemyID,
	)
	return err
}

// ============================================================
// Battle Rewards
// ============================================================

func (db *DB) GetBattleReward(charID, enemyID int64) (*models.BattleReward, error) {
	var r models.BattleReward
	err := db.conn.QueryRow(
		"SELECT id, char_id, enemy_id, title, badge, awarded_at FROM battle_rewards WHERE char_id = ? AND enemy_id = ?",
		charID, enemyID,
	).Scan(&r.ID, &r.CharID, &r.EnemyID, &r.Title, &r.Badge, &r.AwardedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) GetAllBattleRewards(charID int64) ([]models.BattleReward, error) {
	rows, err := db.conn.Query(
		"SELECT id, char_id, enemy_id, title, badge, awarded_at FROM battle_rewards WHERE char_id = ? ORDER BY awarded_at",
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rewards []models.BattleReward
	for rows.Next() {
		var r models.BattleReward
		if err := rows.Scan(&r.ID, &r.CharID, &r.EnemyID, &r.Title, &r.Badge, &r.AwardedAt); err != nil {
			return nil, err
		}
		rewards = append(rewards, r)
	}
	return rewards, nil
}

func (db *DB) InsertBattleReward(r *models.BattleReward) error {
	res, err := db.conn.Exec(
		"INSERT INTO battle_rewards (char_id, enemy_id, title, badge) VALUES (?, ?, ?, ?)",
		r.CharID, r.EnemyID, r.Title, r.Badge,
	)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	return nil
}

// ============================================================
// Battles
// ============================================================

func (db *DB) InsertBattle(b *models.BattleRecord) error {
	res, err := db.conn.Exec(
		`INSERT INTO battles (char_id, enemy_id, enemy_name, result, damage_dealt, damage_taken, accuracy, critical_hits, dodges, fought_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.CharID, b.EnemyID, b.EnemyName, string(b.Result), b.DamageDealt, b.DamageTaken,
		b.Accuracy, b.CriticalHits, b.Dodges, time.Now(),
	)
	if err != nil {
		return err
	}
	b.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetBattleHistory(charID int64, limit int) ([]models.BattleRecord, error) {
	rows, err := db.conn.Query(
		`SELECT id, char_id, enemy_id, enemy_name, result, damage_dealt, damage_taken, accuracy, critical_hits, dodges, fought_at
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
			&b.DamageTaken, &b.Accuracy, &b.CriticalHits, &b.Dodges, &b.FoughtAt); err != nil {
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
