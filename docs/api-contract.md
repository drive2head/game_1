# API Contract

Контракт взаимодействия между frontend и backend.
Модели данных — см. [data-models.md](data-models.md). Термины — см. [glossary.md](glossary.md).

---

## 1. REST API

Базовый путь: `/api`

### POST /api/lobbies

Создание нового лобби.

**Request:**

```json
{
  "nickname": "Алиса"
}
```

**Response (201 Created):**

```json
{
  "lobby": {
    "code": "ABCXYZ",
    "players": [
      { "id": "uuid-1", "nickname": "Алиса", "isAdmin": true, "ready": false }
    ],
    "settings": {
      "minPlayers": 3,
      "maxPlayers": 8,
      "locations": ["Больница", "Школа", "Банк", "..."],
      "voteThreshold": 0.5,
      "votingTimerSeconds": 30
    },
    "state": "waiting",
    "currentSession": null
  },
  "playerId": "uuid-1"
}
```

**Ошибки:**

| Код | Тело | Условие |
|---|---|---|
| 400 | `{"error": "invalid_nickname"}` | Никнейм пустой или длиннее 20 символов. |

---

### POST /api/lobbies/{code}/join

Присоединение к существующему лобби.

**Request:**

```json
{
  "nickname": "Боб"
}
```

**Response (200 OK):**

```json
{
  "lobby": { "...полный объект Lobby..." },
  "playerId": "uuid-2"
}
```

**Ошибки:**

| Код | Тело | Условие |
|---|---|---|
| 400 | `{"error": "invalid_nickname"}` | Никнейм пустой или длиннее 20 символов. |
| 404 | `{"error": "lobby_not_found"}` | Лобби с таким кодом не существует. |
| 409 | `{"error": "lobby_full"}` | В лобби уже `maxPlayers` игроков. |
| 409 | `{"error": "game_in_progress"}` | Сессия уже запущена, присоединение невозможно. |
| 409 | `{"error": "nickname_taken"}` | Никнейм уже занят другим игроком в этом лобби. |

---

### GET /readyz

Health check (уже реализован).

**Response (200 OK):**

```json
{
  "ready": true
}
```

---

## 2. WebSocket API

### Подключение

```
ws://<host>/api/lobbies/{code}/ws?playerId={playerId}
```

- `code` — код лобби
- `playerId` — UUID игрока, полученный при создании/присоединении через REST

**Ошибки подключения (HTTP до upgrade):**

| Код | Условие |
|---|---|
| 404 | Лобби не найдено. |
| 403 | `playerId` не принадлежит этому лобби. |

После успешного upgrade вся коммуникация идёт через JSON-сообщения.

---

### Формат сообщений

Все сообщения (клиент → сервер и сервер → клиент) имеют единый формат:

```json
{
  "type": "message_type",
  "payload": { ... }
}
```

---

### Команды (клиент → сервер)

| type | payload | Описание |
|---|---|---|
| `player_ready` | `{ "ready": boolean }` | Игрок меняет статус готовности. |
| `update_settings` | `{ "settings": LobbySettings }` | Администратор обновляет настройки (только admin). |
| `kick_player` | `{ "targetId": string }` | Администратор кикает игрока (только admin). |
| `transfer_admin` | `{ "targetId": string }` | Администратор передаёт права (только admin). |
| `start_game` | `{}` | Администратор запускает сессию (только admin, все игроки ready, >= minPlayers). |
| `end_turn` | `{}` | Активный игрок завершает свой ход (только active player, фаза `question`). |
| `propose_vote` | `{}` | Игрок предлагает досрочное голосование (фаза `question`). |
| `proposal_vote` | `{ "vote": "for" \| "against" }` | Игрок голосует за/против досрочного голосования (фаза `vote_proposal`). |
| `cast_vote` | `{ "suspectId": string \| null }` | Игрок голосует за подозреваемого или воздерживается (фаза `voting`). |
| `spy_guess` | `{ "location": string }` | Шпион угадывает локацию (фаза `spy_guess`, только spy). |
| `next_session` | `{}` | Администратор запускает следующую сессию (фаза `finished`, только admin). |

---

### События (сервер → клиент)

#### Broadcast (всем в лобби)

| type | payload | Когда |
|---|---|---|
| `lobby_updated` | `{ "lobby": Lobby }` | Любое изменение состояния лобби (игрок зашёл/вышел, настройки, ready). |
| `game_started` | `{ "sessionId": string, "turnOrder": string[], "phase": "question", "currentTurnIndex": 0 }` | Сессия началась. |
| `turn_changed` | `{ "currentTurnIndex": integer }` | Ход перешёл к следующему игроку. |
| `vote_proposal_started` | `{ "initiatorId": string, "threshold": integer }` | Инициировано предложение голосования. |
| `vote_proposal_updated` | `{ "votesFor": string[], "votesAgainst": string[] }` | Изменился счётчик голосов по предложению. |
| `vote_proposal_rejected` | `{}` | Предложение отклонено (порог не достигнут, все проголосовали). |
| `voting_started` | `{ "timerSeconds": integer }` | Началась фаза голосования. |
| `voting_completed` | `{ "votes": Vote[], "suspectId": string \| null }` | Голосование завершено. |
| `spy_guess_phase` | `{ "spyId": string, "locations": string[] }` | Шпион получает право угадать локацию. |
| `session_finished` | `{ "result": SessionResult, "location": string, "spyId": string }` | Сессия завершена, все данные раскрыты. |
| `player_disconnected` | `{ "playerId": string }` | Игрок потерял соединение. |
| `player_reconnected` | `{ "playerId": string }` | Игрок восстановил соединение. |

#### Персональные (одному игроку)

| type | payload | Когда |
|---|---|---|
| `role_assigned` | `{ "role": "spy" \| "regular", "location": string \| null }` | Сессия началась. `location` = `null` для шпиона. |
| `error` | `{ "error": string, "message": string }` | Команда отклонена (нет прав, неверная фаза, и т.д.). |

---

## 3. Коды ошибок WebSocket

| error | Описание |
|---|---|
| `not_admin` | Действие доступно только администратору. |
| `not_active_player` | Действие доступно только активному игроку. |
| `not_spy` | Действие доступно только шпиону. |
| `wrong_phase` | Действие не соответствует текущей фазе сессии. |
| `not_enough_players` | Недостаточно игроков для старта. |
| `players_not_ready` | Не все игроки отметились как готовые. |
| `already_voted` | Игрок уже проголосовал. |
| `invalid_target` | Указанный `targetId`/`suspectId` не найден среди игроков. |
| `invalid_location` | Указанная локация отсутствует в наборе. |
