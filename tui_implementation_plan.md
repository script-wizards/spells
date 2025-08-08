# TUI Implementation Plan - Multi-Pane Interface

## Current State
The existing TUI in `internal/tui/model.go` is a basic single-view interface with:
- Turn counter display
- NPC search functionality (`/` key)
- Basic initiative tracker
- Simple keyboard shortcuts (Space, i, /, Ctrl+C)

## Target: Rich Multi-Pane Interface
Based on `spec.md` lines 167-180, implement the sophisticated layout:

```
┌─ SESSION STATE ────────────┬─ QUICK LOOKUP ─────┬─ ACTIVE EVENTS ───┐
│ Turn: 47  │ XP Earned: 1,247│ [Search: kobol]    │ ⚠ Torch @turn 50  │
│ Party: 4  │ Torch: 3/10     │ → Kobold Shaman    │ ○ Wander @turn 55  │ 
│ Time: 7h40│ Last: 15m ago   │ → Kobold Warren    │ ● Spell ends now   │
├─────────────────────────────┼────────────────────┼────────────────────┤
│ INITIATIVE TRACKER          │ ROOM TRACKER       │ RECENT ACTIONS     │
│ 1. Theron (PC)    HP: 8/12  │ Room 12: Altar     │ Generated orc name │
│ 2. Orc Captain    HP: 15/18 │ ✓ Gold chalice     │ Advanced time 2t   │
│ 3. Gareth (PC)    HP: 6/6   │ ✗ Moldy books      │ Added Room 15      │
│ 4. Orc Warrior    HP: 0/8 💀│ ? Hidden door      │ Rolled wandering   │
└─────────────────────────────┴────────────────────┴────────────────────┘
```

## Implementation Steps

### 1. Layout Architecture
- **New package**: `internal/tui/panes/` for individual pane components
- **Grid system**: Use bubbletea's `lipgloss` for responsive 6-pane layout
- **Pane interface**: Common interface for all panes (Update, View, Focus states)

### 2. Core Panes (Phase 1)
Create individual pane components:

**SessionStatePane** (`session_state.go`)
- Turn counter and advancement
- XP tracking (from `treasure_found` table)  
- Party size tracking
- Torch duration countdown
- Session time elapsed

**InitiativePane** (`initiative.go`) 
- Enhance existing combatant display
- Add HP modification controls
- Status icons (💀 for dead, conditions)
- PC/NPC visual distinction

**QuickLookupPane** (`lookup.go`)
- Enhanced search with live results
- Recent NPC mentions (from `last_mentioned`)
- Location quick-access
- Search history

### 3. Advanced Panes (Phase 2)

**ActiveEventsPane** (`events.go`)
- Timer countdowns (torch, spells, wandering checks)
- Event queue from `time_events` table
- Visual priority indicators (⚠️ ○ ●)
- Turn-based trigger warnings

**RoomTrackerPane** (`room_tracker.go`)
- Current room state from `room_states` table
- Treasure status (taken/remaining) 
- Room contents tracking
- Party action history per room

**RecentActionsPane** (`actions.go`)
- Activity log of recent commands
- Oracle consultations history
- NPC/room interactions
- Time advancement log

### 4. Navigation & Controls

**Tab System**
- Tab between panes with vim keys (h/j/k/l)
- Visual focus indicators
- Modal overlays for detailed editing

**Keyboard Shortcuts** (per spec lines 263-266)
- Vim-style: h/j/k/l navigation
- `/` for search mode
- `:` for command mode (`:timer 10m torch`, `:gen npc`)
- `?` for contextual help
- Space for turn advancement

### 5. Data Integration

**Database Queries**
- Efficient queries for each pane's data needs
- Real-time updates via engine event bus
- Lazy loading for large datasets

**State Management**
- Each pane maintains local state
- Global state coordinator
- Event-driven updates between panes

### 6. Visual Polish

**lipgloss Styling**
- Border styles matching spec aesthetic
- Color themes for different data types
- Status icons and indicators
- Responsive column widths

**Performance**
- Efficient re-renders (only changed panes)
- Debounced search updates
- Minimal database queries

## File Structure
```
internal/tui/
├── model.go              # Main TUI coordinator
├── panes/
│   ├── interface.go      # Pane interface definition
│   ├── session_state.go  # SESSION STATE pane
│   ├── lookup.go         # QUICK LOOKUP pane  
│   ├── events.go         # ACTIVE EVENTS pane
│   ├── initiative.go     # INITIATIVE TRACKER pane
│   ├── room_tracker.go   # ROOM TRACKER pane
│   └── actions.go        # RECENT ACTIONS pane
├── layout.go             # Grid layout management
└── styles.go             # lipgloss styling
```

## Dependencies Already Available
- `charmbracelet/bubbletea` - TUI framework
- `charmbracelet/lipgloss` - Styling (needs to be added)
- Database models in `internal/model/`
- Event system in `internal/engine/`
- Search system in `internal/search/`

## Next Session Tasks
1. Add lipgloss dependency: `go get github.com/charmbracelet/lipgloss`
2. Create pane interface and layout manager
3. Implement SessionStatePane first (highest priority)
4. Test multi-pane rendering and navigation
5. Gradually add remaining panes

This will transform the basic TUI into the rich, multi-pane DM interface specified in the design document.