# Spells: DM Toolkit - Developer Specification

## Overview

`spells` is a terminal-based game master toolkit for tabletop RPGs, optimized for OSR fantasy games but designed to be system-neutral. The tool provides integrated session management, content lookup, random generation, and time tracking through a unified interface that can run as either persistent TUI applications or fast CLI commands.

## Core Design Principles

- **Single binary, multiple interfaces**: One codebase with different entry points
- **System neutral with OSR lean**: No hardcoded rules, configurable mechanics
- **Session persistence**: All data survives between sessions in SQLite
- **Live state sharing**: Multiple processes share real-time session state
- **Minimal friction**: Fast startup, intuitive commands, hot-reload support
- **Composable workflow**: Works standalone or in tmux multiplexer setups

## Architecture

### Technology Stack
- **Language**: Go
- **TUI Framework**: charmbracelet/bubbletea
- **Database**: SQLite with WAL mode for concurrent access
- **File watching**: fsnotify for hot-reload
- **Markdown parsing**: goldmark or similar

### Binary Structure
Single executable with subcommand dispatch:

```bash
spells track     # Long-lived TUI for session management
spells oracle    # Quick oracle consultation interface  
spells gen       # Content generation interface
spells timer     # Time tracking interface
spells full      # Complete integrated workspace
```

### Data Storage

#### Campaign Database (SQLite)
Location: `./campaign.db` in current directory

**Core Tables:**
```sql
-- Session state
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    current_turn INTEGER DEFAULT 0,
    notes TEXT
);

-- NPCs with fuzzy search support
CREATE TABLE npcs (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    location TEXT,
    status TEXT DEFAULT 'neutral', -- ally/neutral/hostile
    motivation TEXT,
    secrets TEXT,
    tags TEXT, -- JSON array
    last_mentioned TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Locations/Places
CREATE TABLE places (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    connections TEXT, -- JSON array of linked places
    current_occupants TEXT,
    dangers TEXT,
    opportunities TEXT,
    tags TEXT, -- JSON array
    last_visited TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Combat tracking
CREATE TABLE combat_encounters (
    id INTEGER PRIMARY KEY,
    session_id INTEGER REFERENCES sessions(id),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE initiative_order (
    id INTEGER PRIMARY KEY,
    encounter_id INTEGER REFERENCES combat_encounters(id),
    name TEXT NOT NULL,
    initiative INTEGER,
    hp_current INTEGER,
    hp_max INTEGER,
    conditions TEXT, -- JSON array
    status TEXT DEFAULT 'active', -- active/inactive/dead
    is_pc BOOLEAN DEFAULT false,
    sort_order INTEGER
);

-- Time tracking and events
CREATE TABLE timers (
    id INTEGER PRIMARY KEY,
    label TEXT NOT NULL,
    duration_minutes INTEGER,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    session_id INTEGER REFERENCES sessions(id)
);

CREATE TABLE time_events (
    id INTEGER PRIMARY KEY,
    trigger_turn INTEGER,
    event_type TEXT, -- 'torch_burnout', 'wandering_check', 'spell_end'
    description TEXT,
    handled BOOLEAN DEFAULT false,
    session_id INTEGER REFERENCES sessions(id)
);

-- Treasure and XP tracking
CREATE TABLE treasure_found (
    id INTEGER PRIMARY KEY,
    session_id INTEGER REFERENCES sessions(id),
    item_description TEXT,
    gold_value INTEGER,
    location TEXT,
    date_found TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    xp_awarded BOOLEAN DEFAULT false
);

-- Room state tracking
CREATE TABLE room_states (
    id INTEGER PRIMARY KEY,
    room_identifier TEXT NOT NULL, -- "Room 12", "Goblin Warren", etc.
    original_contents TEXT, -- What was originally here
    current_contents TEXT, -- What remains after party interaction
    party_actions TEXT, -- What the party did here
    session_id INTEGER REFERENCES sessions(id),
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Random table results cache
CREATE TABLE oracle_results (
    id INTEGER PRIMARY KEY,
    table_name TEXT,
    query_context TEXT,
    result TEXT,
    nested_results TEXT, -- JSON of resolved sub-tables
    session_id INTEGER REFERENCES sessions(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Configuration
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT
);
```

#### External Content Files
- **Markdown files**: NPCs, locations, adventure notes (hot-reloaded)
- **Perchance tables**: Random generation tables (parsed and cached)
- **Images/Assets**: Referenced from markdown, not stored in DB

### Core Modules

#### 1. Session Management (`track` command)
**TUI Interface Layout:**
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

**Key Features:**
- **Time advancement**: Spacebar advances turns, evaluates triggers
- **Fuzzy search**: Real-time NPC/place lookup with recency weighting
- **Initiative management**: Add/remove combatants, track HP/conditions
- **Room state**: Track treasure taken/remaining per location
- **Event queue**: Visual indicators for upcoming triggers

#### 2. Oracle System (`oracle` command)
**CLI Interface:**
```bash
spells oracle "random encounter"     # Consult encounter table
spells oracle "treasure hoard"       # Generate treasure
spells oracle --table goblin_lair    # Specific table
```

**TUI Interface** (when launched standalone):
```
┌─ ORACLE CONSULTATION ─────────────────────────────────────────────┐
│ Query: goblin_patrol                                               │
│                                                                    │
│ Result: 1d4+1 Goblin Warriors                                      │
│ ├─ Number: 1d4+1 → 3 goblins                                       │
│ ├─ HP: 1d6 each → 4, 2, 6                                          │
│ └─ Equipment: Short swords, leather armor                          │
│                                                                    │
│ [Roll Again] [History] [Save Result]                               │
└────────────────────────────────────────────────────────────────────┘
```

**Table Resolution Logic:**
- Parse perchance syntax: `[table_name]`, `{option1|option2}`, `[num]d[sides]`
- Resolve nested references up to 2 levels automatically
- Show unresolved references beyond 2 levels for manual expansion
- Cache results with session context
- Detect circular references and break gracefully

#### 3. Content Generation (`gen` command)
**CLI Interface:**
```bash
spells gen npc                    # Quick NPC with name/trait/motivation
spells gen name --type place      # Location name
spells gen name --type tavern     # Business name  
spells gen quirk                  # Character trait
spells gen complication           # Plot twist
spells gen rumor                  # Tavern gossip
```

**Smart Parsing for NPCs:**
```
Input: "Gareth merchant, suspicious about cult, knows secret passage"
Parsed:
  Name: Gareth
  Role: merchant  
  Personality: suspicious about cult
  Secret: knows secret passage
```

#### 4. Time Management (`timer` command)
**Integrated with session state:**
- Real-world timers for breaks, food orders
- Game-world turn advancement with trigger evaluation
- Torch/light source tracking (configurable duration)
- Spell duration tracking
- Wandering monster check scheduling

#### 5. Content Import System
**Markdown File Watching:**
- Monitor `*.md` files in campaign directory
- Parse NPC/location references automatically
- Support frontmatter for structured data
- Hot-reload changes during active sessions

**Perchance Table Import:**
```bash
spells import table ./encounters.perchance
spells import url https://perchance.org/goblin-generator
```

### User Interface Design

#### TUI Navigation (bubbletea)
- **Vim-style keybindings**: h/j/k/l, /, :, etc.
- **Tab switching**: Between panes and views
- **Quick commands**: `:timer 10m torch`, `:gen npc`, `:search goblin`
- **Contextual help**: ? key shows relevant shortcuts

#### CLI Output Format
- **Structured data**: JSON option for scripting
- **Human readable**: Default colorized output
- **Quiet mode**: Minimal output for automation

### Configuration

#### Global Settings (`~/.config/spells/config.yaml`)
```yaml
# System neutral mechanics
xp_conversion_rate: 1.0  # 1 gold = 1 XP
torch_duration_turns: 10
wandering_check_frequency: 6  # Every 6 turns

# Interface preferences  
default_view: full
auto_advance_time: false
notification_sound: true

# Content paths
global_tables_dir: "~/.config/spells/tables"
templates_dir: "~/.config/spells/templates"
```

#### Campaign Settings (`./spells.yaml`)
```yaml
campaign_name: "Sunken Temple Campaign"
system: "OSE"  # For template selection
party_size: 4

# Custom mechanics for this campaign
torch_duration_turns: 6  # Shorter torches in this dungeon
xp_multiplier: 1.5
```

### Error Handling

#### Database Errors
- **Corruption**: Automatic backup restoration
- **Lock conflicts**: Retry with exponential backoff
- **Migration failures**: Rollback with user notification

#### File System Errors
- **Missing campaign.db**: Auto-create with schema
- **Markdown parse errors**: Log and continue with available data
- **Permission issues**: Clear error messages with resolution steps

#### Network Errors (for imports)
- **Failed downloads**: Retry with timeout
- **Invalid URLs**: Validate before processing
- **Rate limiting**: Respect perchance.org API limits

### Performance Requirements

#### Startup Time
- **CLI commands**: < 100ms cold start
- **TUI applications**: < 500ms to first paint
- **Database operations**: < 50ms for typical queries

#### Memory Usage
- **CLI commands**: < 10MB RSS
- **TUI applications**: < 50MB RSS
- **Large campaigns**: Graceful handling of 1000+ NPCs/locations

#### Responsiveness
- **Fuzzy search**: Results within 100ms
- **File hot-reload**: Changes reflected within 1s
- **Time advancement**: UI update within 50ms

### Testing Strategy

#### Unit Tests
```bash
# Core logic testing
go test ./internal/oracle      # Table resolution
go test ./internal/fuzzy       # Search algorithms  
go test ./internal/parser      # Markdown/perchance parsing
go test ./internal/db          # Database operations
```

#### Integration Tests  
```bash
# End-to-end workflow testing
go test ./e2e/session_flow     # Complete session lifecycle
go test ./e2e/multi_process    # Concurrent access scenarios
go test ./e2e/import_export    # Content import/export
```

#### Performance Tests
```bash
# Load testing with realistic data
go test ./perf/large_campaign  # 1000+ NPCs performance
go test ./perf/concurrent      # Multiple spells processes
go test ./perf/search          # Fuzzy search with large datasets
```

#### Manual Testing Scenarios
1. **New campaign setup**: Initialize, import content, first session
2. **Mid-session workflow**: Combat, time tracking, content lookup
3. **Multi-terminal usage**: tmux panes with live state sharing
4. **Content authoring**: Markdown editing with hot-reload
5. **Data persistence**: Session recovery after unexpected exit

### Security Considerations

#### Local Data Protection
- **SQLite security**: No remote access, file permissions
- **Content parsing**: Sanitize imported markdown/perchance files
- **Process isolation**: No elevated privileges required

#### Input Validation
- **SQL injection**: Prepared statements only
- **File path traversal**: Validate all file operations
- **Memory safety**: Go's built-in protections

### Deployment

#### Build Process
```bash
# Single binary with embedded assets
go build -ldflags="-s -w" -o spells ./cmd/spells
```

#### Installation Options
- **Package managers**: Homebrew, apt, yum
- **Direct download**: GitHub releases with checksums
- **Source compilation**: Go toolchain requirement

#### Platform Support
- **Primary**: Linux, macOS, Windows
- **Terminal requirements**: 256 color support, Unicode
- **Dependencies**: None (static linking)

### Migration Strategy

#### Schema Versioning
- **Database migrations**: Automatic on version upgrade
- **Backward compatibility**: Support N-1 version data
- **Export/import**: JSON format for data portability

#### Configuration Updates
- **Breaking changes**: Migration tools provided
- **Default preservation**: Existing configs respected
- **Documentation**: Clear upgrade instructions

### Monitoring and Debugging

#### Logging
```yaml
# Development mode
log_level: debug
log_file: "./spells.log"

# Production mode  
log_level: info
log_file: "~/.cache/spells/spells.log"
```

#### Debug Tools
```bash
spells debug db-stats      # Database statistics
spells debug perf-profile  # Performance profiling
spells debug export-logs   # Log export for bug reports
```

#### Crash Recovery
- **Automatic backups**: Before schema changes
- **Session recovery**: Restore last known state
- **Data validation**: Integrity checks on startup

### Future Extensibility

#### Plugin Architecture
- **Lua scripting**: Custom table generators
- **Go plugins**: Advanced integrations
- **External tools**: JSON API for third-party tools

#### Content Ecosystem
- **Table sharing**: Community table repository
- **Template library**: Pre-built campaign templates
- **Import/export**: Standard formats for portability

#### Integration Points
- **VTT connectivity**: Roll20, Foundry data exchange
- **Character sheets**: D&D Beyond, PC sheet import
- **Content sources**: Automated SRD importing

---

## Implementation Roadmap

### Phase 1: Core Infrastructure (4-6 weeks)
- SQLite schema and basic CRUD operations
- CLI argument parsing and subcommand dispatch
- Basic TUI framework with bubbletea
- File watching and hot-reload system

### Phase 2: Essential Features (6-8 weeks)  
- Session management and time tracking
- Initiative tracker with combat management
- Fuzzy search for NPCs and locations
- Basic oracle system with dice rolling

### Phase 3: Content Integration (4-6 weeks)
- Markdown file parsing and import
- Perchance table parser and resolver
- Room state tracking system
- Content generation templates

### Phase 4: Polish and Performance (3-4 weeks)
- Multi-process state sharing
- Performance optimization and caching
- Comprehensive error handling
- User experience refinements

### Phase 5: Testing and Release (2-3 weeks)
- Complete test suite implementation
- Documentation and example content
- Release packaging and distribution
- Community feedback integration

**Total Estimated Timeline: 19-27 weeks**

This specification provides a complete roadmap for implementing `spells` as a production-ready DM toolkit. The modular architecture supports both immediate utility and long-term extensibility while maintaining the core design principles of simplicity and effectiveness.

