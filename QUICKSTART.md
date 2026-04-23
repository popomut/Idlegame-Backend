# Backend Quick Start

## 1. Install Go dependencies

```bash
cd C:\workspace_svelte\Idlegame-backend
go mod download
go mod tidy
```

## 2. Run the backend server

```bash
go run main.go
```

You should see:
```
🚀 Server running on http://localhost:3000
```

## 3. Test the API

### Register a user
```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": 1,
  "username": "testuser",
  "email": "test@example.com"
}
```

### Get mining status
```bash
curl -X GET http://localhost:3000/api/mining/status \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Start mining (Copper Ore = ID 1)
```bash
curl -X POST http://localhost:3000/api/mining/start \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{"ore_id": 1}'
```

### Stop mining (after 4+ seconds)
```bash
curl -X POST http://localhost:3000/api/mining/stop \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

## Security Features Implemented

✅ **Server-side timestamp calculation**
  - User cannot spoof local time
  - All mining calculations use UTC timestamps

✅ **JWT Authentication**
  - Tokens expire after 7 days
  - Required for all protected endpoints

✅ **Password hashing**
  - bcrypt with cost=12
  - Rainbow table resistant

✅ **Database-level protection**
  - Ore calculations happen on server
  - Client never calculates final amounts
  - All updates atomic and validated

✅ **Input validation**
  - All endpoints validate request data
  - SQL injection prevention (GORM parameterized queries)

## Database Schema

SQLite file: `idlegame.db` (auto-created)

### Tables
- `users` - Player accounts
- `ore_inventories` - Ore counts per player
- `ore_types` - Ore definitions (static)
- `mining_sessions` - Mining progress tracking
- `activity_logs` - Audit trail

## Why This is Hack-Proof

1. **No client-side math**
   - Server NEVER trusts client input
   - Time calculation: `(server_now - session_start) / mining_time`

2. **Server timestamps**
   - All timestamps stored as UTC
   - Client time is never used

3. **Immutable history**
   - Each mining session records start/end times
   - Activity log tracks all changes
   - Cannot retroactively edit

4. **Atomic transactions**
   - Ore updates happen in single DB transaction
   - No partial updates possible

## Next Steps

1. Connect frontend to these backend endpoints
2. Add JWT token storage in frontend
3. Show offline gains popup when returning
4. Update mining UI to use backend status
