# Feature: Старт игры — распределение ролей и локации

Status: draft
Priority: high
Depends on: [002-websocket-lobby](002-websocket-lobby.md)

## Context

Когда все игроки готовы, администратор запускает игру. Приложение выбирает
случайную локацию, назначает одного шпиона и отправляет каждому игроку его роль.

См. [docs/rules.md](../../../docs/rules.md), разделы 3.3, 4.1.
Бизнес-фича: [docs/features/002-game-start.md](../../../docs/features/002-game-start.md).

## Requirements

### Functional

- Команда `start_game` ДОЛЖНА быть доступна только администратору.
- Сервер ДОЛЖЕН валидировать:
  - Все игроки имеют `ready: true` → иначе WS-ошибка `players_not_ready`.
  - Количество игроков >= `minPlayers` → иначе WS-ошибка `not_enough_players`.
  - Лобби в состоянии `waiting` → иначе WS-ошибка `wrong_phase`.
- При старте сервер ДОЛЖЕН:
  1. Выбрать случайную локацию из `settings.locations`.
  2. Выбрать случайного шпиона из `players[]`.
  3. Сгенерировать случайный порядок ходов (`turnOrder`).
  4. Создать объект `Session` с `phase: "question"`, `currentTurnIndex: 0`.
  5. Перевести `Lobby.state` в `in_game`.
  6. Сбросить `ready: false` у всех игроков.
- Сервер ДОЛЖЕН отправить broadcast-событие `game_started` всем:
  ```json
  {
    "type": "game_started",
    "payload": {
      "sessionId": "uuid",
      "turnOrder": ["playerId1", "playerId2", ...],
      "phase": "question",
      "currentTurnIndex": 0
    }
  }
  ```
- Сервер ДОЛЖЕН отправить персональное событие `role_assigned` каждому игроку:
  - Обычному: `{ "role": "regular", "location": "Больница" }`
  - Шпиону: `{ "role": "spy", "location": null }`

### Non-functional

- Выбор шпиона и локации ДОЛЖЕН использовать `crypto/rand` для непредсказуемости.

## Affected Files

Файлы, которые нужно **создать**:

| Файл | Назначение |
|---|---|
| `backend/game/session.go` | Структура `Session`. Функция `NewSession(players []Player, locations []string) *Session` — создаёт сессию с рандомным шпионом, локацией и порядком ходов. |

Файлы, которые нужно **изменить**:

| Файл | Что изменить |
|---|---|
| `backend/game/lobby.go` | Добавить метод `StartGame() (*Session, error)` — валидация + создание сессии + переключение state. |
| `backend/handler/ws.go` | Добавить обработку команды `start_game`: вызов `lobby.StartGame()`, broadcast `game_started`, персональная отправка `role_assigned`. |

Файлы, которые **НЕ трогать**: `frontend/**`, `backend/handler/lobby.go`, `backend/game/store.go`.

## Data Model

Используются модели из [data-models.md](../../../docs/data-models.md):
- `Session` — создаётся при старте. В рамках этой фичи используются: `id`, `location`, `spyId`, `turnOrder`, `currentTurnIndex`, `phase`.
- Поля `voteProposal`, `votes`, `result` — остаются `null`/пустыми (фазы 2–3).

## API

Полные спецификации — в [api-contract.md](../../../docs/api-contract.md):
- Команда: `start_game`.
- Broadcast: `game_started`.
- Персональное: `role_assigned`.

## Acceptance Criteria

- [ ] `start_game` от не-администратора возвращает ошибку `not_admin`.
- [ ] `start_game` при неготовых игроках возвращает ошибку `players_not_ready`.
- [ ] `start_game` при < minPlayers возвращает ошибку `not_enough_players`.
- [ ] После старта `Lobby.state == "in_game"`.
- [ ] Ровно один игрок получает `role: "spy"`, остальные — `role: "regular"`.
- [ ] Обычные игроки получают название локации, шпион — `null`.
- [ ] `turnOrder` содержит ID всех игроков в случайном порядке.
- [ ] `game_started` broadcast содержит `turnOrder`, `phase`, `currentTurnIndex`.
- [ ] Повторный `start_game` при `state == "in_game"` возвращает `wrong_phase`.

## Edge Cases and Error Handling

- **Ровно minPlayers игроков** — старт разрешён.
- **Админ отменил ready после того, как все были готовы** — `start_game` отклоняется.
- **Игрок отключился между ready и start_game** — его `ready` сбрасывается, старт невозможен.

## Out of Scope

- Фаза вопросов (end_turn) — отдельная фича.
- Голосование — отдельная фича.
- Настройки игры — фаза 4.
