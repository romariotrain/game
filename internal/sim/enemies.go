package sim

// GetPresetEnemies returns 15 simulation enemies tuned for progression checks.
// Zones are assigned by rank: E/D=1, C/B=2, A/S=3.
// Boss flag is set for the strongest enemy per zone (highest rank+power).
func GetPresetEnemies() []EnemyDef {
	enemies := []EnemyDef{
		// Zone 1: E/D
		{Index: 0, Name: "Гоблин", Rank: "E", HP: 195, Attack: 10, AGI: 4, INT: 1, ExpectedMinLevel: 2, ExpectedMaxLevel: 3, Floor: 1, Zone: 1},
		{Index: 1, Name: "Скелет-страж", Rank: "E", HP: 285, Attack: 8, AGI: 5, INT: 2, ExpectedMinLevel: 3, ExpectedMaxLevel: 4, Floor: 2, Zone: 1},
		{Index: 2, Name: "Волк-тень", Rank: "D", HP: 295, Attack: 9, AGI: 8, INT: 3, ExpectedMinLevel: 4, ExpectedMaxLevel: 5, Floor: 3, Zone: 1},
		{Index: 3, Name: "Каменный голем", Rank: "D", HP: 315, Attack: 10, AGI: 4, INT: 5, ExpectedMinLevel: 5, ExpectedMaxLevel: 6, Floor: 4, Zone: 1},

		// Zone 2: C/B
		{Index: 4, Name: "Игрис — Рыцарь Крови", Rank: "C", HP: 370, Attack: 12, AGI: 10, INT: 8, ExpectedMinLevel: 8, ExpectedMaxLevel: 9, Floor: 5, Zone: 2},
		{Index: 5, Name: "Тёмный рыцарь", Rank: "C", HP: 420, Attack: 12, AGI: 9, INT: 6, ExpectedMinLevel: 9, ExpectedMaxLevel: 10, Floor: 6, Zone: 2},
		{Index: 6, Name: "Ядовитый виверн", Rank: "C", HP: 345, Attack: 16, AGI: 12, INT: 4, ExpectedMinLevel: 10, ExpectedMaxLevel: 11, Floor: 7, Zone: 2},
		{Index: 7, Name: "Демон-маг", Rank: "B", HP: 400, Attack: 15, AGI: 11, INT: 14, ExpectedMinLevel: 11, ExpectedMaxLevel: 12, Floor: 8, Zone: 2},
		{Index: 8, Name: "Ледяной великан", Rank: "B", HP: 410, Attack: 17, AGI: 7, INT: 9, ExpectedMinLevel: 12, ExpectedMaxLevel: 13, Floor: 9, Zone: 2},

		// Zone 3: A/S
		{Index: 9, Name: "Барука — Король муравьёв", Rank: "A", HP: 650, Attack: 18, AGI: 16, INT: 12, ExpectedMinLevel: 18, ExpectedMaxLevel: 19, Floor: 10, Zone: 3},
		{Index: 10, Name: "Архидемон", Rank: "A", HP: 745, Attack: 17, AGI: 14, INT: 16, ExpectedMinLevel: 19, ExpectedMaxLevel: 20, Floor: 11, Zone: 3},
		{Index: 11, Name: "Драконид", Rank: "A", HP: 750, Attack: 18, AGI: 18, INT: 10, ExpectedMinLevel: 20, ExpectedMaxLevel: 21, Floor: 12, Zone: 3},
		{Index: 12, Name: "Страж Бездны", Rank: "S", HP: 740, Attack: 20, AGI: 20, INT: 15, ExpectedMinLevel: 21, ExpectedMaxLevel: 22, Floor: 13, Zone: 3},
		{Index: 13, Name: "Монарх Хаоса", Rank: "S", HP: 735, Attack: 21, AGI: 22, INT: 18, ExpectedMinLevel: 22, ExpectedMaxLevel: 23, Floor: 14, Zone: 3},
		{Index: 14, Name: "Монарх Теней", Rank: "S", HP: 760, Attack: 23, AGI: 24, INT: 20, ExpectedMinLevel: 23, ExpectedMaxLevel: 24, Floor: 15, Zone: 3},
	}

	// Mark zone bosses: strongest enemy per zone (highest rank + HP + Attack).
	assignBosses(enemies)
	return enemies
}

func assignBosses(enemies []EnemyDef) {
	// For each zone, find the enemy with highest rank, then highest (HP + Attack).
	zoneBoss := make(map[int]int) // zone -> index of boss candidate
	for i, e := range enemies {
		best, exists := zoneBoss[e.Zone]
		if !exists {
			zoneBoss[e.Zone] = i
			continue
		}
		if rankOrd(e.Rank) > rankOrd(enemies[best].Rank) {
			zoneBoss[e.Zone] = i
		} else if rankOrd(e.Rank) == rankOrd(enemies[best].Rank) {
			if e.HP+e.Attack > enemies[best].HP+enemies[best].Attack {
				zoneBoss[e.Zone] = i
			}
		}
	}
	for _, idx := range zoneBoss {
		enemies[idx].IsBoss = true
	}
}

func rankOrd(rank string) int {
	switch rank {
	case "E":
		return 0
	case "D":
		return 1
	case "C":
		return 2
	case "B":
		return 3
	case "A":
		return 4
	case "S":
		return 5
	default:
		return 0
	}
}
