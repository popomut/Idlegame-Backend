# Backend Architecture Overview

## Project Structure

```
Idlegame-backend/
├── main.go                 # Entry point, Fiber app setup
├── go.mod                  # Go dependencies
├── README.md               # Full documentation
├── QUICKSTART.md           # Quick start guide
├── .gitignore
│
├── database/
│   ├── db.go               # Database connection, migrations, seeding
│   ├── models.go           # GORM models (User, OreInventory, etc.)
│   └── activity.go         # Activity logging helpers
│
├── handlers/
│   ├── auth.go             # Register, Login endpoints
│   ├── user.go             # User profile endpoints
│   ├── mining.go           # Mining logic (core anti-cheat!)
│   └── inventory.go        # Ore inventory endpoint
│
├── middleware/
│   └── auth.go             # JWT authentication middleware
│
└── utils/
    └── auth.go             # Password hashing, JWT generation
```

## Core Security Features

### 1. Server-Side Timestamp Validation
```go
// In handlers/mining.go - CalculateAndSaveOreGains()
now := time.Now().UTC()  // Server time (cannot be spoofed)
elapsed := now.Sub(session.StartedAt)  // Server timestamps only
oresEarned := int(elapsed.Milliseconds()) / ore.MiningTimeMS
```

**Why secure:**
- Client NEVER provides timestamps
- Calculation happens on server
- Database stores immutable UTC times
- Even if client sends fake data, it's ignored

### 2. Database Schema Anti-Cheat
```sql
-- Mining sessions table
CREATE TABLE mining_sessions (
  id INTEGER PRIMARY KEY,
  user_id INTEGER,
  ore_id INTEGER,
  started_at TIMESTAMP,  -- Server-generated, immutable
  ended_at TIMESTAMP,    -- Server-generated
  ores_mined INTEGER,
  status TEXT
);
```

**Why secure:**
- Timestamps created on server insert
- No UPDATE to timestamps allowed
- Client cannot modify historical data

### 3. Activity Audit Trail
```go
// Every ore gain is logged
database.LogActivity(userID, "Mined 1800 Copper Ore")
```

**Benefits:**
- Full audit trail of suspicious activity
- Can detect pattern anomalies
- Easy to spot cheaters

### 4. JWT Authentication
```go
// Token includes exp: 7 days
// Required for all protected endpoints
// Cannot be forged without secret key
```

## Offline Mining Flow (Step-by-Step)

```
Day 1, 2:00 PM:
  User clicks "Mine Copper"
  → POST /api/mining/start { ore_id: 1 }
  → Backend: INSERT mining_sessions 
             { user_id: 5, ore_id: 1, started_at: 2:00 PM UTC, status: 'active' }
  
Day 1, 4:00 PM:
  User closes browser (mining still "active" in DB)
  
Day 2, 10:00 AM (next day, 30 hours later):
  User reopens app
  → Frontend: GET /api/mining/status
  → Backend calculates:
      elapsed = now (10:00 AM UTC) - started_at (2:00 PM UTC yesterday)
      elapsed = 30 hours = 108,000,000 ms
      ores_earned = 108,000,000 / 2000 = 54,000 ores!
      
      UPDATE ore_inventory SET copper_ore = copper_ore + 54000
      INSERT activity_log "Mined 54000 Copper Ore while offline"
  
  → Response to frontend:
      {
        "is_active": true,
        "current_ore": { ... },
        "offline_gains": {
          "was_offline": true,
          "offline_time_ms": 108000000,
          "ores_gained": 54000,
          "ore_name": "Copper Ore"
        }
      }
  
  → Frontend shows popup: "⚡ You earned 54,000 Copper Ore while away!"
```

## Why Client Can't Cheat

**Scenario: User changes system time forward 24 hours**

```
// What user might try:
// Set clock to tomorrow, close browser, reopen

// What happens:
GET /api/mining/status
  // Server checks DB:
  started_at = 2:00 PM (real timestamp, stored before client cheated)
  now = 10:00 AM (server reads its own clock, not client's)
  elapsed = now - started_at  // Uses server's real clock!
  
  // Even if client's clock is 24 hours ahead, server ignores it
  // Calculation uses server's actual time
  // Result: Honest calculation
```

**Scenario: User edits localStorage to fake ore count**

```
// What user might try:
// localStorage.setItem('copper_ore', 999999)

// What happens:
GET /api/inventory/ores
  // Server queries DATABASE, not client storage!
  SELECT copper_ore FROM ore_inventory WHERE user_id = 5
  // Returns real value, ignores client localStorage
  // Result: Honest response
```

**Scenario: User modifies network request to add fake ores**

```
// What user might try:
// POST /api/mining/stop with {"ores_gained": 999999}

// What happens:
// Server validates the request:
var session database.MiningSession
DB.Where("user_id = ? AND status = ?", userID, "active").First(&session)
  
// Recalculates based on timestamps:
oresEarned := int(elapsed.Milliseconds()) / ore.MiningTimeMS
  
// Ignores any client-provided ore count
// Result: Only honest calculated amount is saved
```

## Migration to PostgreSQL

**To switch from SQLite to PostgreSQL:**

1. Install PostgreSQL driver:
```bash
go get gorm.io/driver/postgres
```

2. Update `database/db.go`:
```go
// Replace SQLite:
db, err := gorm.Open(sqlite.Open("idlegame.db"), &gorm.Config{})

// With PostgreSQL:
import "gorm.io/driver/postgres"
db, err := gorm.Open(postgres.Open(
  "user=postgres password=pw dbname=idlegame host=localhost"), 
  &gorm.Config{})
```

3. Everything else stays the same!
   - All models work unchanged
   - All handlers work unchanged
   - All business logic works unchanged

## Testing Endpoints

### Full test sequence:
```bash
# 1. Register
TOKEN=$(curl -s -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player1",
    "email": "player1@example.com",
    "password": "secret123"
  }' | jq -r '.token')

# 2. Start mining
curl -X POST http://localhost:3000/api/mining/start \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ore_id": 1}'

# 3. Wait 4+ seconds...
sleep 5

# 4. Check status (should show mining active)
curl -X GET http://localhost:3000/api/mining/status \
  -H "Authorization: Bearer $TOKEN"

# 5. Stop mining (should show ores gained)
curl -X POST http://localhost:3000/api/mining/stop \
  -H "Authorization: Bearer $TOKEN"

# 6. Check inventory
curl -X GET http://localhost:3000/api/inventory/ores \
  -H "Authorization: Bearer $TOKEN"
```

## Performance Notes

- SQLite handles single-device fine
- PostgreSQL needed for multiplayer/distributed
- All queries indexed (mining_sessions by user + status)
- Ore calculations O(1) (simple math, no loops)

