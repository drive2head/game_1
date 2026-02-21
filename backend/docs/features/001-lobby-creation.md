# Feature: Создание и присоединение к лобби

Status: draft
Priority: high
Depends on: —

## Context

Лобби — отправная точка всего игрового процесса. Прежде чем начать сессию,
игроки должны собраться в одной комнате. Один игрок создаёт лобби и получает
код приглашения, остальные присоединяются по этому коду.

См. [docs/rules.md](../../../docs/rules.md), разделы 3.1–3.3.

## Requirements

### Functional

- Сервер ДОЛЖЕН предоставить эндпоинт `POST /api/lobbies` для создания лобби.
- При создании сервер ДОЛЖЕН сгенерировать уникальный 6-символьный invite code
  (латиница, верхний регистр, например `ABCXYZ`).
- Создатель лобби ДОЛЖЕН автоматически стать первым игроком и администратором (`isAdmin: true`).
- Сервер ДОЛЖЕН сгенерировать `playerId` (UUID v4) для создателя и вернуть его в ответе.
- Лобби ДОЛЖНО быть создано с дефолтными `LobbySettings` (см. [data-models.md](../../../docs/data-models.md)).
- Сервер ДОЛЖЕН предоставить эндпоинт `POST /api/lobbies/{code}/join` для присоединения.
- При присоединении сервер ДОЛЖЕН валидировать:
  - Лобби с данным `code` существует → иначе `404 lobby_not_found`.
  - В лобби меньше `maxPlayers` игроков → иначе `409 lobby_full`.
  - Сессия не запущена (`state == "waiting"`) → иначе `409 game_in_progress`.
  - Никнейм не занят другим игроком в этом лобби → иначе `409 nickname_taken`.
  - Никнейм от 1 до 20 символов → иначе `400 invalid_nickname`.
- Присоединившийся игрок ДОЛЖЕН получить в ответе полный объект `Lobby` и свой `playerId`.

### Non-functional

- Invite code ДОЛЖЕН быть уникальным среди всех текущих активных лобби.
- Генерация invite code НЕ ДОЛЖНА быть предсказуемой (использовать `crypto/rand`).
- In-memory хранилище лобби ДОЛЖНО быть потокобезопасным (`sync.RWMutex`).

## Affected Files

Файлы, которые нужно **создать**:

| Файл | Назначение |
|---|---|
| `backend/game/lobby.go` | Структуры `Player`, `Lobby`, `LobbySettings`, `LobbyState`. Функции `NewLobby`, `AddPlayer`. |
| `backend/game/store.go` | `LobbyStore` — потокобезопасное in-memory хранилище лобби (`map[string]*Lobby`). |
| `backend/handler/lobby.go` | HTTP-хендлеры `CreateLobby`, `JoinLobby`. Регистрация роутов. |

Файлы, которые нужно **изменить**:

| Файл | Что изменить |
|---|---|
| `backend/main.go` | Создать `LobbyStore`, зарегистрировать новые хендлеры на `mux` (по аналогии с `/readyz`). |
| `backend/go.mod` | Может потребоваться обновление, если добавятся зависимости (не ожидается для этой фичи). |

Файлы, которые **НЕ трогать**:

- `frontend/**` — эта фича затрагивает только бэкенд.
- `.github/**` — CI не меняется.

## Data Model

Используются модели из [data-models.md](../../../docs/data-models.md):

- `Player` — создаётся при `POST /api/lobbies` и `POST /api/lobbies/{code}/join`.
- `Lobby` — создаётся при `POST /api/lobbies`.
- `LobbySettings` — вложен в `Lobby`, заполняется дефолтами при создании.

## API

Полные спецификации эндпоинтов — в [api-contract.md](../../../docs/api-contract.md), раздел 1:

- `POST /api/lobbies` → создание
- `POST /api/lobbies/{code}/join` → присоединение

### Пример: создание лобби

```
POST /api/lobbies
Content-Type: application/json

{"nickname": "Алиса"}
```

```
HTTP/1.1 201 Created
Content-Type: application/json

{
  "lobby": {
    "code": "XKRT9M",
    "players": [
      {"id": "550e8400-e29b-41d4-a716-446655440000", "nickname": "Алиса", "isAdmin": true, "ready": false}
    ],
    "settings": {
      "minPlayers": 3,
      "maxPlayers": 8,
      "locations": ["Больница", "Школа", "Банк", "Пляж", "Цирк", "Полицейский участок", "Супермаркет", "Ресторан"],
      "voteThreshold": 0.5,
      "votingTimerSeconds": 30
    },
    "state": "waiting",
    "currentSession": null
  },
  "playerId": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Пример: присоединение

```
POST /api/lobbies/XKRT9M/join
Content-Type: application/json

{"nickname": "Боб"}
```

```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "lobby": {
    "code": "XKRT9M",
    "players": [
      {"id": "550e8400-...", "nickname": "Алиса", "isAdmin": true, "ready": false},
      {"id": "6ba7b810-...", "nickname": "Боб", "isAdmin": false, "ready": false}
    ],
    "settings": { "..." },
    "state": "waiting",
    "currentSession": null
  },
  "playerId": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
}
```

## Acceptance Criteria

- [ ] `POST /api/lobbies` с валидным никнеймом возвращает 201 и объект лобби с invite code.
- [ ] `POST /api/lobbies` с пустым/длинным никнеймом возвращает 400.
- [ ] `POST /api/lobbies/{code}/join` с валидным кодом и никнеймом возвращает 200.
- [ ] `POST /api/lobbies/{code}/join` с несуществующим кодом возвращает 404.
- [ ] `POST /api/lobbies/{code}/join` при полном лобби возвращает 409.
- [ ] `POST /api/lobbies/{code}/join` с занятым никнеймом возвращает 409.
- [ ] Invite code уникален и состоит из 6 символов (A–Z, 0–9).
- [ ] Создатель лобби является администратором.
- [ ] `playerId` — валидный UUID v4.
- [ ] Конкурентные запросы на создание/присоединение не вызывают data race.

## Edge Cases and Error Handling

- **Никнейм из одних пробелов** — считать невалидным (trim перед проверкой длины).
- **Коллизия invite code** — при генерации проверять уникальность, при коллизии генерировать заново (вероятность мала при 36^6 вариантах, но обработать).
- **Максимум лобби** — в MVP не ограничиваем; при необходимости добавить лимит позже.
- **Content-Type не application/json** — возвращать 415 Unsupported Media Type.
- **Тело запроса невалидный JSON** — возвращать 400 Bad Request.

## Out of Scope

- WebSocket-соединение (отдельная фича).
- Управление лобби (кик, передача прав, изменение настроек) — отдельная фича.
- Готовность игроков и старт игры — отдельная фича.
- Персистентность (сохранение лобби на диск/в БД).
