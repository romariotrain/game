package sim

// DefaultArchetypes returns the 5 canonical player archetypes.
func DefaultArchetypes() []Archetype {
	return []Archetype{
		Balanced(),
		INTBuild(),
		STRBuild(),
		HighEffortGrinder(),
		LowEffortCasual(),
	}
}

// Balanced — equal stat distribution, moderate effort.
func Balanced() Archetype {
	return Archetype{
		Name:               "Balanced",
		QuestsPerDay:       3,
		AvgMinutes:         30,
		AvgEffort:          3,
		AvgFriction:        2,
		MinutesStdDev:      10,
		EffortStdDev:       0.8,
		FrictionStdDev:     0.5,
		STRWeight:          0.25,
		AGIWeight:          0.25,
		INTWeight:          0.25,
		STAWeight:          0.25,
		FightsWhenPossible: true,
	}
}

// INTBuild — intelligence-focused, good memory, fewer battles.
func INTBuild() Archetype {
	return Archetype{
		Name:               "INT-build",
		QuestsPerDay:       3,
		AvgMinutes:         35,
		AvgEffort:          4,
		AvgFriction:        2,
		MinutesStdDev:      12,
		EffortStdDev:       0.7,
		FrictionStdDev:     0.5,
		STRWeight:          0.10,
		AGIWeight:          0.15,
		INTWeight:          0.55,
		STAWeight:          0.20,
		FightsWhenPossible: true,
	}
}

// STRBuild — strength-focused, aggressive fighter.
func STRBuild() Archetype {
	return Archetype{
		Name:               "STR-build",
		QuestsPerDay:       4,
		AvgMinutes:         25,
		AvgEffort:          3,
		AvgFriction:        2,
		MinutesStdDev:      8,
		EffortStdDev:       0.8,
		FrictionStdDev:     0.6,
		STRWeight:          0.50,
		AGIWeight:          0.20,
		INTWeight:          0.10,
		STAWeight:          0.20,
		FightsWhenPossible: true,
	}
}

// HighEffortGrinder — plays a lot, high effort quests.
func HighEffortGrinder() Archetype {
	return Archetype{
		Name:               "High-effort Grinder",
		QuestsPerDay:       6,
		AvgMinutes:         40,
		AvgEffort:          4,
		AvgFriction:        2,
		MinutesStdDev:      15,
		EffortStdDev:       0.6,
		FrictionStdDev:     0.5,
		STRWeight:          0.25,
		AGIWeight:          0.25,
		INTWeight:          0.25,
		STAWeight:          0.25,
		FightsWhenPossible: true,
	}
}

// LowEffortCasual — minimal engagement, few short quests.
func LowEffortCasual() Archetype {
	return Archetype{
		Name:               "Low-effort Casual",
		QuestsPerDay:       1,
		AvgMinutes:         15,
		AvgEffort:          2,
		AvgFriction:        1,
		MinutesStdDev:      5,
		EffortStdDev:       0.5,
		FrictionStdDev:     0.3,
		STRWeight:          0.25,
		AGIWeight:          0.25,
		INTWeight:          0.25,
		STAWeight:          0.25,
		FightsWhenPossible: true,
	}
}
