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
		`INSERT INTO enemies (name, description, rank, type, level, hp, attack, floor, zone, is_boss, biome, role, is_transition, target_winrate_min, target_winrate_max)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Name,
		e.Description,
		string(e.Rank),
		string(e.Type),
		e.Level,
		e.HP,
		e.Attack,
		e.Floor,
		e.Zone,
		boolToInt(e.IsBoss),
		e.Biome,
		e.Role,
		boolToInt(e.IsTransition),
		e.TargetWinRateMin,
		e.TargetWinRateMax,
	)
	if err != nil {
		return err
	}
	e.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetAllEnemies() ([]models.Enemy, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, description, rank, type, level, hp, attack, floor, zone, is_boss, biome, role, is_transition, target_winrate_min, target_winrate_max
		 FROM enemies
		 ORDER BY zone, level, id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enemies []models.Enemy
	for rows.Next() {
		var e models.Enemy
		var isBoss int
		var isTransition int
		if err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Description,
			&e.Rank,
			&e.Type,
			&e.Level,
			&e.HP,
			&e.Attack,
			&e.Floor,
			&e.Zone,
			&isBoss,
			&e.Biome,
			&e.Role,
			&isTransition,
			&e.TargetWinRateMin,
			&e.TargetWinRateMax,
		); err != nil {
			return nil, err
		}
		e.IsBoss = isBoss == 1
		e.IsTransition = isTransition == 1
		enemies = append(enemies, e)
	}
	return enemies, nil
}

func (db *DB) GetEnemyCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM enemies").Scan(&count)
	return count, err
}

// EnemyCatalogNeedsReseed checks whether DB enemies differ from preset catalog.
func (db *DB) EnemyCatalogNeedsReseed(preset []models.Enemy) (bool, error) {
	if len(preset) == 0 {
		return false, nil
	}

	current, err := db.GetAllEnemies()
	if err != nil {
		return false, err
	}
	if len(current) != len(preset) {
		return true, nil
	}

	type sig struct {
		Zone   int
		Level  int
		IsBoss bool
	}

	expected := make(map[string]sig, len(preset))
	for _, e := range preset {
		expected[e.Name] = sig{
			Zone:   e.Zone,
			Level:  e.Level,
			IsBoss: e.IsBoss || e.Type == models.EnemyBoss,
		}
	}

	for _, e := range current {
		want, ok := expected[e.Name]
		if !ok {
			return true, nil
		}
		isBoss := e.IsBoss || e.Type == models.EnemyBoss
		if e.Zone != want.Zone || e.Level != want.Level || isBoss != want.IsBoss {
			return true, nil
		}
	}

	return false, nil
}

// ReplaceEnemyCatalog fully reseeds the enemy catalog and clears dependent battle progress.
func (db *DB) ReplaceEnemyCatalog(enemies []models.Enemy) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Keep schema consistent: enemy IDs change after full reseed, so dependent records must reset.
	if _, err := tx.Exec("DELETE FROM battle_rewards"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM enemy_unlocks"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM battles"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM enemies"); err != nil {
		return err
	}

	for i := range enemies {
		e := enemies[i]
		if _, err := tx.Exec(
			`INSERT INTO enemies (name, description, rank, type, level, hp, attack, floor, zone, is_boss, biome, role, is_transition, target_winrate_min, target_winrate_max)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.Name,
			e.Description,
			string(e.Rank),
			string(e.Type),
			e.Level,
			e.HP,
			e.Attack,
			e.Floor,
			e.Zone,
			boolToInt(e.IsBoss || e.Type == models.EnemyBoss),
			e.Biome,
			e.Role,
			boolToInt(e.IsTransition),
			e.TargetWinRateMin,
			e.TargetWinRateMax,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetEnemyByID(id int64) (*models.Enemy, error) {
	var e models.Enemy
	var isBoss int
	var isTransition int
	err := db.conn.QueryRow(
		`SELECT id, name, description, rank, type, level, hp, attack, floor, zone, is_boss, biome, role, is_transition, target_winrate_min, target_winrate_max
		 FROM enemies
		 WHERE id = ?`,
		id,
	).Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.Rank,
		&e.Type,
		&e.Level,
		&e.HP,
		&e.Attack,
		&e.Floor,
		&e.Zone,
		&isBoss,
		&e.Biome,
		&e.Role,
		&isTransition,
		&e.TargetWinRateMin,
		&e.TargetWinRateMax,
	)
	if err != nil {
		return nil, err
	}
	e.IsBoss = isBoss == 1
	e.IsTransition = isTransition == 1
	return &e, nil
}

func (db *DB) GetEnemiesByFloor(floor int) ([]models.Enemy, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, description, rank, type, level, hp, attack, floor, zone, is_boss, biome, role, is_transition, target_winrate_min, target_winrate_max
		 FROM enemies
		 WHERE floor = ?
		 ORDER BY id`,
		floor,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enemies []models.Enemy
	for rows.Next() {
		var e models.Enemy
		var isBoss int
		var isTransition int
		if err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Description,
			&e.Rank,
			&e.Type,
			&e.Level,
			&e.HP,
			&e.Attack,
			&e.Floor,
			&e.Zone,
			&isBoss,
			&e.Biome,
			&e.Role,
			&isTransition,
			&e.TargetWinRateMin,
			&e.TargetWinRateMax,
		); err != nil {
			return nil, err
		}
		e.IsBoss = isBoss == 1
		e.IsTransition = isTransition == 1
		enemies = append(enemies, e)
	}
	return enemies, nil
}

func (db *DB) GetMaxFloor() (int, error) {
	var maxFloor int
	err := db.conn.QueryRow("SELECT COALESCE(MAX(floor), 0) FROM enemies").Scan(&maxFloor)
	return maxFloor, err
}

func (db *DB) NormalizeEnemyZones() error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Preserve seeded zones; only normalize boss flags/type consistency.
	if _, err := tx.Exec(`UPDATE enemies SET type = CASE WHEN is_boss = 1 THEN ? ELSE ? END`, string(models.EnemyBoss), string(models.EnemyRegular)); err != nil {
		return err
	}

	rows, err := tx.Query("SELECT DISTINCT zone FROM enemies ORDER BY zone")
	if err != nil {
		return err
	}
	var zones []int
	for rows.Next() {
		var z int
		if err := rows.Scan(&z); err != nil {
			rows.Close()
			return err
		}
		zones = append(zones, z)
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, zone := range zones {
		rows, err := tx.Query(`SELECT id
			FROM enemies
			WHERE zone = ? AND is_boss = 1
			ORDER BY level DESC, (hp + attack) DESC, id ASC`, zone)
		if err != nil {
			return err
		}
		var bossIDs []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			bossIDs = append(bossIDs, id)
		}
		if err := rows.Close(); err != nil {
			return err
		}

		switch len(bossIDs) {
		case 1:
			if _, err := tx.Exec(`UPDATE enemies SET type = ? WHERE id = ?`, string(models.EnemyBoss), bossIDs[0]); err != nil {
				return err
			}
			continue
		case 0:
			var bossID int64
			err := tx.QueryRow(`SELECT id
			FROM enemies
			WHERE zone = ?
			ORDER BY
				level DESC,
				(hp + attack) DESC,
				hp DESC,
				attack DESC,
				id ASC
			LIMIT 1`, zone).Scan(&bossID)
			if err == sql.ErrNoRows {
				continue
			}
			if err != nil {
				return err
			}
			if _, err := tx.Exec(
				"UPDATE enemies SET is_boss = 1, type = ? WHERE id = ?",
				string(models.EnemyBoss), bossID,
			); err != nil {
				return err
			}
		default:
			keep := bossIDs[0]
			if _, err := tx.Exec(
				"UPDATE enemies SET is_boss = 0, type = ? WHERE zone = ? AND id <> ? AND is_boss = 1",
				string(models.EnemyRegular), zone, keep,
			); err != nil {
				return err
			}
			if _, err := tx.Exec(
				"UPDATE enemies SET is_boss = 1, type = ? WHERE id = ?",
				string(models.EnemyBoss), keep,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (db *DB) GetDefeatedEnemies(charID int64) ([]models.DefeatedEnemy, error) {
	rows, err := db.conn.Query(`
		SELECT e.id, e.name, e.description, e.rank, e.zone, e.is_boss, br.awarded_at
		FROM battle_rewards br
		JOIN enemies e ON e.id = br.enemy_id
		WHERE br.char_id = ?
		ORDER BY e.zone ASC, e.is_boss ASC, e.id ASC
	`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defeated []models.DefeatedEnemy
	for rows.Next() {
		var item models.DefeatedEnemy
		var isBoss int
		var defeatedAt time.Time
		if err := rows.Scan(&item.EnemyID, &item.Name, &item.Description, &item.Rank, &item.Zone, &isBoss, &defeatedAt); err != nil {
			return nil, err
		}
		item.IsBoss = isBoss == 1
		item.DefeatedAt = &defeatedAt
		defeated = append(defeated, item)
	}
	return defeated, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
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
