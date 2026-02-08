# game_1

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
