# Frontend-Backend Integration Guide

## Next Steps to Connect Frontend

### 1. Install API client in Svelte frontend
```bash
cd C:\workspace_svelte\Idlegame
npm install axios
```

### 2. Create API service (`src/services/api.js`)

Create this file in your Svelte project:

```javascript
import axios from 'axios';

const API_BASE = 'http://localhost:3000/api';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Auto-attach JWT token to all requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const authAPI = {
  register: (username, email, password) =>
    api.post('/auth/register', { username, email, password }),
  login: (username, password) =>
    api.post('/auth/login', { username, password }),
};

export const miningAPI = {
  startMining: (oreId) => api.post('/mining/start', { ore_id: oreId }),
  stopMining: () => api.post('/mining/stop'),
  getMiningStatus: () => api.get('/mining/status'),
};

export const inventoryAPI = {
  getOreInventory: () => api.get('/inventory/ores'),
};

export const userAPI = {
  getUser: () => api.get('/user'),
  updateUser: (playerName, playerClass) =>
    api.post('/user/update', { player_name: playerName, player_class: playerClass }),
};
```

### 3. Update game store to use backend (`src/stores/game.js`)

```javascript
import { writable, derived } from 'svelte/store';
import { miningAPI, inventoryAPI } from '../services/api.js';

export const player = writable({
  name: 'Hero',
  level: 1,
  xp: 24,
  xpToNextLevel: 100,
  hp: 85,
  maxHp: 100,
  mana: 42,
  maxMana: 50,
  gold: 150,
  class: 'Apprentice Knight',
});

export const ores = writable({
  copperOre: 5,
  ironOre: 2,
  goldOre: 0,
  mithrilOre: 0,
  diamondOre: 0,
});

// Sync ores with backend
export async function syncOreInventory() {
  try {
    const response = await inventoryAPI.getOreInventory();
    ores.set({
      copperOre: response.data.copper_ore,
      ironOre: response.data.iron_ore,
      goldOre: response.data.gold_ore,
      mithrilOre: response.data.mithril_ore,
      diamondOre: response.data.diamond_ore,
    });
  } catch (error) {
    console.error('Failed to sync ore inventory:', error);
  }
}

export function addOre(oreType) {
  ores.update(function (inv) {
    return { ...inv, [oreType]: (inv[oreType] || 0) + 1 };
  });
}

export const activityLog = writable([
  'You have entered the Realm of Eternity...',
  'The ancient gates creak open before you.',
  'Your adventure begins.',
]);

export function addLogEntry(message) {
  activityLog.update(function (log) {
    const updated = [message, ...log];
    if (updated.length > 50) {
      return updated.slice(0, 50);
    }
    return updated;
  });
}
```

### 4. Update Mining store to use backend (`src/stores/mining.js`)

```javascript
import { writable } from 'svelte/store';
import { miningAPI } from '../services/api.js';
import { addOre, addLogEntry, ores } from './game.js';

export const activeMining = writable(null);
export const miningPopups = writable([]);
export const offlineGains = writable(null);

export async function startMining(oreType, oreName, oreId) {
  try {
    // Tell backend to start mining
    const response = await miningAPI.startMining(oreId);
    
    activeMining.set({ 
      oreType, 
      oreName, 
      progress: 0,
      startedAt: new Date(response.data.started_at),
    });
    
    addLogEntry(`Started mining ${oreName}...`);
  } catch (error) {
    console.error('Failed to start mining:', error);
    addLogEntry('Failed to start mining. Try again.');
  }
}

export async function stopMining() {
  try {
    const response = await miningAPI.stopMining();
    const oredGained = response.data.ores_gained;
    
    activeMining.set(null);
    
    if (oredGained > 0) {
      showMiningPopup(oredGained);
      addLogEntry(`Earned ${oredGained} ores!`);
      
      // Sync inventory with backend
      await syncOreInventory();
    }
  } catch (error) {
    console.error('Failed to stop mining:', error);
    addLogEntry('Failed to stop mining. Try again.');
  }
}

export async function checkMiningStatus() {
  try {
    const response = await miningAPI.getMiningStatus();
    const status = response.data;
    
    // Check for offline gains
    if (status.offline_gains && status.offline_gains.was_offline) {
      const gains = status.offline_gains;
      offlineGains.set({
        wasOffline: true,
        timeMs: gains.offline_time_ms,
        oredGained: gains.ores_gained,
        oreName: gains.ore_name,
      });
      
      addLogEntry(
        `⚡ You earned ${gains.ores_gained} ${gains.ore_name} while away!`
      );
      
      // Update inventory
      await syncOreInventory();
    }
    
    // Update mining status
    if (status.is_active && status.current_ore) {
      activeMining.set({
        oreType: status.current_ore.ore_key,
        oreName: status.current_ore.ore_name,
        startedAt: new Date(status.started_at),
      });
    } else {
      activeMining.set(null);
    }
    
    // Update current ores display
    ores.set({
      copperOre: status.current_ores.copper_ore,
      ironOre: status.current_ores.iron_ore,
      goldOre: status.current_ores.gold_ore,
      mithrilOre: status.current_ores.mithril_ore,
      diamondOre: status.current_ores.diamond_ore,
    });
  } catch (error) {
    console.error('Failed to check mining status:', error);
  }
}

export function showMiningPopup(count = 1) {
  const id = Date.now();
  miningPopups.update(function (popups) {
    return [...popups, { id, count }];
  });

  setTimeout(function () {
    miningPopups.update(function (popups) {
      return popups.filter(function (p) { return p.id !== id; });
    });
  }, 1000);
}

// Call this on app startup to check offline gains
export function initMiningStatus() {
  checkMiningStatus();
}
```

### 5. Add offline gains popup component

Create `src/components/OfflineGainsPopup.svelte`:

```svelte
<script>
  import { offlineGains } from '../stores/mining.js';
</script>

{#if $offlineGains && $offlineGains.wasOffline}
  <div class="offline-popup" in:fadeInScale out:fadeOutScale>
    <div class="popup-content">
      <div class="popup-icon">⚡</div>
      <h2 class="popup-title">Welcome Back!</h2>
      <p class="popup-text">
        You earned <span class="ore-count">{$offlineGains.oredGained}</span>
        <span class="ore-name">{$offlineGains.oreName}</span>
        while away!
      </p>
      <p class="popup-time">
        ({Math.round($offlineGains.timeMs / 1000 / 60)} minutes offline)
      </p>
      <button class="popup-btn" on:click={() => offlineGains.set(null)}>
        Awesome!
      </button>
    </div>
  </div>
{/if}

<style>
  .offline-popup {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .popup-content {
    background: var(--color-bg-panel);
    border: 2px solid var(--color-gold-dim);
    border-radius: 12px;
    padding: 24px;
    text-align: center;
    max-width: 400px;
  }

  .popup-icon {
    font-size: 48px;
    margin-bottom: 12px;
    animation: bounce 0.6s ease-out;
  }

  .popup-title {
    font-family: var(--font-heading);
    font-size: 24px;
    margin-bottom: 8px;
    color: var(--color-text-heading);
  }

  .popup-text {
    font-size: 16px;
    color: var(--color-text);
    margin-bottom: 4px;
  }

  .ore-count {
    font-weight: 700;
    color: var(--color-gold-bright);
  }

  .ore-name {
    color: var(--color-magic-bright);
  }

  .popup-time {
    font-size: 13px;
    color: var(--color-text-muted);
    margin-bottom: 16px;
  }

  .popup-btn {
    padding: 10px 24px;
    background: var(--color-gold-dim);
    color: #000;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .popup-btn:hover {
    background: var(--color-gold-bright);
  }

  @keyframes bounce {
    0% {
      transform: scale(0) rotate(0);
    }
    50% {
      transform: scale(1.2);
    }
    100% {
      transform: scale(1);
    }
  }
</style>
```

### 6. Update App.svelte to init offline gains on mount

```svelte
<script>
  import { onMount } from 'svelte';
  import { initMiningStatus } from './stores/mining.js';
  import OfflineGainsPopup from './components/OfflineGainsPopup.svelte';
  // ... other imports
</script>

<onMount(async function () {
  await initMiningStatus();
});>

<TopBar />
<Sidebar />

<main class="main-content" class:sidebar-expanded={$sidebarOpen}>
  <!-- existing routes -->
</main>

<BottomBar />
<OfflineGainsPopup />
```

## Running Full Stack

### Terminal 1: Backend
```bash
cd C:\workspace_svelte\Idlegame-backend
go run main.go
```

### Terminal 2: Frontend
```bash
cd C:\workspace_svelte\Idlegame
npm run dev
```

Then open **http://localhost:5173** and test mining!

## Testing the Offline Flow

1. **Start mining**
   - Click Mining → Select Copper Ore → Start

2. **Immediately close browser**
   - Close tab/window

3. **Wait 2+ minutes** (or use DevTools to modify clock)

4. **Reopen app**
   - Should show popup: "You earned X ores while away!"
   - Inventory should be updated

5. **Check logs**
   - Backend: `go run main.go` shows all SQL queries
   - Check calculations are correct

