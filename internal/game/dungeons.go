package game

import "solo-leveling/internal/models"

func GetPresetDungeons() []models.Dungeon {
	return []models.Dungeon{
		{
			Name:        "Данж: Новичок",
			Description: "Первое испытание для начинающего охотника. Докажи, что ты достоин звания.",
			Requirements: []models.StatRequirement{
				{StatType: models.StatStrength, MinLevel: 3},
			},
			RewardTitle: "Пробудившийся",
			RewardEXP:   50,
			QuestDefinitions: []models.DungeonQuestDef{
				{Title: "Разминка: 20 отжиманий", Description: "Базовая тренировка тела", Rank: models.RankE, TargetStat: models.StatStrength},
				{Title: "Пробежка 1 км", Description: "Испытание выносливости", Rank: models.RankE, TargetStat: models.StatEndurance},
				{Title: "Прочитать 20 страниц книги", Description: "Тренировка разума", Rank: models.RankE, TargetStat: models.StatIntellect},
			},
		},
		{
			Name:        "Данж: Марафон",
			Description: "Данж для тех, кто покорил свою выносливость. Длинная дистанция ждёт.",
			Requirements: []models.StatRequirement{
				{StatType: models.StatEndurance, MinLevel: 15},
			},
			RewardTitle: "Марафонец",
			RewardEXP:   200,
			QuestDefinitions: []models.DungeonQuestDef{
				{Title: "Пробежать 5 км", Description: "Средняя дистанция", Rank: models.RankB, TargetStat: models.StatEndurance},
				{Title: "100 берпи за день", Description: "Взрывная выносливость", Rank: models.RankA, TargetStat: models.StatEndurance},
				{Title: "Планка 5 минут", Description: "Статическая выносливость", Rank: models.RankC, TargetStat: models.StatEndurance},
				{Title: "Пробежать 10 км", Description: "Финальное испытание", Rank: models.RankA, TargetStat: models.StatEndurance},
			},
		},
		{
			Name:        "Данж: Силач",
			Description: "Только истинная сила позволит пройти этот данж. Стальные мышцы и железная воля.",
			Requirements: []models.StatRequirement{
				{StatType: models.StatStrength, MinLevel: 15},
			},
			RewardTitle: "Титан",
			RewardEXP:   200,
			QuestDefinitions: []models.DungeonQuestDef{
				{Title: "100 отжиманий за подход", Description: "Испытание чистой силы", Rank: models.RankB, TargetStat: models.StatStrength},
				{Title: "50 подтягиваний за день", Description: "Верхняя часть тела", Rank: models.RankA, TargetStat: models.StatStrength},
				{Title: "200 приседаний", Description: "Сила ног", Rank: models.RankB, TargetStat: models.StatStrength},
				{Title: "Тренировка на максимум", Description: "Полная силовая тренировка 1.5 часа", Rank: models.RankA, TargetStat: models.StatStrength},
			},
		},
		{
			Name:        "Данж: Мастер Разума",
			Description: "Данж для интеллектуалов. Знания — это сила другого рода.",
			Requirements: []models.StatRequirement{
				{StatType: models.StatIntellect, MinLevel: 15},
			},
			RewardTitle: "Мудрец",
			RewardEXP:   200,
			QuestDefinitions: []models.DungeonQuestDef{
				{Title: "Прочитать 100 страниц", Description: "Глубокое погружение в знания", Rank: models.RankB, TargetStat: models.StatIntellect},
				{Title: "Пройти онлайн-курс", Description: "Структурированное обучение", Rank: models.RankA, TargetStat: models.StatIntellect},
				{Title: "Написать конспект", Description: "Систематизация знаний", Rank: models.RankC, TargetStat: models.StatIntellect},
				{Title: "Решить 10 сложных задач", Description: "Применение знаний на практике", Rank: models.RankA, TargetStat: models.StatIntellect},
			},
		},
		{
			Name:        "Данж: Ловкач",
			Description: "Скорость и точность — вот что нужно, чтобы пройти этот данж.",
			Requirements: []models.StatRequirement{
				{StatType: models.StatAgility, MinLevel: 15},
			},
			RewardTitle: "Призрак",
			RewardEXP:   200,
			QuestDefinitions: []models.DungeonQuestDef{
				{Title: "Растяжка 30 минут", Description: "Гибкость тела", Rank: models.RankC, TargetStat: models.StatAgility},
				{Title: "Скакалка: 500 прыжков", Description: "Координация и скорость", Rank: models.RankB, TargetStat: models.StatAgility},
				{Title: "Спринт: 10x100м", Description: "Взрывная скорость", Rank: models.RankA, TargetStat: models.StatAgility},
				{Title: "Тренировка реакции", Description: "Час практики на реакцию", Rank: models.RankB, TargetStat: models.StatAgility},
			},
		},
		{
			Name:        "Данж: Универсал",
			Description: "Только гармонично развитый охотник пройдёт это испытание. Все статы должны быть сильны.",
			Requirements: []models.StatRequirement{
				{StatType: models.StatStrength, MinLevel: 10},
				{StatType: models.StatAgility, MinLevel: 10},
				{StatType: models.StatIntellect, MinLevel: 10},
				{StatType: models.StatEndurance, MinLevel: 10},
			},
			RewardTitle: "Универсал",
			RewardEXP:   300,
			QuestDefinitions: []models.DungeonQuestDef{
				{Title: "Полная тренировка тела", Description: "Час силовых упражнений", Rank: models.RankB, TargetStat: models.StatStrength},
				{Title: "Интервальный бег 5 км", Description: "Скорость и выносливость", Rank: models.RankB, TargetStat: models.StatAgility},
				{Title: "Изучить новую тему", Description: "3 часа глубокого обучения", Rank: models.RankB, TargetStat: models.StatIntellect},
				{Title: "Длинная пробежка 8 км", Description: "Испытание воли", Rank: models.RankA, TargetStat: models.StatEndurance},
				{Title: "Медитация 30 минут", Description: "Гармония тела и разума", Rank: models.RankC, TargetStat: models.StatEndurance},
			},
		},
		{
			Name:        "Данж: Легенда",
			Description: "Последнее испытание. Только легендарный охотник с максимальной подготовкой сможет выйти победителем.",
			Requirements: []models.StatRequirement{
				{StatType: models.StatStrength, MinLevel: 20},
				{StatType: models.StatAgility, MinLevel: 20},
				{StatType: models.StatIntellect, MinLevel: 20},
				{StatType: models.StatEndurance, MinLevel: 20},
			},
			RewardTitle: "Легенда",
			RewardEXP:   500,
			QuestDefinitions: []models.DungeonQuestDef{
				{Title: "Марафон 42 км", Description: "Истинное испытание пределов", Rank: models.RankS, TargetStat: models.StatEndurance},
				{Title: "200 отжиманий + 200 приседаний", Description: "Абсолютная сила", Rank: models.RankS, TargetStat: models.StatStrength},
				{Title: "Прочитать книгу за день", Description: "Предельная концентрация", Rank: models.RankS, TargetStat: models.StatIntellect},
				{Title: "Полный день активности", Description: "12 часов без остановки", Rank: models.RankS, TargetStat: models.StatAgility},
				{Title: "Написать план на год", Description: "Стратегическое видение", Rank: models.RankA, TargetStat: models.StatIntellect},
			},
		},
	}
}
