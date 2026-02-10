package game

import (
	"fmt"

	"solo-leveling/internal/models"
)

// ============================================================
// Skills (unchanged)
// ============================================================

type SkillOption struct {
	Name        string
	Description string
	Multiplier  float64
}

func GetSkillOptions(stat models.StatType, level int) []SkillOption {
	catalog := map[models.StatType]map[int][]SkillOption{
		models.StatStrength: {
			3:  {{Name: "Железная хватка", Description: "Увеличивает получение EXP Силы", Multiplier: 1.10}},
			5:  {{Name: "Берсерк", Description: "Мощный прилив силы", Multiplier: 1.15}},
			8:  {{Name: "Титан", Description: "Сила титана течёт в венах", Multiplier: 1.20}},
			10: {{Name: "Разрушитель", Description: "Нет преград, которые не сломать", Multiplier: 1.25}},
			15: {{Name: "Монарх Силы", Description: "Абсолютная мощь", Multiplier: 1.35}},
		},
		models.StatAgility: {
			3:  {{Name: "Быстрые ноги", Description: "Увеличивает получение EXP Ловкости", Multiplier: 1.10}},
			5:  {{Name: "Тень", Description: "Движения быстрее взгляда", Multiplier: 1.15}},
			8:  {{Name: "Фантом", Description: "Неуловимый как призрак", Multiplier: 1.20}},
			10: {{Name: "Молния", Description: "Скорость молнии", Multiplier: 1.25}},
			15: {{Name: "Монарх Скорости", Description: "Время замедляется вокруг", Multiplier: 1.35}},
		},
		models.StatIntellect: {
			3:  {{Name: "Острый ум", Description: "Увеличивает получение EXP Интеллекта", Multiplier: 1.10}},
			5:  {{Name: "Аналитик", Description: "Видит паттерны во всём", Multiplier: 1.15}},
			8:  {{Name: "Стратег", Description: "На три шага впереди", Multiplier: 1.20}},
			10: {{Name: "Мудрец", Description: "Знания бесконечны", Multiplier: 1.25}},
			15: {{Name: "Монарх Разума", Description: "Абсолютный интеллект", Multiplier: 1.35}},
		},
		models.StatEndurance: {
			3:  {{Name: "Толстая кожа", Description: "Увеличивает получение EXP Выносливости", Multiplier: 1.10}},
			5:  {{Name: "Стойкость", Description: "Боль — лишь иллюзия", Multiplier: 1.15}},
			8:  {{Name: "Непробиваемый", Description: "Тело крепче стали", Multiplier: 1.20}},
			10: {{Name: "Бессмертный", Description: "Ничто не сломит волю", Multiplier: 1.25}},
			15: {{Name: "Монарх Воли", Description: "Абсолютная стойкость", Multiplier: 1.35}},
		},
	}

	if statCat, ok := catalog[stat]; ok {
		if options, ok := statCat[level]; ok {
			return options
		}
	}
	return nil
}

func (e *Engine) UnlockSkill(stat models.StatType, level int, optionIndex int) (*models.Skill, error) {
	options := GetSkillOptions(stat, level)
	if optionIndex < 0 || optionIndex >= len(options) {
		return nil, fmt.Errorf("invalid skill option index")
	}
	opt := options[optionIndex]
	skill := &models.Skill{
		CharID:      e.Character.ID,
		Name:        opt.Name,
		Description: opt.Description,
		StatType:    stat,
		Multiplier:  opt.Multiplier,
		UnlockedAt:  level,
	}
	if err := e.DB.CreateSkill(skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func (e *Engine) GetSkills() ([]models.Skill, error) {
	return e.DB.GetSkills(e.Character.ID)
}

func (e *Engine) ToggleSkill(skillID int64, active bool) error {
	return e.DB.ToggleSkill(skillID, active)
}

func (e *Engine) RenameCharacter(name string) error {
	e.Character.Name = name
	return e.DB.UpdateCharacterName(e.Character.ID, name)
}

func HunterRank(level int) string {
	switch {
	case level >= 40:
		return "S-Ранг Охотник"
	case level >= 30:
		return "A-Ранг Охотник"
	case level >= 20:
		return "B-Ранг Охотник"
	case level >= 14:
		return "C-Ранг Охотник"
	case level >= 8:
		return "D-Ранг Охотник"
	default:
		return "E-Ранг Охотник"
	}
}

func HunterRankColor(level int) string {
	switch {
	case level >= 40:
		return "#e74c3c"
	case level >= 30:
		return "#e67e22"
	case level >= 20:
		return "#9b59b6"
	case level >= 14:
		return "#4a7fbf"
	case level >= 8:
		return "#4a9e4a"
	default:
		return "#8a8a8a"
	}
}
