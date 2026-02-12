# Solo Leveling — TODO RPG (Go + Fyne)

Desktop TODO-приложение с RPG-прокачкой.
Главный принцип проекта: прогресс идёт от выполнения задач, а не от боёв.

## Что реализовано сейчас

- EXP начисляется только за квесты и данжи.
- Бои не дают EXP, тратят попытки и дают боевые награды (титулы/бейджи/unlock).
- Вкладка `Сегодня` собрана как компактный игровой цикл: персонаж -> следующий враг -> streak -> квесты.
- Запуск боёв выполняется только с вкладки `Сегодня` (карточка следующего врага).
- Есть вкладки `Сегодня`, `Задания`, `Прогресс`, `Достижения`, `Данжи`.
- Данные хранятся в SQLite (`~/.solo-leveling/game.db`), миграции создаются автоматически.

## Стек

- Go `1.25.5` (см. `go.mod`)
- Fyne `v2.7.2`
- SQLite (`github.com/mattn/go-sqlite3`, нужен CGO)

## Запуск

```bash
cd "путь/к/проекту"
go mod tidy
CGO_ENABLED=1 go run .
```

## Ключевые правила прогрессии

- EXP квеста считается формулой:
  - `EXP = round(minutes*0.6 + effort*4 + friction*3)`
- Ранг квеста определяется по EXP:
  - `E: <=10`, `D: <=18`, `C: <=28`, `B: <=40`, `A: <=55`, `S: >55`
- Попытки за квест:
  - `EXP < 15 -> +1`
  - `EXP 15..30 -> +2`
  - `EXP > 30 -> +3`
- Максимум попыток: `8`.
- Уровень стата: `ExpForLevel(level) = 50 + (level-1)*30`.
- Бои (включая боссов) EXP не дают.

## Текущий UI

### Режимы вкладок

`MinimalMode=true` (по умолчанию):
- `Сегодня`
- `Задания`
- `Прогресс`
- `Достижения`
- `Данжи`

`MinimalMode=false`:
- `Охотник`
- `Задания`
- `Данжи`
- `Статистика`
- `Достижения`
- `История`

### Экран «Сегодня»

Собран в `internal/ui/tabs/today.go`:

- Верхний блок (2 колонки):
  - карточка персонажа (большой портрет + мета + 4 стата с барами);
  - карточка «Следующий враг» (иконка/арт, ранг, HP/ATK, first-win reward, попытки, CTA боя).
- Портрет персонажа загружается из `assets/avatar.png` (если есть), автоматически подрезается по прозрачным полям и дополнительно увеличивается crop-zoom внутри бокса.
- Под верхним блоком: компактная строка streak.
- Ниже: «Задания на сегодня» в accordion по группам (`Главные`, `Средние`, `Быстрые`).

Условия CTA на «Следующий враг»:
- бой отключён feature-флагом -> кнопка недоступна;
- попытки `0` -> бой недоступен;
- если есть активные задания и сегодня не выполнено минимум 1 -> показывается требование выполнить задание.
- при нажатии CTA бой стартует сразу (без отдельной вкладки Tower).

## Задания и импорт JSON

Во вкладке `Задания`:
- создание задания вручную;
- завершение/провал/удаление;
- импорт JSON (кнопка `Импорт JSON`).

Поддерживаемый JSON (основной формат):

```json
[
  {
    "title": "Название",
    "desc": "Описание",
    "congratulations": "Поздравление",
    "minutes": 25,
    "effort": 3,
    "friction": 2,
    "stat": "INT",
    "is_daily": false
  }
]
```

Поддерживаются алиасы/legacy-поля:
- `name` вместо `title`
- `description` вместо `desc`
- `stats` как legacy fallback для статов

## Боевая система и прогресс врагов

- Линейная прогрессия: 15 врагов в фиксированной последовательности.
- Доступен только текущий враг последовательности (next unlock после первой победы).
- Все бои работают на Visual Memory-механике (`internal/game/combat/memory/memory.go`):
  - обычные враги: поле `6x6`,
  - боссы: поле `8x8`,
  - сложность задаётся количеством подсвеченных клеток и временем показа (не размером поля).
- Количество клеток:
  - базово по рангу врага: `E:6 D:8 C:10 B:12 A:14 S:16`,
  - `INT` уменьшает цель на `INT/3`,
  - минимум `4`,
  - для босса: `+2` к базе и ещё `+3` после редукции.
- Время показа: `2.5s + INT*0.05s`, clamp `[2.0..4.0]`.
- Раунд:
  - игрок выбирает до лимита `cellsToShow`,
  - точность = `correct/cellsToShow`,
  - урон игрока = `(10 + STR*2) * accuracy`, крит: `AGI*1.5%`, множитель `x1.5`,
  - урон врага = `random(ATK*0.8..ATK*1.2) - STA*0.25`, минимум `1`.
- Один запуск боя тратит `1` попытку, раунды внутри боя попытки не тратят.
- EXP за бои по-прежнему не начисляется.
- Босс-фаза `Pressure Puzzle` удалена; босс теперь проходит те же memory-раунды до победы/поражения (`internal/game/boss.go`, `internal/game/combat/boss/boss.go`).
- За первую победу над врагом выдаются:
  - титул,
  - бейдж,
  - разблокировка следующего врага.

## Данжи

- Пресеты инициализируются при старте (`InitDungeons`).
- Статусы: `locked -> available -> in_progress -> completed`.
- Вход в данж создаёт набор связанных квестов.
- За завершение данжа:
  - бонусный EXP во все статы,
  - титул данжа,
  - достижение `first_dungeon`.

## Достижения

Сидируются 4 базовых достижения:
- `first_task`
- `first_battle`
- `streak_7`
- `first_dungeon`

UI: вкладка `Достижения` (`internal/ui/tabs/achievements.go`).

## Hunter Profile (состояние на текущем этапе)

- Таблица и backend-методы для `hunter_profile` реализованы.
- Rule-based/LLM-слой рекомендаций в `internal/game/profile.go` присутствует.
- Отдельной активной UI-формы профиля в текущих вкладках нет.

## Feature flags

`internal/config/features.go`:

```go
type Features struct {
    MinimalMode bool
    Combat      bool
    Events      bool
}
```

Дефолт:
- `MinimalMode = true`
- `Combat = true`
- `Events = false`

## База данных (18 таблиц)

- `character`
- `hunter_profile`
- `ai_profile`
- `ai_suggestions`
- `achievements`
- `stat_levels`
- `quests`
- `skills`
- `daily_quest_templates`
- `daily_activity`
- `dungeons`
- `dungeon_quests`
- `completed_dungeons`
- `enemies`
- `streak_titles`
- `battles`
- `enemy_unlocks`
- `battle_rewards`

## Структура проекта

```text
.
├── main.go
├── assets/
│   ├── avatar.png
│   └── enemies/
├── internal/
│   ├── ai/
│   │   ├── suggestions/
│   │   └── ollama/
│   ├── config/
│   ├── database/
│   ├── game/
│   │   └── combat/
│   │       ├── memory/
│   │       └── boss/
│   ├── models/
│   └── ui/
│       ├── app.go
│       ├── theme.go
│       ├── components/
│       └── tabs/
└── README.md
```

## Тесты

```bash
go test ./...
```

Ключевые тесты:
- `internal/models/quest_exp_test.go`
- `internal/game/tower_test.go`
- `internal/game/combat/memory/memory_test.go`
- `internal/game/ai_rank_test.go`
- `internal/ai/suggestions/parse_test.go`
- `internal/ai/ollama/parse_test.go`
