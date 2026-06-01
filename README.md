# Spyfall (Находка шпиона)

Многопользовательская веб-игра на дедукцию и блеф.

## Документация

| Документ | Описание |
|---|---|
| [docs/workflow.md](docs/workflow.md) | Рабочий процесс разработки, роли, жизненный цикл фичи |
| [docs/rules.md](docs/rules.md) | Правила игры |
| [docs/architecture.md](docs/architecture.md) | Архитектура, стек, структура кода |
| [docs/data-models.md](docs/data-models.md) | Модели данных |
| [docs/api-contract.md](docs/api-contract.md) | API-контракт (REST + WebSocket) |
| [docs/glossary.md](docs/glossary.md) | Глоссарий терминов |

## Local development

### Backend (Go)

```bash
cd backend
go run .
```

The server listens on `http://localhost:8080`. You can override the port with `PORT`.

Health check:

```bash
curl http://localhost:8080/readyz
```

### Frontend (React + Vite)

```bash
cd frontend
pnpm install
pnpm run dev
```

The dev server runs at `http://localhost:5173` by default.
