package game

import "solo-leveling/internal/models"

func presetExpeditionTask(title string, description string, exp int, stat models.StatType, target int) models.ExpeditionTask {
	if target <= 0 {
		target = 1
	}
	if exp <= 0 {
		exp = 20
	}
	return models.ExpeditionTask{
		Title:           title,
		Description:     description,
		ProgressCurrent: 0,
		ProgressTarget:  target,
		RewardEXP:       exp,
		TargetStat:      stat,
	}
}

func GetPresetExpeditions() []models.Expedition {
	return []models.Expedition{
		{
			Name:         "Экспедиция: Пробуждение",
			Description:  "Долгосрочный старт: выстроить базовую дисциплину и ритм.",
			RewardEXP:    50,
			RewardStats:  map[models.StatType]int{models.StatStrength: 10, models.StatEndurance: 10},
			IsRepeatable: false,
			Status:       models.ExpeditionActive,
			Tasks: []models.ExpeditionTask{
				presetExpeditionTask("Разминка: 20 отжиманий", "Базовая активация тела.", 20, models.StatStrength, 1),
				presetExpeditionTask("Пробежка 1 км", "Поднять общий тонус и выносливость.", 20, models.StatEndurance, 1),
				presetExpeditionTask("Чтение 20 страниц", "Мини-сессия фокуса для разума.", 20, models.StatIntellect, 1),
			},
		},
		{
			Name:         "Экспедиция: Марафонская База",
			Description:  "Серия задач на устойчивую выносливость и темп.",
			RewardEXP:    200,
			RewardStats:  map[models.StatType]int{models.StatEndurance: 60},
			IsRepeatable: false,
			Status:       models.ExpeditionActive,
			Tasks: []models.ExpeditionTask{
				presetExpeditionTask("Пробежать 5 км", "Стабильный средний темп.", 120, models.StatEndurance, 1),
				presetExpeditionTask("100 берпи за день", "Взрывная выносливость и контроль дыхания.", 200, models.StatEndurance, 1),
				presetExpeditionTask("Планка 5 минут", "Статическая выдержка корпуса.", 70, models.StatEndurance, 1),
				presetExpeditionTask("Пробежать 10 км", "Финал блока с длинной дистанцией.", 200, models.StatEndurance, 1),
			},
		},
		{
			Name:         "Экспедиция: Силовой Контур",
			Description:  "Программа развития силы по верхнему и нижнему контуру.",
			RewardEXP:    200,
			RewardStats:  map[models.StatType]int{models.StatStrength: 60},
			IsRepeatable: false,
			Status:       models.ExpeditionActive,
			Tasks: []models.ExpeditionTask{
				presetExpeditionTask("100 отжиманий за подход", "Проверка силовой выносливости.", 120, models.StatStrength, 1),
				presetExpeditionTask("50 подтягиваний за день", "Тяговой объём под контроль техники.", 200, models.StatStrength, 1),
				presetExpeditionTask("200 приседаний", "Силовой блок для ног.", 120, models.StatStrength, 1),
				presetExpeditionTask("Тренировка 90 минут", "Полный силовой протокол.", 200, models.StatStrength, 1),
			},
		},
		{
			Name:         "Экспедиция: Мастер Разума",
			Description:  "Интеллектуальный блок: обучение, конспекты и применение.",
			RewardEXP:    200,
			RewardStats:  map[models.StatType]int{models.StatIntellect: 60},
			IsRepeatable: false,
			Status:       models.ExpeditionActive,
			Tasks: []models.ExpeditionTask{
				presetExpeditionTask("Прочитать 100 страниц", "Глубокое чтение без отвлечений.", 120, models.StatIntellect, 1),
				presetExpeditionTask("Пройти онлайн-курс", "Структурированное обучение до финала.", 200, models.StatIntellect, 1),
				presetExpeditionTask("Написать конспект", "Систематизировать материал в собственную структуру.", 70, models.StatIntellect, 1),
				presetExpeditionTask("Решить 10 сложных задач", "Применение знаний на практике.", 200, models.StatIntellect, 1),
			},
		},
		{
			Name:         "Экспедиция: Реакция и Контроль",
			Description:  "Цикл на скорость реакции, гибкость и координацию.",
			RewardEXP:    200,
			RewardStats:  map[models.StatType]int{models.StatAgility: 60},
			IsRepeatable: false,
			Status:       models.ExpeditionActive,
			Tasks: []models.ExpeditionTask{
				presetExpeditionTask("Растяжка 30 минут", "Мобильность и контроль амплитуды.", 70, models.StatAgility, 1),
				presetExpeditionTask("Скакалка: 500 прыжков", "Координация и темп дыхания.", 120, models.StatAgility, 1),
				presetExpeditionTask("Спринт: 10x100 м", "Взрывной скоростной блок.", 200, models.StatAgility, 1),
				presetExpeditionTask("Тренировка реакции", "60 минут упражнений на отклик.", 120, models.StatAgility, 1),
			},
		},
		{
			Name:        "Экспедиция: Универсал",
			Description: "Баланс всех характеристик в одном долгом цикле.",
			RewardEXP:   300,
			RewardStats: map[models.StatType]int{
				models.StatStrength:  25,
				models.StatAgility:   25,
				models.StatIntellect: 25,
				models.StatEndurance: 25,
			},
			IsRepeatable: false,
			Status:       models.ExpeditionActive,
			Tasks: []models.ExpeditionTask{
				presetExpeditionTask("Полная тренировка тела", "Комплекс на 60 минут.", 120, models.StatStrength, 1),
				presetExpeditionTask("Интервальный бег 5 км", "Скорость и выносливость вместе.", 120, models.StatAgility, 1),
				presetExpeditionTask("Изучить новую тему", "3 часа углублённого фокуса.", 120, models.StatIntellect, 1),
				presetExpeditionTask("Длинная пробежка 8 км", "Проверка запаса выносливости.", 200, models.StatEndurance, 1),
				presetExpeditionTask("Медитация 30 минут", "Стабилизация внимания и восстановления.", 70, models.StatEndurance, 1),
			},
		},
		{
			Name:        "Экспедиция: Легенда",
			Description: "Финальная многозадачная экспедиция для топ-уровня.",
			RewardEXP:   500,
			RewardStats: map[models.StatType]int{
				models.StatStrength:  60,
				models.StatAgility:   60,
				models.StatIntellect: 60,
				models.StatEndurance: 60,
			},
			IsRepeatable: false,
			Status:       models.ExpeditionActive,
			Tasks: []models.ExpeditionTask{
				presetExpeditionTask("Марафон 42 км", "Испытание пределов выносливости.", 350, models.StatEndurance, 1),
				presetExpeditionTask("200 отжиманий + 200 приседаний", "Пиковый силовой объём.", 350, models.StatStrength, 1),
				presetExpeditionTask("Прочитать книгу за день", "Предельная концентрация.", 350, models.StatIntellect, 1),
				presetExpeditionTask("Полный день активности", "12 часов без срывов плана.", 350, models.StatAgility, 1),
				presetExpeditionTask("Написать план на год", "Стратегия на длинную дистанцию.", 200, models.StatIntellect, 1),
			},
		},
	}
}
