package database

import "fmt"

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS character (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL DEFAULT 'Hunter',
		attempts INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS hunter_profile (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		about TEXT NOT NULL DEFAULT '',
		goals TEXT NOT NULL DEFAULT '',
		priorities TEXT NOT NULL DEFAULT '',
		time_budget TEXT NOT NULL DEFAULT '',
		physical_constraints TEXT NOT NULL DEFAULT '',
		psychological_constraints TEXT NOT NULL DEFAULT '',
		day_routine TEXT NOT NULL DEFAULT '',
		primary_places TEXT NOT NULL DEFAULT '',
		dislikes TEXT NOT NULL DEFAULT '',
		support_style TEXT NOT NULL DEFAULT '',
		extra_details TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(char_id)
	);

	CREATE TABLE IF NOT EXISTS ai_profile (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		profile_text TEXT NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

		CREATE TABLE IF NOT EXISTS ai_suggestions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			raw_json TEXT NOT NULL,
			model TEXT NOT NULL DEFAULT '',
			error TEXT
		);

		CREATE TABLE IF NOT EXISTS achievements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			obtained_at DATETIME,
			is_unlocked INTEGER NOT NULL DEFAULT 0
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
		congratulations TEXT NOT NULL DEFAULT '',
		exp INTEGER NOT NULL DEFAULT 20,
		rank TEXT NOT NULL DEFAULT 'E',
		target_stat TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			is_daily INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER,
			expedition_id INTEGER,
			expedition_task_id INTEGER
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
		congratulations TEXT NOT NULL DEFAULT '',
		exp INTEGER NOT NULL DEFAULT 20,
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

		CREATE TABLE IF NOT EXISTS expeditions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			deadline DATETIME,
			reward_exp INTEGER NOT NULL DEFAULT 0,
			reward_stats TEXT NOT NULL DEFAULT '{}',
			is_repeatable INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS expedition_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expedition_id INTEGER NOT NULL REFERENCES expeditions(id),
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			is_completed INTEGER NOT NULL DEFAULT 0,
			progress_current INTEGER NOT NULL DEFAULT 0,
			progress_target INTEGER NOT NULL DEFAULT 1,
			reward_exp INTEGER NOT NULL DEFAULT 20,
			target_stat TEXT NOT NULL DEFAULT 'strength',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS completed_expeditions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			char_id INTEGER NOT NULL REFERENCES character(id),
			expedition_id INTEGER NOT NULL REFERENCES expeditions(id),
			completed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

	CREATE TABLE IF NOT EXISTS enemies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		rank TEXT NOT NULL DEFAULT 'E',
		type TEXT NOT NULL DEFAULT 'regular',
		level INTEGER NOT NULL DEFAULT 1,
		hp INTEGER NOT NULL DEFAULT 100,
		attack INTEGER NOT NULL DEFAULT 10,
		floor INTEGER NOT NULL DEFAULT 1,
		zone INTEGER NOT NULL DEFAULT 1,
		is_boss INTEGER NOT NULL DEFAULT 0,
		biome TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT 'NORMAL',
		is_transition INTEGER NOT NULL DEFAULT 0,
		target_winrate_min REAL NOT NULL DEFAULT 0,
		target_winrate_max REAL NOT NULL DEFAULT 0
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_enemies_name_unique ON enemies(name);

	CREATE TABLE IF NOT EXISTS streak_titles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		title TEXT NOT NULL,
		streak_days INTEGER NOT NULL,
		awarded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(char_id, streak_days)
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
		fought_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);


	
	
	
	CREATE TABLE IF NOT EXISTS enemy_unlocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		enemy_id INTEGER NOT NULL REFERENCES enemies(id),
		unlocked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(char_id, enemy_id)
	);

	CREATE TABLE IF NOT EXISTS battle_rewards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		char_id INTEGER NOT NULL REFERENCES character(id),
		enemy_id INTEGER NOT NULL REFERENCES enemies(id),
		title TEXT NOT NULL DEFAULT '',
		badge TEXT NOT NULL DEFAULT '',
		awarded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(char_id, enemy_id)
	);
	`
	_, err := db.conn.Exec(schema)
	if err != nil {
		return err
	}

	// Safe ALTER TABLE migrations for existing databases
	migrations := []string{
		"ALTER TABLE character ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE enemies ADD COLUMN floor INTEGER NOT NULL DEFAULT 1",
		"ALTER TABLE quests ADD COLUMN congratulations TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE quests ADD COLUMN exp INTEGER NOT NULL DEFAULT 20",
		"ALTER TABLE quests ADD COLUMN expedition_id INTEGER",
		"ALTER TABLE quests ADD COLUMN expedition_task_id INTEGER",
		"ALTER TABLE daily_quest_templates ADD COLUMN congratulations TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE daily_quest_templates ADD COLUMN exp INTEGER NOT NULL DEFAULT 20",
		"ALTER TABLE expedition_tasks ADD COLUMN reward_exp INTEGER NOT NULL DEFAULT 20",
		"ALTER TABLE expedition_tasks ADD COLUMN target_stat TEXT NOT NULL DEFAULT 'strength'",
		"ALTER TABLE character ADD COLUMN active_title TEXT NOT NULL DEFAULT ''",
		`UPDATE quests SET exp = CASE UPPER(rank)
			WHEN 'S' THEN 350
			WHEN 'A' THEN 200
			WHEN 'B' THEN 120
			WHEN 'C' THEN 70
			WHEN 'D' THEN 40
			ELSE 20
		END WHERE exp <= 0`,
		`UPDATE daily_quest_templates SET exp = CASE UPPER(rank)
			WHEN 'S' THEN 350
			WHEN 'A' THEN 200
			WHEN 'B' THEN 120
			WHEN 'C' THEN 70
			WHEN 'D' THEN 40
			ELSE 20
		END WHERE exp <= 0`,
		"UPDATE quests SET expedition_id = dungeon_id WHERE expedition_id IS NULL AND dungeon_id IS NOT NULL",
	}
	for _, m := range migrations {
		db.conn.Exec(m) // ignore errors (column already exists)
	}

	if err := db.migrateDungeonDataToExpeditions(); err != nil {
		return err
	}
	if err := db.migrateQuestDungeonLinksToExpeditions(); err != nil {
		return err
	}

	if err := db.addColumnIfMissing("enemies", "zone", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := db.addColumnIfMissing("enemies", "is_boss", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := db.addColumnIfMissing("enemies", "level", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := db.addColumnIfMissing("enemies", "biome", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := db.addColumnIfMissing("enemies", "role", "TEXT NOT NULL DEFAULT 'NORMAL'"); err != nil {
		return err
	}
	if err := db.addColumnIfMissing("enemies", "is_transition", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := db.addColumnIfMissing("enemies", "target_winrate_min", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := db.addColumnIfMissing("enemies", "target_winrate_max", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if _, err := db.conn.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_enemies_name_unique ON enemies(name)"); err != nil {
		return err
	}
	if err := db.NormalizeEnemyZones(); err != nil {
		return err
	}

	return nil
}

func (db *DB) migrateDungeonDataToExpeditions() error {
	if !db.tableExists("dungeons") {
		return nil
	}

	var expeditionCount int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM expeditions").Scan(&expeditionCount); err != nil {
		return err
	}
	if expeditionCount > 0 {
		return nil
	}

	_, err := db.conn.Exec(`
		INSERT INTO expeditions (id, name, description, deadline, reward_exp, reward_stats, is_repeatable, status, created_at, updated_at)
		SELECT
			id,
			name,
			description,
			NULL AS deadline,
			reward_exp,
			'{}' AS reward_stats,
			0 AS is_repeatable,
			CASE status
				WHEN 'completed' THEN 'completed'
				ELSE 'active'
			END AS status,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		FROM dungeons
	`)
	if err != nil {
		return err
	}

	if db.tableExists("dungeon_quests") {
		_, err = db.conn.Exec(`
			INSERT INTO expedition_tasks (
				expedition_id, title, description, is_completed, progress_current, progress_target,
				reward_exp, target_stat, created_at, updated_at
			)
			SELECT
				dq.dungeon_id,
				dq.title,
				dq.description,
				CASE
					WHEN e.status = 'completed' THEN 1
					ELSE 0
				END AS is_completed,
				CASE
					WHEN e.status = 'completed' THEN 1
					ELSE 0
				END AS progress_current,
				1 AS progress_target,
				COALESCE(dq.exp, 20) AS reward_exp,
				COALESCE(dq.target_stat, 'strength') AS target_stat,
				CURRENT_TIMESTAMP,
				CURRENT_TIMESTAMP
			FROM dungeon_quests dq
			JOIN expeditions e ON e.id = dq.dungeon_id
			ORDER BY dq.id
		`)
		if err != nil {
			return err
		}
	}

	if db.tableExists("completed_dungeons") {
		_, err = db.conn.Exec(`
			INSERT OR IGNORE INTO completed_expeditions (char_id, expedition_id, completed_at)
			SELECT char_id, dungeon_id, completed_at
			FROM completed_dungeons
		`)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) migrateQuestDungeonLinksToExpeditions() error {
	if !db.columnExistsFast("quests", "expedition_id") || !db.columnExistsFast("quests", "dungeon_id") {
		return nil
	}
	_, err := db.conn.Exec("UPDATE quests SET expedition_id = dungeon_id WHERE expedition_id IS NULL AND dungeon_id IS NOT NULL")
	return err
}

func (db *DB) addColumnIfMissing(table string, column string, definition string) error {
	exists, err := db.columnExists(table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = db.conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	return err
}

func (db *DB) columnExists(table string, column string) (bool, error) {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var cType string
		var notNull int
		var dfltValue any
		var pk int
		if err := rows.Scan(&cid, &name, &cType, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, nil
}

func (db *DB) columnExistsFast(table string, column string) bool {
	ok, err := db.columnExists(table, column)
	if err != nil {
		return false
	}
	return ok
}

func (db *DB) tableExists(name string) bool {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", name).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
