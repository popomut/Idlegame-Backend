# Idlegame Backend

Go backend server for Realm of Eternity idle game using SQLite.

## Setup

### Prerequisites
- Go 1.21+
- SQLite3 (bundled with GORM)

### Install dependencies
```bash
go mod download
go mod tidy
```

### Run the server
```bash
go run main.go
```

Server will start at `http://localhost:3000`

## Architecture

### Database (SQLite → PostgreSQL compatible)

All code is structured to easily migrate from SQLite to PostgreSQL:
- GORM handles database abstraction
- SQL queries are vendor-agnostic
- No SQLite-specific features used

### Security

**Mining Anti-Cheat:**
- All calculations use **server-side timestamps** (UTC)
- Client time is **never trusted**
- Ore calculations: `(now - session_start) / mining_time_ms`
- Timestamps stored in database, immutable

**Authentication:**
- Passwords hashed with bcrypt (never stored plaintext)
- JWT tokens (7-day expiration)
- Authorization header required for all protected endpoints

**API Protection:**
- CORS enabled (localhost:5173 only for now)
- All endpoints validate JWT
- Input validation on all requests

## API Endpoints

### Authentication (Public)
```
POST /api/auth/register
  { "username", "email", "password" }

POST /api/auth/login
  { "username", "password" }
```

### User (Protected)
```
GET /api/user
  Returns current user profile

POST /api/user/update
  { "player_name", "player_class" }
```

### Mining (Protected)
```
POST /api/mining/start
  { "ore_id" } → Starts mining session

POST /api/mining/stop
  Stops mining, saves ores, returns offline gains

GET /api/mining/status
  Returns current mining status + offline gains summary
```

### Inventory (Protected)
```
GET /api/inventory/ores
  Returns ore counts
```

## Offline Mining Flow

1. **User starts mining**
   - Backend creates `MiningSession` with `started_at = NOW()`
   
2. **User closes browser**
   - Session remains in database as `status = 'active'`
   
3. **User returns 2 hours later**
   - Frontend calls `GET /api/mining/status`
   - Backend calculates: `(now - started_at) / mining_time_ms`
   - Ore inventory automatically updated
   - Response includes `offline_gains` summary
   - Frontend shows popup: "You earned 3600 Copper Ore while offline!"

## Migration to PostgreSQL

To migrate from SQLite to PostgreSQL:

1. Replace `gorm.io/driver/sqlite` with `gorm.io/driver/postgres`
2. Update connection string in `database/db.go`
3. All models and logic remain unchanged!

Example:
```go
// SQLite (current)
db, err := gorm.Open(sqlite.Open("idlegame.db"), &gorm.Config{})

// PostgreSQL (future)
db, err := gorm.Open(postgres.Open("user=postgres password=pw dbname=idlegame host=localhost"), &gorm.Config{})
```

## Development

### Database file
- Location: `./idlegame.db`
- Automatically created on first run
- Contains all game data

### Environment variables
- `JWT_SECRET`: JWT signing key (default: hardcoded for dev)
- Set in production!

### Logs
- Server logs all database operations
- Mining calculations logged to `ActivityLog`

## Security Checklist

- [x] Passwords hashed (bcrypt)
- [x] JWT authentication
- [x] Server-side timestamp validation
- [x] Database-level calculation (no client trust)
- [x] Input validation
- [x] CORS restricted
- [x] No plaintext secrets in code
- [ ] HTTPS/TLS (add in production)
- [ ] Rate limiting (add if needed)
- [ ] Audit logging (comprehensive)

