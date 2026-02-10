package database

import (
	"database/sql"
	"time"

	"solo-leveling/internal/models"
)

func (db *DB) GetHunterProfile(charID int64) (*models.HunterProfile, error) {
	row := db.conn.QueryRow(`
		SELECT char_id, about, goals, priorities, time_budget, physical_constraints,
		       psychological_constraints, day_routine, primary_places, dislikes,
		       support_style, extra_details, created_at, updated_at
		FROM hunter_profile
		WHERE char_id = ?
	`, charID)

	var p models.HunterProfile
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	if err := row.Scan(
		&p.CharID, &p.About, &p.Goals, &p.Priorities, &p.TimeBudget, &p.PhysicalConstraints,
		&p.PsychologicalConstraints, &p.DayRoutine, &p.PrimaryPlaces, &p.Dislikes,
		&p.SupportStyle, &p.ExtraDetails, &createdAt, &updatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}
	return &p, nil
}

func (db *DB) SaveHunterProfile(p *models.HunterProfile) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		INSERT INTO hunter_profile (
			char_id, about, goals, priorities, time_budget, physical_constraints,
			psychological_constraints, day_routine, primary_places, dislikes,
			support_style, extra_details, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(char_id) DO UPDATE SET
			about = excluded.about,
			goals = excluded.goals,
			priorities = excluded.priorities,
			time_budget = excluded.time_budget,
			physical_constraints = excluded.physical_constraints,
			psychological_constraints = excluded.psychological_constraints,
			day_routine = excluded.day_routine,
			primary_places = excluded.primary_places,
			dislikes = excluded.dislikes,
			support_style = excluded.support_style,
			extra_details = excluded.extra_details,
			updated_at = excluded.updated_at
	`,
		p.CharID, p.About, p.Goals, p.Priorities, p.TimeBudget, p.PhysicalConstraints,
		p.PsychologicalConstraints, p.DayRoutine, p.PrimaryPlaces, p.Dislikes,
		p.SupportStyle, p.ExtraDetails, now, now,
	)
	return err
}
