# Idle Game Backend

Go (Fiber) backend for an idle mining game using SQLite (pure-Go driver) and JWT auth.

## Stack
- Go 1.21
- Fiber v2 web framework
- GORM + glebarez/sqlite (no CGO)
- JWT auth (golang-jwt/jwt v5)

## Layout
- `main.go` — server entry, routes, CORS
- `database/` — DB init, models, activity logs
- `handlers/` — auth, user, mining, inventory
- `middleware/` — JWT auth
- `utils/` — auth helpers

## Replit Setup
- Server listens on `0.0.0.0:5000` (was `localhost:3000` upstream).
- CORS allows all origins (dev) so the Replit preview proxy works.
- Workflow `Server` runs `go run .` and waits for port 5000.
- Deployment: VM target (SQLite needs persistent disk), `go build -o server .` then `./server`.
- SQLite DB file: `idlegame.db` (created in working dir on first run).
