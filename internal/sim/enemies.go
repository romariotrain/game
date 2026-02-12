package sim

import "math"

type zoneTemplate struct {
	Zone     int
	Biome    string
	MinLevel int
	MaxLevel int
	Names    []string
}

// GetPresetEnemies returns 50 simulation enemies (5 zones x 10 slots).
func GetPresetEnemies() []EnemyDef {
	zones := []zoneTemplate{
		{
			Zone:     1,
			Biome:    "swamp",
			MinLevel: 2,
			MaxLevel: 6,
			Names: []string{
				"Квакающий Разведчик",
				"Болотный Пиявочник",
				"Гнилотный Шаман",
				"Трясинный Волк",
				"Моховой Голем",
				"Токсичный Удильщик",
				"Ведьмин Грибник",
				"Слизень-Разъедатель",
				"Пасть Топи",
				"Хозяйка Туманов Морра",
			},
		},
		{
			Zone:     2,
			Biome:    "ruins",
			MinLevel: 7,
			MaxLevel: 12,
			Names: []string{
				"Ржавый Страж Портала",
				"Пыльный Скелет-Рыцарь",
				"Летучая Моль Проклятий",
				"Крипт-Охотник",
				"Костяной Арбалетчик",
				"Каменный Идол",
				"Тень Архива",
				"Пожиратель Реликвий",
				"Жрец Разломанных Печатей",
				"Архонт Руин Кальдрос",
			},
		},
		{
			Zone:     3,
			Biome:    "frost",
			MinLevel: 13,
			MaxLevel: 18,
			Names: []string{
				"Снежный Падальщик",
				"Морозный Пехотинец",
				"Ледяная Гарпия",
				"Вьюжный Волк",
				"Осколочный Голем",
				"Северный Берсерк",
				"Хрустальный Охотник",
				"Ледяной Колдун",
				"Белый Йети",
				"Король Вьюги Хельгрим",
			},
		},
		{
			Zone:     4,
			Biome:    "volcanic",
			MinLevel: 19,
			MaxLevel: 24,
			Names: []string{
				"Пепельный Разбойник",
				"Обугленный Скелет",
				"Лавовый Плевун",
				"Огненный Гончий",
				"Шлаковый Голем",
				"Жрец Пепла",
				"Крылатый Угольник",
				"Демон Искр",
				"Плавильщик Костей",
				"Владыка Разломов Азгар",
			},
		},
		{
			Zone:     5,
			Biome:    "void",
			MinLevel: 25,
			MaxLevel: 30,
			Names: []string{
				"Безликий Смотритель",
				"Паразит Пустоты",
				"Теневой Дуэлянт",
				"Пожиратель Света",
				"Хор Бездны",
				"Клеймённый Инквизитор",
				"Рыцарь Нулевой Тени",
				"Коготь Монарха",
				"Оракул Тишины",
				"Монарх Бездны Ноктэрн",
			},
		},
	}

	enemies := make([]EnemyDef, 0, 50)
	index := 0
	floor := 1
	for _, zone := range zones {
		for i, name := range zone.Names {
			slot := i + 1
			level := levelForSlot(zone.MinLevel, zone.MaxLevel, slot)
			role := roleForSlot(slot)
			targetMin, targetMax := targetWinRateForSlot(slot)
			isBoss := slot == 10

			baseHP := 160 + level*24
			baseATK := 8 + int(math.Round(float64(level)*0.9))
			power := rolePowerMultiplier(slot)

			enemy := EnemyDef{
				Index:  index,
				Name:   name,
				Rank:   rankForZoneAndSlot(zone.Zone, slot),
				Role:   role,
				Biome:  zone.Biome,
				Level:  level,
				HP:     int(math.Round(float64(baseHP) * power)),
				Attack: int(math.Round(float64(baseATK) * (0.80 + power*0.35))),
				AGI:    level + slot%3,
				INT:    level/2 + slot,

				ExpectedMinLevel: maxInt(1, level-1),
				ExpectedMaxLevel: level + 1,
				Floor:            floor,
				Zone:             zone.Zone,
				IsBoss:           isBoss,
				IsTransition:     slot <= 2,
				TargetWinRateMin: targetMin,
				TargetWinRateMax: targetMax,
			}
			if enemy.HP < 1 {
				enemy.HP = 1
			}
			if enemy.Attack < 1 {
				enemy.Attack = 1
			}

			enemies = append(enemies, enemy)
			index++
			floor++
		}
	}

	ensureOneBossPerZone(enemies)
	return enemies
}

func ensureOneBossPerZone(enemies []EnemyDef) {
	zoneBoss := map[int]int{}
	for i := range enemies {
		if enemies[i].IsBoss {
			zoneBoss[enemies[i].Zone]++
		}
	}
	for i := range enemies {
		if zoneBoss[enemies[i].Zone] > 0 {
			continue
		}
		if slotFromFloor(enemies[i].Floor) == 10 {
			enemies[i].IsBoss = true
			zoneBoss[enemies[i].Zone]++
		}
	}
}

func slotFromFloor(floor int) int {
	mod := floor % 10
	if mod == 0 {
		return 10
	}
	return mod
}

func levelForSlot(minLevel, maxLevel, slot int) int {
	pattern := []int{0, 1, 1, 2, 2, 3, 4, 4, 5, 5}
	if slot < 1 {
		slot = 1
	}
	if slot > len(pattern) {
		slot = len(pattern)
	}
	level := minLevel + pattern[slot-1]
	if level > maxLevel {
		level = maxLevel
	}
	return level
}

func roleForSlot(slot int) string {
	switch slot {
	case 1:
		return "TRANSITION"
	case 2:
		return "TRANSITION_ELITE"
	case 3:
		return "NORMAL"
	case 4:
		return "HARD"
	case 5:
		return "EASY"
	case 6:
		return "HARD"
	case 7:
		return "ELITE"
	case 8:
		return "NORMAL"
	case 9:
		return "MINIBOSS"
	case 10:
		return "BOSS"
	default:
		return "NORMAL"
	}
}

func rolePowerMultiplier(slot int) float64 {
	switch slot {
	case 1:
		return 1.16
	case 2:
		return 1.24
	case 3:
		return 1.00
	case 4:
		return 1.08
	case 5:
		return 0.90
	case 6:
		return 1.10
	case 7:
		return 1.20
	case 8:
		return 1.02
	case 9:
		return 1.30
	case 10:
		return 1.38
	default:
		return 1.00
	}
}

func targetWinRateForSlot(slot int) (float64, float64) {
	switch slot {
	case 1:
		return 15, 25
	case 2:
		return 12, 20
	case 3:
		return 30, 45
	case 4:
		return 20, 30
	case 5:
		return 45, 60
	case 6:
		return 20, 30
	case 7:
		return 12, 20
	case 8:
		return 30, 45
	case 9:
		return 8, 15
	case 10:
		return 5, 12
	default:
		return 30, 45
	}
}

func rankForZoneAndSlot(zone, slot int) string {
	switch zone {
	case 1:
		if slot <= 5 {
			return "E"
		}
		return "D"
	case 2:
		if slot <= 5 {
			return "C"
		}
		return "B"
	case 3:
		if slot <= 4 {
			return "B"
		}
		if slot <= 8 {
			return "A"
		}
		return "S"
	case 4:
		if slot <= 5 {
			return "A"
		}
		return "S"
	default:
		return "S"
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
