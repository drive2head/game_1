# Feature: WebSocket-соединение, управление лобби в реальном времени

Status: draft
Priority: high
Depends on: [001-lobby-creation](001-lobby-creation.md)

## Context

После создания/присоединения к лобби через REST, клиент устанавливает WebSocket-соединение
для получения обновлений в реальном времени: появление новых игроков, готовность, выход.
Также необходимо обеспечить переподключение при обрыве связи во время игры.

См. [docs/rules.md](../../../docs/rules.md), разделы 3.1, 3.3.
Бизнес-фича: [docs/features/001-lobby-management.md](../../../docs/features/001-lobby-management.md).

## Requirements

### Functional

- Сервер ДОЛЖЕН предоставить WebSocket-эндпоинт `GET /api/lobbies/{code}/ws?playerId={id}`.
- При подключении сервер ДОЛЖЕН валидировать:
  - Лобби с данным `code` существует → иначе HTTP 404.
  - `playerId` принадлежит этому лобби → иначе HTTP 403.
- После upgrade сервер ДОЛЖЕН отправить клиенту текущее состояние лобби (`lobby_updated`).
- При подключении нового клиента сервер ДОЛЖЕН отметить игрока как `online: true` и
  разослать `lobby_updated` всем.
- При обрыве соединения сервер ДОЛЖЕН отметить игрока как `online: false` и разослать
  `player_disconnected` всем.
- При повторном подключении (`online: false` → WS connect) сервер ДОЛЖЕН отметить
  `online: true` и разослать `player_reconnected` всем.
- Команда `leave_lobby`:
  - ДОЛЖНА работать только в фазе `waiting`.
  - Сервер ДОЛЖЕН удалить игрока из `players[]`, закрыть его WS-соединение.
  - Если выходит администратор, права ДОЛЖНЫ перейти к следующему игроку.
  - Если лобби стало пустым, сервер ДОЛЖЕН удалить его из `LobbyStore`.
  - Сервер ДОЛЖЕН разослать `player_left` и `lobby_updated` остальным.
- Команда `player_ready`:
  - Игрок ДОЛЖЕН мочь менять `ready` между `true` и `false`.
  - Сервер ДОЛЖЕН разослать `lobby_updated` всем.
- **Переподключение во время игры** (модифицированный `POST /api/lobbies/{code}/join`):
  - Если `state == "in_game"` и `nickname` совпадает с `nickname` отключённого игрока,
    сервер ДОЛЖЕН вернуть существующий `playerId` этого игрока.
  - Клиент затем подключается по WS и получает актуальное состояние.
  - Если `nickname` не совпадает ни с одним отключённым игроком → `409 game_in_progress`.

### Non-functional

- WebSocket hub ДОЛЖЕН быть потокобезопасным.
- Отправка broadcast-сообщений НЕ ДОЛЖНА блокировать обработку входящих команд.
- Ping/pong для детекции мёртвых соединений (интервал 15 секунд).

## Affected Files

Файлы, которые нужно **создать**:

| Файл | Назначение |
|---|---|
| `backend/hub/hub.go` | `Hub` — управление WS-соединениями: register, unregister, broadcast. Хранит `map[string]*Client` (ключ — playerId). |
| `backend/hub/client.go` | `Client` — обёртка над WS-соединением: чтение, запись, ping/pong. |
| `backend/handler/ws.go` | HTTP-хендлер для WS upgrade + роутинг входящих WS-команд (`leave_lobby`, `player_ready`). |
| `backend/game/location.go` | Предустановленный набор локаций (список строк). |

Файлы, которые нужно **изменить**:

| Файл | Что изменить |
|---|---|
| `backend/main.go` | Зарегистрировать WS-эндпоинт. Создать Hub и передать в хендлеры. |
| `backend/game/lobby.go` | Добавить методы: `RemovePlayer`, `SetPlayerOnline`, `TransferAdmin`. Добавить поле `online` в `Player`. |
| `backend/handler/lobby.go` | Модифицировать `JoinLobby`: при `in_game` + совпадении nickname с offline-игроком → reconnect. |
| `backend/go.mod` | Добавить зависимость для WebSocket (`nhooyr.io/websocket`). |

Файлы, которые **НЕ трогать**: `frontend/**`, `.github/**`.

## Data Model

Используются модели из [data-models.md](../../../docs/data-models.md):
- `Player` — добавлено поле `online`.
- `Lobby` — поле `state` определяет доступность `leave_lobby`.

## API

Полные спецификации — в [api-contract.md](../../../docs/api-contract.md):
- WebSocket upgrade: раздел 2, «Подключение».
- Команды: `leave_lobby`, `player_ready`.
- События: `lobby_updated`, `player_left`, `player_disconnected`, `player_reconnected`.

## Acceptance Criteria

- [ ] WS-соединение устанавливается по `/api/lobbies/{code}/ws?playerId={id}`.
- [ ] При подключении клиент получает текущее состояние лобби.
- [ ] Все клиенты в лобби получают `lobby_updated` при изменениях.
- [ ] `leave_lobby` удаляет игрока и освобождает место (только фаза `waiting`).
- [ ] При выходе администратора права передаются следующему игроку.
- [ ] Пустое лобби удаляется из хранилища.
- [ ] Обрыв соединения → `player_disconnected` broadcast.
- [ ] Переподключение → `player_reconnected` broadcast.
- [ ] Reconnect во время игры через `POST /join` + совпадение nickname → возвращает существующий `playerId`.
- [ ] Новый игрок не может подключиться к запущенной игре.
- [ ] `player_ready` переключает `ready` и рассылает обновление.

## Edge Cases and Error Handling

- **Двойное WS-соединение от одного playerId** — закрыть предыдущее, оставить новое.
- **leave_lobby во время игры** — отклонить с ошибкой `wrong_phase`.
- **WS-команда от неизвестного playerId** — закрыть соединение.
- **Reconnect с nickname, совпадающим с online-игроком** — `409 nickname_taken`.

## Out of Scope

- Управление лобби (кик, передача прав по инициативе админа, настройки) — фаза 4.
- Фаза вопросов, голосование — фазы 2–3.
